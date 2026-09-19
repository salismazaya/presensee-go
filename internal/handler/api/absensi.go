package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"presensee/internal/middleware"
	"presensee/internal/model"

	"gorm.io/gorm"
)

type AbsensiHandler struct {
	DB *gorm.DB
}

func (h *AbsensiHandler) GetAbsensies(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	dateStr := r.URL.Query().Get("date")
	kelasIDStr := r.URL.Query().Get("kelas_id")
	if dateStr == "" || kelasIDStr == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"detail": "Parameter date dan kelas_id wajib"})
		return
	}

	kelasID, err := strconv.ParseUint(kelasIDStr, 10, 32)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"detail": "kelas_id tidak valid"})
		return
	}

	parsedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		if t, e := time.Parse("02-01-2006", dateStr); e == nil {
			parsedDate = t
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"detail": "Tanggal tidak valid"})
			return
		}
	}

	var siswas []model.Siswa
	h.DB.Where("kelas_id = ?", kelasID).Find(&siswas)

	var absensies []model.Absensi
	h.DB.Where("date = ? AND siswa_id IN (?)", parsedDate, h.DB.Model(&model.Siswa{}).Select("id").Where("kelas_id = ?", kelasID)).Find(&absensies)

	statusMap := make(map[uint]string)
	for _, a := range absensies {
		statusMap[a.SiswaID] = string(a.FinalStatus())
	}

	result := make(map[uint]*string)
	for _, s := range siswas {
		if st, ok := statusMap[s.ID]; ok {
			stVal := st
			result[s.ID] = &stVal
		} else {
			result[s.ID] = nil
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"data": result})
}

func (h *AbsensiHandler) GetAbsensiProgress(w http.ResponseWriter, r *http.Request) {
	kelasIDStr := r.URL.Query().Get("kelas_id")
	datesStr := r.URL.Query().Get("dates")
	if kelasIDStr == "" || datesStr == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"detail": "kelas_id dan dates wajib"})
		return
	}

	kelasID, _ := strconv.ParseUint(kelasIDStr, 10, 32)
	dates := strings.Split(datesStr, ",")
	if len(dates) >= 32 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"detail": "terlalu banyak input tanggal"})
		return
	}

	var totalSiswa int64
	h.DB.Model(&model.Siswa{}).Where("kelas_id = ?", kelasID).Count(&totalSiswa)

	result := make(map[string]map[string]any)

	for _, dStr := range dates {
		dStr = strings.TrimSpace(dStr)
		if dStr == "" {
			continue
		}
		t, err := time.Parse("2006-01-02", dStr)
		if err != nil {
			if parsed, e := time.Parse("02-01-2006", dStr); e == nil {
				t = parsed
			} else {
				continue
			}
		}

		var totalAbsensi int64
		var totalTidakMasuk int64

		h.DB.Model(&model.Absensi{}).
			Joins("JOIN siswas ON siswas.id = absensis.siswa_id").
			Where("siswas.kelas_id = ? AND absensis.date = ?", kelasID, t).
			Count(&totalAbsensi)

		h.DB.Model(&model.Absensi{}).
			Joins("JOIN siswas ON siswas.id = absensis.siswa_id").
			Where("siswas.kelas_id = ? AND absensis.date = ? AND absensis._status != ?", kelasID, t, model.StatusHadir).
			Count(&totalTidakMasuk)

		result[dStr] = map[string]any{
			"total_tidak_masuk": totalTidakMasuk,
			"is_complete":       totalAbsensi == totalSiswa && totalSiswa > 0,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"data": result})
}
