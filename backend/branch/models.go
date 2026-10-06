package branch

import "time"

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
