package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"presensee/internal/middleware"
	"presensee/internal/model"

	"gorm.io/gorm"
)

type CommonHandler struct {
	DB *gorm.DB
}

func (h *CommonHandler) Ping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"data": "pong"})
}

func (h *CommonHandler) Me(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"detail": "Unauthorized"})
		return
	}

	var kelasID *uint
	if user.Type != nil {
		switch *user.Type {
		case model.TypeWaliKelas:
			var k model.Kelas
			if err := h.DB.Where("wali_kelas_id = ? AND active = ?", user.ID, true).First(&k).Error; err == nil {
				kelasID = &k.ID
			}
		case model.TypeSekretaris:
			var k model.Kelas
			if err := h.DB.Joins("JOIN kelas_sekretaris ON kelas_sekretaris.kelas_id = kelas.id").
				Where("kelas_sekretaris.user_id = ? AND kelas.active = ?", user.ID, true).First(&k).Error; err == nil {
				kelasID = &k.ID
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"data": map[string]any{
			"username": user.Username,
			"type":     user.Type,
			"kelas":    kelasID,
		},
	})
}

var indonesianMonths = [...]string{
	"", "Januari", "Februari", "Maret", "April", "Mei", "Juni",
	"Juli", "Agustus", "September", "Oktober", "November", "Desember",
}

type BulanResponse struct {
	Bulan          string `json:"bulan"`
	BulanHumanize  string `json:"bulan_humanize"`
}

func (h *CommonHandler) Bulan(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"detail": "Unauthorized"})
		return
	}

	var dates []time.Time
	q := h.DB.Model(&model.Absensi{}).Select("DISTINCT date").Order("date ASC")

	if user.Type != nil && *user.Type != model.TypeKesiswaan && !user.IsSuperuser {
		q = q.Joins("JOIN siswas ON siswas.id = absensis.siswa_id").
			Joins("JOIN kelas ON kelas.id = siswas.kelas_id")
		if *user.Type == model.TypeWaliKelas {
			q = q.Where("kelas.wali_kelas_id = ?", user.ID)
		} else if *user.Type == model.TypeSekretaris {
			q = q.Joins("JOIN kelas_sekretaris ON kelas_sekretaris.kelas_id = kelas.id").
				Where("kelas_sekretaris.user_id = ?", user.ID)
		}
	}

	q.Pluck("date", &dates)

	seen := make(map[string]bool)
	var result []BulanResponse

	for _, d := range dates {
		m := int(d.Month())
		y := d.Year()
		code := fmt.Sprintf("%02d-%02d", m, y%100)
		if seen[code] {
			continue
		}
		seen[code] = true

		monthName := "Unknown"
		if m >= 1 && m <= 12 {
			monthName = indonesianMonths[m]
		}
		human := fmt.Sprintf("%s %d", monthName, y)
		result = append(result, BulanResponse{
			Bulan:         code,
			BulanHumanize: human,
		})
	}

	if result == nil {
		result = []BulanResponse{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"data": result})
}

func (h *CommonHandler) Siswas(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil || (user.Type != nil && *user.Type != model.TypeGuruPiket && !user.IsSuperuser) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"detail": "Forbidden"})
		return
	}

	var siswas []model.Siswa
	h.DB.Preload("Kelas").Joins("JOIN kelas ON kelas.id = siswas.kelas_id AND kelas.active = true").Find(&siswas)

	res := make(map[uint]map[string]any)
	for _, s := range siswas {
		res[s.ID] = map[string]any{
			"name":     s.FullName,
			"kelas":    s.Kelas.Name,
			"kelas_id": s.KelasID,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"data": res})
}
