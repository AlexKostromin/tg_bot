-- +goose Up

CREATE TABLE time_slots (
    id SERIAL PRIMARY KEY,
    tutor_id INTEGER REFERENCES,
    subject_id INTEGER REFERENCES subjects(id),
    class_group_id INTEGER REFERENCES class_group(id),
    slot_date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    is_available BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT chk_time_order CHECK ( end_time > start_time )
);

CREATE INDEX idx_slots_date ON time_slots(slot_date);
CREATE INDEX idx_slots_group ON time_slots(class_group_id);
CREATE INDEX idx_slots_available ON time_slots(is_available) WHERE is_available = TRUE;


-- +goose Down
DROP TABLE time_slots;