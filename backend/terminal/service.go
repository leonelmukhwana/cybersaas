package terminal

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
	"cybersaas/backend/subscription"
)

type Service struct {
	repo *Repository
	subscription *subscription.Service
}

func NewService(repo *Repository, subscriptionService *subscription.Service) *Service {
	return &Service{
			repo:         repo,
			subscription: subscriptionService,
	}
}
// ============================================================
// LICENCE KEYS
// ============================================================

func (s *Service) GenerateLicenceKey(
	ctx context.Context,
	tenantID string,
	userID string,
	req GenerateLicenceKeyRequest,
) (*GenerateLicenceKeyResponse, error) {
	if tenantID == "" {
		return nil, errors.New("tenant context is required")
	}

	if userID == "" {
		return nil, errors.New("user context is required")
	}

	req.BranchID = strings.TrimSpace(req.BranchID)

	if req.BranchID == "" {
		return nil, errors.New("branch_id is required")
	}

	if err := s.repo.VerifyBranchBelongsToTenant(
		ctx,
		tenantID,
		req.BranchID,
	); err != nil {
		return nil, err
	}

	// Default licence lifetime: 24 hours.
	expiresInHours := 24

	if req.ExpiresInHours != nil {
		expiresInHours = *req.ExpiresInHours
	}

	if expiresInHours < 1 {
		return nil, errors.New(
			"expires_in_hours must be at least 1",
		)
	}

	if expiresInHours > 168 {
		return nil, errors.New(
			"licence key cannot expire more than 168 hours from creation",
		)
	}

	expiresAt := time.Now().Add(
		time.Duration(expiresInHours) * time.Hour,
	)

	key, plainKey, err := s.repo.CreateLicenceKey(
		ctx,
		tenantID,
		req.BranchID,
		userID,
		expiresAt,
	)
	if err != nil {
		return nil, err
	}

	return &GenerateLicenceKeyResponse{
		ID:         key.ID,
		BranchID:   key.BranchID,
		BranchName: key.BranchName,
		LicenceKey: plainKey,
		ExpiresAt:  *key.ExpiresAt,
		CreatedAt:  key.CreatedAt,
	}, nil
}

func (s *Service) ListLicenceKeys(
	ctx context.Context,
	tenantID string,
	limit int,
	offset int,
) (*LicenceKeyListResponse, error) {
	limit, offset = normalizePagination(limit, offset)

	keys, total, err := s.repo.ListLicenceKeys(
		ctx,
		tenantID,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}

	return &LicenceKeyListResponse{
		LicenceKeys: keys,
		Total:       total,
	}, nil
}

func (s *Service) RevokeLicenceKey(
	ctx context.Context,
	tenantID string,
	keyID string,
) error {
	keyID = strings.TrimSpace(keyID)

	if keyID == "" {
		return errors.New("licence key id is required")
	}

	return s.repo.RevokeLicenceKey(
		ctx,
		tenantID,
		keyID,
	)
}

// ============================================================
// TERMINALS
// ============================================================

func (s *Service) RegisterTerminal(
    ctx context.Context,
    req RegisterTerminalRequest,
) (*RegisterTerminalResponse, error) {
    req.LicenceKey = strings.TrimSpace(req.LicenceKey)
    req.MachineName = strings.TrimSpace(req.MachineName)
    req.DeviceID = strings.TrimSpace(req.DeviceID)

    if err := validateTerminalInput(req); err != nil {
        return nil, err
    }

    tenantID, err := s.repo.GetTenantIDByLicenceKey(ctx, req.LicenceKey)
    if err != nil {
        return nil, err
    }

    if s.subscription == nil {
        return nil, errors.New("subscription service is not configured")
    }

    if err := s.subscription.CanCreateTerminal(ctx, tenantID); err != nil {
        return nil, err
    }

    return s.repo.RegisterTerminal(ctx, req)
}

func (s *Service) AuthenticateTerminal(
    ctx context.Context,
    credentialHash string,
) (string, string, string, error) {
    if credentialHash == "" {
        return "", "", "", errors.New(
            "credential hash is required",
        )
    }

    return s.repo.AuthenticateTerminal(
        ctx,
        credentialHash,
    )
}


