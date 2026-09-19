package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"presensee/internal/database"
	"presensee/internal/handler/admin"
	"presensee/internal/handler/api"
	"presensee/internal/handler/setup"
	"presensee/internal/middleware"
	"presensee/internal/model"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "presensee.db" // default SQLite untuk development
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	db, err := database.Connect(dsn)
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}

	// Auto-migrate schema on start
	if err := model.AutoMigrate(db); err != nil {
		log.Fatalf("Database migration error: %v", err)
	}

	r := chi.NewRouter()

	// Global Middlewares
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Setup Wizard Routes
	setupHandler := &setup.SetupHandler{DB: db}
	r.Get("/setup", setupHandler.SetupPage)
	r.Post("/setup", setupHandler.SetupPost)

	// Admin Panel Routes
	adminHandler := &admin.AdminHandler{DB: db}
	r.Route("/admin", func(ar chi.Router) {
		adminHandler.RegisterRoutes(ar)
	})

	// File Storage Routes
	r.Get("/files/{file_id}", api.ServeFiles)

	// REST API Routes
	authHandler := &api.AuthHandler{DB: db}
	commonHandler := &api.CommonHandler{DB: db}
	syncHandler := &api.SyncHandler{DB: db}
	absensiHandler := &api.AbsensiHandler{DB: db}
	jadwalHandler := &api.JadwalHandler{DB: db}
	piketHandler := &api.PiketHandler{DB: db}
	rekapHandler := &api.RekapHandler{DB: db}

	r.Route("/api", func(apiRouter chi.Router) {
		// Public
		apiRouter.Get("/ping", commonHandler.Ping)
		apiRouter.Post("/login", authHandler.Login)

		// Protected (Bearer Token)
		apiRouter.Group(func(pr chi.Router) {
			pr.Use(middleware.AuthBearer(db))

			pr.Get("/me", commonHandler.Me)
			pr.Post("/change-password", authHandler.ChangePassword)

			// Offline-First Sync & Data
			pr.Get("/data", syncHandler.GetData)
			pr.Post("/upload", syncHandler.Upload)
			pr.Post("/compressed-upload", syncHandler.CompressedUpload)

			// Absensi & Progress
			pr.Get("/absensi", absensiHandler.GetAbsensies)
			pr.Get("/absensi/progress", absensiHandler.GetAbsensiProgress)

			// Features
			pr.Get("/bulan", commonHandler.Bulan)
			pr.Get("/siswas", commonHandler.Siswas)
			pr.Get("/jadwal", jadwalHandler.GetJadwal)
			pr.Post("/piket/upload", piketHandler.PiketUpload)
			pr.Get("/get-rekap", rekapHandler.GetRekap)
		})
	})

	// Frontend SPA Serving
	distDir := "./frontend/dist"
	if _, err := os.Stat(distDir); err == nil {
		fs := http.FileServer(http.Dir(distDir))

		r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			// Check if file exists in dist
			target := filepath.Join(distDir, path)
			info, err := os.Stat(target)
			if err == nil && !info.IsDir() {
				fs.ServeHTTP(w, r)
				return
			}

			// SPA Fallback: check cookie for guru_piket
			cookie, _ := r.Cookie("user_type")
			indexFile := "index.html"
			if cookie != nil && cookie.Value == "guru_piket" {
				indexFile = "piket.html"
			}

			indexPath := filepath.Join(distDir, indexFile)
			if _, err := os.Stat(indexPath); err == nil {
				http.ServeFile(w, r, indexPath)
				return
			}

			// If dist not built yet
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`
				<html><body style="font-family:sans-serif;text-align:center;padding:50px;">
				<h2>Presensee Go Backend Running 🚀</h2>
				<p>Admin Panel: <a href="/admin">/admin</a> | Setup: <a href="/setup">/setup</a></p>
				<p style="color:#666;">Frontend SPA dist belum di-build (jalankan <code>cd frontend && bun run build</code>)</p>
				</body></html>
			`))
		})
	} else {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
		})
	}

	addr := fmt.Sprintf("0.0.0.0:%s", port)
	log.Printf("Presensee Go server running at http://%s (DB: %s)", addr, dsn)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
