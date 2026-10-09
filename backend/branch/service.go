package branch

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"cybersaas/backend/subscription"

	"github.com/google/uuid"
)

type Service struct {
	repo            *Repository
	subscriptionSvc *subscription.Service
}

func NewService(
	repo *Repository,
	subscriptionSvc *subscription.Service,
) *Service {
	return &Service{
		repo:            repo,
		subscriptionSvc: subscriptionSvc,
	}
}

// -----------------------------------------------------------------------------
// CREATE BRANCH
// -----------------------------------------------------------------------------

func (s *Service) CreateBranch(
	ctx context.Context,
	userID string,
	businessName string,
	name string,
	address *string,
	phone *string,
	ipAddress *string,
) (Branch, error) {
	businessName = strings.TrimSpace(businessName)
	name = strings.TrimSpace(name)

	if name == "" {
		return Branch{}, errors.New("branch name is required")
	}

	tenantID, ownerName, ownerPhone, ownerEmail, err :=
		s.repo.GetUserOwnerDetails(ctx, userID)

	if err != nil {
		return Branch{}, err
	}

	// -------------------------------------------------------------------------
	// FIRST BRANCH / FIRST BUSINESS ONBOARDING
	// -------------------------------------------------------------------------

	if tenantID == nil || *tenantID == "" {
		if businessName == "" {
			return Branch{}, errors.New(
				"business_name is required when creating the first branch",
			)
		}

		if ownerPhone == nil || strings.TrimSpace(*ownerPhone) == "" {
			return Branch{}, errors.New(
				"owner phone is required before creating the business",
			)
		}

		newTenantID := uuid.New().String()

		email := ""
		if ownerEmail != nil {
			email = strings.TrimSpace(*ownerEmail)
		}

		if err := s.repo.CreateTenant(
			ctx,
			newTenantID,
			businessName,
			ownerName,
			strings.TrimSpace(*ownerPhone),
			email,
		); err != nil {
			return Branch{}, fmt.Errorf(
				"create tenant: %w",
				err,
			)
		}

		if err := s.repo.UpdateUserTenant(
			ctx,
			userID,
			newTenantID,
		); err != nil {
			return Branch{}, fmt.Errorf(
				"attach owner to tenant: %w",
				err,
			)
		}

		tenantID = &newTenantID
	}

	// -------------------------------------------------------------------------
	// SUBSCRIPTION
	// -------------------------------------------------------------------------
	//
	// The owner can complete the initial business onboarding and then
	// subscribe. Once a subscription exists, branch creation is controlled
	// by the subscription service.

	if s.subscriptionSvc != nil && tenantID != nil && *tenantID != "" {
		if err := s.subscriptionSvc.CanCreateBranch(
			ctx,
			*tenantID,
		); err != nil {
			return Branch{}, err
		}
	}

	// -------------------------------------------------------------------------
	// CREATE
	// -------------------------------------------------------------------------

	b, err := s.repo.CreateBranch(
		ctx,
		*tenantID,
		name,
		address,
		phone,
	)

	if err != nil {
		return Branch{}, err
	}

	// -------------------------------------------------------------------------
	// AUDIT
	// -------------------------------------------------------------------------

	tenant := *tenantID
	branchID := b.ID

	_ = s.repo.CreateAuditLog(
		ctx,
		&tenant,
		&branchID,
		userID,
		"owner",
		"branch.created",
		"branch",
		b.ID,
		nil,
		fmt.Sprintf(
			`{"name":%q,"status":%q}`,
			b.Name,
			b.Status,
		),
		"Branch created",
		ipAddress,
	)

	return b, nil
}

// -----------------------------------------------------------------------------
// LIST
// -----------------------------------------------------------------------------

func (s *Service) ListBranches(
	ctx context.Context,
	tenantID string,
) ([]Branch, error) {
	if strings.TrimSpace(tenantID) == "" {
		return nil, errors.New("tenant context is required")
	}

	return s.repo.ListBranches(ctx, tenantID)
}

// -----------------------------------------------------------------------------
// GET
// -----------------------------------------------------------------------------

func (s *Service) GetBranch(
	ctx context.Context,
	tenantID string,
	branchID string,
) (Branch, error) {
	if _, err := uuid.Parse(branchID); err != nil {
		return Branch{}, errors.New("invalid branch id")
	}

	return s.repo.GetBranch(ctx, tenantID, branchID)
}

// -----------------------------------------------------------------------------
// UPDATE
// -----------------------------------------------------------------------------

func (s *Service) UpdateBranch(
	ctx context.Context,
	userID string,
	tenantID string,
	branchID string,
	name string,
	address *string,
	phone *string,
	ipAddress *string,
) (Branch, error) {
	name = strings.TrimSpace(name)

	if name == "" {
		return Branch{}, errors.New("branch name is required")
	}

	oldBranch, err := s.repo.GetBranch(
		ctx,
		tenantID,
		branchID,
	)

	if err != nil {
		return Branch{}, err
	}

	b, err := s.repo.UpdateBranch(
		ctx,
		tenantID,
		branchID,
		name,
		address,
		phone,
	)

	if err != nil {
		return Branch{}, err
	}

	tenant := tenantID
	branch := branchID

	_ = s.repo.CreateAuditLog(
		ctx,
		&tenant,
		&branch,
		userID,
		"owner",
		"branch.updated",
		"branch",
		branchID,
		fmt.Sprintf(
			`{"name":%q,"status":%q}`,
			oldBranch.Name,
			oldBranch.Status,
		),
		fmt.Sprintf(
			`{"name":%q,"status":%q}`,
			b.Name,
			b.Status,
		),
		"Branch details updated",
		ipAddress,
	)

	return b, nil
}

