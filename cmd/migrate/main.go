package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"presensee/internal/database"
	"presensee/internal/migration"
	"presensee/internal/model"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "presensee.db"
	}

	db, err := database.Connect(dsn)
	if err != nil {
		log.Fatalf("DB connect: %v", err)
	}

	cmd := "up"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	switch cmd {
	case "up":
		fmt.Println("Running migrations up...")
		if err := migration.MigrateUp(db); err != nil {
			log.Fatalf("migrate up: %v", err)
		}
		fmt.Println("Done.")

	case "down":
		steps := 1
		if len(os.Args) > 2 {
			steps, err = strconv.Atoi(os.Args[2])
			if err != nil {
				log.Fatalf("invalid steps: %v", err)
			}
		}
		fmt.Printf("Rolling back %d migration(s)...\n", steps)
		if err := migration.MigrateDown(db, steps); err != nil {
			log.Fatalf("migrate down: %v", err)
		}
		fmt.Println("Done.")

	case "status":
		statuses, err := migration.MigrateStatus(db)
		if err != nil {
			log.Fatalf("migrate status: %v", err)
		}
		fmt.Printf("%-8s %-35s %-10s %s\n", "VERSION", "NAME", "APPLIED", "AT")
		for _, s := range statuses {
			at := ""
			if s.AppliedAt != nil {
				at = s.AppliedAt.Format("2006-01-02 15:04:05")
			}
			applied := "no"
			if s.Applied {
				applied = "yes"
			}
			fmt.Printf("%-8d %-35s %-10s %s\n", s.Version, s.Name, applied, at)
		}

	case "createsuperuser", "user":
		if len(os.Args) < 4 {
			log.Fatalf("Usage: migrate createsuperuser <username> <password>")
		}
		username := strings.ToLower(strings.TrimSpace(os.Args[2]))
		password := os.Args[3]

		var user model.User
		errUser := db.Where("LOWER(TRIM(username)) = ?", username).First(&user).Error
		if errUser != nil {
			user = model.User{
				Username:    username,
				FullName:    "Super Administrator",
				IsSuperuser: true,
				IsStaff:     true,
				IsActive:    true,
			}
		} else {
			user.IsSuperuser = true
			user.IsStaff = true
			user.IsActive = true
		}
		if err := user.SetPassword(password); err != nil {
			log.Fatalf("set password: %v", err)
		}
		if errUser != nil {
			if err := db.Create(&user).Error; err != nil {
				log.Fatalf("create user: %v", err)
			}
			fmt.Printf("✓ Superuser '%s' created successfully.\n", username)
		} else {
			if err := db.Save(&user).Error; err != nil {
				log.Fatalf("update user: %v", err)
			}
			fmt.Printf("✓ Superuser '%s' password updated successfully.\n", username)
		}

	default:
		fmt.Fprintf(os.Stderr, "Usage: migrate [up|down [N]|status|createsuperuser <user> <pass>]\n")
		os.Exit(1)
	}
}
