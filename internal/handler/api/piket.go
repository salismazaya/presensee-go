package api

import (
	"encoding/json"
	"net/http"
	"sort"
	"time"

	"presensee/internal/middleware"
	"presensee/internal/model"

	"gorm.io/gorm"
)

type PiketHandler struct {
	DB *gorm.DB
}

type PiketScanItem struct {
	Siswa     uint   `json:"siswa"`
	Type      string `json:"type"` // "absen_masuk" / "absen_pulang"
	Timestamp int64  `json:"timestamp"`
}

func (h *PiketHandler) PiketUpload(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil || (user.Type != nil && *user.Type != model.TypeGuruPiket && !user.IsSuperuser) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"detail": "Forbidden"})
		return
	}

	var items []PiketScanItem
	if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"detail": "Format request tidak valid"})
		return
	}

	// Sort so absen_masuk comes before absen_pulang
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].Type == "absen_masuk" && items[j].Type != "absen_masuk"
	})

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			itemTime := time.Unix(item.Timestamp, 0)
			date := time.Date(itemTime.Year(), itemTime.Month(), itemTime.Day(), 0, 0, 0, 0, time.Local)

			var dayCol string
			switch date.Weekday() {
			case time.Monday:
				dayCol = "senin"
			case time.Tuesday:
				dayCol = "selasa"
			case time.Wednesday:
				dayCol = "rabu"
			case time.Thursday:
				dayCol = "kamis"
			case time.Friday:
				dayCol = "jumat"
			case time.Saturday:
				dayCol = "sabtu"
			default:
				continue
			}

			var siswa model.Siswa
			if err := tx.First(&siswa, item.Siswa).Error; err != nil {
				continue
			}

			var session model.AbsensiSession
			errSess := tx.Joins("JOIN absensi_session_kelas ON absensi_session_kelas.absensi_session_id = absensi_sessions.id").
				Where("absensi_session_kelas.kelas_id = ? AND "+dayCol+" = true", siswa.KelasID).
				First(&session).Error
			if errSess != nil {
				continue
			}

			var existing model.Absensi
			errFind := tx.Where("siswa_id = ? AND date = ?", siswa.ID, date).First(&existing).Error

			if item.Type == "absen_pulang" {
				if errFind == nil {
					existing.Status = model.StatusHadir
					existing.UpdatedAt = time.Now()
					tx.Save(&existing)
				}
			} else if item.Type == "absen_masuk" && errFind != nil {
				// Parse jam keluar
				jk, _ := time.Parse("15:04", session.JamKeluar)
				if jk.IsZero() {
					jk, _ = time.Parse("15:04:05", session.JamKeluar)
				}
				expiredAt := time.Date(date.Year(), date.Month(), date.Day(), jk.Hour(), jk.Minute(), jk.Second(), 0, time.Local)

				newAbs := model.Absensi{
					Date:          date,
					SiswaID:       siswa.ID,
					ByID:          &user.ID,
					WaitExpiredAt: &expiredAt,
					Status:        model.StatusTunggu,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				}
				tx.Create(&newAbs)
			}
		}
		return nil
	})

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"detail": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"data": map[string]any{
			"invalids": []any{},
		},
	})
}
