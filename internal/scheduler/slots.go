package scheduler

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

// SlotScheduler каждый день создаёт слоты на 14-й день вперёд,
// чтобы окно в 2 недели всегда было заполнено.
// Слоты создаются для ВСЕХ репетиторов из таблицы tutors.
type SlotScheduler struct {
	db *sqlx.DB
}

func NewSlotScheduler(db *sqlx.DB) *SlotScheduler {
	return &SlotScheduler{db: db}
}

func (s *SlotScheduler) Run(ctx context.Context) {
	// При старте — заполнить пропущенные дни за все 14 дней
	s.generateSlots(ctx, 0, 13)

	// Вычисляем время до следующей полуночи
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 1, 0, now.Location())
	timer := time.NewTimer(time.Until(next))

	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			s.generateSlots(ctx, 13, 13)
			// Следующий тик — через сутки
			timer.Reset(24 * time.Hour)
		}
	}
}

// generateSlots создаёт слоты с CURRENT_DATE+fromDay по CURRENT_DATE+toDay
// для всех репетиторов. INSERT ... ON CONFLICT DO NOTHING — не дублирует.
func (s *SlotScheduler) generateSlots(ctx context.Context, fromDay, toDay int) {
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO time_slots (tutor_id, subject_id, class_group_id, slot_date, start_time, end_time)
		SELECT t.id, s.id, g.id, d::date,
		       make_time(h, 0, 0), make_time(h + 1, 0, 0)
		FROM tutors t
		CROSS JOIN generate_series(CURRENT_DATE + ($1::int), CURRENT_DATE + ($2::int), interval '1 day') AS d
		CROSS JOIN generate_series(9, 19) AS h
		CROSS JOIN subjects s
		CROSS JOIN class_groups g
		WHERE EXTRACT(DOW FROM d::date) <> 0
		ON CONFLICT DO NOTHING`,
		fromDay, toDay,
	)
	if err != nil {
		log.Error().Err(err).Msg("scheduler: failed to generate slots")
		return
	}
	rows, _ := res.RowsAffected()
	if rows > 0 {
		log.Info().Int64("created", rows).Msg("scheduler: new slots generated")
	}
}
