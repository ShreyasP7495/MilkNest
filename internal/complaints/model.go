package complaints

import "time"

type Complaint struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	OrderID        *string    `json:"order_id,omitempty"`
	Type           string     `json:"type"`
	Description    string     `json:"description"`
	Status         string     `json:"status"`
	ResolutionNote *string    `json:"resolution_note,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
}

type CreateRequest struct {
	OrderID     string `json:"order_id"    validate:"omitempty,uuid"`
	Type        string `json:"type"        validate:"required,oneof=late_delivery damaged_product other"`
	Description string `json:"description" validate:"required,min=5,max=2000"`
}

type ResolveRequest struct {
	Note string `json:"note" validate:"required,min=2,max=500"`
}
