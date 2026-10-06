package service

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"strings"

	"github.com/google/uuid"
)

type ServiceManager struct {
	repository *Repository
}

func NewService(repository *Repository) *ServiceManager {
	return &ServiceManager{
		repository: repository,
	}
}

func (s *ServiceManager) Create(
	ctx context.Context,
	tenantID string,
	userID string,
	userRole string,
	request CreateServiceRequest,
) (*Service, error) {
	branchID := strings.TrimSpace(request.BranchID)

	if branchID == "" {
		return nil, errors.New("branch id is required")
	}

	name := normalizeName(request.Name)

	if name == "" {
		return nil, errors.New("service name is required")
	}

	if len(name) > 200 {
		return nil, errors.New("service name must not exceed 200 characters")
	}

	description := normalizeOptional(request.Description)

	if description != nil && len(*description) > 5000 {
		return nil, errors.New("description must not exceed 5000 characters")
	}

	price, err := normalizePrice(request.Price)
	if err != nil {
		return nil, err
	}

	if err := s.repository.VerifyBranchBelongsToTenant(
		ctx,
		branchID,
		tenantID,
	); err != nil {
		return nil, err
	}

	serviceID := uuid.NewString()

	service, err := s.repository.Create(
		ctx,
		serviceID,
		tenantID,
		branchID,
		name,
		description,
		price,
	)
	if err != nil {
		return nil, err
	}

	newValue, _ := json.Marshal(map[string]interface{}{
		"id":          service.ID,
		"branch_id":   service.BranchID,
		"name":        service.Name,
		"description": service.Description,
		"price":       service.Price,
		"active":      service.Active,
	})

	_ = s.repository.CreateAuditLog(
		ctx,
		tenantID,
		branchID,
		userID,
		userRole,
		"service.created",
		"service",
		service.ID,
		nil,
		newValue,
		nil,
	)

	return service, nil
}

func (s *ServiceManager) Get(
	ctx context.Context,
	tenantID string,
	userID string,
	userRole string,
	serviceID string,
	branchID string,
) (*Service, error) {
	serviceID = strings.TrimSpace(serviceID)
	branchID = strings.TrimSpace(branchID)

	if serviceID == "" {
		return nil, errors.New("service id is required")
	}

	if branchID == "" {
		return nil, errors.New("branch id is required")
	}

	if userRole == "attendant" {
		if err := s.repository.VerifyAttendantBranchAccess(
			ctx,
			userID,
			tenantID,
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
		serviceID,
		tenantID,
		branchID,
	)
}

func (s *ServiceManager) List(
	ctx context.Context,
	tenantID string,
	userID string,
	userRole string,
	branchID string,
	activeOnly bool,
	limit int,
	offset int,
) (*ServiceListResponse, error) {
	branchID = strings.TrimSpace(branchID)

	if branchID == "" {
		return nil, errors.New("branch id is required")
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

	if userRole == "attendant" {
		if err := s.repository.VerifyAttendantBranchAccess(
			ctx,
			userID,
			tenantID,
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

	services, err := s.repository.List(
		ctx,
		tenantID,
		branchID,
		activeOnly,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}

	total, err := s.repository.Count(
		ctx,
		tenantID,
		branchID,
		activeOnly,
	)
	if err != nil {
		return nil, err
	}

	return &ServiceListResponse{
		Services: services,
		Total:    total,
	}, nil
}

func (s *ServiceManager) Update(
	ctx context.Context,
	tenantID string,
	userID string,
	userRole string,
	serviceID string,
	branchID string,
	request UpdateServiceRequest,
) (*Service, error) {
	if userRole != "owner" {
		return nil, errors.New("only the owner can update services")
	}

	branchID = strings.TrimSpace(branchID)
	serviceID = strings.TrimSpace(serviceID)

	if branchID == "" {
		return nil, errors.New("branch id is required")
	}

	if serviceID == "" {
		return nil, errors.New("service id is required")
	}

	name := normalizeName(request.Name)

	if name == "" {
		return nil, errors.New("service name is required")
	}

	if len(name) > 200 {
		return nil, errors.New("service name must not exceed 200 characters")
	}

	description := normalizeOptional(request.Description)

	if description != nil && len(*description) > 5000 {
		return nil, errors.New("description must not exceed 5000 characters")
	}

	price, err := normalizePrice(request.Price)
	if err != nil {
		return nil, err
	}

	if err := s.repository.VerifyBranchBelongsToTenant(
		ctx,
		branchID,
		tenantID,
	); err != nil {
		return nil, err
	}

	oldService, err := s.repository.Get(
		ctx,
		serviceID,
		tenantID,
		branchID,
	)
	if err != nil {
		return nil, err
	}

	service, err := s.repository.Update(
		ctx,
		serviceID,
		tenantID,
		branchID,
		name,
		description,
		price,
	)
	if err != nil {
		return nil, err
	}

	oldValue, _ := json.Marshal(oldService)
	newValue, _ := json.Marshal(service)

	_ = s.repository.CreateAuditLog(
		ctx,
		tenantID,
		branchID,
		userID,
		userRole,
		"service.updated",
		"service",
		service.ID,
		oldValue,
		newValue,
		nil,
	)

	return service, nil
}

func (s *ServiceManager) ChangeStatus(
	ctx context.Context,
	tenantID string,
	userID string,
	userRole string,
	serviceID string,
	branchID string,
	active bool,
) (*Service, error) {
	if userRole != "owner" {
		return nil, errors.New("only the owner can change service status")
	}

	branchID = strings.TrimSpace(branchID)
	serviceID = strings.TrimSpace(serviceID)

	if branchID == "" {
		return nil, errors.New("branch id is required")
	}

	if serviceID == "" {
		return nil, errors.New("service id is required")
	}

	if err := s.repository.VerifyBranchBelongsToTenant(
		ctx,
		branchID,
		tenantID,
	); err != nil {
		return nil, err
	}

	oldService, err := s.repository.Get(
		ctx,
		serviceID,
		tenantID,
		branchID,
	)
	if err != nil {
		return nil, err
	}

	service, err := s.repository.UpdateStatus(
		ctx,
		serviceID,
		tenantID,
		branchID,
		active,
	)
	if err != nil {
		return nil, err
	}

	oldValue, _ := json.Marshal(oldService)
	newValue, _ := json.Marshal(service)

	_ = s.repository.CreateAuditLog(
		ctx,
		tenantID,
		branchID,
		userID,
		userRole,
		"service.status_changed",
		"service",
		service.ID,
		oldValue,
		newValue,
		nil,
	)

	return service, nil
}

func normalizeName(value string) string {
	return strings.Join(
		strings.Fields(strings.TrimSpace(value)),
		" ",
	)
}

func normalizeOptional(value *string) *string {
	if value == nil {
		return nil
	}

	cleaned := strings.TrimSpace(*value)

	if cleaned == "" {
		return nil
	}

	return &cleaned
}

func normalizePrice(value string) (string, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return "", errors.New("price is required")
	}

	if strings.HasPrefix(value, "-") || strings.Contains(value, "e") || strings.Contains(value, "E") {
		return "", errors.New("price must be a valid non-negative decimal number")
	}

	rat, ok := new(big.Rat).SetString(value)
	if !ok {
		return "", errors.New("price must be a valid number")
	}

	if rat.Sign() < 0 {
		return "", errors.New("price cannot be negative")
	}

	decimalParts := strings.Split(value, ".")

	if len(decimalParts) == 2 && len(decimalParts[1]) > 2 {
		return "", errors.New("price can have at most 2 decimal places")
	}

	return value, nil
}
