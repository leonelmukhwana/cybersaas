package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type TokenGenerator interface {
	Generate(userID string, tenantID string, role string) (string, error)
}

type EmailSender interface {
	Send(to string, subject string, body string) error
}

type Service struct {
	repo        *Repository
	tokenMaker  TokenGenerator
	resetURL    string
	emailSender EmailSender
}

func NewService(
	repo *Repository,
	tokenMaker TokenGenerator,
	resetURL string,
) *Service {
	return &Service{
		repo:       repo,
		tokenMaker: tokenMaker,
		resetURL:   strings.TrimRight(resetURL, "/"),
	}
}

func (s *Service) SetEmailSender(sender EmailSender) {
	s.emailSender = sender
}

func (s *Service) Register(
	ctx context.Context,
	req RegisterRequest,
) (AuthResponse, error) {

	fullName := strings.TrimSpace(req.FullName)
	email := NormalizeEmail(req.Email)
	phone := strings.TrimSpace(req.Phone)

	if fullName == "" {
		return AuthResponse{}, errors.New("full name is required")
	}

	if email == "" {
		return AuthResponse{}, errors.New("email is required")
	}

	if phone == "" {
		return AuthResponse{}, errors.New("phone is required")
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		return AuthResponse{}, err
	}

	user, err := s.repo.CreateOwner(
		ctx,
		fullName,
		email,
		phone,
		hash,
	)

	if err != nil {
		return AuthResponse{}, err
	}

	// A newly registered Cyber Owner does not have a tenant yet.
	tenantID := ""

	token, err := s.tokenMaker.Generate(
		user.ID,
		tenantID,
		user.Role,
	)

	if err != nil {
		return AuthResponse{}, err
	}

	return AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *Service) Login(
	ctx context.Context,
	req LoginRequest,
) (AuthResponse, error) {

	email := NormalizeEmail(req.Email)

	user, passwordHash, err := s.repo.FindUserByEmail(
		ctx,
		email,
	)

	if err != nil {
		return AuthResponse{}, errors.New("invalid email or password")
	}

	if user.Status != "active" {
		return AuthResponse{}, errors.New("account is not active")
	}

	valid, err := VerifyPassword(
		req.Password,
		passwordHash,
	)

	if err != nil || !valid {
		return AuthResponse{}, errors.New("invalid email or password")
	}

	if err := s.repo.UpdateLastLogin(ctx, user.ID); err != nil {
		return AuthResponse{}, err
	}

	tenantID := ""

	if user.TenantID != nil {
		tenantID = *user.TenantID
	}

	token, err := s.tokenMaker.Generate(
		user.ID,
		tenantID,
		user.Role,
	)

	if err != nil {
		return AuthResponse{}, err
	}

	return AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

func generateResetToken() (string, error) {
	bytes := make([]byte, 32)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

func hashResetToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (s *Service) ForgotPassword(
	ctx context.Context,
	email string,
) (string, error) {

	email = NormalizeEmail(email)

	user, _, err := s.repo.FindUserByEmail(ctx, email)

	// Never reveal whether the email exists.
	if err != nil {
		return "", nil
	}

	rawToken, err := generateResetToken()
	if err != nil {
		return "", err
	}

	tokenHash := hashResetToken(rawToken)

	// Reset links are valid for 12 hours.
	expiresAt := time.Now().UTC().Add(12 * time.Hour)

	if err := s.repo.CreatePasswordResetToken(
		ctx,
		user.ID,
		tokenHash,
		expiresAt,
	); err != nil {
		return "", err
	}

	resetLink := fmt.Sprintf(
		"%s/reset-password?token=%s",
		s.resetURL,
		rawToken,
	)

	if s.emailSender == nil {
		return "", errors.New("email service is not configured")
	}

	body := fmt.Sprintf(
		"Hello %s,\n\n"+
			"We received a request to reset your CyberSaaS password.\n\n"+
			"Reset your password using the link below:\n\n"+
			"%s\n\n"+
			"This link expires in 12 hours and can only be used once.\n\n"+
			"If you did not request a password reset, you can safely ignore this email.\n\n"+
			"Regards,\n"+
			"CyberSaaS",
		user.FullName,
		resetLink,
	)

	if err := s.emailSender.Send(
		user.Email,
		"CyberSaaS Password Reset",
		body,
	); err != nil {
		return "", fmt.Errorf(
			"failed to send password reset email: %w",
			err,
		)
	}

	// Do not return the raw reset link/token through the API.
	return "", nil
}

func (s *Service) ResetPassword(
	ctx context.Context,
	req ResetPasswordRequest,
) error {

	token := strings.TrimSpace(req.Token)

	if token == "" {
		return ErrInvalidResetToken
	}

	tokenHash := hashResetToken(token)

	userID, err := s.repo.FindValidResetToken(
		ctx,
		tokenHash,
	)

	if err != nil {
		return ErrInvalidResetToken
	}

	passwordHash, err := HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	return s.repo.ResetPassword(
		ctx,
		userID,
		tokenHash,
		passwordHash,
	)
}

func IsUUID(value string) bool {
	_, err := uuid.Parse(value)
	return err == nil
}
