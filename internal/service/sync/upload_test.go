package sync_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"presensee/internal/database"
	"presensee/internal/handler/api"
	"presensee/internal/middleware"
	"presensee/internal/migration"
	"presensee/internal/model"
	"presensee/internal/service/lzstring"

	"github.com/go-chi/chi/v5"
)

func TestCompressedUploadIntegration(t *testing.T) {
	dbPath := "test_upload_sync.db"
	_ = os.Remove(dbPath)
	defer os.Remove(dbPath)

	db, err := database.Connect(dbPath)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}

	if err := migration.MigrateUp(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// 1. Create User Wali Kelas
	wali := model.User{
		Username:    "guru_wali",
		FullName:    "Guru Wali",
		Type:        func() *model.UserType { ut := model.TypeWaliKelas; return &ut }(),
		Token:       "test_wali_token",
		IsActive:    true,
		IsStaff:     false,
		IsSuperuser: false,
	}
	wali.SetPassword("wali123")
	db.Create(&wali)

	// 2. Create Kelas
	kelas := model.Kelas{
		Name:        "XII MIPA 1",
		Active:      true,
		WaliKelasID: &wali.ID,
	}
	db.Create(&kelas)

	// 3. Create Siswa
	siswa := model.Siswa{
		FullName: "Ahmad Santoso",
		KelasID:  kelas.ID,
		NIS:      "12345",
		NISN:     "0012345678",
	}
	db.Create(&siswa)

	// Setup API router
	r := chi.NewRouter()
	syncH := &api.SyncHandler{DB: db}

	// Mock middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			token := req.Header.Get("Authorization")
			if token == "Bearer test_wali_token" {
				ctx := context.WithValue(req.Context(), middleware.UserContextKey, &wali)
				req = req.WithContext(ctx)
			}
			next.ServeHTTP(w, req)
		})
	})

	r.Post("/api/compressed-upload", syncH.CompressedUpload)

	// 4. Prepare Staging payload as sent by frontend
	today := time.Now().Format("2006-01-02")
	absenDataJSON, _ := json.Marshal(map[string]any{
		"kelas":           kelas.ID,
		"siswa":           siswa.ID,
		"status":          "hadir",
		"previous_status": nil,
		"date":            today,
		"updated_at":      time.Now().Unix(),
	})

	stagingItems := []map[string]string{
		{
			"action": "absen",
			"data":   string(absenDataJSON),
		},
	}

	rawJSON, _ := json.Marshal(stagingItems)
	compressedStr := lzstring.CompressToBase64(string(rawJSON))

	reqBody, _ := json.Marshal(map[string]string{
		"data": compressedStr,
	})

	req := httptest.NewRequest("POST", "/api/compressed-upload", bytes.NewReader(reqBody))
	req.Header.Set("Authorization", "Bearer test_wali_token")
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	t.Logf("Response Code: %d, Body: %s", rec.Code, rec.Body.String())
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Verify absensi was recorded in DB
	var absensi model.Absensi
	if err := db.Where("siswa_id = ?", siswa.ID).First(&absensi).Error; err != nil {
		t.Fatalf("absensi not found in DB: %v", err)
	}
	if absensi.Status != model.StatusHadir {
		t.Fatalf("expected status 'hadir', got '%s'", absensi.Status)
	}
}