// -----------------------------------------------------------------------------
// STATUS
// -----------------------------------------------------------------------------

func (s *Service) ChangeStatus(
	ctx context.Context,
	userID string,
	tenantID string,
	branchID string,
	status string,
	ipAddress *string,
) (Branch, error) {
	status = strings.ToLower(strings.TrimSpace(status))

	if status != "active" && status != "inactive" {
		return Branch{}, errors.New(
			"status must be active or inactive",
		)
	}

	oldBranch, err := s.repo.GetBranch(
		ctx,
		tenantID,
		branchID,
	)

	if err != nil {
		return Branch{}, err
	}

	if oldBranch.Status == status {
		return oldBranch, nil
	}

	b, err := s.repo.UpdateBranchStatus(
		ctx,
		tenantID,
		branchID,
		status,
	)

	if err != nil {
		return Branch{}, err
	}

	tenant := tenantID
	branch := branchID

	_ = s.repo.CreateAuditLog(
		ctx,
		&tenant,
		&branch,
		userID,
		"owner",
		"branch.status_changed",
		"branch",
		branchID,
		fmt.Sprintf(`{"status":%q}`, oldBranch.Status),
		fmt.Sprintf(`{"status":%q}`, b.Status),
		"Branch status changed",
		ipAddress,
	)

	return b, nil
}

var ErrInvalidBillingConfig = errors.New("invalid billing configuration")

func (s *Service) GetBillingConfig(
	ctx context.Context,
	tenantID string,
	branchID string,
) (BillingConfig, error) {
	if strings.TrimSpace(tenantID) == "" {
		return BillingConfig{}, errors.New("tenant context is required")
	}

	if _, err := uuid.Parse(branchID); err != nil {
		return BillingConfig{}, errors.New("invalid branch id")
	}

	return s.repo.GetBillingConfig(ctx, tenantID, branchID)
}

func (s *Service) UpdateBillingConfig(
	ctx context.Context,
	userID string,
	tenantID string,
	branchID string,
	req UpdateBillingConfigRequest,
	ipAddress *string,
) (BillingConfig, error) {
	if strings.TrimSpace(tenantID) == "" {
		return BillingConfig{}, errors.New("tenant context is required")
	}

	if strings.TrimSpace(userID) == "" {
		return BillingConfig{}, errors.New("user context is required")
	}

	if _, err := uuid.Parse(branchID); err != nil {
		return BillingConfig{}, fmt.Errorf(
			"%w: invalid branch id",
			ErrInvalidBillingConfig,
		)
	}

	if req.RatePerMinute < 0 {
		return BillingConfig{}, fmt.Errorf(
			"%w: rate_per_minute cannot be negative",
			ErrInvalidBillingConfig,
		)
	}

	if req.MinimumCharge < 0 {
		return BillingConfig{}, fmt.Errorf(
			"%w: minimum_charge cannot be negative",
			ErrInvalidBillingConfig,
		)
	}

	if req.BillingIntervalMinutes <= 0 {
		return BillingConfig{}, fmt.Errorf(
			"%w: billing_interval_minutes must be greater than zero",
			ErrInvalidBillingConfig,
		)
	}

	req.RoundingMode = strings.ToLower(
		strings.TrimSpace(req.RoundingMode),
	)
	switch req.RoundingMode {
	case "up", "down", "nearest":
	default:
		return BillingConfig{}, fmt.Errorf(
			"%w: rounding_mode must be up, down, or nearest",
			ErrInvalidBillingConfig,
		)
	}

	req.Currency = strings.ToUpper(
		strings.TrimSpace(req.Currency),
	)
	if len(req.Currency) < 3 || len(req.Currency) > 10 {
		return BillingConfig{}, fmt.Errorf(
			"%w: currency must contain between 3 and 10 characters",
			ErrInvalidBillingConfig,
		)
	}

	config, err := s.repo.UpdateBillingConfig(
		ctx,
		tenantID,
		branchID,
		req,
	)
	if err != nil {
		return BillingConfig{}, err
	}

	tenant := tenantID
	branch := branchID

	_ = s.repo.CreateAuditLog(
		ctx,
		&tenant,
		&branch,
		userID,
		"owner",
		"branch.billing_config_updated",
		"billing_config",
		config.ID,
		nil,
		fmt.Sprintf(
			`{"rate_per_minute":%v,"minimum_charge":%v,"billing_interval_minutes":%d,"rounding_mode":%q,"currency":%q}`,
			config.RatePerMinute,
			config.MinimumCharge,
			config.BillingIntervalMinutes,
			config.RoundingMode,
			config.Currency,
		),
		"Branch billing configuration updated",
		ipAddress,
	)

	return config, nil
}
