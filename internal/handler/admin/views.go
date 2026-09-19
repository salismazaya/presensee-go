package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"presensee/internal/model"
	"presensee/internal/service/excel"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// Dashboard View
func (h *AdminHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	var totalSiswa, totalKelas, totalUser, totalAbsensiToday int64
	today := time.Now().Format("2006-01-02")

	h.DB.Model(&model.Siswa{}).Count(&totalSiswa)
	h.DB.Model(&model.Kelas{}).Where("active = true").Count(&totalKelas)
	h.DB.Model(&model.User{}).Where("is_active = true").Count(&totalUser)
	h.DB.Model(&model.Absensi{}).Where("date = ?", today).Count(&totalAbsensiToday)

	renderTemplate(w, "dashboard", map[string]any{
		"Title":              "Dashboard Admin",
		"TotalSiswa":         totalSiswa,
		"TotalKelas":         totalKelas,
		"TotalUser":          totalUser,
		"TotalAbsensiToday":  totalAbsensiToday,
		"ActiveMenu":         "dashboard",
	})
}

// Users Handlers
func (h *AdminHandler) UsersList(w http.ResponseWriter, r *http.Request) {
	var users []model.User
	h.DB.Preload("WaliKelas").Preload("SekretarisKelas").Order("id DESC").Find(&users)

	renderTemplate(w, "users_list", map[string]any{
		"Title":      "Manajemen Pengguna",
		"Users":      users,
		"ActiveMenu": "users",
	})
}

func (h *AdminHandler) UserCreatePage(w http.ResponseWriter, r *http.Request) {
	var kelass []model.Kelas
	h.DB.Where("active = true").Find(&kelass)

	renderTemplate(w, "user_form", map[string]any{
		"Title":      "Tambah Pengguna",
		"User":       model.User{},
		"Kelass":     kelass,
		"IsNew":      true,
		"ActiveMenu": "users",
	})
}

func (h *AdminHandler) UserCreatePost(w http.ResponseWriter, r *http.Request) {
	username := strings.ToLower(strings.TrimSpace(r.FormValue("username")))
	password := r.FormValue("password")
	fullName := r.FormValue("full_name")
	userTypeStr := r.FormValue("type")
	isStaff := r.FormValue("is_staff") == "on" || r.FormValue("is_staff") == "true"
	isSuperuser := r.FormValue("is_superuser") == "on" || r.FormValue("is_superuser") == "true"
	kelasIDStr := r.FormValue("kelas_id")

	var uType *model.UserType
	if userTypeStr != "" && userTypeStr != "----" {
		t := model.UserType(userTypeStr)
		uType = &t
	}

	user := model.User{
		Username:    username,
		FullName:    fullName,
		Type:        uType,
		IsStaff:     isStaff,
		IsSuperuser: isSuperuser,
		IsActive:    true,
	}
	if err := user.SetPassword(password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		if kelasIDStr != "" && kelasIDStr != "----" && uType != nil {
			kID, _ := strconv.ParseUint(kelasIDStr, 10, 32)
			if *uType == model.TypeWaliKelas {
				tx.Model(&model.Kelas{}).Where("id = ?", kID).Update("wali_kelas_id", user.ID)
			} else if *uType == model.TypeSekretaris {
				var k model.Kelas
				if err := tx.First(&k, kID).Error; err == nil {
					tx.Model(&k).Association("Sekretaris").Append(&user)
				}
			}
		}
		return nil
	})

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

func (h *AdminHandler) UserEditPage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var user model.User
	if err := h.DB.Preload("WaliKelas").Preload("SekretarisKelas").First(&user, id).Error; err != nil {
		http.NotFound(w, r)
		return
	}

	var kelass []model.Kelas
	h.DB.Where("active = true").Find(&kelass)

	var selectedKelasID uint
	if user.WaliKelas != nil {
		selectedKelasID = user.WaliKelas.ID
	} else if len(user.SekretarisKelas) > 0 {
		selectedKelasID = user.SekretarisKelas[0].ID
	}

	renderTemplate(w, "user_form", map[string]any{
		"Title":           "Edit Pengguna: " + user.DisplayName(),
		"User":            user,
		"Kelass":          kelass,
		"SelectedKelasID": selectedKelasID,
		"IsNew":           false,
		"ActiveMenu":      "users",
	})
}

