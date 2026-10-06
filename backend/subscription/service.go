package subscription

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	TrialBranchLimit   = 1
	TrialTerminalLimit = 5
	TrialDurationDays  = 7
)

var (
	ErrSubscriptionExpired  = errors.New("subscription has expired")
	ErrSubscriptionInactive = errors.New("subscription is not active")
	ErrTrialExpired         = errors.New("free trial has expired")
	ErrPlanRequired         = errors.New("subscription plan is required")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// ------------------------------------------------------------
// PLANS
// ------------------------------------------------------------

func (s *Service) GetPlans(ctx context.Context) ([]Plan, error) {
	return s.repo.GetPlans(ctx)
}

func (s *Service) GetPlan(ctx context.Context, planID string) (Plan, error) {
	return s.repo.GetPlan(ctx, planID)
}

// ------------------------------------------------------------
// SUBSCRIPTION OVERVIEW
// ------------------------------------------------------------

func (s *Service) GetOverview(
	ctx context.Context,
	tenantID string,
) (SubscriptionOverview, error) {
	if tenantID == "" {
		return SubscriptionOverview{}, errors.New("tenant ID is required")
	}

	subscription, err := s.repo.GetSubscriptionByTenant(ctx, tenantID)
	if err != nil {
		return SubscriptionOverview{}, err
	}

	branches, terminals, err := s.repo.GetUsage(ctx, tenantID)
	if err != nil {
		return SubscriptionOverview{}, err
	}

	active := s.IsSubscriptionActive(subscription)
	expired := s.IsSubscriptionExpired(subscription)

	var (
		plan                  *Plan
		branchesLimit         int
		terminalsLimit        int
		mustSelectPackage     bool
		suggestedMonthlyPrice string
	)

	// --------------------------------------------------------
	// TRIAL
	// --------------------------------------------------------

	if subscription.IsTrial {
		branchesLimit = TrialBranchLimit
		terminalsLimit = TrialTerminalLimit

		mustSelectPackage = subscription.PlanID == nil ||
			*subscription.PlanID == ""

		if subscription.PlanID != nil && *subscription.PlanID != "" {
			selectedPlan, planErr := s.repo.GetPlan(
				ctx,
				*subscription.PlanID,
			)

			if planErr == nil {
				plan = &selectedPlan
				suggestedMonthlyPrice = selectedPlan.MonthlyPrice
				mustSelectPackage = false
			}
		}

		return SubscriptionOverview{
			Subscription:           subscription,
			Plan:                   plan,
			BranchesUsed:           branches,
			BranchesLimit:          branchesLimit,
			TerminalsUsed:          terminals,
			TerminalsLimit:         terminalsLimit,
			IsActive:               active,
			IsExpired:              expired,
			IsTrial:                true,
			IsTrialExpired:         expired,
			IsLifetime:             subscription.IsLifetime,
			MustSelectPackage:      mustSelectPackage,
			SuggestedMonthlyAmount: suggestedMonthlyPrice,
		}, nil
	}

	// --------------------------------------------------------
	// PAID SUBSCRIPTION
	// --------------------------------------------------------

	if subscription.PlanID != nil &&
		*subscription.PlanID != "" {

		selectedPlan, err := s.repo.GetPlan(
			ctx,
			*subscription.PlanID,
		)
		if err != nil {
			return SubscriptionOverview{}, err
		}

		plan = &selectedPlan
		branchesLimit = selectedPlan.IncludedBranches
		terminalsLimit = selectedPlan.IncludedTerminals
		suggestedMonthlyPrice = selectedPlan.MonthlyPrice
	}

	return SubscriptionOverview{
		Subscription:           subscription,
		Plan:                   plan,
		BranchesUsed:           branches,
		BranchesLimit:          branchesLimit,
		TerminalsUsed:          terminals,
		TerminalsLimit:         terminalsLimit,
		IsActive:               active,
		IsExpired:              expired,
		IsTrial:                false,
		IsTrialExpired:         false,
		IsLifetime:             subscription.IsLifetime,
		MustSelectPackage:      plan == nil,
		SuggestedMonthlyAmount: suggestedMonthlyPrice,
	}, nil
}

// ------------------------------------------------------------
// ACTIVE CHECK
//
// This is the central subscription-access rule.
//
// IMPORTANT:
// This check is for WEB/SaaS access.
//
// It must NOT be used by the terminal client to lock a PC.
// ------------------------------------------------------------

func (s *Service) IsSubscriptionActive(
	subscription Subscription,
) bool {
	now := time.Now().UTC()

	if subscription.IsTrial {
		if subscription.Status != "active" {
			return false
		}

		if subscription.TrialEndsAt == nil {
			return false
		}

		return now.Before(*subscription.TrialEndsAt)
	}

	if subscription.IsLifetime {
		return subscription.Status == "active"
	}

	if subscription.Status != "active" {
		return false
	}

	if subscription.CurrentPeriodEnd == nil {
		return false
	}

	return now.Before(*subscription.CurrentPeriodEnd)
}

// ------------------------------------------------------------
// EXPIRY CHECK
// ------------------------------------------------------------

func (s *Service) IsSubscriptionExpired(
	subscription Subscription,
) bool {
	now := time.Now().UTC()

	if subscription.IsTrial {
		return subscription.TrialEndsAt != nil &&
			!now.Before(*subscription.TrialEndsAt)
	}

	if subscription.IsLifetime {
		return false
	}

	return subscription.CurrentPeriodEnd != nil &&
		!now.Before(*subscription.CurrentPeriodEnd)
}

// ------------------------------------------------------------
// ACCESS CHECK
//
// Used by web dashboard/API operations.
//
// Trial expiration blocks access.
// Paid subscription expiration blocks access.
//
// Terminal clients must NOT use this method to lock terminals.
// ------------------------------------------------------------

func (s *Service) RequireActiveSubscription(
	ctx context.Context,
	tenantID string,
) error {
	if tenantID == "" {
		return errors.New("tenant ID is required")
	}

	subscription, err := s.repo.GetSubscriptionByTenant(
		ctx,
		tenantID,
	)
	if err != nil {
		return err
	}

	if s.IsSubscriptionActive(subscription) {
		return nil
	}

	if subscription.IsTrial {
		if s.IsSubscriptionExpired(subscription) {
			return ErrTrialExpired
		}

		return ErrSubscriptionInactive
	}

	if s.IsSubscriptionExpired(subscription) {
		return ErrSubscriptionExpired
	}

	return ErrSubscriptionInactive
}

// ------------------------------------------------------------
// BRANCH LIMIT
// ------------------------------------------------------------

func (s *Service) CanCreateBranch(
	ctx context.Context,
	tenantID string,
) error {
	if tenantID == "" {
		return errors.New("tenant ID is required")
	}

	subscription, err := s.repo.GetSubscriptionByTenant(
		ctx,
		tenantID,
	)
	if err != nil {
		return err
	}

	if !s.IsSubscriptionActive(subscription) {
		if subscription.IsTrial &&
			s.IsSubscriptionExpired(subscription) {
			return ErrTrialExpired
		}

		return ErrSubscriptionInactive
	}

	branches, _, err := s.repo.GetUsage(
		ctx,
		tenantID,
	)
	if err != nil {
		return err
	}

	if subscription.IsTrial {
		if branches >= TrialBranchLimit {
			return fmt.Errorf(
				"trial branch limit reached: free trial allows %d branch",
				TrialBranchLimit,
			)
		}

		return nil
	}

	if subscription.PlanID == nil ||
		*subscription.PlanID == "" {
		return ErrPlanRequired
	}

	plan, err := s.repo.GetPlan(
		ctx,
		*subscription.PlanID,
	)
	if err != nil {
		return err
	}

	if branches >= plan.IncludedBranches {
		return fmt.Errorf(
			"branch limit reached: plan allows %d branches",
			plan.IncludedBranches,
		)
	}

	return nil
}

// ------------------------------------------------------------
// TERMINAL LIMIT
// ------------------------------------------------------------

func (s *Service) CanCreateTerminal(
	ctx context.Context,
	tenantID string,
) error {
	if tenantID == "" {
		return errors.New("tenant ID is required")
	}

	subscription, err := s.repo.GetSubscriptionByTenant(
		ctx,
		tenantID,
	)
	if err != nil {
		return err
	}

	if !s.IsSubscriptionActive(subscription) {
		if subscription.IsTrial &&
			s.IsSubscriptionExpired(subscription) {
			return ErrTrialExpired
		}

		return ErrSubscriptionInactive
	}

	_, terminals, err := s.repo.GetUsage(
		ctx,
		tenantID,
	)
	if err != nil {
		return err
	}

	if subscription.IsTrial {
		if terminals >= TrialTerminalLimit {
			return fmt.Errorf(
				"trial terminal limit reached: free trial allows %d terminals",
				TrialTerminalLimit,
			)
		}

		return nil
	}

	if subscription.PlanID == nil ||
		*subscription.PlanID == "" {
		return ErrPlanRequired
	}

	plan, err := s.repo.GetPlan(
		ctx,
		*subscription.PlanID,
	)
	if err != nil {
		return err
	}

	if terminals >= plan.IncludedTerminals {
		return fmt.Errorf(
			"terminal limit reached: plan allows %d terminals",
			plan.IncludedTerminals,
		)
	}

	return nil
}

// ------------------------------------------------------------
// SELECT PLAN
// ------------------------------------------------------------

func (s *Service) SelectPlan(
	ctx context.Context,
	tenantID string,
	planID string,
) (Subscription, error) {
	if tenantID == "" {
		return Subscription{}, errors.New(
			"tenant ID is required",
		)
	}

	if planID == "" {
		return Subscription{}, errors.New(
			"plan ID is required",
		)
	}

	exists, err := s.repo.TenantExists(
		ctx,
		tenantID,
	)
	if err != nil {
		return Subscription{}, err
	}

	if !exists {
		return Subscription{}, errors.New(
			"tenant does not exist",
		)
	}

	plan, err := s.repo.GetPlan(
		ctx,
		planID,
	)
	if err != nil {
		return Subscription{}, err
	}

	if !plan.IsActive {
		return Subscription{}, errors.New(
			"subscription plan is inactive",
		)
	}

	subscription, err := s.repo.GetSubscriptionByTenant(
		ctx,
		tenantID,
	)
	if err != nil {
		return Subscription{}, err
	}

	if subscription.Status == "cancelled" {
		return Subscription{}, errors.New(
			"subscription has been cancelled",
		)
	}

	if subscription.IsTrial &&
		s.IsSubscriptionExpired(subscription) {
		return Subscription{}, ErrTrialExpired
	}

	return s.repo.SelectPlan(
		ctx,
		tenantID,
		planID,
	)
}

// ------------------------------------------------------------
// CREATE SUBSCRIPTION
// ------------------------------------------------------------

func (s *Service) CreateSubscription(
	ctx context.Context,
	tenantID string,
	planID string,
) (Subscription, error) {
	return s.SelectPlan(
		ctx,
		tenantID,
		planID,
	)
}

// ------------------------------------------------------------
// PAYMENT HISTORY
// ------------------------------------------------------------

func (s *Service) GetPayments(
	ctx context.Context,
	tenantID string,
	limit int,
	offset int,
) ([]SubscriptionPayment, error) {
	if limit <= 0 {
		limit = 50
	}

	if limit > 100 {
		limit = 100
	}

	if offset < 0 {
		offset = 0
	}

	return s.repo.GetSubscriptionPayments(
		ctx,
		tenantID,
		limit,
		offset,
	)
}

// ------------------------------------------------------------
// LEDGER HISTORY
// ------------------------------------------------------------

func (s *Service) GetLedger(
	ctx context.Context,
	tenantID string,
	limit int,
	offset int,
) ([]LedgerEntry, error) {
	if limit <= 0 {
		limit = 50
	}

	if limit > 100 {
		limit = 100
	}

	if offset < 0 {
		offset = 0
	}

	return s.repo.GetLedger(
		ctx,
		tenantID,
		limit,
		offset,
	)
}

// ------------------------------------------------------------
// CREATE PAYMENT
// ------------------------------------------------------------

func (s *Service) CreatePayment(
	ctx context.Context,
	tenantID string,
	phoneNumber string,
	paymentMethod string,
) (SubscriptionPayment, error) {
	if tenantID == "" {
		return SubscriptionPayment{}, errors.New(
			"tenant ID is required",
		)
	}

	if phoneNumber == "" {
		return SubscriptionPayment{}, errors.New(
			"phone number is required",
		)
	}

	subscription, err := s.repo.GetSubscriptionByTenant(
		ctx,
		tenantID,
	)
	if err != nil {
		return SubscriptionPayment{}, err
	}

	if subscription.IsTrial &&
		s.IsSubscriptionExpired(subscription) {
		return SubscriptionPayment{}, ErrTrialExpired
	}

	if subscription.PlanID == nil ||
		*subscription.PlanID == "" {
		return SubscriptionPayment{}, ErrPlanRequired
	}

	plan, err := s.repo.GetPlan(
		ctx,
		*subscription.PlanID,
	)
	if err != nil {
		return SubscriptionPayment{}, err
	}

	if !plan.IsActive {
		return SubscriptionPayment{}, errors.New(
			"subscription plan is not active",
		)
	}

	if plan.MonthlyPrice == "" ||
		plan.MonthlyPrice == "0" {
		return SubscriptionPayment{}, errors.New(
			"subscription plan has an invalid price",
		)
	}

	if plan.IsLifetime {
		return SubscriptionPayment{}, errors.New(
			"lifetime subscriptions require the lifetime payment flow",
		)
	}

	if paymentMethod == "" {
		paymentMethod = "mpesa_stk"
	}

	switch paymentMethod {
	case "mpesa_stk", "manual", "bank", "other":
	default:
		return SubscriptionPayment{}, errors.New(
			"invalid subscription payment method",
		)
	}

	return s.repo.CreateSubscriptionPayment(
		ctx,
		subscription.ID,
		tenantID,
		plan.MonthlyPrice,
		phoneNumber,
		paymentMethod,
	)
}

// ------------------------------------------------------------
// PLATFORM ADMIN — SUBSCRIPTION PRICING
// ------------------------------------------------------------

func (s *Service) GetAllPlans(
	ctx context.Context,
) ([]Plan, error) {
	return s.repo.GetAllPlans(ctx)
}

func (s *Service) UpdatePlan(
	ctx context.Context,
	planID string,
	plan Plan,
	userID string,
) (Plan, error) {
	if planID == "" {
		return Plan{}, errors.New("plan ID is required")
	}

	if userID == "" {
		return Plan{}, errors.New("user ID is required")
	}

	plan.Name = strings.TrimSpace(plan.Name)
	plan.ExtraBranchRate = strings.TrimSpace(plan.ExtraBranchRate)
	plan.ExtraTerminalRate = strings.TrimSpace(plan.ExtraTerminalRate)
	plan.MonthlyPrice = strings.TrimSpace(plan.MonthlyPrice)

	if plan.Name == "" {
		return Plan{}, errors.New("plan name is required")
	}

	if plan.IncludedBranches < 0 {
		return Plan{}, errors.New(
			"included branches cannot be negative",
		)
	}

	if plan.IncludedTerminals < 0 {
		return Plan{}, errors.New(
			"included terminals cannot be negative",
		)
	}

	if plan.ExtraBranchRate == "" {
		return Plan{}, errors.New(
			"extra branch rate is required",
		)
	}

	if plan.ExtraTerminalRate == "" {
		return Plan{}, errors.New(
			"extra terminal rate is required",
		)
	}

	if plan.MonthlyPrice == "" {
		return Plan{}, errors.New(
			"monthly price is required",
		)
	}

	return s.repo.UpdatePlan(
		ctx,
		planID,
		plan,
		userID,
	)
}

// ------------------------------------------------------------
// PLATFORM ADMIN — CREATE SUBSCRIPTION PLAN
// ------------------------------------------------------------

func (s *Service) CreatePlan(
	ctx context.Context,
	plan Plan,
	userID string,
) (Plan, error) {
	if userID == "" {
		return Plan{}, errors.New("user ID is required")
	}

	plan.Name = strings.TrimSpace(plan.Name)
	plan.ExtraBranchRate = strings.TrimSpace(plan.ExtraBranchRate)
	plan.ExtraTerminalRate = strings.TrimSpace(plan.ExtraTerminalRate)
	plan.MonthlyPrice = strings.TrimSpace(plan.MonthlyPrice)

	if plan.Name == "" {
		return Plan{}, errors.New("plan name is required")
	}

	if plan.IncludedBranches < 0 {
		return Plan{}, errors.New(
			"included branches cannot be negative",
		)
	}

	if plan.IncludedTerminals < 0 {
		return Plan{}, errors.New(
			"included terminals cannot be negative",
		)
	}

	if plan.ExtraBranchRate == "" {
		return Plan{}, errors.New(
			"extra branch rate is required",
		)
	}

	if plan.ExtraTerminalRate == "" {
		return Plan{}, errors.New(
			"extra terminal rate is required",
		)
	}

	if plan.MonthlyPrice == "" {
		return Plan{}, errors.New(
			"monthly price is required",
		)
	}

	return s.repo.CreatePlan(
		ctx,
		plan,
		userID,
	)
}

// ------------------------------------------------------------
// WEB ACCESS
// ------------------------------------------------------------

type AccessCheckResult struct {
	Allowed bool
	Reason  string
}

func (s *Service) CheckWebAccess(
	ctx context.Context,
	tenantID string,
) (bool, string, error) {
	err := s.RequireActiveSubscription(ctx, tenantID)

	if err == nil {
		return true, "", nil
	}

	switch {
	case errors.Is(err, ErrTrialExpired):
		return false, "trial_expired", nil

	case errors.Is(err, ErrSubscriptionExpired):
		return false, "subscription_expired", nil

	case errors.Is(err, ErrSubscriptionInactive):
		return false, "subscription_inactive", nil

	default:
		return false, "", err
	}
}
