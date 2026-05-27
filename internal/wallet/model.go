package wallet

import "time"

type Wallet struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	BalancePaise int64     `json:"balance_paise"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Transaction struct {
	ID                string    `json:"id"`
	WalletID          string    `json:"wallet_id"`
	Type              string    `json:"type"`
	AmountPaise       int64     `json:"amount_paise"`
	RefType           string    `json:"ref_type"`
	RefID             string    `json:"ref_id"`
	BalanceAfterPaise int64     `json:"balance_after_paise"`
	CreatedAt         time.Time `json:"created_at"`
}

type TopupRequest struct {
	AmountPaise int64 `json:"amount_paise" validate:"required,gt=0"`
}