func (h *AdminHandler) UserEditPost(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var user model.User
	if err := h.DB.First(&user, id).Error; err != nil {
		http.NotFound(w, r)
		return
	}

	username := strings.ToLower(strings.TrimSpace(r.FormValue("username")))
	password := r.FormValue("password")
	fullName := r.FormValue("full_name")
	userTypeStr := r.FormValue("type")
	isStaff := r.FormValue("is_staff") == "on" || r.FormValue("is_staff") == "true"
	isSuperuser := r.FormValue("is_superuser") == "on" || r.FormValue("is_superuser") == "true"
	kelasIDStr := r.FormValue("kelas_id")

	var uType *model.UserType
	if userTypeStr != "" && userTypeStr != "----" {
		t := model.UserType(userTypeStr)
		uType = &t
	}

	user.Username = username
	user.FullName = fullName
	user.Type = uType
	user.IsStaff = isStaff
	user.IsSuperuser = isSuperuser
	if password != "" {
		user.SetPassword(password)
	}

	h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&user).Error; err != nil {
			return err
		}

		// Clear previous roles in classes
		tx.Model(&model.Kelas{}).Where("wali_kelas_id = ?", user.ID).Update("wali_kelas_id", nil)
		var kAll []model.Kelas
		tx.Find(&kAll)
		for _, k := range kAll {
			tx.Model(&k).Association("Sekretaris").Delete(&user)
		}

		if kelasIDStr != "" && kelasIDStr != "----" && uType != nil {
			kID, _ := strconv.ParseUint(kelasIDStr, 10, 32)
			if *uType == model.TypeWaliKelas {
				tx.Model(&model.Kelas{}).Where("id = ?", kID).Update("wali_kelas_id", user.ID)
			} else if *uType == model.TypeSekretaris {
				var k model.Kelas
				if err := tx.First(&k, kID).Error; err == nil {
					tx.Model(&k).Association("Sekretaris").Append(&user)
				}
			}
		}
		return nil
	})

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

func (h *AdminHandler) UserDelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.DB.Delete(&model.User{}, id)
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

// Kelas Handlers
func (h *AdminHandler) KelasList(w http.ResponseWriter, r *http.Request) {
	var kelass []model.Kelas
	h.DB.Preload("WaliKelas").Preload("Sekretaris").Order("active DESC, name ASC").Find(&kelass)

	renderTemplate(w, "kelas_list", map[string]any{
		"Title":      "Manajemen Kelas",
		"Kelass":     kelass,
		"ActiveMenu": "kelas",
	})
}

func (h *AdminHandler) KelasCreatePage(w http.ResponseWriter, r *http.Request) {
	var waliUsers []model.User
	var sekUsers []model.User
	h.DB.Where("type = ?", model.TypeWaliKelas).Find(&waliUsers)
	h.DB.Where("type = ?", model.TypeSekretaris).Find(&sekUsers)

	renderTemplate(w, "kelas_form", map[string]any{
		"Title":      "Tambah Kelas",
		"Kelas":      model.Kelas{Active: true},
		"WaliUsers":  waliUsers,
		"SekUsers":   sekUsers,
		"IsNew":      true,
		"ActiveMenu": "kelas",
	})
}

func (h *AdminHandler) KelasCreatePost(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.FormValue("name"))
	active := r.FormValue("active") == "on" || r.FormValue("active") == "true"
	waliIDStr := r.FormValue("wali_kelas_id")

	var waliID *uint
	if waliIDStr != "" && waliIDStr != "0" {
		id, _ := strconv.ParseUint(waliIDStr, 10, 32)
		uID := uint(id)
		waliID = &uID
	}

	kelas := model.Kelas{
		Name:        name,
		Active:      active,
		WaliKelasID: waliID,
	}

	h.DB.Create(&kelas)
	http.Redirect(w, r, "/admin/kelas", http.StatusSeeOther)
}

