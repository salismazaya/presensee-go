package database

import (
	"fmt"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ponytail: auto-detect driver by DSN prefix.
func Connect(dsn string) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch {
	case strings.HasPrefix(dsn, "postgres://"), strings.HasPrefix(dsn, "postgresql://"):
		dialector = postgres.Open(dsn)
	case strings.Contains(dsn, "@tcp("), strings.HasPrefix(dsn, "mysql://"):
		dialector = mysql.Open(strings.TrimPrefix(dsn, "mysql://"))
	default:
		dialector = sqlite.Open(dsn)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}