// AuthenticateTerminalRequest is used by the HTTP terminal
// authentication endpoint.
//
// The PC sends the permanent plaintext credential to this
// endpoint over HTTPS in production. The service hashes it
// before looking it up in terminal_credentials.
func (s *Service) AuthenticateTerminalRequest(
	ctx context.Context,
	req TerminalAuthRequest,
) (*TerminalAuthResponse, error) {
	req.Credential = strings.TrimSpace(req.Credential)

	if req.Credential == "" {
		return nil, errors.New("credential is required")
	}

	hash := sha256.Sum256([]byte(req.Credential))
	credentialHash := hex.EncodeToString(hash[:])

	terminalID, tenantID, branchID, err :=
		s.repo.AuthenticateTerminal(
			ctx,
			credentialHash,
		)

	if err != nil {
		return nil, errors.New(
			"invalid or revoked terminal credential",
		)
	}

	return &TerminalAuthResponse{
		TerminalID: terminalID,
		TenantID:   tenantID,
		BranchID:   branchID,
	}, nil
}

func (s *Service) GetTerminal(
	ctx context.Context,
	tenantID string,
	terminalID string,
) (*Terminal, error) {
	return s.repo.GetTerminal(
		ctx,
		tenantID,
		terminalID,
	)
}

func (s *Service) ListTerminals(
	ctx context.Context,
	tenantID string,
	limit int,
	offset int,
) (*TerminalListResponse, error) {
	limit, offset = normalizePagination(limit, offset)

	terminals, total, err := s.repo.ListTerminals(
		ctx,
		tenantID,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}

	return &TerminalListResponse{
		Terminals: terminals,
		Total:     total,
	}, nil
}

func (s *Service) RenameTerminal(
	ctx context.Context,
	tenantID string,
	terminalID string,
	req RenameTerminalRequest,
) error {
	req.MachineName = strings.TrimSpace(req.MachineName)

	if req.MachineName == "" {
		return errors.New("machine_name is required")
	}

	if len(req.MachineName) > 200 {
		return errors.New("machine_name is too long")
	}

	return s.repo.UpdateTerminalName(
		ctx,
		tenantID,
		terminalID,
		req.MachineName,
	)
}

func (s *Service) ChangeStatus(
	ctx context.Context,
	tenantID string,
	terminalID string,
	req ChangeStatusRequest,
) error {
	status := strings.ToLower(strings.TrimSpace(req.Status))

	switch status {
	case "pending", "active", "disabled":
		// valid
	default:
		return errors.New(
			"status must be pending, active, or disabled",
		)
	}

	return s.repo.ChangeTerminalStatus(
		ctx,
		tenantID,
		terminalID,
		status,
	)
}

func (s *Service) MoveTerminal(
	ctx context.Context,
	tenantID string,
	terminalID string,
	req MoveTerminalRequest,
) error {
	req.BranchID = strings.TrimSpace(req.BranchID)

	if req.BranchID == "" {
		return errors.New("branch_id is required")
	}

	if err := s.repo.VerifyBranchBelongsToTenant(
		ctx,
		tenantID,
		req.BranchID,
	); err != nil {
		return err
	}

	return s.repo.MoveTerminal(
		ctx,
		tenantID,
		terminalID,
		req.BranchID,
	)
}

// ============================================================
// TERMINAL HEARTBEAT
// ============================================================

func (s *Service) Heartbeat(
	ctx context.Context,
	terminalID string,
	req TerminalHeartbeatRequest,
) (*TerminalHeartbeatResponse, error) {
	terminalID = strings.TrimSpace(terminalID)

	if terminalID == "" {
		return nil, errors.New("terminal_id is required")
	}

	if req.MachineName != nil {
		value := strings.TrimSpace(*req.MachineName)

		if value != "" {
			if len(value) > 200 {
				return nil, errors.New("machine_name is too long")
			}

			req.MachineName = &value
		} else {
			req.MachineName = nil
		}
	}

	if req.DeviceIdentifier != nil {
		value := strings.TrimSpace(*req.DeviceIdentifier)

		if value != "" {
			if len(value) > 255 {
				return nil, errors.New(
					"device_identifier is too long",
				)
			}

			req.DeviceIdentifier = &value
		} else {
			req.DeviceIdentifier = nil
		}
	}

	if err := s.repo.Heartbeat(
		ctx,
		terminalID,
		req.MachineName,
		req.DeviceIdentifier,
	); err != nil {
		return nil, err
	}

	state, err := s.repo.GetControlState(
		ctx,
		terminalID,
	)
	if err != nil {
		return nil, err
	}

	return &TerminalHeartbeatResponse{
		TerminalID:     terminalID,
		ServerTime:     time.Now().UTC(),
		DesiredState:   state.DesiredState,
		CommandVersion: state.CommandVersion,
	}, nil
}

