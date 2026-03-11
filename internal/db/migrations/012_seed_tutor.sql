-- +goose Up

-- Начальный репетитор; tg_chat_id заполняется через админку или API.
INSERT INTO tutors (full_name) VALUES ('Ольга')
ON CONFLICT DO NOTHING;

-- Привязываем ко всем предметам и группам
INSERT INTO tutor_subjects (tutor_id, subject_id)
SELECT t.id, s.id FROM tutors t CROSS JOIN subjects s
ON CONFLICT DO NOTHING;

INSERT INTO tutor_groups (tutor_id, class_group_id)
SELECT t.id, g.id FROM tutors t CROSS JOIN class_groups g
ON CONFLICT DO NOTHING;

-- +goose Down

DELETE FROM tutor_groups WHERE tutor_id = (SELECT id FROM tutors WHERE full_name = 'Ольга');
DELETE FROM tutor_subjects WHERE tutor_id = (SELECT id FROM tutors WHERE full_name = 'Ольга');
DELETE FROM tutors WHERE full_name = 'Ольга';
