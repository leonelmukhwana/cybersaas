package customer

import "time"

type Customer struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenant_id"`
	BranchID       string    `json:"branch_id"`
	CustomerType   string    `json:"customer_type"`
	FullName       string    `json:"full_name"`
	Phone          *string   `json:"phone,omitempty"`
	IDNumber       *string   `json:"id_number,omitempty"`
	ParentName     *string   `json:"parent_name,omitempty"`
	ParentPhone    *string   `json:"parent_phone,omitempty"`
	ParentIDNumber *string   `json:"parent_id_number,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreateCustomerRequest struct {
	ID             string  `json:"id" binding:"required"`
	BranchID       string  `json:"branch_id" binding:"required"`
	CustomerType   string  `json:"customer_type" binding:"required"`
	FullName       string  `json:"full_name" binding:"required"`
	Phone          *string `json:"phone"`
	IDNumber       *string `json:"id_number"`
	ParentName     *string `json:"parent_name"`
	ParentPhone    *string `json:"parent_phone"`
	ParentIDNumber *string `json:"parent_id_number"`
}

type UpdateCustomerRequest struct {
	CustomerType   string  `json:"customer_type" binding:"required"`
	FullName       string  `json:"full_name" binding:"required"`
	Phone          *string `json:"phone"`
	IDNumber       *string `json:"id_number"`
	ParentName     *string `json:"parent_name"`
	ParentPhone    *string `json:"parent_phone"`
	ParentIDNumber *string `json:"parent_id_number"`
}

type CustomerListResponse struct {
	Customers []Customer `json:"customers"`
	Total     int64      `json:"total"`
}