// ============================================================
// TERMINAL CONTROL
// ============================================================

// GetControlState returns the latest lock/unlock command
// for an authenticated terminal.
//
// This is intentionally independent of SaaS subscription status.
// Subscription expiry must never modify terminal control state.
func (s *Service) GetControlState(
	ctx context.Context,
	terminalID string,
) (*TerminalControlState, error) {
	terminalID = strings.TrimSpace(terminalID)

	if terminalID == "" {
		return nil, errors.New("terminal_id is required")
	}

	return s.repo.GetControlState(
		ctx,
		terminalID,
	)
}

// SetControlState is used by the Cyber Owner dashboard
// to request that a specific terminal be locked or unlocked.
//
// The terminal itself does not use this method.
// The terminal reads the resulting state through Heartbeat
// or GetControlState.
func (s *Service) SetControlState(
	ctx context.Context,
	tenantID string,
	terminalID string,
	state string,
) (*TerminalControlState, error) {
	tenantID = strings.TrimSpace(tenantID)
	terminalID = strings.TrimSpace(terminalID)
	state = strings.ToLower(strings.TrimSpace(state))

	if tenantID == "" {
		return nil, errors.New("tenant context is required")
	}

	if terminalID == "" {
		return nil, errors.New("terminal_id is required")
	}

	switch state {
	case "locked", "unlocked":
		// valid
	default:
		return nil, errors.New(
			"state must be locked or unlocked",
		)
	}

	if err := s.repo.VerifyTerminalBelongsToTenant(
		ctx,
		terminalID,
		tenantID,
	); err != nil {
		return nil, err
	}

	return s.repo.SetControlState(
		ctx,
		terminalID,
		state,
	)
}

// ============================================================
// DEREGISTER
// ============================================================

func (s *Service) DeregisterTerminal(
	ctx context.Context,
	tenantID string,
	terminalID string,
) error {
	tenantID = strings.TrimSpace(tenantID)
	terminalID = strings.TrimSpace(terminalID)

	if tenantID == "" {
		return errors.New("tenant context is required")
	}

	if terminalID == "" {
		return errors.New("terminal_id is required")
	}

	return s.repo.DeregisterTerminal(
		ctx,
		tenantID,
		terminalID,
	)
}

// The C# terminal client will retrieve the command, execute it,
// and acknowledge it so the same command is not executed again.
func (s *Service) QueueCommand(
	ctx context.Context,
	tenantID string,
	terminalID string,
	command TerminalCommand,
) (*TerminalCommandResponse, error) {
	tenantID = strings.TrimSpace(tenantID)
	terminalID = strings.TrimSpace(terminalID)
	command = TerminalCommand(strings.ToLower(strings.TrimSpace(string(command))))

	if tenantID == "" {
		return nil, errors.New("tenant context is required")
	}

	if terminalID == "" {
		return nil, errors.New("terminal_id is required")
	}

	switch command {
	case CommandRestart, CommandShutdown:
		// valid
	default:
		return nil, errors.New(
			"command must be restart or shutdown",
		)
	}

	// Security check: the terminal must belong to the
	// authenticated owner's tenant.
	if err := s.repo.VerifyTerminalBelongsToTenant(
		ctx,
		terminalID,
		tenantID,
	); err != nil {
		return nil, err
	}

	return s.repo.QueueCommand(
		ctx,
		terminalID,
		command,
	)
}

func (s *Service) GetPendingCommand(
	ctx context.Context,
	terminalID string,
) (*TerminalCommandResponse, error) {
	terminalID = strings.TrimSpace(terminalID)

	if terminalID == "" {
		return nil, errors.New("terminal_id is required")
	}

	return s.repo.GetPendingCommand(
		ctx,
		terminalID,
	)
}

