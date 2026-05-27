package utils

import "time"

// IST returns the Asia/Kolkata location. Falls back to a fixed +5:30 offset if
// the system tzdata is unavailable (e.g. minimal containers without tzdata).
func IST() *time.Location {
	if loc, err := time.LoadLocation("Asia/Kolkata"); err == nil {
		return loc
	}
	return time.FixedZone("IST", 5*3600+30*60)
}

// NowIST is a convenience for "right now in IST".
func NowIST() time.Time { return time.Now().In(IST()) }

// TomorrowIST returns midnight tomorrow in IST (used for delivery_date).
func TomorrowIST() time.Time {
	now := NowIST()
	return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, IST())
}
