package repository

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

type Tutor struct {
	ID        int       `db:"id"         json:"id"`
	FullName  string    `db:"full_name"  json:"full_name"`
	TgChatID  *int64    `db:"tg_chat_id" json:"tg_chat_id,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type TutorRepository struct {
	db *sqlx.DB
}

func NewTutorRepository(db *sqlx.DB) *TutorRepository {
	return &TutorRepository{db: db}
}

// GetAll возвращает всех репетиторов. Используется в админ-панели
// для списка репетиторов и в выпадающих списках при создании слота.
func (r *TutorRepository) GetAll(ctx context.Context) ([]Tutor, error) {
	var tutors []Tutor
	err := r.db.SelectContext(ctx, &tutors,
		`SELECT id, full_name, tg_chat_id, created_at FROM tutors ORDER BY full_name`)
	return tutors, err
}

// GetByID возвращает одного репетитора. Используется при редактировании в админке.
func (r *TutorRepository) GetByID(ctx context.Context, id int) (*Tutor, error) {
	var t Tutor
	err := r.db.GetContext(ctx, &t,
		`SELECT id, full_name, tg_chat_id, created_at FROM tutors WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// Create добавляет нового репетитора. tgChatID может быть nil,
// если репетитор пока не привязан к Telegram-аккаунту.
func (r *TutorRepository) Create(ctx context.Context, fullName string, tgChatID *int64) (*Tutor, error) {
	var t Tutor
	err := r.db.QueryRowxContext(ctx,
		`INSERT INTO tutors (full_name, tg_chat_id) VALUES ($1, $2)
		 RETURNING id, full_name, tg_chat_id, created_at`,
		fullName, tgChatID,
	).StructScan(&t)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// Update обновляет ФИО и/или chat ID репетитора.
func (r *TutorRepository) Update(ctx context.Context, id int, fullName string, tgChatID *int64) (*Tutor, error) {
	var t Tutor
	err := r.db.QueryRowxContext(ctx,
		`UPDATE tutors SET full_name = $1, tg_chat_id = $2 WHERE id = $3
		 RETURNING id, full_name, tg_chat_id, created_at`,
		fullName, tgChatID, id,
	).StructScan(&t)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// Delete удаляет репетитора. CASCADE в миграции автоматически
// удалит записи из tutor_subjects и tutor_groups.
func (r *TutorRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM tutors WHERE id = $1`, id)
	return err
}

// SetSubjects заменяет все предметы репетитора за одну транзакцию.
func (r *TutorRepository) SetSubjects(ctx context.Context, tutorID int, subjectIDs []int) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `DELETE FROM tutor_subjects WHERE tutor_id = $1`, tutorID)
	if err != nil {
		return err
	}
	for _, sid := range subjectIDs {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO tutor_subjects (tutor_id, subject_id) VALUES ($1, $2)`, tutorID, sid)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// SetGroups заменяет все группы классов репетитора за одну транзакцию.
func (r *TutorRepository) SetGroups(ctx context.Context, tutorID int, groupIDs []int) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `DELETE FROM tutor_groups WHERE tutor_id = $1`, tutorID)
	if err != nil {
		return err
	}
	for _, gid := range groupIDs {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO tutor_groups (tutor_id, class_group_id) VALUES ($1, $2)`, tutorID, gid)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// GetSubjectIDs возвращает ID предметов, привязанных к репетитору.
func (r *TutorRepository) GetSubjectIDs(ctx context.Context, tutorID int) ([]int, error) {
	var ids []int
	err := r.db.SelectContext(ctx, &ids,
		`SELECT subject_id FROM tutor_subjects WHERE tutor_id = $1`, tutorID)
	return ids, err
}

// GetGroupIDs возвращает ID групп классов, привязанных к репетитору.
func (r *TutorRepository) GetGroupIDs(ctx context.Context, tutorID int) ([]int, error) {
	var ids []int
	err := r.db.SelectContext(ctx, &ids,
		`SELECT class_group_id FROM tutor_groups WHERE tutor_id = $1`, tutorID)
	return ids, err
}
