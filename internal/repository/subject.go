package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type Subject struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

type SubjectRepository struct {
	db *sqlx.DB
}

func NewSubjectRepository(db *sqlx.DB) *SubjectRepository {
	return &SubjectRepository{db: db}
}

// GetByGroupID возвращает предметы, доступные для группы класса.
// Запрос через связующую таблицу subject_groups.
func (r *SubjectRepository) GetByGroupID(ctx context.Context, groupID int) ([]Subject, error) {
	var subjects []Subject
	err := r.db.SelectContext(ctx, &subjects, `
        SELECT s.id, s.name
        FROM subjects s
        JOIN subject_groups sg ON sg.subject_id = s.id
        WHERE sg.class_group_id = $1
        ORDER BY s.name`,
		groupID,
	)
	return subjects, err
}

// GetAll возвращает все предметы.
func (r *SubjectRepository) GetAll(ctx context.Context) ([]Subject, error) {
	var subjects []Subject
	err := r.db.SelectContext(ctx, &subjects, `SELECT id, name FROM subjects ORDER BY name`)
	return subjects, err
}

// Create создаёт новый предмет.
func (r *SubjectRepository) Create(ctx context.Context, name string) (*Subject, error) {
	var s Subject
	err := r.db.QueryRowxContext(ctx,
		`INSERT INTO subjects (name) VALUES ($1) RETURNING id, name`, name).StructScan(&s)
	return &s, err
}

// Update обновляет предмет.
func (r *SubjectRepository) Update(ctx context.Context, id int, name string) (*Subject, error) {
	var s Subject
	err := r.db.QueryRowxContext(ctx,
		`UPDATE subjects SET name = $1 WHERE id = $2 RETURNING id, name`, name, id).StructScan(&s)
	return &s, err
}

// Delete удаляет предмет.
func (r *SubjectRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM subjects WHERE id = $1`, id)
	return err
}

// --- ClassGroups ---

type ClassGroup struct {
	ID   int    `db:"id"   json:"id"`
	Name string `db:"name" json:"name"`
}

// GetAllGroups возвращает все группы классов.
func (r *SubjectRepository) GetAllGroups(ctx context.Context) ([]ClassGroup, error) {
	var groups []ClassGroup
	err := r.db.SelectContext(ctx, &groups, `SELECT id, name FROM class_groups ORDER BY name`)
	return groups, err
}
