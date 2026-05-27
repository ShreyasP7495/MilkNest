package orders

import "time"

// Order mirrors API Spec §7.
type Order struct {
	ID                string    `json:"id"`
	SubscriptionID    *string   `json:"subscription_id,omitempty"`
	UserID            string    `json:"user_id"`
	ProductID         string    `json:"product_id"`
	HubID             string    `json:"hub_id"`
	AddressID         string    `json:"address_id"`
	DeliveryDate      time.Time `json:"delivery_date"`
	Quantity          int       `json:"quantity"`
	AmountPaise       int64     `json:"amount_paise"`
	Status            string    `json:"status"`
	PaymentID         *string   `json:"payment_id,omitempty"`
	DeliveryPartnerID *string   `json:"delivery_partner_id,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