func (h *AdminHandler) KelasEditPage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var kelas model.Kelas
	if err := h.DB.Preload("WaliKelas").Preload("Sekretaris").Preload("Siswas").First(&kelas, id).Error; err != nil {
		http.NotFound(w, r)
		return
	}

	var waliUsers []model.User
	var sekUsers []model.User
	h.DB.Where("type = ?", model.TypeWaliKelas).Find(&waliUsers)
	h.DB.Where("type = ?", model.TypeSekretaris).Find(&sekUsers)

	renderTemplate(w, "kelas_form", map[string]any{
		"Title":      "Edit Kelas: " + kelas.Name,
		"Kelas":      kelas,
		"WaliUsers":  waliUsers,
		"SekUsers":   sekUsers,
		"IsNew":      false,
		"ActiveMenu": "kelas",
	})
}

func (h *AdminHandler) KelasEditPost(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var kelas model.Kelas
	if err := h.DB.First(&kelas, id).Error; err != nil {
		http.NotFound(w, r)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	active := r.FormValue("active") == "on" || r.FormValue("active") == "true"
	waliIDStr := r.FormValue("wali_kelas_id")

	var waliID *uint
	if waliIDStr != "" && waliIDStr != "0" {
		id, _ := strconv.ParseUint(waliIDStr, 10, 32)
		uID := uint(id)
		waliID = &uID
	}

	kelas.Name = name
	kelas.Active = active
	kelas.WaliKelasID = waliID

	h.DB.Save(&kelas)
	http.Redirect(w, r, "/admin/kelas", http.StatusSeeOther)
}

func (h *AdminHandler) KelasDelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.DB.Delete(&model.Kelas{}, id)
	http.Redirect(w, r, "/admin/kelas", http.StatusSeeOther)
}

type NaikKelasRequest struct {
	OldKelasID   uint   `json:"old_kelas_id"`
	NewKelasName string `json:"new_kelas_name"`
}

func (h *AdminHandler) NaikKelas(w http.ResponseWriter, r *http.Request) {
	var req NaikKelasRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "NOT_OK", http.StatusBadRequest)
		return
	}

	if req.OldKelasID == 0 || req.NewKelasName == "" {
		http.Error(w, "NOT_OK", http.StatusBadRequest)
		return
	}

	var newKelasID uint
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		var oldKelas model.Kelas
		if err := tx.First(&oldKelas, req.OldKelasID).Error; err != nil {
			return err
		}

		newKelas := model.Kelas{
			Name:   req.NewKelasName,
			Active: true,
		}
		if err := tx.Create(&newKelas).Error; err != nil {
			return err
		}

		oldKelas.Active = false
		if err := tx.Save(&oldKelas).Error; err != nil {
			return err
		}

		if err := tx.Model(&model.Siswa{}).Where("kelas_id = ?", oldKelas.ID).Update("kelas_id", newKelas.ID).Error; err != nil {
			return err
		}

		newKelasID = newKelas.ID
		return nil
	})

	if err != nil {
		http.Error(w, "NOT_OK", http.StatusInternalServerError)
		return
	}

	w.Write([]byte(strconv.Itoa(int(newKelasID))))
}

