package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrEmailExists       = errors.New("email already exists")
	ErrInvalidResetToken = errors.New("invalid or expired reset token")
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

// CreateOwner creates a Cyber Owner account.
//
// A Cyber Owner does NOT get a tenant during registration.
// The owner will create their cyber branches later.
//
// tenant_id is therefore NULL at this stage.
func (r *Repository) CreateOwner(
	ctx context.Context,
	fullName string,
	email string,
	phone string,
	passwordHash string,
) (User, error) {

	var user User

	err := r.db.QueryRow(
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
			'owner',
			'active'
		)
		RETURNING
			id,
			tenant_id,
			full_name,
			email,
			COALESCE(phone, ''),
			role::text,
			status::text,
			last_login_at,
			created_at,
			updated_at
		`,
		uuid.New(),
		fullName,
		email,
		phone,
		passwordHash,
	).Scan(
		&user.ID,
		&user.TenantID,
		&user.FullName,
		&user.Email,
		&user.Phone,
		&user.Role,
		&user.Status,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return User{}, ErrEmailExists
		}

		return User{}, err
	}

	return user, nil
}

func (r *Repository) FindUserByEmail(
	ctx context.Context,
	email string,
) (User, string, error) {

	var user User
	var passwordHash string

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			tenant_id,
			full_name,
			email,
			COALESCE(phone, ''),
			password_hash,
			role::text,
			status::text,
			last_login_at,
			created_at,
			updated_at
		FROM users
		WHERE LOWER(email) = LOWER($1)
		LIMIT 1
		`,
		email,
	).Scan(
		&user.ID,
		&user.TenantID,
		&user.FullName,
		&user.Email,
		&user.Phone,
		&passwordHash,
		&user.Role,
		&user.Status,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, "", ErrUserNotFound
	}

	if err != nil {
		return User{}, "", err
	}

	return user, passwordHash, nil
}

func (r *Repository) UpdateLastLogin(
	ctx context.Context,
	userID string,
) error {

	_, err := r.db.Exec(
		ctx,
		`
		UPDATE users
		SET
			last_login_at = NOW(),
			updated_at = NOW()
		WHERE id = $1
		`,
		userID,
	)

	return err
}

func (r *Repository) CreatePasswordResetToken(
	ctx context.Context,
	userID string,
	tokenHash string,
	expiresAt time.Time,
) error {

	_, err := r.db.Exec(
		ctx,
		`
		INSERT INTO password_reset_tokens (
			user_id,
			token_hash,
			expires_at
		)
		VALUES ($1, $2, $3)
		`,
		userID,
		tokenHash,
		expiresAt,
	)

	return err
}

func (r *Repository) FindValidResetToken(
	ctx context.Context,
	tokenHash string,
) (string, error) {

	var userID string

	err := r.db.QueryRow(
		ctx,
		`
		SELECT user_id
		FROM password_reset_tokens
		WHERE token_hash = $1
		  AND expires_at > NOW()
		  AND used_at IS NULL
		LIMIT 1
		`,
		tokenHash,
	).Scan(&userID)

	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrInvalidResetToken
	}

	if err != nil {
		return "", err
	}

	return userID, nil
}

func (r *Repository) ResetPassword(
	ctx context.Context,
	userID string,
	tokenHash string,
	passwordHash string,
) error {

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	result, err := tx.Exec(
		ctx,
		`
		UPDATE password_reset_tokens
		SET used_at = NOW()
		WHERE token_hash = $1
		  AND user_id = $2
		  AND expires_at > NOW()
		  AND used_at IS NULL
		`,
		tokenHash,
		userID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() != 1 {
		return ErrInvalidResetToken
	}

	_, err = tx.Exec(
		ctx,
		`
		UPDATE users
		SET
			password_hash = $1,
			updated_at = NOW()
		WHERE id = $2
		`,
		passwordHash,
		userID,
	)

	if err != nil {
		return err
	}

	// Invalidate any other outstanding reset links.
	_, err = tx.Exec(
		ctx,
		`
		UPDATE password_reset_tokens
		SET used_at = NOW()
		WHERE user_id = $1
		  AND used_at IS NULL
		  AND token_hash <> $2
		`,
		userID,
		tokenHash,
	)

	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
