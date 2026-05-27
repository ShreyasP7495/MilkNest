package admin

// DashboardStats mirrors API Spec §12.1.
type DashboardStats struct {
	TotalUsers          int64 `json:"total_users"`
	ActiveSubscriptions int64 `json:"active_subscriptions"`
	DailyOrders         int64 `json:"daily_orders"`
	RevenuePaise        int64 `json:"revenue_paise"`
	// Revenue is the same number in rupees (convenience for legacy UI).
	Revenue float64 `json:"revenue"`
}
