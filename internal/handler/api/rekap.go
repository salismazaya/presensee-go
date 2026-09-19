package api

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"presensee/internal/middleware"
	"presensee/internal/model"
	"presensee/internal/service/excel"

	"gorm.io/gorm"
)

type CachedFile struct {
	Filename    string
	ContentType string
	Content     []byte
	ExpiresAt   time.Time
}

var (
	fileStoreMu sync.RWMutex
	fileStore   = make(map[string]CachedFile)
)

func StoreFile(fileID string, filename string, contentType string, content []byte, ttl time.Duration) {
	fileStoreMu.Lock()
	defer fileStoreMu.Unlock()
	fileStore[fileID] = CachedFile{
		Filename:    filename,
		ContentType: contentType,
		Content:     content,
		ExpiresAt:   time.Now().Add(ttl),
	}
}

func GetFile(fileID string) (*CachedFile, bool) {
	fileStoreMu.RLock()
	defer fileStoreMu.RUnlock()
	cf, ok := fileStore[fileID]
	if !ok || time.Now().After(cf.ExpiresAt) {
		return nil, false
	}
	return &cf, true
}

func ServeFiles(w http.ResponseWriter, r *http.Request) {
	fileID := stringsTrimPrefix(r.URL.Path, "/files/")
	cf, ok := GetFile(fileID)
	if !ok {
		http.Error(w, "File tidak ditemukan atau sudah kadaluarsa", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", cf.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", cf.Filename))
	w.Write(cf.Content)
}

func stringsTrimPrefix(s, prefix string) string {
	if len(s) >= len(prefix) && s[:len(prefix)] == prefix {
		return s[len(prefix):]
	}
	return s
}

type RekapHandler struct {
	DB *gorm.DB
}

func (h *RekapHandler) GetRekap(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"detail": "Unauthorized"})
		return
	}

	bulanStr := r.URL.Query().Get("bulan")
	kelasStr := r.URL.Query().Get("kelas")
	tahunStr := r.URL.Query().Get("tahun")

	bulan, _ := strconv.Atoi(bulanStr)
	kelasID, _ := strconv.Atoi(kelasStr)
	tahun, _ := strconv.Atoi(tahunStr)

	if tahun <= 99 {
		tahun += 2000
	}

	var kelas model.Kelas
	if err := h.DB.Preload("Sekretaris").First(&kelas, kelasID).Error; err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"detail": "kelas not found"})
		return
	}

	canAccess := user.IsSuperuser
	if user.Type != nil {
		switch *user.Type {
		case model.TypeKesiswaan:
			canAccess = true
		case model.TypeWaliKelas:
			if kelas.WaliKelasID != nil && *kelas.WaliKelasID == user.ID {
				canAccess = true
			}
		case model.TypeSekretaris:
			for _, s := range kelas.Sekretaris {
				if s.ID == user.ID {
					canAccess = true
					break
				}
			}
		}
	}

	if !canAccess {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"detail": "kelas not found"})
		return
	}

	excelBytes, filename, err := excel.ExportAbsensi(h.DB, uint(kelasID), tahun, bulan)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"detail": err.Error()})
		return
	}

	hash := md5.Sum([]byte(fmt.Sprintf("%d_%d_%d_%d", kelasID, tahun, bulan, len(excelBytes))))
	fileID := fmt.Sprintf("rekap-%s", hex.EncodeToString(hash[:]))

	StoreFile(fileID, filename, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", excelBytes, 24*time.Hour)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"data": map[string]string{"file_id": fileID},
	})
}
