package customer

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"cybersaas/backend/idempotency"
)

type Service struct {
	repository    *Repository
	idempotency   *idempotency.Repository
	encryptionKey []byte
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository:    repository,
		idempotency:   idempotency.NewRepository(),
		encryptionKey: loadEncryptionKey(),
	}
}

func (s *Service) Create(
	ctx context.Context,
	tenantID string,
	userID string,
	role string,
	clientOperationID string,
	request CreateCustomerRequest,
) (*Customer, error) {
	if tenantID == "" {
		return nil, errors.New("tenant context is required")
	}

	if !uuidRegex.MatchString(request.ID) {
		return nil, errors.New("invalid customer id")
	}

	if !uuidRegex.MatchString(request.BranchID) {
		return nil, errors.New("invalid branch id")
	}

	if !uuidRegex.MatchString(clientOperationID) {
		return nil, errors.New("invalid client operation id")
	}

	customerType := strings.ToLower(strings.TrimSpace(request.CustomerType))

	if customerType != "adult" && customerType != "child" {
		return nil, errors.New("customer type must be adult or child")
	}

	fullName := strings.TrimSpace(request.FullName)

	if fullName == "" {
		return nil, errors.New("full name is required")
	}

	if len(fullName) > 200 {
		return nil, errors.New("full name is too long")
	}

	if role == "attendant" {
		if err := s.repository.VerifyAttendantBranchAccess(
			ctx,
			tenantID,
			userID,
			request.BranchID,
		); err != nil {
			return nil, err
		}
	} else {
		if err := s.repository.VerifyBranchBelongsToTenant(
			ctx,
			request.BranchID,
			tenantID,
		); err != nil {
			return nil, err
		}
	}

	fmt.Println("CUSTOMER CREATE: branch access verified")

	if customerType == "child" {
		if request.ParentName == nil ||
			strings.TrimSpace(*request.ParentName) == "" {
			return nil, errors.New("parent name is required for child customers")
		}

		if request.ParentPhone == nil ||
			strings.TrimSpace(*request.ParentPhone) == "" {
			return nil, errors.New("parent phone is required for child customers")
		}
	}

	requestHash, err := idempotency.HashRequest(request)
	if err != nil {
		return nil, err
	}

	tx, err := s.repository.BeginTx(ctx)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	branchID := request.BranchID
	userIDValue := userID

	fmt.Println("CUSTOMER CREATE: starting idempotency")

	record, created, err := s.idempotency.GetOrCreate(
		ctx,
		tx,
		tenantID,
		&branchID,
		&userIDValue,
		clientOperationID,
		"customer.create",
		requestHash,
	)
	if err != nil {
		return nil, err
	}

	fmt.Println("CUSTOMER CREATE: idempotency completed")

	if !created {
		return s.customerFromIdempotencyRecord(record)
	}

	customer := Customer{
		ID:           request.ID,
		TenantID:     tenantID,
		BranchID:     request.BranchID,
		CustomerType: customerType,
		FullName:     fullName,
		Phone:        normalizeOptional(request.Phone),
		ParentName:   normalizeOptional(request.ParentName),
		ParentPhone:  normalizeOptional(request.ParentPhone),
	}

	var (
		idHash            *string
		encryptedID       *string
		parentIDHash      *string
		encryptedParentID *string
	)

	fmt.Println("CUSTOMER CREATE: checking customer ID")

	if request.IDNumber != nil &&
		strings.TrimSpace(*request.IDNumber) != "" {

		normalized := normalizeID(*request.IDNumber)
		hash := hashID(normalized)

		encrypted, err := s.encryptID(normalized)
		if err != nil {
			return nil, err
		}

		idHash = &hash
		encryptedID = &encrypted

		existing, err := s.repository.FindByIDHashTx(
			ctx,
			tx,
			tenantID,
			request.BranchID,
			hash,
		)

		if err != nil && !errors.Is(err, ErrCustomerNotFound) {
			return nil, err
		}

		if existing != nil {
			return nil, ErrCustomerAlreadyExists
		}
	}

	fmt.Println("CUSTOMER CREATE: customer ID check completed")

	if request.ParentIDNumber != nil &&
		strings.TrimSpace(*request.ParentIDNumber) != "" {

		fmt.Println("CUSTOMER CREATE: hashing/encrypting parent ID")

		normalized := normalizeID(*request.ParentIDNumber)
		hash := hashID(normalized)

		encrypted, err := s.encryptID(normalized)
		if err != nil {
			return nil, err
		}

		parentIDHash = &hash
		encryptedParentID = &encrypted
	}

	fmt.Println("CUSTOMER CREATE: inserting customer")

	if err := s.repository.CreateTx(
		ctx,
		tx,
		customer,
		idHash,
		encryptedID,
		parentIDHash,
		encryptedParentID,
	); err != nil {
		return nil, err
	}

	fmt.Println("CUSTOMER CREATE: customer inserted")

	response := customer

	fmt.Println("CUSTOMER CREATE: completing idempotency")

	if err := s.idempotency.Complete(
		ctx,
		tx,
		record.ID,
		201,
		response,
		"customer",
		customer.ID,
	); err != nil {
		return nil, err
	}

	fmt.Println("CUSTOMER CREATE: idempotency response completed")

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	fmt.Println("CUSTOMER CREATE: transaction committed")

	return &customer, nil
}

