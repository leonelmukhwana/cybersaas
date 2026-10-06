package platform

import "context"

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}

// ============================================================
// DASHBOARD
// ============================================================

func (s *Service) GetDashboardStats(
	ctx context.Context,
) (DashboardStats, error) {
	return s.repository.GetDashboardStats(ctx)
}

// ============================================================
// CYBER OWNERS
// ============================================================

func (s *Service) GetCyberOwners(
	ctx context.Context,
	search string,
	limit int,
	offset int,
) (CyberOwnerListResponse, error) {
	if limit <= 0 {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	if offset < 0 {
		offset = 0
	}

	return s.repository.GetCyberOwners(
		ctx,
		search,
		limit,
		offset,
	)
}

// ============================================================
// SUBSCRIPTIONS
// ============================================================

func (s *Service) GetSubscriptions(
	ctx context.Context,
	search string,
	limit int,
	offset int,
) (SubscriptionListResponse, error) {
	if limit <= 0 {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	if offset < 0 {
		offset = 0
	}

	return s.repository.GetSubscriptions(
		ctx,
		search,
		limit,
		offset,
	)
}

// ============================================================
// REVENUE
// ============================================================

func (s *Service) GetRevenue(
	ctx context.Context,
	search string,
	limit int,
	offset int,
) (RevenueListResponse, error) {
	if limit <= 0 {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	if offset < 0 {
		offset = 0
	}

	return s.repository.GetRevenue(
		ctx,
		search,
		limit,
		offset,
	)
}

// ============================================================
// AUDIT LOGS
// ============================================================

func (s *Service) GetAuditLogs(
	ctx context.Context,
	search string,
	limit int,
	offset int,
) (AuditLogResponse, error) {
	if limit <= 0 {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	if offset < 0 {
		offset = 0
	}

	return s.repository.GetAuditLogs(
		ctx,
		search,
		limit,
		offset,
	)
}
