-- +goose Up
CREATE UNIQUE INDEX IF NOT EXISTS idx_slots_unique
ON time_slots(tutor_id, subject_id, class_group_id, slot_date, start_time);

-- +goose Down
DROP INDEX IF EXISTS idx_slots_unique;