func (s *Service) Get(
	ctx context.Context,
	tenantID string,
	userID string,
	role string,
	customerID string,
	branchID string,
) (*Customer, error) {
	if tenantID == "" {
		return nil, errors.New("tenant context is required")
	}

	if branchID == "" {
		return nil, errors.New("branch id is required")
	}

	if role == "attendant" {
		if err := s.repository.VerifyAttendantBranchAccess(
			ctx,
			tenantID,
			userID,
			branchID,
		); err != nil {
			return nil, err
		}
	} else {
		if err := s.repository.VerifyBranchBelongsToTenant(
			ctx,
			branchID,
			tenantID,
		); err != nil {
			return nil, err
		}
	}

	return s.repository.Get(
		ctx,
		customerID,
		tenantID,
		branchID,
	)
}

func (s *Service) List(
	ctx context.Context,
	tenantID string,
	userID string,
	role string,
	branchID string,
	search string,
	limit int,
	offset int,
) (*CustomerListResponse, error) {
	if tenantID == "" {
		return nil, errors.New("tenant context is required")
	}

	if branchID == "" {
		return nil, errors.New("branch id is required")
	}

	if role == "attendant" {
		if err := s.repository.VerifyAttendantBranchAccess(
			ctx,
			tenantID,
			userID,
			branchID,
		); err != nil {
			return nil, err
		}
	} else {
		if err := s.repository.VerifyBranchBelongsToTenant(
			ctx,
			branchID,
			tenantID,
		); err != nil {
			return nil, err
		}
	}

	if limit <= 0 {
		limit = 50
	}

	if limit > 100 {
		limit = 100
	}

	if offset < 0 {
		offset = 0
	}

	customers, total, err := s.repository.List(
		ctx,
		tenantID,
		branchID,
		search,
		limit,
		offset,
	)

	if err != nil {
		return nil, err
	}

	return &CustomerListResponse{
		Customers: customers,
		Total:     total,
	}, nil
}

