-- migrations/007_triggers.sql
-- +goose Up

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER trg_bookings_updated_at
    BEFORE UPDATE ON bookings
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION mark_slot_unavailable()
RETURNS TRIGGER AS $$
BEGIN
UPDATE time_slots SET is_available = FALSE WHERE id = NEW.slot_id;
RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER trg_booking_created
    AFTER INSERT ON bookings
    FOR EACH ROW EXECUTE FUNCTION mark_slot_unavailable();

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION release_slot_on_cancel()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'cancelled' AND OLD.status != 'cancelled' THEN
UPDATE time_slots SET is_available = TRUE WHERE id = NEW.slot_id;
END IF;
RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER trg_booking_cancelled
    AFTER UPDATE ON bookings
    FOR EACH ROW EXECUTE FUNCTION release_slot_on_cancel();

-- +goose Down
DROP TRIGGER IF EXISTS trg_booking_cancelled   ON bookings;
DROP TRIGGER IF EXISTS trg_booking_created     ON bookings;
DROP TRIGGER IF EXISTS trg_bookings_updated_at ON bookings;
DROP FUNCTION IF EXISTS release_slot_on_cancel;
DROP FUNCTION IF EXISTS mark_slot_unavailable;
DROP FUNCTION IF EXISTS set_updated_at;
