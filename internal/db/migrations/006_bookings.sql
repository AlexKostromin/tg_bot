-- migrations/006_bookings.sql
-- +goose Up
CREATE TABLE bookings (
                          id         SERIAL PRIMARY KEY,
                          user_id    INTEGER REFERENCES users(id)      ON DELETE CASCADE,
                          slot_id    INTEGER REFERENCES time_slots(id) ON DELETE RESTRICT,
    -- RESTRICT: нельзя удалить слот, если на него есть бронь.
                          status     VARCHAR(20) DEFAULT 'pending'
                              CHECK (status IN ('pending', 'confirmed', 'cancelled', 'completed')),
                          comment    TEXT,
                          booked_at  TIMESTAMPTZ DEFAULT NOW(),
                          updated_at TIMESTAMPTZ DEFAULT NOW(),
                          CONSTRAINT unique_slot_booking UNIQUE (slot_id)
    -- Один слот — одна бронь. Второй конкурентный запрос
    -- получит UNIQUE VIOLATION прямо на уровне базы.
);
CREATE INDEX idx_bookings_user   ON bookings(user_id);
CREATE INDEX idx_bookings_status ON bookings(status);
-- +goose Down
DROP TABLE bookings;