func (s *Service) Update(
	ctx context.Context,
	tenantID string,
	userID string,
	role string,
	customerID string,
	branchID string,
	clientOperationID string,
	request UpdateCustomerRequest,
) (*Customer, error) {
	if tenantID == "" {
		return nil, errors.New("tenant context is required")
	}

	if !uuidRegex.MatchString(clientOperationID) {
		return nil, errors.New("invalid client operation id")
	}

	if !uuidRegex.MatchString(customerID) {
		return nil, errors.New("invalid customer id")
	}

	if !uuidRegex.MatchString(branchID) {
		return nil, errors.New("invalid branch id")
	}

	customerType := strings.ToLower(strings.TrimSpace(request.CustomerType))

	if customerType != "adult" && customerType != "child" {
		return nil, errors.New("customer type must be adult or child")
	}

	fullName := strings.TrimSpace(request.FullName)

	if fullName == "" {
		return nil, errors.New("full name is required")
	}

	if customerType == "child" {
		if request.ParentName == nil ||
			strings.TrimSpace(*request.ParentName) == "" {
			return nil, errors.New("parent name is required for child customers")
		}

		if request.ParentPhone == nil ||
			strings.TrimSpace(*request.ParentPhone) == "" {
			return nil, errors.New("parent phone is required for child customers")
		}
	}

	if role == "attendant" {
		if err := s.repository.VerifyAttendantBranchAccess(
			ctx,
			tenantID,
			userID,
			branchID,
		); err != nil {
			return nil, err
		}
	} else {
		if err := s.repository.VerifyBranchBelongsToTenant(
			ctx,
			branchID,
			tenantID,
		); err != nil {
			return nil, err
		}
	}

	requestHash, err := idempotency.HashRequest(request)
	if err != nil {
		return nil, err
	}

	tx, err := s.repository.BeginTx(ctx)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	branchIDValue := branchID
	userIDValue := userID

	record, created, err := s.idempotency.GetOrCreate(
		ctx,
		tx,
		tenantID,
		&branchIDValue,
		&userIDValue,
		clientOperationID,
		"customer.update",
		requestHash,
	)
	if err != nil {
		return nil, err
	}

	if !created {
		return s.customerFromIdempotencyRecord(record)
	}

	customer := Customer{
		ID:           customerID,
		TenantID:     tenantID,
		BranchID:     branchID,
		CustomerType: customerType,
		FullName:     fullName,
		Phone:        normalizeOptional(request.Phone),
		ParentName:   normalizeOptional(request.ParentName),
		ParentPhone:  normalizeOptional(request.ParentPhone),
	}

	var (
		idHash            *string
		encryptedID       *string
		parentIDHash      *string
		encryptedParentID *string
	)

	if request.IDNumber != nil &&
		strings.TrimSpace(*request.IDNumber) != "" {

		normalized := normalizeID(*request.IDNumber)
		hash := hashID(normalized)

		existing, err := s.repository.FindByIDHashTx(
			ctx,
			tx,
			tenantID,
			branchID,
			hash,
		)

		if err != nil && !errors.Is(err, ErrCustomerNotFound) {
			return nil, err
		}

		if existing != nil && existing.ID != customerID {
			return nil, ErrCustomerAlreadyExists
		}

		encrypted, err := s.encryptID(normalized)
		if err != nil {
			return nil, err
		}

		idHash = &hash
		encryptedID = &encrypted
	}

	if request.ParentIDNumber != nil &&
		strings.TrimSpace(*request.ParentIDNumber) != "" {

		normalized := normalizeID(*request.ParentIDNumber)
		hash := hashID(normalized)

		encrypted, err := s.encryptID(normalized)
		if err != nil {
			return nil, err
		}

		parentIDHash = &hash
		encryptedParentID = &encrypted
	}

	if err := s.repository.UpdateTx(
		ctx,
		tx,
		customer,
		idHash,
		encryptedID,
		parentIDHash,
		encryptedParentID,
	); err != nil {
		return nil, err
	}

	updatedCustomer, err := s.repository.GetTx(
		ctx,
		tx,
		customerID,
		tenantID,
		branchID,
	)
	if err != nil {
		return nil, err
	}

	if err := s.idempotency.Complete(
		ctx,
		tx,
		record.ID,
		200,
		*updatedCustomer,
		"customer",
		updatedCustomer.ID,
	); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return updatedCustomer, nil
}

func (s *Service) customerFromIdempotencyRecord(
	record *idempotency.Record,
) (*Customer, error) {
	if record == nil {
		return nil, errors.New("idempotency record is missing")
	}

	if record.ResponseStatus == nil ||
		len(record.ResponseBody) == 0 {
		return nil, idempotency.ErrResponseNotReady
	}

	var customer Customer

	if err := json.Unmarshal(
		record.ResponseBody,
		&customer,
	); err != nil {
		return nil, err
	}

	return &customer, nil
}

func normalizeOptional(value *string) *string {
	if value == nil {
		return nil
	}

	value = func() *string {
		valueCopy := strings.TrimSpace(*value)

		if valueCopy == "" {
			return nil
		}

		return &valueCopy
	}()

	return value
}

func normalizeID(value string) string {
	return strings.ToUpper(
		strings.ReplaceAll(
			strings.TrimSpace(value),
			" ",
			"",
		),
	)
}

func hashID(value string) string {
	hash := sha256.Sum256([]byte(value))
	return fmt.Sprintf("%x", hash)
}

func loadEncryptionKey() []byte {
	value := strings.TrimSpace(
		os.Getenv("CUSTOMER_ID_ENCRYPTION_KEY"),
	)

	if value == "" {
		panic("CUSTOMER_ID_ENCRYPTION_KEY is required")
	}

	decoded, err := base64.StdEncoding.DecodeString(value)

	if err == nil && len(decoded) == 32 {
		return decoded
	}

	if len(value) == 64 {
		decoded, err := base64.RawStdEncoding.DecodeString(value)

		if err == nil && len(decoded) == 32 {
			return decoded
		}
	}

	panic(
		"CUSTOMER_ID_ENCRYPTION_KEY must contain a base64-encoded 32-byte AES-256 key",
	)
}

func (s *Service) encryptID(value string) (string, error) {
	block, err := aes.NewCipher(s.encryptionKey)

	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)

	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())

	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(
		nonce,
		nonce,
		[]byte(value),
		nil,
	)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

var uuidRegex = func() *regexp.Regexp {
	return regexp.MustCompile(
		`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`,
	)
}()
