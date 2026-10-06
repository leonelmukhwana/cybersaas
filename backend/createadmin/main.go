package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"cybersaas/backend/auth"
	"cybersaas/backend/config"
	"cybersaas/backend/database"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func main() {
	// Load the backend .env file.
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not loaded; using environment variables")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fullName := strings.TrimSpace(os.Getenv("ADMIN_FULL_NAME"))
	email := auth.NormalizeEmail(os.Getenv("ADMIN_EMAIL"))
	phone := strings.TrimSpace(os.Getenv("ADMIN_PHONE"))
	password := os.Getenv("ADMIN_PASSWORD")

	if fullName == "" {
		log.Fatal("ADMIN_FULL_NAME is required")
	}

	if email == "" {
		log.Fatal("ADMIN_EMAIL is required")
	}

	if phone == "" {
		log.Fatal("ADMIN_PHONE is required")
	}

	if len(password) < 8 {
		log.Fatal("ADMIN_PASSWORD must be at least 8 characters")
	}

	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		log.Fatalf("failed to hash password: %v", err)
	}

	ctx := context.Background()

	var userID uuid.UUID

	err = db.QueryRow(
		ctx,
		`
		INSERT INTO users (
			id,
			tenant_id,
			full_name,
			email,
			phone,
			password_hash,
			role,
			status
		)
		VALUES (
			$1,
			NULL,
			$2,
			$3,
			$4,
			$5,
			'platform_admin',
			'active'
		)
		RETURNING id
		`,
		uuid.New(),
		fullName,
		email,
		phone,
		passwordHash,
	).Scan(&userID)

	if err != nil {
		log.Fatalf("failed to create platform admin: %v", err)
	}

	fmt.Println("Platform Admin created successfully.")
	fmt.Println("ID:", userID)
	fmt.Println("Email:", email)
	fmt.Println("Role: platform_admin")
}
