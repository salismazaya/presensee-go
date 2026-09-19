package model

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/pbkdf2"
	"gorm.io/gorm"
)

type UserType string

const (
	TypeWaliKelas  UserType = "wali_kelas"
	TypeKesiswaan  UserType = "kesiswaan"
	TypeSekretaris UserType = "sekretaris"
	TypeGuruPiket  UserType = "guru_piket"
)

type User struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Username    string    `gorm:"size:50;uniqueIndex;not null" json:"username"`
	Password    string    `gorm:"size:255;not null" json:"-"`
	FullName    string    `gorm:"size:100" json:"full_name"`
	Type        *UserType `gorm:"size:20;index" json:"type"`
	Token       string    `gorm:"size:50;index" json:"token,omitempty"`
	IsSuperuser bool      `gorm:"default:false" json:"is_superuser"`
	IsStaff     bool      `gorm:"default:false" json:"is_staff"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	Photo       string    `gorm:"size:255" json:"photo,omitempty"`
	DateJoined  time.Time `gorm:"autoCreateTime" json:"date_joined"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	WaliKelas       *Kelas  `gorm:"foreignKey:WaliKelasID" json:"wali_kelas,omitempty"`
	SekretarisKelas []Kelas `gorm:"many2many:kelas_sekretaris;" json:"sekretaris_kelas,omitempty"`
}

func (u *User) SetPassword(raw string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return nil
}

func (u *User) CheckPassword(raw string) bool {
	if strings.HasPrefix(u.Password, "pbkdf2_sha256$") {
		parts := strings.Split(u.Password, "$")
		if len(parts) == 4 {
			iter, err := strconv.Atoi(parts[1])
			if err == nil {
				salt := parts[2]
				expectedHash := parts[3]
				key := pbkdf2.Key([]byte(raw), []byte(salt), iter, 32, sha256.New)
				computedHash := base64.StdEncoding.EncodeToString(key)
				if subtle.ConstantTimeCompare([]byte(computedHash), []byte(expectedHash)) == 1 {
					return true
				}
			}
		}
	}
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(raw)) == nil
}

func (u *User) DisplayName() string {
	if u.FullName != "" {
		return u.FullName
	}
	return u.Username
}

func GenerateToken(length int) string {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return uuid.New().String()[:length]
	}
	return hex.EncodeToString(bytes)
}

type Kelas struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:50;uniqueIndex;not null" json:"name"`
	Active      bool      `gorm:"default:true;index" json:"active"`
	WaliKelasID *uint     `gorm:"uniqueIndex" json:"wali_kelas_id"`
	WaliKelas   *User     `gorm:"foreignKey:WaliKelasID;constraint:OnDelete:SET NULL" json:"wali_kelas,omitempty"`
	Sekretaris  []User    `gorm:"many2many:kelas_sekretaris;" json:"sekretaris,omitempty"`
	Siswas      []Siswa   `gorm:"foreignKey:KelasID;constraint:OnDelete:RESTRICT" json:"siswas,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Siswa struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	FullName  string    `gorm:"size:100;not null;index" json:"fullname"`
	KelasID   uint      `gorm:"not null;index" json:"kelas_id"`
	Kelas     Kelas     `gorm:"foreignKey:KelasID;constraint:OnDelete:RESTRICT" json:"kelas"`
	NIS       string    `gorm:"size:20;index" json:"nis"`
	NISN      string    `gorm:"size:20;index" json:"nisn"`
	Photo     string    `gorm:"size:255" json:"photo,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AbsensiStatus string

const (
	StatusHadir  AbsensiStatus = "hadir"
	StatusSakit  AbsensiStatus = "sakit"
	StatusIzin   AbsensiStatus = "izin"
	StatusAlfa   AbsensiStatus = "alfa"
	StatusBolos  AbsensiStatus = "bolos"
	StatusTunggu AbsensiStatus = "tunggu"
)

type Absensi struct {
	ID            uint          `gorm:"primaryKey" json:"id"`
	Date          time.Time     `gorm:"type:date;not null;uniqueIndex:idx_date_siswa" json:"date"`
	SiswaID       uint          `gorm:"not null;uniqueIndex:idx_date_siswa" json:"siswa_id"`
	Siswa         Siswa         `gorm:"foreignKey:SiswaID;constraint:OnDelete:RESTRICT" json:"siswa"`
	Status        AbsensiStatus `gorm:"column:_status;size:30;not null" json:"status"`
	WaitExpiredAt *time.Time    `gorm:"index" json:"wait_expired_at,omitempty"`
	ByID          *uint         `gorm:"index" json:"by_id"`
	By            *User         `gorm:"foreignKey:ByID;constraint:OnDelete:SET NULL" json:"by,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

func (a *Absensi) FinalStatus() AbsensiStatus {
	if a.Status == StatusTunggu {
		if a.WaitExpiredAt != nil && time.Now().After(*a.WaitExpiredAt) {
			return StatusBolos
		}
	}
	return a.Status
}

type KunciAbsensi struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Date      time.Time `gorm:"type:date;not null;uniqueIndex:idx_kunci_kelas_date" json:"date"`
	KelasID   uint      `gorm:"not null;uniqueIndex:idx_kunci_kelas_date" json:"kelas_id"`
	Kelas     Kelas     `gorm:"foreignKey:KelasID;constraint:OnDelete:CASCADE" json:"kelas"`
	Locked    bool      `gorm:"default:true" json:"locked"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AbsensiSession struct {
	ID                  uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Senin               bool      `gorm:"default:false" json:"senin"`
	Selasa              bool      `gorm:"default:false" json:"selasa"`
	Rabu                bool      `gorm:"default:false" json:"rabu"`
	Kamis               bool      `gorm:"default:false" json:"kamis"`
	Jumat               bool      `gorm:"default:false" json:"jumat"`
	Sabtu               bool      `gorm:"default:false" json:"sabtu"`
	JamMasuk            string    `gorm:"size:8;not null" json:"jam_masuk"`
	JamMasukToleransi   int64     `gorm:"default:900" json:"jam_masuk_toleransi_seconds"` // default 15 menit
	JamKeluarMulaiAbsen string    `gorm:"size:8" json:"jam_keluar_mulai_absen"`
	JamKeluar           string    `gorm:"size:8;not null" json:"jam_keluar"`
	Kelas               []Kelas   `gorm:"many2many:absensi_session_kelas;" json:"kelas"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

func (s *AbsensiSession) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

type DataSekolah struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	NamaSekolah      string    `gorm:"size:100;not null" json:"nama_sekolah"`
	LogoSekolah      string    `gorm:"size:255" json:"logo_sekolah"`
	DeskripsiSekolah string    `gorm:"type:text" json:"deskripsi_sekolah"`
	KopSekolah       string    `gorm:"size:255" json:"kop_sekolah"`
	NamaAplikasi     string    `gorm:"size:50;default:'Presensee'" json:"nama_aplikasi"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (DataSekolah) TableName() string {
	return "data_sekolah"
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&Kelas{},
		&Siswa{},
		&Absensi{},
		&KunciAbsensi{},
		&AbsensiSession{},
		&DataSekolah{},
	)
}
