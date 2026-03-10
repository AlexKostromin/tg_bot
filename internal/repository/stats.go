package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type DashboardStats struct {
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

type StatsRepository struct {
	db *sqlx.DB
}

func NewStatsRepository(db *sqlx.DB) *StatsRepository {
	return &StatsRepository{db: db}
}

func (r *StatsRepository) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	var stats DashboardStats

	r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM bookings WHERE booked_at::date = CURRENT_DATE`).Scan(&stats.BookingsToday)
	r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM bookings`).Scan(&stats.BookingsTotal)
	r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM users`).Scan(&stats.UsersTotal)
	r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM bookings WHERE status = 'confirmed'`).Scan(&stats.ConfirmedTotal)

	rows, err := r.db.QueryContext(ctx,
		`SELECT booked_at::date AS date, COUNT(*) AS count
		 FROM bookings
		 WHERE booked_at >= CURRENT_DATE - INTERVAL '30 days'
		 GROUP BY booked_at::date
		 ORDER BY date`)
	if err != nil {
		return &stats, nil
	}
	defer rows.Close()

	for rows.Next() {
		var p ChartDataPoint
		rows.Scan(&p.Date, &p.Count)
		stats.BookingsChart = append(stats.BookingsChart, p)
	}

	return &stats, nil
}
