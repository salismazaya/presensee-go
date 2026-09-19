package migration

var Migrations = []Migration{
	{
		Version: 1,
		Name:    "create_users",
		Up: `CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			full_name TEXT,
			type TEXT,
			token TEXT,
			is_superuser INTEGER DEFAULT 0,
			is_staff INTEGER DEFAULT 0,
			is_active INTEGER DEFAULT 1,
			photo TEXT,
			date_joined DATETIME,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		Down: `DROP TABLE IF EXISTS users`,
	},
	{
		Version: 2,
		Name:    "create_kelas",
		Up: `CREATE TABLE IF NOT EXISTS kelas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			active INTEGER DEFAULT 1,
			wali_kelas_id INTEGER UNIQUE REFERENCES users(id),
			created_at DATETIME,
			updated_at DATETIME
		)`,
		Down: `DROP TABLE IF EXISTS kelas`,
	},
	{
		Version: 3,
		Name:    "create_kelas_sekretaris",
		Up: `CREATE TABLE IF NOT EXISTS kelas_sekretaris (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			kelas_id INTEGER REFERENCES kelas(id),
			user_id INTEGER REFERENCES users(id),
			UNIQUE(kelas_id, user_id)
		)`,
		Down: `DROP TABLE IF EXISTS kelas_sekretaris`,
	},
	{
		Version: 4,
		Name:    "create_siswas",
		Up: `CREATE TABLE IF NOT EXISTS siswas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			full_name TEXT NOT NULL,
			kelas_id INTEGER NOT NULL REFERENCES kelas(id),
			nis TEXT,
			nisn TEXT,
			photo TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		Down: `DROP TABLE IF EXISTS siswas`,
	},
	{
		Version: 5,
		Name:    "create_absensis",
		Up: `CREATE TABLE IF NOT EXISTS absensis (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date DATETIME NOT NULL,
			siswa_id INTEGER NOT NULL REFERENCES siswas(id),
			_status TEXT NOT NULL,
			wait_expired_at DATETIME,
			by_id INTEGER REFERENCES users(id),
			created_at DATETIME,
			updated_at DATETIME,
			UNIQUE(date, siswa_id)
		)`,
		Down: `DROP TABLE IF EXISTS absensis`,
	},
	{
		Version: 6,
		Name:    "create_kunci_absensis",
		Up: `CREATE TABLE IF NOT EXISTS kunci_absensis (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date DATETIME NOT NULL,
			kelas_id INTEGER NOT NULL REFERENCES kelas(id),
			locked INTEGER DEFAULT 1,
			created_at DATETIME,
			updated_at DATETIME,
			UNIQUE(date, kelas_id)
		)`,
		Down: `DROP TABLE IF EXISTS kunci_absensis`,
	},
	{
		Version: 7,
		Name:    "create_absensi_sessions",
		Up: `CREATE TABLE IF NOT EXISTS absensi_sessions (
			id TEXT PRIMARY KEY,
			senin INTEGER DEFAULT 0,
			selasa INTEGER DEFAULT 0,
			rabu INTEGER DEFAULT 0,
			kamis INTEGER DEFAULT 0,
			jumat INTEGER DEFAULT 0,
			sabtu INTEGER DEFAULT 0,
			jam_masuk TEXT NOT NULL,
			jam_masuk_toleransi INTEGER DEFAULT 900,
			jam_keluar_mulai_absen TEXT,
			jam_keluar TEXT NOT NULL,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		Down: `DROP TABLE IF EXISTS absensi_sessions`,
	},
	{
		Version: 8,
		Name:    "create_absensi_session_kelas",
		Up: `CREATE TABLE IF NOT EXISTS absensi_session_kelas (
			absensi_session_id TEXT NOT NULL REFERENCES absensi_sessions(id),
			kelas_id INTEGER NOT NULL REFERENCES kelas(id),
			PRIMARY KEY (absensi_session_id, kelas_id)
		)`,
		Down: `DROP TABLE IF EXISTS absensi_session_kelas`,
	},
	{
		Version: 9,
		Name:    "create_data_sekolah",
		Up: `CREATE TABLE IF NOT EXISTS data_sekolah (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			nama_sekolah TEXT NOT NULL,
			logo_sekolah TEXT,
			deskripsi_sekolah TEXT,
			kop_sekolah TEXT,
			nama_aplikasi TEXT DEFAULT 'Presensee',
			created_at DATETIME,
			updated_at DATETIME
		)`,
		Down: `DROP TABLE IF EXISTS data_sekolah`,
	},
}