// Siswa Handlers
func (h *AdminHandler) SiswaList(w http.ResponseWriter, r *http.Request) {
	var siswas []model.Siswa
	kelasIDStr := r.URL.Query().Get("kelas_id")
	search := r.URL.Query().Get("search")

	q := h.DB.Preload("Kelas").Order("kelas_id ASC, full_name ASC")
	if kelasIDStr != "" {
		q = q.Where("kelas_id = ?", kelasIDStr)
	}
	if search != "" {
		q = q.Where("full_name LIKE ? OR nis LIKE ? OR nisn LIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	q.Find(&siswas)

	var kelass []model.Kelas
	h.DB.Where("active = true").Find(&kelass)

	renderTemplate(w, "siswa_list", map[string]any{
		"Title":           "Manajemen Siswa",
		"Siswas":          siswas,
		"Kelass":          kelass,
		"SelectedKelasID": kelasIDStr,
		"Search":          search,
		"ActiveMenu":      "siswa",
	})
}

func (h *AdminHandler) SiswaCreatePage(w http.ResponseWriter, r *http.Request) {
	var kelass []model.Kelas
	h.DB.Where("active = true").Find(&kelass)

	renderTemplate(w, "siswa_form", map[string]any{
		"Title":      "Tambah Siswa",
		"Siswa":      model.Siswa{},
		"Kelass":     kelass,
		"IsNew":      true,
		"ActiveMenu": "siswa",
	})
}

func (h *AdminHandler) SiswaCreatePost(w http.ResponseWriter, r *http.Request) {
	fullName := strings.TrimSpace(r.FormValue("full_name"))
	kelasID, _ := strconv.ParseUint(r.FormValue("kelas_id"), 10, 32)
	nis := strings.TrimSpace(r.FormValue("nis"))
	nisn := strings.TrimSpace(r.FormValue("nisn"))

	siswa := model.Siswa{
		FullName: fullName,
		KelasID:  uint(kelasID),
		NIS:      nis,
		NISN:     nisn,
	}

	h.DB.Create(&siswa)
	http.Redirect(w, r, "/admin/siswa", http.StatusSeeOther)
}

func (h *AdminHandler) SiswaEditPage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var siswa model.Siswa
	if err := h.DB.First(&siswa, id).Error; err != nil {
		http.NotFound(w, r)
		return
	}

	var kelass []model.Kelas
	h.DB.Where("active = true").Find(&kelass)

	renderTemplate(w, "siswa_form", map[string]any{
		"Title":      "Edit Siswa: " + siswa.FullName,
		"Siswa":      siswa,
		"Kelass":     kelass,
		"IsNew":      false,
		"ActiveMenu": "siswa",
	})
}

func (h *AdminHandler) SiswaEditPost(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var siswa model.Siswa
	if err := h.DB.First(&siswa, id).Error; err != nil {
		http.NotFound(w, r)
		return
	}

	fullName := strings.TrimSpace(r.FormValue("full_name"))
	kelasID, _ := strconv.ParseUint(r.FormValue("kelas_id"), 10, 32)
	nis := strings.TrimSpace(r.FormValue("nis"))
	nisn := strings.TrimSpace(r.FormValue("nisn"))

	siswa.FullName = fullName
	siswa.KelasID = uint(kelasID)
	siswa.NIS = nis
	siswa.NISN = nisn

	h.DB.Save(&siswa)
	http.Redirect(w, r, "/admin/siswa", http.StatusSeeOther)
}

func (h *AdminHandler) SiswaDelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.DB.Delete(&model.Siswa{}, id)
	http.Redirect(w, r, "/admin/siswa", http.StatusSeeOther)
}

// Absensi Handlers
func (h *AdminHandler) AbsensiList(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}
	kelasIDStr := r.URL.Query().Get("kelas_id")

	var absensies []model.Absensi
	q := h.DB.Preload("Siswa.Kelas").Preload("By").Where("date = ?", dateStr).Order("siswa_id ASC")
	if kelasIDStr != "" {
		q = q.Joins("JOIN siswas ON siswas.id = absensis.siswa_id").Where("siswas.kelas_id = ?", kelasIDStr)
	}
	q.Find(&absensies)

	var kelass []model.Kelas
	h.DB.Where("active = true").Find(&kelass)

	renderTemplate(w, "absensi_list", map[string]any{
		"Title":           "Monitoring Absensi",
		"Absensies":       absensies,
		"Kelass":          kelass,
		"SelectedDate":    dateStr,
		"SelectedKelasID": kelasIDStr,
		"ActiveMenu":      "absensi",
	})
}

// Kunci Handlers
func (h *AdminHandler) KunciList(w http.ResponseWriter, r *http.Request) {
	var kuncis []model.KunciAbsensi
	h.DB.Preload("Kelas").Order("date DESC, kelas_id ASC").Find(&kuncis)

	var kelass []model.Kelas
	h.DB.Where("active = true").Find(&kelass)

	renderTemplate(w, "kunci_list", map[string]any{
		"Title":      "Kunci Absensi",
		"Kuncis":     kuncis,
		"Kelass":     kelass,
		"ActiveMenu": "kunci",
	})
}

