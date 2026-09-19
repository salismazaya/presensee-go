package admin

import (
	"net/http"
	"strings"

	"presensee/internal/model"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type AdminHandler struct {
	DB *gorm.DB
}

func (h *AdminHandler) RegisterRoutes(r chi.Router) {
	r.Get("/login", h.LoginPage)
	r.Post("/login", h.LoginPost)
	r.Get("/logout", h.Logout)

	// Protected routes
	r.Group(func(pr chi.Router) {
		pr.Use(h.AdminAuthMiddleware)

		pr.Get("/", h.Dashboard)
		pr.Get("/dashboard", h.Dashboard)

		// Users
		pr.Get("/users", h.UsersList)
		pr.Get("/users/create", h.UserCreatePage)
		pr.Post("/users/create", h.UserCreatePost)
		pr.Get("/users/{id}/edit", h.UserEditPage)
		pr.Post("/users/{id}/edit", h.UserEditPost)
		pr.Post("/users/{id}/delete", h.UserDelete)

		// Kelas
		pr.Get("/kelas", h.KelasList)
		pr.Get("/kelas/create", h.KelasCreatePage)
		pr.Post("/kelas/create", h.KelasCreatePost)
		pr.Get("/kelas/{id}/edit", h.KelasEditPage)
		pr.Post("/kelas/{id}/edit", h.KelasEditPost)
		pr.Post("/kelas/{id}/delete", h.KelasDelete)
		pr.Post("/naik-kelas", h.NaikKelas)

		// Siswa
		pr.Get("/siswa", h.SiswaList)
		pr.Get("/siswa/create", h.SiswaCreatePage)
		pr.Post("/siswa/create", h.SiswaCreatePost)
		pr.Get("/siswa/{id}/edit", h.SiswaEditPage)
		pr.Post("/siswa/{id}/edit", h.SiswaEditPost)
		pr.Post("/siswa/{id}/delete", h.SiswaDelete)

		// Absensi
		pr.Get("/absensi", h.AbsensiList)

		// Kunci Absensi
		pr.Get("/kunci", h.KunciList)
		pr.Post("/kunci/toggle", h.KunciToggle)

		// Sessions / Jadwal QR
		pr.Get("/sessions", h.SessionsList)
		pr.Get("/sessions/create", h.SessionCreatePage)
		pr.Post("/sessions/create", h.SessionCreatePost)
		pr.Get("/sessions/{id}/edit", h.SessionEditPage)
		pr.Post("/sessions/{id}/edit", h.SessionEditPost)
		pr.Post("/sessions/{id}/delete", h.SessionDelete)

		// Data Sekolah / Config
		pr.Get("/data-sekolah", h.DataSekolahPage)
		pr.Post("/data-sekolah", h.DataSekolahPost)

		// Import & Export
		pr.Get("/import-siswa", h.ImportSiswaPage)
		pr.Post("/import-siswa", h.ImportSiswaPost)
		pr.Get("/export-absensi", h.ExportAbsensiPage)
		pr.Post("/export-absensi", h.ExportAbsensiPost)

		// Kartu Siswa
		pr.Get("/kartu-siswa", h.KartuSiswaPage)
	})
}

func (h *AdminHandler) AdminAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("admin_token")
		if err != nil || cookie.Value == "" {
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}

		var user model.User
		if err := h.DB.Where("token = ? AND is_active = ? AND (is_superuser = ? OR is_staff = ?)", cookie.Value, true, true, true).First(&user).Error; err != nil {
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (h *AdminHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "login", map[string]any{
		"Title": "Login Admin Presensee",
		"Error": r.URL.Query().Get("error"),
	})
}

func (h *AdminHandler) LoginPost(w http.ResponseWriter, r *http.Request) {
	username := strings.ToLower(strings.TrimSpace(r.FormValue("username")))
	password := r.FormValue("password")

	var user model.User
	if err := h.DB.Where("LOWER(TRIM(username)) = ? AND is_active = ?", username, true).First(&user).Error; err != nil {
		http.Redirect(w, r, "/admin/login?error=Username+atau+password+salah", http.StatusSeeOther)
		return
	}

	if !user.IsSuperuser && !user.IsStaff {
		http.Redirect(w, r, "/admin/login?error=Akses+admin+panel+ditolak", http.StatusSeeOther)
		return
	}

	if !user.CheckPassword(password) {
		http.Redirect(w, r, "/admin/login?error=Username+atau+password+salah", http.StatusSeeOther)
		return
	}

	token := model.GenerateToken(32)
	h.DB.Model(&user).Update("token", token)

	http.SetCookie(w, &http.Cookie{
		Name:     "admin_token",
		Value:    token,
		Path:     "/admin",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

func (h *AdminHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "admin_token",
		Value:    "",
		Path:     "/admin",
		MaxAge:   -1,
		HttpOnly: true,
	})
	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}
