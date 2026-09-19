package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"presensee/internal/database"
	"presensee/internal/migration"
	"presensee/internal/model"

	"github.com/google/uuid"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "presensee.db"
	}

	db, err := database.Connect(dsn)
	if err != nil {
		log.Fatalf("connect DB: %v", err)
	}

	// Ensure migration up
	if err := migration.MigrateUp(db); err != nil {
		log.Fatalf("migrate up: %v", err)
	}

	fmt.Println("Seeding dummy data...")

	// 1. Data Sekolah
	var dataSekolah model.DataSekolah
	if err := db.First(&dataSekolah).Error; err != nil {
		dataSekolah = model.DataSekolah{
			NamaSekolah:      "SMA Negeri 1 Nusantara",
			NamaAplikasi:     "Presensee",
			DeskripsiSekolah: "Sistem Absensi Digital Offline-First & QR Code Presensee Go",
		}
		db.Create(&dataSekolah)
	} else {
		dataSekolah.NamaSekolah = "SMA Negeri 1 Nusantara"
		dataSekolah.NamaAplikasi = "Presensee"
		db.Save(&dataSekolah)
	}

	// 2. Users Helper
	createUser := func(username, fullName, password string, uType *model.UserType, isStaff, isSuper bool) model.User {
		var u model.User
		if err := db.Where("LOWER(TRIM(username)) = ?", username).First(&u).Error; err != nil {
			u = model.User{
				Username:    username,
				FullName:    fullName,
				Type:        uType,
				IsStaff:     isStaff,
				IsSuperuser: isSuper,
				IsActive:    true,
			}
			u.SetPassword(password)
			db.Create(&u)
		} else {
			u.FullName = fullName
			u.Type = uType
			u.IsStaff = isStaff
			u.IsSuperuser = isSuper
			u.IsActive = true
			u.SetPassword(password)
			db.Save(&u)
		}
		return u
	}

	typeWK := model.TypeWaliKelas
	typeSek := model.TypeSekretaris
	typeKes := model.TypeKesiswaan
	typePik := model.TypeGuruPiket

	// Superusers
	createUser("admin", "Super Administrator", "admin123", nil, true, true)
	createUser("salismazaya", "Salis Mazaya", "admin123", nil, true, true)

	// Wali Kelas
	wali1 := createUser("pak_budi", "Budi Santoso, S.Pd.", "password123", &typeWK, false, false)
	wali2 := createUser("bu_siti", "Siti Aminah, M.Pd.", "password123", &typeWK, false, false)
	wali3 := createUser("darto", "Darto Hendrawan, S.Si.", "password123", &typeWK, false, false)

	// Sekretaris
	sek1 := createUser("ani_sekretaris", "Ani Suryani", "password123", &typeSek, false, false)
	sek2 := createUser("budi_sekretaris", "Budi Pratama", "password123", &typeSek, false, false)

	// Kesiswaan & Guru Piket
	createUser("pak_joko", "Joko Widodo, S.Pd. (Kesiswaan)", "password123", &typeKes, false, false)
	createUser("bu_rini", "Rini Astuti, S.Pd. (Guru Piket)", "password123", &typePik, false, false)

	// 3. Classes
	createKelas := func(name string, waliID *uint, sekIDs []uint) model.Kelas {
		var k model.Kelas
		if err := db.Where("name = ?", name).First(&k).Error; err != nil {
			k = model.Kelas{
				Name:        name,
				Active:      true,
				WaliKelasID: waliID,
			}
			db.Create(&k)
		} else {
			k.Active = true
			k.WaliKelasID = waliID
			db.Save(&k)
		}

		if len(sekIDs) > 0 {
			var sekUsers []model.User
			db.Where("id IN ?", sekIDs).Find(&sekUsers)
			db.Model(&k).Association("Sekretaris").Replace(sekUsers)
		}
		return k
	}

	kelas1 := createKelas("XII IPA 1", &wali1.ID, []uint{sek1.ID})
	kelas2 := createKelas("XII IPA 2", &wali2.ID, []uint{sek2.ID})
	kelas3 := createKelas("XI MIPA 1", &wali3.ID, nil)
	kelas4 := createKelas("X MIPA 1", nil, nil)

	// 4. Siswas
	siswaList := []struct {
		Nama  string
		NIS   string
		NISN  string
		Kelas uint
	}{
		// XII IPA 1
		{"Achmad Fauzi", "20240101", "0051234501", kelas1.ID},
		{"Aditya Pratama", "20240102", "0051234502", kelas1.ID},
		{"Aisyah Nurul", "20240103", "0051234503", kelas1.ID},
		{"Bagus Ramadhan", "20240104", "0051234504", kelas1.ID},
		{"Cahya Maulana", "20240105", "0051234505", kelas1.ID},
		{"Dewi Anggraini", "20240106", "0051234506", kelas1.ID},
		{"Dimas Arya", "20240107", "0051234507", kelas1.ID},
		{"Eka Saputri", "20240108", "0051234508", kelas1.ID},
		{"Fajar Hidayat", "20240109", "0051234509", kelas1.ID},
		{"Gita Permata", "20240110", "0051234510", kelas1.ID},

		// XII IPA 2
		{"Hafiz Al-Fatih", "20240201", "0051234601", kelas2.ID},
		{"Indah Lestari", "20240202", "0051234602", kelas2.ID},
		{"Irfan Hakim", "20240203", "0051234603", kelas2.ID},
		{"Kiki Amelia", "20240204", "0051234604", kelas2.ID},
		{"Luthfi Zaki", "20240205", "0051234605", kelas2.ID},
		{"Mutiara Rahma", "20240206", "0051234606", kelas2.ID},
		{"Naufal Azmi", "20240207", "0051234607", kelas2.ID},
		{"Putri Maharani", "20240208", "0051234608", kelas2.ID},

		// XI MIPA 1 (Wali: darto)
		{"Jajang Nurjaman", "20250101", "0061234501", kelas3.ID},
		{"Asep Sunandar", "20250102", "0061234502", kelas3.ID},
		{"Cecep Supriatna", "20250103", "0061234503", kelas3.ID},
		{"Dadang Konelo", "20250104", "0061234504", kelas3.ID},
		{"Entin Rostini", "20250105", "0061234505", kelas3.ID},
		{"Fitri Handayani", "20250106", "0061234506", kelas3.ID},
		{"Gugun Gunawan", "20250107", "0061234507", kelas3.ID},

		// X MIPA 1
		{"Rizky Febian", "20260101", "0071234501", kelas4.ID},
		{"Siti Badriah", "20260102", "0071234502", kelas4.ID},
		{"Tegar Septian", "20260103", "0071234503", kelas4.ID},
	}

	var createdSiswas []model.Siswa
	for _, s := range siswaList {
		var existing model.Siswa
		if err := db.Where("full_name = ? AND kelas_id = ?", s.Nama, s.Kelas).First(&existing).Error; err != nil {
			newS := model.Siswa{
				FullName: s.Nama,
				KelasID:  s.Kelas,
				NIS:      s.NIS,
				NISN:     s.NISN,
			}
			db.Create(&newS)
			createdSiswas = append(createdSiswas, newS)
		} else {
			createdSiswas = append(createdSiswas, existing)
		}
	}

	// 5. Absensi Session (Jadwal QR Guru Piket)
	var sessionCount int64
	db.Model(&model.AbsensiSession{}).Count(&sessionCount)
	if sessionCount == 0 {
		sess := model.AbsensiSession{
			ID:                  uuid.New(),
			Senin:               true,
			Selasa:              true,
			Rabu:                true,
			Kamis:               true,
			Jumat:               true,
			Sabtu:               false,
			JamMasuk:            "07:00",
			JamMasukToleransi:   900,
			JamKeluarMulaiAbsen: "13:30",
			JamKeluar:           "14:30",
			Kelas:               []model.Kelas{kelas1, kelas2, kelas3, kelas4},
		}
		db.Create(&sess)
	}

	// 6. Absensi Dummy Data (Hari ini & kemarin)
	today := time.Now().Truncate(24 * time.Hour)
	yesterday := today.AddDate(0, 0, -1)

	statuses := []model.AbsensiStatus{
		model.StatusHadir,
		model.StatusHadir,
		model.StatusHadir,
		model.StatusSakit,
		model.StatusIzin,
		model.StatusHadir,
		model.StatusHadir,
		model.StatusAlfa,
		model.StatusHadir,
		model.StatusHadir,
	}

	recordAbsensi := func(s model.Siswa, d time.Time, st model.AbsensiStatus, byID uint) {
		var a model.Absensi
		if err := db.Where("date = ? AND siswa_id = ?", d, s.ID).First(&a).Error; err != nil {
			a = model.Absensi{
				Date:      d,
				SiswaID:   s.ID,
				Status:    st,
				ByID:      &byID,
				CreatedAt: d,
				UpdatedAt: d,
			}
			db.Create(&a)
		}
	}

	for i, s := range createdSiswas {
		stToday := statuses[i%len(statuses)]
		stYest := statuses[(i+2)%len(statuses)]
		recordAbsensi(s, today, stToday, wali1.ID)
		recordAbsensi(s, yesterday, stYest, wali1.ID)
	}

	fmt.Println("✓ Dummy data successfully seeded!")
	fmt.Printf("  • %d Users (admin, pak_budi, bu_siti, darto, ani_sekretaris, pak_joko, bu_rini)\n", 7)
	fmt.Printf("  • %d Classes (XII IPA 1, XII IPA 2, XI MIPA 1, X MIPA 1)\n", 4)
	fmt.Printf("  • %d Students across classes\n", len(siswaList))
	fmt.Printf("  • Absensi sessions & attendance records generated\n")
}
