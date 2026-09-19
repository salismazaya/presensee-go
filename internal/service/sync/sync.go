package sync

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"presensee/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ActionItem struct {
	Action string `json:"action"` // "absen", "lock", "unlock"
	Data   string `json:"data"`   // JSON string of payload
}

type UploadPayload struct {
	Data []ActionItem `json:"data"`
}

type AbsenPayload struct {
	Siswa          uint    `json:"siswa"`
	Date           string  `json:"date"`
	Status         string  `json:"status"`
	PreviousStatus *string `json:"previous_status,omitempty"`
	UpdatedAt      *int64  `json:"updated_at,omitempty"`
}

type LockPayload struct {
	Kelas uint   `json:"kelas"`
	Date  string `json:"date"`
}

type ConflictSide struct {
	DisplayName   string `json:"display_name"`
	AbsensiStatus string `json:"absensi_status"`
}

type ConflictInfo struct {
	Type           string       `json:"type"`
	AbsensiID      uint         `json:"absensi_id"`
	AbsensiSiswa   string       `json:"absensi_siswa"`
	AbsensiSiswaID uint         `json:"absensi_siswa_id"`
	AbsensiKelasID uint         `json:"absensi_kelas_id"`
	AbsensiDate    string       `json:"absensi_date"`
	Other          ConflictSide `json:"other"`
	Self           ConflictSide `json:"self"`
}

type SyncResult struct {
	Conflicts []ConflictInfo `json:"conflicts"`
}

func parseFlexDate(dateStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	layouts := []string{
		"02-01-06",
		"2-1-06",
		"02-01-2006",
		"2-1-2006",
		"2006-01-02",
		"2006/01/02",
		"02/01/2006",
	}

	var parsed time.Time
	var err error
	for _, l := range layouts {
		if t, e := time.Parse(l, dateStr); e == nil {
			parsed = t
			err = nil
			break
		} else {
			err = e
		}
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date: %s", dateStr)
	}

	if parsed.Year() < 100 {
		parsed = parsed.AddDate(2000, 0, 0)
	}

	if parsed.Year() < 2020 || parsed.After(time.Now().Add(24*time.Hour)) {
		return time.Time{}, fmt.Errorf("date out of range: %s", dateStr)
	}

	return time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.Local), nil
}

