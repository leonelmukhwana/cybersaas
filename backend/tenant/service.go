package tenant

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrInvalidUserID   = errors.New("invalid user id")
	ErrInvalidTenantID = errors.New("invalid tenant id")
	ErrInvalidBusiness = errors.New("business name is required")
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}

// ------------------------------------------------------------
// CREATE
// ------------------------------------------------------------

func (s *Service) CreateTenant(
	ctx context.Context,
	userID string,
	request CreateTenantRequest,
) (Tenant, error) {

	if _, err := uuid.Parse(userID); err != nil {
		return Tenant{}, ErrInvalidUserID
	}

	request.BusinessName = strings.TrimSpace(request.BusinessName)

	if request.BusinessName == "" {
		return Tenant{}, ErrInvalidBusiness
	}

	if len(request.BusinessName) < 2 {
		return Tenant{}, errors.New("business name must contain at least 2 characters")
	}

	if len(request.BusinessName) > 200 {
		return Tenant{}, errors.New("business name is too long")
	}

	return s.repository.CreateTenant(
		ctx,
		userID,
		request.BusinessName,
	)
}

// ------------------------------------------------------------
// GET CURRENT OWNER TENANT
// ------------------------------------------------------------

func (s *Service) GetMyTenant(
	ctx context.Context,
	userID string,
) (Tenant, error) {

	if _, err := uuid.Parse(userID); err != nil {
		return Tenant{}, ErrInvalidUserID
	}

	return s.repository.GetTenantByOwnerID(
		ctx,
		userID,
	)
}

// ------------------------------------------------------------
// GET BY ID
// ------------------------------------------------------------

func (s *Service) GetTenant(
	ctx context.Context,
	userID string,
	tenantID string,
) (Tenant, error) {

	if _, err := uuid.Parse(userID); err != nil {
		return Tenant{}, ErrInvalidUserID
	}

	if _, err := uuid.Parse(tenantID); err != nil {
		return Tenant{}, ErrInvalidTenantID
	}

	tenant, err := s.repository.GetTenantByID(
		ctx,
		tenantID,
	)

	if err != nil {
		return Tenant{}, err
	}

	// Verify that the authenticated owner actually owns
	// this tenant.
	myTenant, err := s.repository.GetTenantByOwnerID(
		ctx,
		userID,
	)

	if err != nil {
		return Tenant{}, err
	}

	if myTenant.ID != tenant.ID {
		return Tenant{}, errors.New("tenant access denied")
	}

	return tenant, nil
}

// ------------------------------------------------------------
// UPDATE
// ------------------------------------------------------------

func (s *Service) UpdateTenant(
	ctx context.Context,
	userID string,
	request UpdateTenantRequest,
) (Tenant, error) {

	if _, err := uuid.Parse(userID); err != nil {
		return Tenant{}, ErrInvalidUserID
	}

	request.BusinessName = strings.TrimSpace(request.BusinessName)

	if request.BusinessName == "" {
		return Tenant{}, ErrInvalidBusiness
	}

	if len(request.BusinessName) < 2 {
		return Tenant{}, errors.New("business name must contain at least 2 characters")
	}

	if len(request.BusinessName) > 200 {
		return Tenant{}, errors.New("business name is too long")
	}

	return s.repository.UpdateTenant(
		ctx,
		userID,
		request.BusinessName,
	)
}
