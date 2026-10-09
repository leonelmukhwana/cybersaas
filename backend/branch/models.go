package branch

import (
	"time"
)

type Branch struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Name      string    `json:"name"`
	Address   *string   `json:"address,omitempty"`
	Phone     *string   `json:"phone,omitempty"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateBranchRequest struct {
	BusinessName string  `json:"business_name"`
	Name         string  `json:"name"`
	Address      *string `json:"address,omitempty"`
	Phone        *string `json:"phone,omitempty"`
}

type UpdateBranchRequest struct {
	Name    string  `json:"name"`
	Address *string `json:"address,omitempty"`
	Phone   *string `json:"phone,omitempty"`
}

type ChangeStatusRequest struct {
	Status string `json:"status"`
}

type BranchListResponse struct {
	Branches []Branch `json:"branches"`
	Total    int64    `json:"total"`
}

type BillingConfig struct {
	ID                     string    `json:"id"`
	TenantID               string    `json:"tenant_id"`
	BranchID               string    `json:"branch_id"`
	RatePerMinute          float64   `json:"rate_per_minute"`
	MinimumCharge          float64   `json:"minimum_charge"`
	BillingIntervalMinutes int       `json:"billing_interval_minutes"`
	RoundingMode           string    `json:"rounding_mode"`
	Currency               string    `json:"currency"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

type UpdateBillingConfigRequest struct {
	RatePerMinute          float64 `json:"rate_per_minute" binding:"required"`
	MinimumCharge          float64 `json:"minimum_charge" binding:"required"`
	BillingIntervalMinutes int     `json:"billing_interval_minutes" binding:"required"`
	RoundingMode           string  `json:"rounding_mode" binding:"required"`
	Currency               string  `json:"currency" binding:"required"`
}
