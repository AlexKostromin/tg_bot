// internal/api/dto/stats.go

package dto

type StatsResponse struct {
	BookingsToday  int              `json:"bookings_today"`
	BookingsTotal  int              `json:"bookings_total"`
	UsersTotal     int              `json:"users_total"`
	ConfirmedTotal int              `json:"confirmed_total"`
	BookingsChart  []ChartDataPoint `json:"bookings_chart"`
}

type ChartDataPoint struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}
