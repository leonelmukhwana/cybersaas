package attendant

import "time"

type Attendant struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	FullName   string    `json:"full_name"`
	Email      *string   `json:"email,omitempty"`
	Phone      *string   `json:"phone,omitempty"`
	Status     string    `json:"status"`
	BranchID   *string   `json:"branch_id,omitempty"`
	BranchName *string   `json:"branch_name,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type CreateAttendantRequest struct {
	FullName string  `json:"full_name"`
	Email    *string `json:"email,omitempty"`
	Phone    *string `json:"phone,omitempty"`
	Password string  `json:"password"`
	BranchID string  `json:"branch_id"`
}

type UpdateAttendantRequest struct {
	FullName string  `json:"full_name"`
	Email    *string `json:"email,omitempty"`
	Phone    *string `json:"phone,omitempty"`
	Password *string `json:"password,omitempty"`
}

type ChangeStatusRequest struct {
	Status string `json:"status"`
}

type AssignBranchRequest struct {
	BranchID string `json:"branch_id"`
}

type AttendantListResponse struct {
	Attendants []Attendant `json:"attendants"`
	Total      int64       `json:"total"`
}
