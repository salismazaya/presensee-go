package migration

import (
	"fmt"
	"sort"
	"time"

	"gorm.io/gorm"
)

type Migration struct {
	Version int
	Name    string
	Up      string
	Down    string
}

type MigrationStatus struct {
	Version   int
	Name      string
	Applied   bool
	AppliedAt *time.Time
}

type schemaMigration struct {
	Version   int       `gorm:"primaryKey"`
	Name      string    `gorm:"not null"`
	AppliedAt time.Time `gorm:"not null"`
}

func ensureTable(db *gorm.DB) error {
	return db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version   INTEGER PRIMARY KEY,
		name      TEXT NOT NULL,
		applied_at DATETIME NOT NULL
	)`).Error
}

// MigrateUp applies all pending migrations in version order.
func MigrateUp(db *gorm.DB) error {
	if err := ensureTable(db); err != nil {
		return fmt.Errorf("migration: ensure table: %w", err)
	}

	applied, err := appliedVersions(db)
	if err != nil {
		return err
	}

	sorted := sortedMigrations()
	for _, m := range sorted {
		if applied[m.Version] {
			continue
		}
		if err := runUp(db, m); err != nil {
			return fmt.Errorf("migration %03d_%s up: %w", m.Version, m.Name, err)
		}
		fmt.Printf("  ✓ %03d_%s\n", m.Version, m.Name)
	}
	return nil
}

// MigrateDown rolls back the last `steps` applied migrations.
func MigrateDown(db *gorm.DB, steps int) error {
	if err := ensureTable(db); err != nil {
		return fmt.Errorf("migration: ensure table: %w", err)
	}

	var rows []schemaMigration
	if err := db.Raw("SELECT version, name, applied_at FROM schema_migrations ORDER BY version DESC LIMIT ?", steps).Scan(&rows).Error; err != nil {
		return fmt.Errorf("migration: list applied: %w", err)
	}

	byVersion := migrationMap()
	for _, row := range rows {
		m, ok := byVersion[row.Version]
		if !ok {
			return fmt.Errorf("migration %03d not found in code", row.Version)
		}
		if err := runDown(db, m); err != nil {
			return fmt.Errorf("migration %03d_%s down: %w", m.Version, m.Name, err)
		}
		fmt.Printf("  ↓ %03d_%s\n", m.Version, m.Name)
	}
	return nil
}

// MigrateStatus returns status of every registered migration.
func MigrateStatus(db *gorm.DB) ([]MigrationStatus, error) {
	if err := ensureTable(db); err != nil {
		return nil, err
	}

	var rows []schemaMigration
	if err := db.Raw("SELECT version, name, applied_at FROM schema_migrations ORDER BY version").Scan(&rows).Error; err != nil {
		return nil, err
	}

	appliedMap := make(map[int]time.Time, len(rows))
	for _, r := range rows {
		appliedMap[r.Version] = r.AppliedAt
	}

	sorted := sortedMigrations()
	out := make([]MigrationStatus, len(sorted))
	for i, m := range sorted {
		ms := MigrationStatus{Version: m.Version, Name: m.Name}
		if t, ok := appliedMap[m.Version]; ok {
			ms.Applied = true
			ms.AppliedAt = &t
		}
		out[i] = ms
	}
	return out, nil
}

func runUp(db *gorm.DB, m Migration) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(m.Up).Error; err != nil {
			return err
		}
		return tx.Exec("INSERT INTO schema_migrations (version, name, applied_at) VALUES (?, ?, ?)",
			m.Version, m.Name, time.Now()).Error
	})
}

func runDown(db *gorm.DB, m Migration) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(m.Down).Error; err != nil {
			return err
		}
		return tx.Exec("DELETE FROM schema_migrations WHERE version = ?", m.Version).Error
	})
}

func appliedVersions(db *gorm.DB) (map[int]bool, error) {
	var versions []int
	if err := db.Raw("SELECT version FROM schema_migrations").Scan(&versions).Error; err != nil {
		return nil, fmt.Errorf("migration: read applied: %w", err)
	}
	m := make(map[int]bool, len(versions))
	for _, v := range versions {
		m[v] = true
	}
	return m, nil
}

func sortedMigrations() []Migration {
	out := make([]Migration, len(Migrations))
	copy(out, Migrations)
	sort.Slice(out, func(i, j int) bool { return out[i].Version < out[j].Version })
	return out
}

func migrationMap() map[int]Migration {
	m := make(map[int]Migration, len(Migrations))
	for _, mg := range Migrations {
		m[mg.Version] = mg
	}
	return m
}
