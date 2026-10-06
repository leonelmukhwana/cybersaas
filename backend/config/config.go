package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	AppEnv        string
	Port          string
	DatabaseURL   string
	JWTSecret     string
	EncryptionKey string

	// Email / SMTP
	SMTPHost      string
	SMTPPort      int
	SMTPUsername  string
	SMTPPassword  string
	SMTPFromEmail string
	SMTPFromName  string
	FrontendURL   string
}

func Load() (*Config, error) {
	smtpPort := 587

	if value := os.Getenv("SMTP_PORT"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > 65535 {
			return nil, fmt.Errorf("SMTP_PORT must be a valid port number")
		}
		smtpPort = parsed
	}

	cfg := &Config{
		AppEnv:        os.Getenv("APP_ENV"),
		Port:          os.Getenv("PORT"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		EncryptionKey: os.Getenv("ENCRYPTION_KEY"),

		SMTPHost:      os.Getenv("SMTP_HOST"),
		SMTPPort:      smtpPort,
		SMTPUsername:  os.Getenv("SMTP_USERNAME"),
		SMTPPassword:  os.Getenv("SMTP_PASSWORD"),
		SMTPFromEmail: os.Getenv("SMTP_FROM_EMAIL"),
		SMTPFromName:  os.Getenv("SMTP_FROM_NAME"),
		FrontendURL:   os.Getenv("FRONTEND_URL"),
	}

	if cfg.AppEnv == "" {
		cfg.AppEnv = "development"
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	if len(cfg.EncryptionKey) != 32 {
		return nil, fmt.Errorf("ENCRYPTION_KEY must be exactly 32 characters")
	}

	// SMTP is required in production.
	if cfg.AppEnv == "production" {
		if cfg.SMTPHost == "" {
			return nil, fmt.Errorf("SMTP_HOST is required in production")
		}

		if cfg.SMTPUsername == "" {
			return nil, fmt.Errorf("SMTP_USERNAME is required in production")
		}

		if cfg.SMTPPassword == "" {
			return nil, fmt.Errorf("SMTP_PASSWORD is required in production")
		}

		if cfg.SMTPFromEmail == "" {
			return nil, fmt.Errorf("SMTP_FROM_EMAIL is required in production")
		}

		if cfg.SMTPFromName == "" {
			return nil, fmt.Errorf("SMTP_FROM_NAME is required in production")
		}

		if cfg.FrontendURL == "" {
			return nil, fmt.Errorf("FRONTEND_URL is required in production")
		}
	}

	return cfg, nil
}
