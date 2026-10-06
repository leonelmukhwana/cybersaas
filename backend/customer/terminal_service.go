package customer

import (
	"context"
	"strings"
	"time"
)

type TerminalCustomer struct {
	ID           string    `json:"id"`
	BranchID     string    `json:"branch_id"`
	CustomerType string    `json:"customer_type"`
	FullName     string    `json:"full_name"`
	Phone        *string   `json:"phone,omitempty"`
	IDNumberHash *string   `json:"id_number_hash,omitempty"`
	ParentID     *string   `json:"parent_id,omitempty"`
	ParentName   *string   `json:"parent_name,omitempty"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type TerminalCustomerLookupRequest struct {
	SearchType string `json:"search_type"`
	SearchTerm string `json:"search_term"`

	// Used for adult lookup.
	IDNumber *string `json:"id_number,omitempty"`

	// Used for child lookup.
	FullName *string `json:"full_name,omitempty"`
}

type TerminalCustomerLookupResponse struct {
	Customer *TerminalCustomer `json:"customer"`
}

type TerminalCustomerSyncResponse struct {
	Customers []TerminalCustomer `json:"customers"`
	Total     int64              `json:"total"`
}

func (s *Service) LookupForTerminal(
	ctx context.Context,
	tenantID string,
	branchID string,
	searchType string,
	searchTerm string,
) (*TerminalCustomerLookupResponse, error) {
	tenantID = strings.TrimSpace(tenantID)
	branchID = strings.TrimSpace(branchID)
	searchType = strings.ToLower(strings.TrimSpace(searchType))
	searchTerm = strings.TrimSpace(searchTerm)

	if tenantID == "" {
		return nil, ErrBranchAccessDenied
	}

	if branchID == "" {
		return nil, ErrBranchAccessDenied
	}

	if searchTerm == "" {
		return nil, ErrCustomerNotFound
	}

	if searchType != "adult" && searchType != "child" {
		return nil, ErrCustomerNotFound
	}

	if err := s.repository.VerifyBranchBelongsToTenant(
		ctx,
		branchID,
		tenantID,
	); err != nil {
		return nil, err
	}

	var (
		customer *TerminalCustomer
		err      error
	)

	switch searchType {
	case "adult":
		idNumber := normalizeID(searchTerm)

		if idNumber == "" {
			return nil, ErrCustomerNotFound
		}

		hash := hashID(idNumber)

		customer, err =
			s.repository.FindTerminalCustomerByIDHash(
				ctx,
				tenantID,
				branchID,
				hash,
			)

	case "child":
		customer, err =
			s.repository.FindTerminalChildByName(
				ctx,
				tenantID,
				branchID,
				searchTerm,
			)
	}

	if err != nil {
		return nil, err
	}

	return &TerminalCustomerLookupResponse{
		Customer: customer,
	}, nil
}
