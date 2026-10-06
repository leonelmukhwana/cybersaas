package tenant

import "time"

type Tenant struct {
	ID           string    `json:"id"`
	BusinessName string    `json:"business_name"`
	OwnerName    string    `json:"owner_name"`
	Phone        string    `json:"phone"`
	Email        *string   `json:"email,omitempty"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type CreateTenantRequest struct {
	BusinessName string `json:"business_name" binding:"required,min=2,max=200"`
}

type UpdateTenantRequest struct {
	BusinessName string `json:"business_name" binding:"required,min=2,max=200"`
}