func (h *AdminHandler) KunciToggle(w http.ResponseWriter, r *http.Request) {
	kelasID, _ := strconv.ParseUint(r.FormValue("kelas_id"), 10, 32)
	dateStr := r.FormValue("date")
	locked := r.FormValue("locked") == "true" || r.FormValue("locked") == "1"

	d, _ := time.Parse("2006-01-02", dateStr)
	var k model.KunciAbsensi
	err := h.DB.Where("kelas_id = ? AND date = ?", kelasID, d).First(&k).Error
	if err != nil {
		k = model.KunciAbsensi{
			KelasID: uint(kelasID),
			Date:    d,
			Locked:  locked,
		}
		h.DB.Create(&k)
	} else {
		k.Locked = locked
		h.DB.Save(&k)
	}

	http.Redirect(w, r, "/admin/kunci", http.StatusSeeOther)
}

// Sessions Handlers
func (h *AdminHandler) SessionsList(w http.ResponseWriter, r *http.Request) {
	var sessions []model.AbsensiSession
	h.DB.Preload("Kelas").Find(&sessions)

	renderTemplate(w, "sessions_list", map[string]any{
		"Title":      "Jadwal Sesi Absensi (QR)",
		"Sessions":   sessions,
		"ActiveMenu": "sessions",
	})
}

func (h *AdminHandler) SessionCreatePage(w http.ResponseWriter, r *http.Request) {
	var kelass []model.Kelas
	h.DB.Where("active = true").Find(&kelass)

	renderTemplate(w, "session_form", map[string]any{
		"Title":      "Tambah Sesi Jadwal QR",
		"Session":    model.AbsensiSession{},
		"Kelass":     kelass,
		"IsNew":      true,
		"ActiveMenu": "sessions",
	})
}

func (h *AdminHandler) SessionCreatePost(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	jamMasuk := r.FormValue("jam_masuk")
	jamKeluar := r.FormValue("jam_keluar")
	jamKeluarMulai := r.FormValue("jam_keluar_mulai_absen")
	kelasIDs := r.Form["kelas_ids"]

	session := model.AbsensiSession{
		Senin:               r.FormValue("senin") == "on",
		Selasa:              r.FormValue("selasa") == "on",
		Rabu:                r.FormValue("rabu") == "on",
		Kamis:               r.FormValue("kamis") == "on",
		Jumat:               r.FormValue("jumat") == "on",
		Sabtu:               r.FormValue("sabtu") == "on",
		JamMasuk:            jamMasuk,
		JamKeluar:           jamKeluar,
		JamKeluarMulaiAbsen: jamKeluarMulai,
		JamMasukToleransi:   900,
	}

	h.DB.Create(&session)

	for _, kID := range kelasIDs {
		id, _ := strconv.ParseUint(kID, 10, 32)
		var k model.Kelas
		if err := h.DB.First(&k, id).Error; err == nil {
			h.DB.Model(&session).Association("Kelas").Append(&k)
		}
	}

	http.Redirect(w, r, "/admin/sessions", http.StatusSeeOther)
}

func (h *AdminHandler) SessionEditPage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var session model.AbsensiSession
	if err := h.DB.Preload("Kelas").First(&session, "id = ?", id).Error; err != nil {
		http.NotFound(w, r)
		return
	}

	var kelass []model.Kelas
	h.DB.Where("active = true").Find(&kelass)

	renderTemplate(w, "session_form", map[string]any{
		"Title":      "Edit Sesi Jadwal QR",
		"Session":    session,
		"Kelass":     kelass,
		"IsNew":      false,
		"ActiveMenu": "sessions",
	})
}

func (h *AdminHandler) SessionEditPost(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var session model.AbsensiSession
	if err := h.DB.First(&session, "id = ?", id).Error; err != nil {
		http.NotFound(w, r)
		return
	}

	r.ParseForm()
	session.Senin = r.FormValue("senin") == "on"
	session.Selasa = r.FormValue("selasa") == "on"
	session.Rabu = r.FormValue("rabu") == "on"
	session.Kamis = r.FormValue("kamis") == "on"
	session.Jumat = r.FormValue("jumat") == "on"
	session.Sabtu = r.FormValue("sabtu") == "on"
	session.JamMasuk = r.FormValue("jam_masuk")
	session.JamKeluar = r.FormValue("jam_keluar")
	session.JamKeluarMulaiAbsen = r.FormValue("jam_keluar_mulai_absen")

	h.DB.Save(&session)

	// Replace associated classes
	h.DB.Model(&session).Association("Kelas").Clear()
	for _, kID := range r.Form["kelas_ids"] {
		idVal, _ := strconv.ParseUint(kID, 10, 32)
		var k model.Kelas
		if err := h.DB.First(&k, idVal).Error; err == nil {
			h.DB.Model(&session).Association("Kelas").Append(&k)
		}
	}

	http.Redirect(w, r, "/admin/sessions", http.StatusSeeOther)
}

