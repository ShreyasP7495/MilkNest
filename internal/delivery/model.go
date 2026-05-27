package delivery

import "time"

type Tracking struct {
	ID                string    `json:"id"`
	OrderID           string    `json:"order_id"`
	DeliveryPartnerID string    `json:"delivery_partner_id"`
	Latitude          float64   `json:"latitude"`
	Longitude         float64   `json:"longitude"`
	Status            string    `json:"status"`
	RecordedAt        time.Time `json:"recorded_at"`
}

// AssignRequest mirrors API Spec §9.1.
type AssignRequest struct {
	OrderID           string `json:"order_id"             validate:"required,uuid"`
	DeliveryPartnerID string `json:"delivery_partner_id"  validate:"required,uuid"`
}

// StatusUpdateRequest mirrors API Spec §9.2.
type StatusUpdateRequest struct {
	OrderID   string  `json:"order_id"  validate:"required,uuid"`
	Status    string  `json:"status"    validate:"required,oneof=picked in_transit delivered failed"`
	Latitude  float64 `json:"latitude"  validate:"required,latitude"`
	Longitude float64 `json:"longitude" validate:"required,longitude"`
}

// TrackResponse mirrors API Spec §9.3.
type TrackResponse struct {
	OrderID       string        `json:"order_id"`
	CurrentStatus string        `json:"current_status"`
	LiveLocation  *LiveLocation `json:"live_location,omitempty"`
}

type LiveLocation struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}
