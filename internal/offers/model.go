package offers

import "time"

// Offer mirrors the offers table. The ladder (50->+2, 100->+5) is seeded but
// can be modified by an admin without code changes.
type Offer struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	MinQuantity  int       `json:"min_quantity"`
	FreeQuantity int       `json:"free_quantity"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
}
