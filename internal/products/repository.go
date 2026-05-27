package products

import (
	"context"
	"database/sql"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

const baseSelect = `
	SELECT p.id, p.name, p.brand, p.unit_size_ml, p.price_paise, p.image_url, p.active, p.created_at,
	       COALESCE(SUM(i.quantity - i.reserved_quantity), 0) AS available
	  FROM products p
	  LEFT JOIN inventory i ON i.product_id = p.id`

func (r *Repository) List(ctx context.Context) ([]Product, error) {
	q := baseSelect + ` WHERE p.active = TRUE GROUP BY p.id ORDER BY p.name`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Product, 0)
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Brand, &p.UnitSizeML, &p.PricePaise, &p.ImageURL, &p.Active, &p.CreatedAt, &p.StockQuantity); err != nil {
			return nil, err
		}
		p.Price = float64(p.PricePaise) / 100
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *Repository) Get(ctx context.Context, id string) (*Product, error) {
	q := baseSelect + ` WHERE p.id = $1 GROUP BY p.id`
	var p Product
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&p.ID, &p.Name, &p.Brand, &p.UnitSizeML, &p.PricePaise, &p.ImageURL, &p.Active, &p.CreatedAt, &p.StockQuantity,
	)
	if err != nil {
		return nil, err
	}
	p.Price = float64(p.PricePaise) / 100
	return &p, nil
}
