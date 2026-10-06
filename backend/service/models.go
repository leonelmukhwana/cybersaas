package service

import "time"

type Service struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	BranchID    string    `json:"branch_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Price       string    `json:"price"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateServiceRequest struct {
	BranchID    string  `json:"branch_id" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
	Price       string  `json:"price" binding:"required"`
}

type UpdateServiceRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
	Price       string  `json:"price" binding:"required"`
}

type ChangeStatusRequest struct {
	Active bool `json:"active"`
}

type ServiceListResponse struct {
	Services []Service `json:"services"`
	Total    int64     `json:"total"`
}
