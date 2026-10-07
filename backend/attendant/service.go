package attendant

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"cybersaas/backend/auth"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(
	ctx context.Context,
	tenantID string,
	ownerID string,
	req CreateAttendantRequest,
) (Attendant, error) {
	req.FullName = strings.TrimSpace(req.FullName)

	if req.FullName == "" {
		return Attendant{}, errors.New("full name is required")
	}

	if len(req.FullName) < 2 {
		return Attendant{}, errors.New("full name is too short")
	}

	if len(req.Password) < 8 {
		return Attendant{}, errors.New("password must be at least 8 characters")
	}

	if strings.TrimSpace(req.BranchID) == "" {
		return Attendant{}, errors.New("branch_id is required")
	}

	if req.Email != nil {
		email := strings.TrimSpace(strings.ToLower(*req.Email))
		if email == "" {
			req.Email = nil
		} else {
			req.Email = &email
		}
	}

	if req.Phone != nil {
		phone := strings.TrimSpace(*req.Phone)

		if phone == "" {
			req.Phone = nil
		} else if !isKenyanPhone(phone) {
			return Attendant{}, errors.New("invalid Kenyan phone number")
		} else {
			req.Phone = &phone
		}
	}

	if err := s.repo.VerifyBranchBelongsToTenant(
		ctx,
		tenantID,
		req.BranchID,
	); err != nil {
		return Attendant{}, err
	}

	if err := s.repo.VerifyBranchHasNoActiveAttendant(
		ctx,
		tenantID,
		req.BranchID,
	); err != nil {
		return Attendant{}, err
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		return Attendant{}, fmt.Errorf("hash password: %w", err)
	}

	attendant, err := s.repo.CreateAttendant(
		ctx,
		tenantID,
		req,
		passwordHash,
	)
	if err != nil {
		return Attendant{}, err
	}

	if err := s.repo.AssignBranch(
		ctx,
		tenantID,
		attendant.ID,
		req.BranchID,
		ownerID,
	); err != nil {
		return Attendant{}, err
	}

	attendant.BranchID = &req.BranchID

	if err := s.repo.CreateAuditLog(
		ctx,
		tenantID,
		ownerID,
		"attendant_created",
		attendant.ID,
		"Cyber attendant created",
	); err != nil {
		return Attendant{}, fmt.Errorf("create audit log: %w", err)
	}

	return s.repo.GetAttendant(ctx, tenantID, attendant.ID)
}

func (s *Service) Get(
	ctx context.Context,
	tenantID string,
	id string,
) (Attendant, error) {
	return s.repo.GetAttendant(ctx, tenantID, id)
}

func (s *Service) List(
	ctx context.Context,
	tenantID string,
) (AttendantListResponse, error) {
	attendants, err := s.repo.ListAttendants(ctx, tenantID)
	if err != nil {
		return AttendantListResponse{}, err
	}

	total, err := s.repo.CountAttendants(ctx, tenantID)
	if err != nil {
		return AttendantListResponse{}, err
	}

	if attendants == nil {
		attendants = []Attendant{}
	}

	return AttendantListResponse{
		Attendants: attendants,
		Total:      total,
	}, nil
}

func (s *Service) Update(
	ctx context.Context,
	tenantID string,
	ownerID string,
	id string,
	req UpdateAttendantRequest,
) error {
	req.FullName = strings.TrimSpace(req.FullName)

	if req.FullName == "" {
		return errors.New("full name is required")
	}

	if req.Email != nil {
		email := strings.TrimSpace(strings.ToLower(*req.Email))

		if email == "" {
			req.Email = nil
		} else {
			req.Email = &email
		}
	}

	if req.Phone != nil {
		phone := strings.TrimSpace(*req.Phone)

		if phone == "" {
			req.Phone = nil
		} else if !isKenyanPhone(phone) {
			return errors.New("invalid Kenyan phone number")
		} else {
			req.Phone = &phone
		}
	}

	var passwordHash *string

	if req.Password != nil {
		password := strings.TrimSpace(*req.Password)

		if password == "" {
			return errors.New("password cannot be empty")
		}

		if len(password) < 8 {
			return errors.New("password must be at least 8 characters")
		}

		hash, err := auth.HashPassword(password)
		if err != nil {
			return fmt.Errorf("hash password: %w", err)
		}

		passwordHash = &hash
	}

	if err := s.repo.UpdateAttendant(
		ctx,
		tenantID,
		id,
		req,
		passwordHash,
	); err != nil {
		return err
	}

	return s.repo.CreateAuditLog(
		ctx,
		tenantID,
		ownerID,
		"attendant_updated",
		id,
		"Cyber attendant details updated",
	)
}

func (s *Service) ChangeStatus(
	ctx context.Context,
	tenantID string,
	ownerID string,
	id string,
	status string,
) error {
	status = strings.ToLower(strings.TrimSpace(status))

	switch status {
	case "active", "inactive", "suspended":
	default:
		return errors.New("invalid attendant status")
	}

	if err := s.repo.UpdateStatus(
		ctx,
		tenantID,
		id,
		status,
	); err != nil {
		return err
	}

	return s.repo.CreateAuditLog(
		ctx,
		tenantID,
		ownerID,
		"attendant_status_changed",
		id,
		"Attendant status changed to "+status,
	)
}

func (s *Service) AssignBranch(
	ctx context.Context,
	tenantID string,
	ownerID string,
	id string,
	branchID string,
) error {
	if strings.TrimSpace(branchID) == "" {
		return errors.New("branch_id is required")
	}

	if err := s.repo.AssignBranch(
		ctx,
		tenantID,
		id,
		branchID,
		ownerID,
	); err != nil {
		return err
	}

	return s.repo.CreateAuditLog(
		ctx,
		tenantID,
		ownerID,
		"attendant_branch_assigned",
		id,
		"Attendant assigned to branch "+branchID,
	)
}

func (s *Service) UnassignBranch(
	ctx context.Context,
	tenantID string,
	ownerID string,
	id string,
) error {
	if err := s.repo.UnassignBranch(
		ctx,
		tenantID,
		id,
	); err != nil {
		return err
	}

	return s.repo.CreateAuditLog(
		ctx,
		tenantID,
		ownerID,
		"attendant_branch_unassigned",
		id,
		"Attendant unassigned from branch",
	)
}

func isKenyanPhone(phone string) bool {
	pattern := regexp.MustCompile(`^(07|01)[0-9]{8}$|^254(7|1)[0-9]{8}$`)
	return pattern.MatchString(phone)
}