func (s *Service) AcknowledgeCommand(
	ctx context.Context,
	terminalID string,
	commandID string,
	status string,
	errorMessage *string,
) (*TerminalCommandResponse, error) {
	terminalID = strings.TrimSpace(terminalID)
	commandID = strings.TrimSpace(commandID)
	status = strings.ToLower(strings.TrimSpace(status))

	if terminalID == "" {
		return nil, errors.New("terminal_id is required")
	}

	if commandID == "" {
		return nil, errors.New("command_id is required")
	}

	switch status {
	case "executed", "failed":
	default:
		return nil, errors.New("status must be executed or failed")
	}

	return s.repo.AcknowledgeCommand(
		ctx,
		terminalID,
		commandID,
		status,
		errorMessage,
	)
}

// ============================================================
// ATTENDANT TERMINALS
// ============================================================

// ListAttendantTerminals returns only terminals from the
// attendant's currently assigned branch.
func (s *Service) ListAttendantTerminals(
	ctx context.Context,
	tenantID string,
	userID string,
	limit int,
	offset int,
) (*TerminalListResponse, error) {
	tenantID = strings.TrimSpace(tenantID)
	userID = strings.TrimSpace(userID)

	if tenantID == "" {
		return nil, errors.New("tenant context is required")
	}

	if userID == "" {
		return nil, errors.New("user context is required")
	}

	limit, offset = normalizePagination(limit, offset)

	terminals, total, err := s.repo.ListAttendantTerminals(
		ctx,
		tenantID,
		userID,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}

	return &TerminalListResponse{
		Terminals: terminals,
		Total:     total,
	}, nil
}

func (s *Service) GetAttendantTerminal(
	ctx context.Context,
	tenantID string,
	userID string,
	terminalID string,
) (*Terminal, error) {
	tenantID = strings.TrimSpace(tenantID)
	userID = strings.TrimSpace(userID)
	terminalID = strings.TrimSpace(terminalID)

	if tenantID == "" {
		return nil, errors.New("tenant context is required")
	}

	if userID == "" {
		return nil, errors.New("user context is required")
	}

	if terminalID == "" {
		return nil, errors.New("terminal_id is required")
	}

	return s.repo.GetAttendantTerminal(
		ctx,
		tenantID,
		userID,
		terminalID,
	)
}

func (s *Service) SetAttendantControlState(
	ctx context.Context,
	tenantID string,
	userID string,
	terminalID string,
	state string,
) (*TerminalControlState, error) {
	tenantID = strings.TrimSpace(tenantID)
	userID = strings.TrimSpace(userID)
	terminalID = strings.TrimSpace(terminalID)
	state = strings.ToLower(strings.TrimSpace(state))

	if tenantID == "" {
		return nil, errors.New("tenant context is required")
	}

	if userID == "" {
		return nil, errors.New("user context is required")
	}

	if terminalID == "" {
		return nil, errors.New("terminal_id is required")
	}

	switch state {
	case "locked", "unlocked":
	default:
		return nil, errors.New(
			"state must be locked or unlocked",
		)
	}

	return s.repo.SetAttendantControlState(
		ctx,
		tenantID,
		userID,
		terminalID,
		state,
	)
}

func (s *Service) QueueAttendantCommand(
	ctx context.Context,
	tenantID string,
	userID string,
	terminalID string,
	command TerminalCommand,
) (*TerminalCommandResponse, error) {
	tenantID = strings.TrimSpace(tenantID)
	userID = strings.TrimSpace(userID)
	terminalID = strings.TrimSpace(terminalID)
	command = TerminalCommand(
		strings.ToLower(strings.TrimSpace(string(command))),
	)

	if tenantID == "" {
		return nil, errors.New("tenant context is required")
	}

	if userID == "" {
		return nil, errors.New("user context is required")
	}

	if terminalID == "" {
		return nil, errors.New("terminal_id is required")
	}

	switch command {
	case CommandRestart, CommandShutdown:
	default:
		return nil, errors.New(
			"command must be restart or shutdown",
		)
	}

	return s.repo.QueueAttendantCommand(
		ctx,
		tenantID,
		userID,
		terminalID,
		command,
	)
}

// ============================================================
// PAGINATION
// ============================================================

func normalizePagination(
	limit int,
	offset int,
) (int, int) {
	if limit <= 0 {
		limit = 50
	}

	if limit > 100 {
		limit = 100
	}

	if offset < 0 {
		offset = 0
	}

	return limit, offset
}

// ============================================================
// COMPILE-TIME SAFETY HELPERS
// ============================================================

var _ = fmt.Sprintf
