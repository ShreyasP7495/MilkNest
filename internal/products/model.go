package products

import "time"

// Product mirrors API Spec §5.
type Product struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Brand      string `json:"brand"`
	UnitSizeML int    `json:"unit_size_ml"`
	// PricePaise is the canonical money field.
	PricePaise int64 `json:"price_paise"`
	// Price is a derived convenience field in rupees for legacy clients matching
	// the API Spec sample response.
	Price    float64 `json:"price"`
	ImageURL *string `json:"image_url,omitempty"`
	// StockQuantity is the sum of available (quantity - reserved) inventory
	// across all hubs.
	StockQuantity int       `json:"stock_quantity"`
	Active        bool      `json:"active"`
	CreatedAt     time.Time `json:"created_at"`
}
