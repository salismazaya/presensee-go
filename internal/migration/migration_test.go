package migration_test

import (
	"os"
	"testing"

	"presensee/internal/database"
	"presensee/internal/migration"
)

func TestMigrationUpDownStatus(t *testing.T) {
	dbPath := "test_migration.db"
	_ = os.Remove(dbPath)
	defer os.Remove(dbPath)

	db, err := database.Connect(dbPath)
	if err != nil {
		t.Fatalf("database connect: %v", err)
	}

	// 1. Initial Status -> none applied
	statuses, err := migration.MigrateStatus(db)
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if len(statuses) != len(migration.Migrations) {
		t.Fatalf("expected %d migrations, got %d", len(migration.Migrations), len(statuses))
	}
	for _, s := range statuses {
		if s.Applied {
			t.Fatalf("expected migration %d not applied yet", s.Version)
		}
	}

	// 2. Migrate Up
	if err := migration.MigrateUp(db); err != nil {
		t.Fatalf("migrate up failed: %v", err)
	}

	// Verify all applied
	statuses, err = migration.MigrateStatus(db)
	if err != nil {
		t.Fatalf("status after up failed: %v", err)
	}
	for _, s := range statuses {
		if !s.Applied {
			t.Fatalf("expected migration %d to be applied", s.Version)
		}
	}

	// 3. Migrate Down 3 steps
	if err := migration.MigrateDown(db, 3); err != nil {
		t.Fatalf("migrate down failed: %v", err)
	}

	statuses, err = migration.MigrateStatus(db)
	if err != nil {
		t.Fatalf("status after down failed: %v", err)
	}
	// Last 3 should be unapplied
	for _, s := range statuses {
		if s.Version > len(migration.Migrations)-3 {
			if s.Applied {
				t.Fatalf("expected migration %d to be unapplied", s.Version)
			}
		} else {
			if !s.Applied {
				t.Fatalf("expected migration %d to still be applied", s.Version)
			}
		}
	}

	// 4. Migrate Up again -> idempotency
	if err := migration.MigrateUp(db); err != nil {
		t.Fatalf("second migrate up failed: %v", err)
	}
}