func ProcessSync(db *gorm.DB, user *model.User, payload UploadPayload) (*SyncResult, error) {
	// Sort actions: unlock (0) -> absen (1) -> lock (2)
	actions := make([]ActionItem, len(payload.Data))
	copy(actions, payload.Data)
	sort.SliceStable(actions, func(i, j int) bool {
		score := func(a string) int {
			switch a {
			case "unlock":
				return 0
			case "absen":
				return 1
			case "lock":
				return 2
			default:
				return 3
			}
		}
		return score(actions[i].Action) < score(actions[j].Action)
	})

	conflicts := make([]ConflictInfo, 0)

	err := db.Transaction(func(tx *gorm.DB) error {
		// Cache in-memory lock state: "YYYY-MM-DD_kelasID" -> locked
		lockMap := make(map[string]bool)

		for _, item := range actions {
			switch item.Action {
			case "lock", "unlock":
				var lp LockPayload
				if err := json.Unmarshal([]byte(item.Data), &lp); err != nil {
					continue
				}
				d, err := parseFlexDate(lp.Date)
				if err != nil {
					continue
				}

				// Check permission: wali kelas or superuser
				var kelas model.Kelas
				if err := tx.Preload("WaliKelas").First(&kelas, lp.Kelas).Error; err != nil {
					return errors.New("kelas not found")
				}

				if !user.IsSuperuser && (kelas.WaliKelasID == nil || *kelas.WaliKelasID != user.ID) {
					return errors.New("ditolak: hanya wali kelas yang dapat mengunci absensi")
				}

				isLocked := item.Action == "lock"
				kunci := model.KunciAbsensi{
					Date:    d,
					KelasID: lp.Kelas,
					Locked:  isLocked,
				}

				if err := tx.Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "date"}, {Name: "kelas_id"}},
					DoUpdates: clause.AssignmentColumns([]string{"locked", "updated_at"}),
				}).Create(&kunci).Error; err != nil {
					return err
				}

				lockKey := fmt.Sprintf("%s_%d", d.Format("2006-01-02"), lp.Kelas)
				lockMap[lockKey] = isLocked

			case "absen":
				var ap AbsenPayload
				if err := json.Unmarshal([]byte(item.Data), &ap); err != nil {
					continue
				}
				d, err := parseFlexDate(ap.Date)
				if err != nil {
					continue
				}

				var siswa model.Siswa
				if err := tx.Preload("Kelas.WaliKelas").Preload("Kelas.Sekretaris").First(&siswa, ap.Siswa).Error; err != nil {
					continue
				}

				// Check access permission
				canAccess := user.IsSuperuser
				if !canAccess && siswa.Kelas.WaliKelasID != nil && *siswa.Kelas.WaliKelasID == user.ID {
					canAccess = true
				}
				if !canAccess {
					for _, sek := range siswa.Kelas.Sekretaris {
						if sek.ID == user.ID {
							canAccess = true
							break
						}
					}
				}
				if !canAccess {
					continue
				}

				// Check lock status
				dateKey := d.Format("2006-01-02")
				lockKey := fmt.Sprintf("%s_%d", dateKey, siswa.KelasID)
				locked, exists := lockMap[lockKey]
				if !exists {
					var k model.KunciAbsensi
					if err := tx.Where("date = ? AND kelas_id = ?", d, siswa.KelasID).First(&k).Error; err == nil {
						locked = k.Locked
					}
					lockMap[lockKey] = locked
				}

				if locked {
					return fmt.Errorf("absen tanggal %s sedang dikunci", ap.Date)
				}

				updatedAt := time.Now()
				if ap.UpdatedAt != nil {
					updatedAt = time.Unix(*ap.UpdatedAt, 0)
				}

				var existing model.Absensi
				errFind := tx.Preload("By").Where("date = ? AND siswa_id = ?", d, siswa.ID).First(&existing).Error

				if errors.Is(errFind, gorm.ErrRecordNotFound) {
					newAbsensi := model.Absensi{
						Date:      d,
						SiswaID:   siswa.ID,
						Status:    model.AbsensiStatus(ap.Status),
						ByID:      &user.ID,
						CreatedAt: updatedAt,
						UpdatedAt: updatedAt,
					}
					if err := tx.Create(&newAbsensi).Error; err != nil {
						return err
					}
				} else if errFind == nil {
					currentStatus := string(existing.FinalStatus())
					newStatus := ap.Status

					if existing.ByID == nil {
						existing.Status = model.AbsensiStatus(newStatus)
						existing.ByID = &user.ID
						existing.UpdatedAt = updatedAt
						tx.Save(&existing)
					} else {
						isSameStatus := currentStatus == newStatus
						isSameUser := *existing.ByID == user.ID

						if !isSameStatus && isSameUser {
							existing.Status = model.AbsensiStatus(newStatus)
							existing.UpdatedAt = updatedAt
							tx.Save(&existing)
						} else if !isSameStatus && !isSameUser && ap.PreviousStatus == nil {
							conflicts = append(conflicts, ConflictInfo{
								Type:           "absensi",
								AbsensiID:      existing.ID,
								AbsensiSiswa:   siswa.FullName,
								AbsensiSiswaID: siswa.ID,
								AbsensiKelasID: siswa.KelasID,
								AbsensiDate:    dateKey,
								Other: ConflictSide{
									DisplayName:   existing.By.DisplayName(),
									AbsensiStatus: currentStatus,
								},
								Self: ConflictSide{
									DisplayName:   user.DisplayName(),
									AbsensiStatus: newStatus,
								},
							})
						} else if !isSameStatus && !isSameUser && ap.PreviousStatus != nil {
							if *ap.PreviousStatus == currentStatus || currentStatus == string(model.StatusTunggu) {
								existing.Status = model.AbsensiStatus(newStatus)
								existing.ByID = &user.ID
								existing.UpdatedAt = updatedAt
								tx.Save(&existing)
							} else {
								conflicts = append(conflicts, ConflictInfo{
									Type:           "absensi",
									AbsensiID:      existing.ID,
									AbsensiSiswa:   siswa.FullName,
									AbsensiSiswaID: siswa.ID,
									AbsensiKelasID: siswa.KelasID,
									AbsensiDate:    dateKey,
									Other: ConflictSide{
										DisplayName:   existing.By.DisplayName(),
										AbsensiStatus: currentStatus,
									},
									Self: ConflictSide{
										DisplayName:   user.DisplayName(),
										AbsensiStatus: newStatus,
									},
								})
							}
						}
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &SyncResult{Conflicts: conflicts}, nil
}
