package dump

import (
	"fmt"
	"strings"

	"presensee/internal/model"

	"gorm.io/gorm"
)

// GenerateDump creates minimized SQL dump for sql.js offline client.
func GenerateDump(db *gorm.DB, user *model.User) (string, error) {
	var kelass []model.Kelas
	var siswas []model.Siswa
	var absensies []model.Absensi
	var kunciList []model.KunciAbsensi

	// Query based on role
	kelasQuery := db.Model(&model.Kelas{}).Where("active = ?", true)
	siswaQuery := db.Model(&model.Siswa{}).Joins("JOIN kelas ON kelas.id = siswas.kelas_id AND kelas.active = true")
	absensiQuery := db.Model(&model.Absensi{}).Joins("JOIN siswas ON siswas.id = absensis.siswa_id").Joins("JOIN kelas ON kelas.id = siswas.kelas_id AND kelas.active = true")
	kunciQuery := db.Model(&model.KunciAbsensi{}).Joins("JOIN kelas ON kelas.id = kunci_absensis.kelas_id AND kelas.active = true").Where("locked = ?", true)

	if user.Type != nil {
		switch *user.Type {
		case model.TypeKesiswaan:
			// Full access to all active
		case model.TypeWaliKelas:
			kelasQuery = kelasQuery.Where("wali_kelas_id = ?", user.ID)
			siswaQuery = siswaQuery.Where("kelas.wali_kelas_id = ?", user.ID)
			absensiQuery = absensiQuery.Where("kelas.wali_kelas_id = ?", user.ID)
		case model.TypeSekretaris:
			kelasQuery = kelasQuery.Joins("JOIN kelas_sekretaris ON kelas_sekretaris.kelas_id = kelas.id AND kelas_sekretaris.user_id = ?", user.ID)
			siswaQuery = siswaQuery.Joins("JOIN kelas_sekretaris ON kelas_sekretaris.kelas_id = kelas.id AND kelas_sekretaris.user_id = ?", user.ID)
			absensiQuery = absensiQuery.Joins("JOIN kelas_sekretaris ON kelas_sekretaris.kelas_id = kelas.id AND kelas_sekretaris.user_id = ?", user.ID)
		default:
			return "", nil
		}
	} else if !user.IsSuperuser {
		return "", nil
	}

	if err := kelasQuery.Find(&kelass).Error; err != nil {
		return "", err
	}
	if err := siswaQuery.Find(&siswas).Error; err != nil {
		return "", err
	}
	if err := absensiQuery.Find(&absensies).Error; err != nil {
		return "", err
	}
	if err := kunciQuery.Find(&kunciList).Error; err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString("BEGIN TRANSACTION;\n")
	sb.WriteString(`CREATE TABLE kelas (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL);` + "\n")
	sb.WriteString(`CREATE TABLE siswa (id INTEGER PRIMARY KEY AUTOINCREMENT, fullname TEXT NOT NULL, kelas_id INTEGER NOT NULL, FOREIGN KEY (kelas_id) REFERENCES kelas(id) ON DELETE RESTRICT);` + "\n")
	sb.WriteString(`CREATE TABLE absensi (id INTEGER PRIMARY KEY AUTOINCREMENT, date DATE NOT NULL, siswa_id INTEGER NOT NULL, status TEXT NOT NULL, previous_status TEXT, updated_at INTEGER, FOREIGN KEY (siswa_id) REFERENCES siswa(id) ON DELETE RESTRICT, UNIQUE(date, siswa_id));` + "\n")
	sb.WriteString(`CREATE TABLE kunci_absensi (id INTEGER PRIMARY KEY AUTOINCREMENT, date DATE NOT NULL, kelas_id INTEGER NOT NULL);` + "\n")
	sb.WriteString(`CREATE TABLE kelas_sekretaris (id INTEGER PRIMARY KEY AUTOINCREMENT, kelas_id INTEGER NOT NULL, siswa_id INTEGER NOT NULL, FOREIGN KEY (kelas_id) REFERENCES kelas(id) ON DELETE CASCADE, FOREIGN KEY (siswa_id) REFERENCES siswa(id) ON DELETE CASCADE, UNIQUE(kelas_id, siswa_id));` + "\n")

	// Batch Kelas
	if len(kelass) > 0 {
		var vals []string
		for _, k := range kelass {
			vals = append(vals, fmt.Sprintf("(%d, '%s')", k.ID, escapeSQL(k.Name)))
		}
		sb.WriteString(fmt.Sprintf("INSERT INTO \"kelas\" VALUES %s;\n", strings.Join(vals, ", ")))
	}

	// Batch Siswa
	if len(siswas) > 0 {
		var vals []string
		for _, s := range siswas {
			vals = append(vals, fmt.Sprintf("(%d, '%s', %d)", s.ID, escapeSQL(s.FullName), s.KelasID))
		}
		sb.WriteString(fmt.Sprintf("INSERT INTO \"siswa\" VALUES %s;\n", strings.Join(vals, ", ")))
	}

	// Batch Absensi
	if len(absensies) > 0 {
		var vals []string
		for _, a := range absensies {
			dateStr := a.Date.Format("2006-01-02")
			vals = append(vals, fmt.Sprintf("(%d, '%s', %d, '%s', NULL, %d)", a.ID, dateStr, a.SiswaID, a.FinalStatus(), a.UpdatedAt.Unix()))
		}
		sb.WriteString(fmt.Sprintf("INSERT INTO \"absensi\" VALUES %s;\n", strings.Join(vals, ", ")))
	}

	// Batch Kunci
	if len(kunciList) > 0 {
		var vals []string
		for _, k := range kunciList {
			dateStr := k.Date.Format("2006-01-02")
			vals = append(vals, fmt.Sprintf("(%d, '%s', %d)", k.ID, dateStr, k.KelasID))
		}
		sb.WriteString(fmt.Sprintf("INSERT INTO \"kunci_absensi\" VALUES %s;\n", strings.Join(vals, ", ")))
	}

	sb.WriteString("COMMIT;")
	return sb.String(), nil
}

func escapeSQL(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
