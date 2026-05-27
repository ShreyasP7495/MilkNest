package addresses

import "time"

// Address mirrors API Spec §4.
type Address struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	HouseNumber string    `json:"house_number"`
	Street      string    `json:"street"`
	City        string    `json:"city"`
	State       string    `json:"state"`
	Pincode     string    `json:"pincode"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	IsDefault   bool      `json:"is_default"`
	CreatedAt   time.Time `json:"created_at"`
}

type UpsertRequest struct {
	HouseNumber string  `json:"house_number" validate:"required,max=120"`
	Street      string  `json:"street"       validate:"required,max=240"`
	City        string  `json:"city"         validate:"required,max=80"`
	State       string  `json:"state"        validate:"required,max=80"`
	Pincode     string  `json:"pincode"      validate:"required,len=6"`
	Latitude    float64 `json:"latitude"     validate:"required,latitude"`
	Longitude   float64 `json:"longitude"    validate:"required,longitude"`
	IsDefault   bool    `json:"is_default"`
}
