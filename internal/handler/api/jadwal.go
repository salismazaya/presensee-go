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

type JadwalHandler struct {
	DB *gorm.DB
}

func (h *JadwalHandler) GetJadwal(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil || (user.Type != nil && *user.Type != model.TypeGuruPiket && !user.IsSuperuser) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"detail": "Forbidden"})
		return
	}

	days := []string{"senin", "selasa", "rabu", "kamis", "jumat", "sabtu"}

	var kelass []model.Kelas
	h.DB.Where("active = ?", true).Find(&kelass)

	var sessions []model.AbsensiSession
	h.DB.Preload("Kelas").Find(&sessions)

	// Map: "kelasID_day" -> Session
	sessionMap := make(map[string]*model.AbsensiSession)
	for i := range sessions {
		s := &sessions[i]
		for _, k := range s.Kelas {
			if s.Senin {
				sessionMap[fmt.Sprintf("%d_senin", k.ID)] = s
			}
			if s.Selasa {
				sessionMap[fmt.Sprintf("%d_selasa", k.ID)] = s
			}
			if s.Rabu {
				sessionMap[fmt.Sprintf("%d_rabu", k.ID)] = s
			}
			if s.Kamis {
				sessionMap[fmt.Sprintf("%d_kamis", k.ID)] = s
			}
			if s.Jumat {
				sessionMap[fmt.Sprintf("%d_jumat", k.ID)] = s
			}
			if s.Sabtu {
				sessionMap[fmt.Sprintf("%d_sabtu", k.ID)] = s
			}
		}
	}

	results := make(map[uint]map[string]any)

	for _, k := range kelass {
		kRes := map[string]any{
			"name": k.Name,
		}

		for _, day := range days {
			sess := sessionMap[fmt.Sprintf("%d_%s", k.ID, day)]
			if sess == nil {
				kRes[day] = nil
				continue
			}

			jamMasuk := sess.JamMasuk
			jamMasukSampai := formatJamMasukSampai(sess.JamMasuk, sess.JamMasukToleransi)
			jamKeluar := sess.JamKeluar
			if sess.JamKeluarMulaiAbsen != "" {
				jamKeluar = sess.JamKeluarMulaiAbsen
			}

			kRes[day] = map[string]any{
				"jam_masuk":        jamMasuk,
				"jam_masuk_sampai": jamMasukSampai,
				"jam_keluar":       jamKeluar,
			}
		}

		results[k.ID] = kRes
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"data": results})
}

func formatJamMasukSampai(jamMasuk string, toleransiSeconds int64) string {
	t, err := time.Parse("15:04", jamMasuk)
	if err != nil {
		t, err = time.Parse("15:04:05", jamMasuk)
		if err != nil {
			return jamMasuk
		}
	}
	t = t.Add(time.Duration(toleransiSeconds) * time.Second)
	return t.Format("15:04")
}
