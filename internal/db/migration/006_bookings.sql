-- +goose Up

CREATE TABLE bookings (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    clot_id INTEGER REFERENCES time_slots(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'pending' CHECK ( status IN ('pending', 'confirmed', 'cancelled', 'completed')),
    comment TEXT,
    booked_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT unique_slot_booking UNIQUE (slot_id)
);

CREATE INDEX idx_bookings_user ON bookings(user_id);
CREATE INDEX idx_bookings_status ON bookings(status);


-- +goose Down

DROP TABLE bookings;