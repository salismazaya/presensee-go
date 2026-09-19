package excel

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"presensee/internal/model"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type ImportResult struct {
	Success int
	Skipped int
}

func ImportSiswa(db *gorm.DB, reader io.Reader) (*ImportResult, error) {
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca file excel: %w", err)
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca sheet: %w", err)
	}

	res := &ImportResult{}

	err = db.Transaction(func(tx *gorm.DB) error {
		for i, row := range rows {
			if i == 0 || len(row) < 2 {
				continue // skip header / baris kosong
			}

			rawNama := strings.TrimSpace(row[0])
			rawKelas := strings.TrimSpace(row[1])
			if rawNama == "" || rawKelas == "" {
				continue
			}

			var rawNIS, rawNISN string
			if len(row) > 2 {
				rawNIS = strings.TrimSpace(row[2])
			}
			if len(row) > 3 {
				rawNISN = strings.TrimSpace(row[3])
			}

			if rawNIS != "" {
				var count int64
				tx.Model(&model.Siswa{}).Where("nis = ?", rawNIS).Count(&count)
				if count > 0 {
					res.Skipped++
					continue
				}
			}

			if rawNISN != "" {
				var count int64
				tx.Model(&model.Siswa{}).Where("nisn = ?", rawNISN).Count(&count)
				if count > 0 {
					res.Skipped++
					continue
				}
			}

			var kelas model.Kelas
			if err := tx.Where("name = ?", rawKelas).FirstOrCreate(&kelas, model.Kelas{Name: rawKelas, Active: true}).Error; err != nil {
				return err
			}

			siswa := model.Siswa{
				FullName: rawNama,
				KelasID:  kelas.ID,
				NIS:      rawNIS,
				NISN:     rawNISN,
			}
			if err := tx.Create(&siswa).Error; err != nil {
				return err
			}
			res.Success++
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return res, nil
}

func ExportAbsensi(db *gorm.DB, kelasID uint, year int, month int) ([]byte, string, error) {
	var kelas model.Kelas
	if err := db.First(&kelas, kelasID).Error; err != nil {
		return nil, "", fmt.Errorf("kelas tidak ditemukan")
	}

	var siswas []model.Siswa
	if err := db.Where("kelas_id = ?", kelasID).Order("full_name ASC").Find(&siswas).Error; err != nil {
		return nil, "", err
	}

	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 1, -1)
	numDays := endDate.Day()

	var absensies []model.Absensi
	db.Where("siswa_id IN (?) AND date >= ? AND date <= ?",
		db.Model(&model.Siswa{}).Select("id").Where("kelas_id = ?", kelasID),
		startDate, endDate).Find(&absensies)

	// Map: "siswaID_day" -> status
	statusMap := make(map[string]string)
	for _, a := range absensies {
		key := fmt.Sprintf("%d_%d", a.SiswaID, a.Date.Day())
		statusMap[key] = string(a.FinalStatus())
	}

	f := excelize.NewFile()
	defer f.Close()

	sheet := "Rekap Absensi"
	f.SetSheetName("Sheet1", sheet)

	// Header Baris 1: Judul
	totalCols := 3 + numDays + 5
	lastColName, _ := excelize.ColumnNumberToName(totalCols)
	f.MergeCell(sheet, "A1", fmt.Sprintf("%s1", lastColName))
	f.SetCellValue(sheet, "A1", fmt.Sprintf("LAPORAN ABSENSI KELAS %s - PERIODE %d/%d", strings.ToUpper(kelas.Name), month, year))

	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	f.SetCellStyle(sheet, "A1", fmt.Sprintf("%s1", lastColName), titleStyle)

	// Header Baris 2: Kolom
	headers := []string{"No", "NIS", "Nama Siswa"}
	for d := 1; d <= numDays; d++ {
		headers = append(headers, fmt.Sprintf("%d", d))
	}
	headers = append(headers, "H", "S", "I", "A", "B")

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"4F81BD"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	for colIdx, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 2)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	centerStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	// Data rows
	for rowIdx, s := range siswas {
		rowNum := rowIdx + 3
		f.SetCellValue(sheet, fmt.Sprintf("A%d", rowNum), rowIdx+1)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", rowNum), s.NIS)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", rowNum), s.FullName)

		counts := map[string]int{"hadir": 0, "sakit": 0, "izin": 0, "alfa": 0, "bolos": 0}

		for d := 1; d <= numDays; d++ {
			colName, _ := excelize.ColumnNumberToName(3 + d)
			cell := fmt.Sprintf("%s%d", colName, rowNum)
			st, ok := statusMap[fmt.Sprintf("%d_%d", s.ID, d)]
			if ok && st != "" {
				short := strings.ToUpper(st[:1])
				f.SetCellValue(sheet, cell, short)
				counts[st]++
			} else {
				f.SetCellValue(sheet, cell, "-")
			}
			f.SetCellStyle(sheet, cell, cell, centerStyle)
		}

		// Summary columns
		summaryOffset := 3 + numDays
		f.SetCellValue(sheet, fmt.Sprintf("%s%d", getColName(summaryOffset+1), rowNum), counts["hadir"])
		f.SetCellValue(sheet, fmt.Sprintf("%s%d", getColName(summaryOffset+2), rowNum), counts["sakit"])
		f.SetCellValue(sheet, fmt.Sprintf("%s%d", getColName(summaryOffset+3), rowNum), counts["izin"])
		f.SetCellValue(sheet, fmt.Sprintf("%s%d", getColName(summaryOffset+4), rowNum), counts["alfa"])
		f.SetCellValue(sheet, fmt.Sprintf("%s%d", getColName(summaryOffset+5), rowNum), counts["bolos"])
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("Absensi_%s_%d_%d.xlsx", strings.ReplaceAll(kelas.Name, " ", "_"), year, month)
	return buf.Bytes(), filename, nil
}

func getColName(n int) string {
	name, _ := excelize.ColumnNumberToName(n)
	return name
}