func (h *AdminHandler) SessionDelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.DB.Delete(&model.AbsensiSession{}, "id = ?", id)
	http.Redirect(w, r, "/admin/sessions", http.StatusSeeOther)
}

// Data Sekolah Handler
func (h *AdminHandler) DataSekolahPage(w http.ResponseWriter, r *http.Request) {
	var data model.DataSekolah
	h.DB.Last(&data)

	renderTemplate(w, "data_sekolah", map[string]any{
		"Title":      "Data & Pengaturan Sekolah",
		"Data":       data,
		"ActiveMenu": "data_sekolah",
	})
}

func (h *AdminHandler) DataSekolahPost(w http.ResponseWriter, r *http.Request) {
	var data model.DataSekolah
	h.DB.Last(&data)

	data.NamaSekolah = r.FormValue("nama_sekolah")
	data.NamaAplikasi = r.FormValue("nama_aplikasi")
	data.DeskripsiSekolah = r.FormValue("deskripsi_sekolah")

	if data.ID == 0 {
		h.DB.Create(&data)
	} else {
		h.DB.Save(&data)
	}

	http.Redirect(w, r, "/admin/data-sekolah", http.StatusSeeOther)
}

// Import & Export Handlers
func (h *AdminHandler) ImportSiswaPage(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "import_siswa", map[string]any{
		"Title":      "Import Data Siswa dari Excel",
		"ActiveMenu": "import",
	})
}

func (h *AdminHandler) ImportSiswaPost(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("excel_file")
	if err != nil {
		http.Error(w, "File excel wajib diunggah", http.StatusBadRequest)
		return
	}
	defer file.Close()

	res, err := excel.ImportSiswa(h.DB, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	renderTemplate(w, "import_siswa", map[string]any{
		"Title":      "Import Data Siswa dari Excel",
		"SuccessMsg": fmt.Sprintf("Berhasil mengimpor %d siswa! (Dilewati: %d siswa duplikat)", res.Success, res.Skipped),
		"ActiveMenu": "import",
	})
}

func (h *AdminHandler) ExportAbsensiPage(w http.ResponseWriter, r *http.Request) {
	var kelass []model.Kelas
	h.DB.Where("active = true").Find(&kelass)

	renderTemplate(w, "export_absensi", map[string]any{
		"Title":        "Export Laporan Absensi (.xlsx)",
		"Kelass":       kelass,
		"CurrentYear":  time.Now().Year(),
		"CurrentMonth": int(time.Now().Month()),
		"ActiveMenu":   "export",
	})
}

func (h *AdminHandler) ExportAbsensiPost(w http.ResponseWriter, r *http.Request) {
	kelasID, _ := strconv.ParseUint(r.FormValue("kelas_id"), 10, 32)
	month, _ := strconv.Atoi(r.FormValue("month"))
	year, _ := strconv.Atoi(r.FormValue("year"))

	bytes, filename, err := excel.ExportAbsensi(h.DB, uint(kelasID), year, month)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Write(bytes)
}

func (h *AdminHandler) KartuSiswaPage(w http.ResponseWriter, r *http.Request) {
	kelasID := r.URL.Query().Get("kelas_id")
	var siswas []model.Siswa
	q := h.DB.Preload("Kelas").Limit(100)
	if kelasID != "" {
		q = q.Where("kelas_id = ?", kelasID)
	}
	q.Find(&siswas)

	var data model.DataSekolah
	h.DB.Last(&data)

	renderTemplate(w, "kartu_siswa", map[string]any{
		"Title":       "Cetak Kartu Siswa",
		"Siswas":      siswas,
		"DataSekolah": data,
	})
}
