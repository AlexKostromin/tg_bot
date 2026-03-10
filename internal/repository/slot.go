package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type TimeSlot struct {
	ID           int       `db:"id"`
	TutorID      int       `db:"tutor_id"`
	SubjectID    int       `db:"subject_id"`
	ClassGroupID int       `db:"class_group_id"`
	SlotDate     time.Time `db:"slot_date"`
	StartTime    string    `db:"start_time"`
	EndTime      string    `db:"end_time"`
	IsAvailable  bool      `db:"is_available"`
	CreatedAt    time.Time `db:"created_at"`
}

type SlotRepository struct {
	db *sqlx.DB
}

func NewSlotRepository(db *sqlx.DB) *SlotRepository {
	return &SlotRepository{db: db}
}

type dateRow struct {
	SlotDate time.Time `db:"slot_date"`
}

// GetAvailableDates возвращает уникальные даты, на которые есть свободные слоты
// для данной группы и предмета. Используется в боте для первого шага записи:
// ученик сначала видит список дат, потом выбирает конкретное время.
func (r *SlotRepository) GetAvailableDates(ctx context.Context, groupID, subjectID int) ([]time.Time, error) {
	var rows []dateRow
	err := r.db.SelectContext(ctx, &rows, `
        SELECT DISTINCT slot_date
        FROM time_slots
        WHERE class_group_id = $1
          AND subject_id     = $2
          AND is_available   = TRUE
          AND slot_date      >= CURRENT_DATE
        ORDER BY slot_date`,
		groupID, subjectID,
	)
	if err != nil {
		return nil, err
	}
	dates := make([]time.Time, len(rows))
	for i, row := range rows {
		dates[i] = row.SlotDate
	}
	return dates, nil
}

// GetByID возвращает один слот по ID.
// Используется после создания брони, чтобы показать ученику дату и время.
func (r *SlotRepository) GetByID(ctx context.Context, id int) (*TimeSlot, error) {
	var row TimeSlot
	err := r.db.GetContext(ctx, &row, `SELECT * FROM time_slots WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	return &row, nil
}
// GetAvailableSlots возвращает свободные слоты на конкретную дату для группы и предмета.
// Второй шаг записи: ученик выбрал дату — теперь видит доступное время.
func (r *SlotRepository) GetAvailableSlots(ctx context.Context, groupID, subjectID int, date time.Time) ([]TimeSlot, error) {
	var slots []TimeSlot
	err := r.db.SelectContext(ctx, &slots, `SELECT id, start_time, end_time FROM time_slots WHERE class_group_id = $1 AND subject_id = $2 AND slot_date = $3 AND is_available = TRUE  ORDER BY start_time`, groupID, subjectID, date)
	return slots, err
}

// --- Admin API types and methods ---

type SlotFilters struct {
	Date      string
	GroupID   int
	Available string
	Offset    int
	Limit     int
}

type CreateSlotParams struct {
	TutorID      int
	SubjectID    int
	ClassGroupID int
	SlotDate     string
	StartTime    string
	EndTime      string
}

// ListWithFilters возвращает слоты с фильтрацией и пагинацией для админки.
func (r *SlotRepository) ListWithFilters(ctx context.Context, f SlotFilters) ([]TimeSlot, int, error) {
	query := `SELECT * FROM time_slots WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM time_slots WHERE 1=1`
	var args []interface{}
	idx := 1

	if f.Date != "" {
		query += fmt.Sprintf(` AND slot_date = $%d`, idx)
		countQuery += fmt.Sprintf(` AND slot_date = $%d`, idx)
		args = append(args, f.Date)
		idx++
	}
	if f.GroupID > 0 {
		query += fmt.Sprintf(` AND class_group_id = $%d`, idx)
		countQuery += fmt.Sprintf(` AND class_group_id = $%d`, idx)
		args = append(args, f.GroupID)
		idx++
	}
	if f.Available == "true" {
		query += ` AND is_available = TRUE`
		countQuery += ` AND is_available = TRUE`
	} else if f.Available == "false" {
		query += ` AND is_available = FALSE`
		countQuery += ` AND is_available = FALSE`
	}

	var total int
	r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)

	query += fmt.Sprintf(` ORDER BY slot_date, start_time LIMIT $%d OFFSET $%d`, idx, idx+1)
	args = append(args, f.Limit, f.Offset)

	var slots []TimeSlot
	err := r.db.SelectContext(ctx, &slots, query, args...)
	return slots, total, err
}

// Create создаёт новый слот.
func (r *SlotRepository) Create(ctx context.Context, p CreateSlotParams) (*TimeSlot, error) {
	var s TimeSlot
	err := r.db.QueryRowxContext(ctx, `
		INSERT INTO time_slots (tutor_id, subject_id, class_group_id, slot_date, start_time, end_time)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING *`,
		p.TutorID, p.SubjectID, p.ClassGroupID, p.SlotDate, p.StartTime, p.EndTime,
	).StructScan(&s)
	return &s, err
}

// Update обновляет слот по ID.
func (r *SlotRepository) Update(ctx context.Context, id int, p CreateSlotParams) (*TimeSlot, error) {
	var s TimeSlot
	err := r.db.QueryRowxContext(ctx, `
		UPDATE time_slots
		SET tutor_id = $1, subject_id = $2, class_group_id = $3,
		    slot_date = $4, start_time = $5, end_time = $6
		WHERE id = $7
		RETURNING *`,
		p.TutorID, p.SubjectID, p.ClassGroupID, p.SlotDate, p.StartTime, p.EndTime, id,
	).StructScan(&s)
	return &s, err
}

// Delete удаляет слот по ID.
func (r *SlotRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM time_slots WHERE id = $1`, id)
	return err
}
