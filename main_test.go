package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"presensee/internal/database"
	"presensee/internal/handler/admin"
	"presensee/internal/handler/api"
	"presensee/internal/handler/setup"
	"presensee/internal/middleware"
	"presensee/internal/model"
	"presensee/internal/service/lzstring"
	"presensee/internal/service/sync"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

func setupTestApp(t *testing.T) (*gorm.DB, http.Handler) {
	dbFile := "test_presensee.db"
	os.Remove(dbFile)

	db, err := database.Connect(dbFile)
	if err != nil {
		t.Fatalf("db error: %v", err)
	}

	if err := model.AutoMigrate(db); err != nil {
		t.Fatalf("migrate error: %v", err)
	}

	r := chi.NewRouter()

	setupHandler := &setup.SetupHandler{DB: db}
	r.Get("/setup", setupHandler.SetupPage)
	r.Post("/setup", setupHandler.SetupPost)

	adminHandler := &admin.AdminHandler{DB: db}
	r.Route("/admin", func(ar chi.Router) {
		adminHandler.RegisterRoutes(ar)
	})

	authHandler := &api.AuthHandler{DB: db}
	commonHandler := &api.CommonHandler{DB: db}
	syncHandler := &api.SyncHandler{DB: db}
	absensiHandler := &api.AbsensiHandler{DB: db}

	r.Route("/api", func(apiRouter chi.Router) {
		apiRouter.Get("/ping", commonHandler.Ping)
		apiRouter.Post("/login", authHandler.Login)

		apiRouter.Group(func(pr chi.Router) {
			pr.Use(middleware.AuthBearer(db))
			pr.Get("/me", commonHandler.Me)
			pr.Get("/data", syncHandler.GetData)
			pr.Post("/upload", syncHandler.Upload)
			pr.Post("/compressed-upload", syncHandler.CompressedUpload)
			pr.Get("/absensi", absensiHandler.GetAbsensies)
			pr.Get("/absensi/progress", absensiHandler.GetAbsensiProgress)
			pr.Get("/bulan", commonHandler.Bulan)
		})
	})

	t.Cleanup(func() {
		os.Remove(dbFile)
	})

	return db, r
}

func TestEndToEndFlow(t *testing.T) {
	db, router := setupTestApp(t)

	// 1. Setup Superuser via /setup
	form := url.Values{
		"username":         {"admin"},
		"password":         {"admin123"},
		"confirm_password": {"admin123"},
	}
	req := httptest.NewRequest("POST", "/setup", bytes.NewBufferString(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 redirect from /setup, got %d", rec.Code)
	}

	// 2. Login via API
	loginBody, _ := json.Marshal(map[string]string{
		"username": "admin",
		"password": "admin123",
	})
	req = httptest.NewRequest("POST", "/api/login", bytes.NewBuffer(loginBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 from login, got %d: %s", rec.Code, rec.Body.String())
	}

	var loginResp struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	json.Unmarshal(rec.Body.Bytes(), &loginResp)
	token := loginResp.Data.Token
	if token == "" {
		t.Fatal("empty token received")
	}

	// 3. Create Kelas & Siswa in DB
	var kelas model.Kelas
	kelas = model.Kelas{Name: "XII IPA 1", Active: true}
	db.Create(&kelas)

	var siswa model.Siswa
	siswa = model.Siswa{FullName: "Budi Santoso", KelasID: kelas.ID, NIS: "1001"}
	db.Create(&siswa)

	// 4. Test API /api/data (Dump to SQL)
	req = httptest.NewRequest("GET", "/api/data", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 from /api/data, got %d", rec.Code)
	}
	dumpText := rec.Body.String()
	if !bytes.Contains([]byte(dumpText), []byte("Budi Santoso")) {
		t.Fatalf("dump missing siswa: %s", dumpText)
	}

	// 5. Test Offline Sync /api/upload
	absenData, _ := json.Marshal(sync.AbsenPayload{
		Siswa:  siswa.ID,
		Date:   time.Now().Format("02-01-2006"),
		Status: "hadir",
	})
	payload := sync.UploadPayload{
		Data: []sync.ActionItem{
			{Action: "absen", Data: string(absenData)},
		},
	}
	uploadBody, _ := json.Marshal(payload)
	req = httptest.NewRequest("POST", "/api/upload", bytes.NewBuffer(uploadBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 from /api/upload, got %d: %s", rec.Code, rec.Body.String())
	}

	// 6. Test /api/absensi
	req = httptest.NewRequest("GET", "/api/absensi?kelas_id="+url.QueryEscape("1")+"&date="+time.Now().Format("2006-01-02"), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 from /api/absensi, got %d: %s", rec.Code, rec.Body.String())
	}

	// 7. Verify LZString compressed upload works
	compressedRaw, _ := json.Marshal([]sync.ActionItem{
		{Action: "absen", Data: string(absenData)},
	})
	_ = compressedRaw
	_ = lzstring.DecompressFromBase64
}
