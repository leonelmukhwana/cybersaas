package idempotency

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrKeyConflict = errors.New(
		"idempotency key was already used with a different request",
	)
	ErrResponseNotReady = errors.New(
		"idempotency response is not available",
	)
)

type Record struct {
	ID             string
	TenantID       *string
	BranchID       *string
	UserID         *string
	Key            string
	Operation      string
	RequestHash    *string
	ResponseStatus *int
	ResponseBody   []byte
	ResourceType   *string
	ResourceID     *string
}

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

func HashRequest(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)

	return hex.EncodeToString(hash[:]), nil
}

// Get returns an existing idempotency record.
func (r *Repository) Get(
	ctx context.Context,
	tx pgx.Tx,
	tenantID string,
	key string,
) (*Record, error) {
	var record Record

	err := tx.QueryRow(
		ctx,
		`
		SELECT
			id,
			tenant_id,
			branch_id,
			user_id,
			key,
			operation,
			request_hash,
			response_status,
			response_body,
			resource_type,
			resource_id
		FROM idempotency_keys
		WHERE tenant_id = $1
		  AND key = $2
		`,
		tenantID,
		key,
	).Scan(
		&record.ID,
		&record.TenantID,
		&record.BranchID,
		&record.UserID,
		&record.Key,
		&record.Operation,
		&record.RequestHash,
		&record.ResponseStatus,
		&record.ResponseBody,
		&record.ResourceType,
		&record.ResourceID,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &record, nil
}

// GetOrCreate atomically claims an idempotency key.
//
// created=true means this request owns the key and may perform
// the business operation.
//
// created=false means another request already owns the key.
func (r *Repository) GetOrCreate(
	ctx context.Context,
	tx pgx.Tx,
	tenantID string,
	branchID *string,
	userID *string,
	key string,
	operation string,
	requestHash string,
) (*Record, bool, error) {
	id := uuid.NewString()

	var insertedID string

	err := tx.QueryRow(
		ctx,
		`
		INSERT INTO idempotency_keys (
			id,
			tenant_id,
			branch_id,
			user_id,
			key,
			operation,
			request_hash
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7
		)
		ON CONFLICT (tenant_id, key)
		DO NOTHING
		RETURNING id
		`,
		id,
		tenantID,
		branchID,
		userID,
		key,
		operation,
		requestHash,
	).Scan(&insertedID)

	if err == nil {
		return &Record{
			ID:          insertedID,
			Key:         key,
			Operation:   operation,
			RequestHash: &requestHash,
		}, true, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, false, err
	}

	record, err := r.Get(
		ctx,
		tx,
		tenantID,
		key,
	)

	if err != nil {
		return nil, false, err
	}

	if record == nil {
		return nil, false, errors.New(
			"idempotency record could not be retrieved",
		)
	}

	if record.Operation != operation ||
		record.RequestHash == nil ||
		*record.RequestHash != requestHash {
		return nil, false, ErrKeyConflict
	}

	return record, false, nil
}

func (r *Repository) Create(
	ctx context.Context,
	tx pgx.Tx,
	tenantID string,
	branchID *string,
	userID *string,
	key string,
	operation string,
	requestHash string,
	id string,
) (*Record, error) {
	record := &Record{
		ID:          id,
		Key:         key,
		Operation:   operation,
		RequestHash: &requestHash,
	}

	_, err := tx.Exec(
		ctx,
		`
		INSERT INTO idempotency_keys (
			id,
			tenant_id,
			branch_id,
			user_id,
			key,
			operation,
			request_hash
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7
		)
		`,
		id,
		tenantID,
		branchID,
		userID,
		key,
		operation,
		requestHash,
	)

	if err != nil {
		return nil, err
	}

	return record, nil
}

func (r *Repository) Complete(
	ctx context.Context,
	tx pgx.Tx,
	id string,
	status int,
	response any,
	resourceType string,
	resourceID string,
) error {
	responseBody, err := json.Marshal(response)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		ctx,
		`
		UPDATE idempotency_keys
		SET
			response_status = $1,
			response_body = $2::jsonb,
			resource_type = $3,
			resource_id = $4
		WHERE id = $5
		`,
		status,
		string(responseBody),
		resourceType,
		resourceID,
		id,
	)

	return err
}
