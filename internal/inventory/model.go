package inventory

import "time"

type Inventory struct {
	ID               string    `json:"id"`
	ProductID        string    `json:"product_id"`
	HubID            string    `json:"hub_id"`
	Quantity         int       `json:"quantity"`
	ReservedQuantity int       `json:"reserved_quantity"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// UpdateRequest is the body of POST /admin/inventory/update.
type UpdateRequest struct {
	ProductID string `json:"product_id" validate:"required,uuid"`
	HubID     string `json:"hub_id"     validate:"required,uuid"`
	// Quantity is the new absolute on-hand stock. Reserved quantity is left
	// untouched.
	Quantity int `json:"quantity" validate:"gte=0"`
}
