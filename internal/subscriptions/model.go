package subscriptions

import "time"

// Subscription mirrors API Spec §6.
type Subscription struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	ProductID    string    `json:"product_id"`
	HubID        string    `json:"hub_id"`
	AddressID    string    `json:"address_id"`
	Quantity     int       `json:"quantity"`
	FreeQuantity int       `json:"free_quantity"`
	Frequency    string    `json:"frequency"`
	DeliveryTime string    `json:"delivery_time"`
	StartDate    time.Time `json:"start_date"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateRequest struct {
	ProductID    string `json:"product_id"    validate:"required,uuid"`
	HubID        string `json:"hub_id"        validate:"required,uuid"`
	AddressID    string `json:"address_id"    validate:"required,uuid"`
	Quantity     int    `json:"quantity"      validate:"required,gt=0"`
	Frequency    string `json:"frequency"     validate:"required,oneof=daily weekly monthly"`
	DeliveryTime string `json:"delivery_time" validate:"required"`
	StartDate    string `json:"start_date"    validate:"required,datetime=2006-01-02"`
}

type UpdateRequest struct {
	Quantity     *int    `json:"quantity"      validate:"omitempty,gt=0"`
	DeliveryTime *string `json:"delivery_time" validate:"omitempty"`
	AddressID    *string `json:"address_id"    validate:"omitempty,uuid"`
}
