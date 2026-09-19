package setup

import (
	"html/template"
	"net/http"
	"strings"

	"presensee/internal/model"

	"gorm.io/gorm"
)

type SetupHandler struct {
	DB *gorm.DB
}

var setupTemplate = template.Must(template.New("setup").Parse(`
<!DOCTYPE html>
<html lang="id" data-theme="light">
<head>
    <meta charset="UTF-8">
    <title>Setup Awal - Presensee</title>
    <link href="https://cdn.jsdelivr.net/npm/daisyui@4.12.10/dist/full.min.css" rel="stylesheet" type="text/css" />
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-base-200 min-h-screen flex items-center justify-center p-4">
    <div class="card w-full max-w-md bg-base-100 shadow-xl border border-base-300 rounded-3xl p-8">
        <div class="text-center mb-6">
            <div class="inline-flex w-16 h-16 rounded-2xl bg-primary items-center justify-center text-primary-content font-bold text-3xl mb-3">P</div>
            <h1 class="text-2xl font-bold">Setup Awal Presensee</h1>
            <p class="text-sm text-base-content/60">Buat akun Superuser (Admin Utama)</p>
        </div>
        {{if .Error}}
        <div class="alert alert-error mb-4 text-sm py-2">
            <span>{{.Error}}</span>
        </div>
        {{end}}
        <form method="POST" action="/setup" class="space-y-4">
            <div class="form-control">
                <label class="label"><span class="label-text font-medium">Username Admin</span></label>
                <input type="text" name="username" required autofocus class="input input-bordered w-full" placeholder="admin" />
            </div>
            <div class="form-control">
                <label class="label"><span class="label-text font-medium">Password</span></label>
                <input type="password" name="password" required class="input input-bordered w-full" placeholder="Minimal 6 karakter" />
            </div>
            <div class="form-control">
                <label class="label"><span class="label-text font-medium">Konfirmasi Password</span></label>
                <input type="password" name="confirm_password" required class="input input-bordered w-full" placeholder="Ulangi password" />
            </div>
            <button type="submit" class="btn btn-primary w-full mt-4">Inisialisasi & Buat Admin</button>
        </form>
    </div>
</body>
</html>
`))

func (h *SetupHandler) SetupPage(w http.ResponseWriter, r *http.Request) {
	var count int64
	h.DB.Model(&model.User{}).Where("is_superuser = true").Count(&count)
	if count > 0 {
		http.Error(w, "Setup sudah dilakukan", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	setupTemplate.Execute(w, map[string]any{
		"Error": r.URL.Query().Get("error"),
	})
}

func (h *SetupHandler) SetupPost(w http.ResponseWriter, r *http.Request) {
	var count int64
	h.DB.Model(&model.User{}).Where("is_superuser = true").Count(&count)
	if count > 0 {
		http.Error(w, "Setup sudah dilakukan", http.StatusForbidden)
		return
	}

	username := strings.ToLower(strings.TrimSpace(r.FormValue("username")))
	password := r.FormValue("password")
	confirm := r.FormValue("confirm_password")

	if username == "" || password == "" {
		http.Redirect(w, r, "/setup?error=Semua+field+wajib+diisi", http.StatusSeeOther)
		return
	}

	if password != confirm {
		http.Redirect(w, r, "/setup?error=Konfirmasi+password+tidak+cocok", http.StatusSeeOther)
		return
	}

	if len(password) < 6 {
		http.Redirect(w, r, "/setup?error=Password+minimal+6+karakter", http.StatusSeeOther)
		return
	}

	// Auto-migrate schema
	if err := model.AutoMigrate(h.DB); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	adminUser := model.User{
		Username:    username,
		FullName:    "Super Administrator",
		IsSuperuser: true,
		IsStaff:     true,
		IsActive:    true,
	}
	if err := adminUser.SetPassword(password); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.DB.Create(&adminUser).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create default DataSekolah if not exists
	var dataCount int64
	h.DB.Model(&model.DataSekolah{}).Count(&dataCount)
	if dataCount == 0 {
		h.DB.Create(&model.DataSekolah{
			NamaSekolah:  "Sekolah Indonesia",
			NamaAplikasi: "Presensee",
		})
	}

	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}
