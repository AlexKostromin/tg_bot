-- +goose Up

-- При создании брони — блокируем ВСЕ слоты на это же время у этого репетитора
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION mark_slot_unavailable()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE time_slots SET is_available = FALSE
    WHERE (tutor_id, slot_date, start_time) = (
        SELECT tutor_id, slot_date, start_time
        FROM time_slots WHERE id = NEW.slot_id
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- При отмене брони — разблокируем слоты, только если нет других активных броней на это время
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION release_slot_on_cancel()
RETURNS TRIGGER AS $$
DECLARE
    v_tutor_id   INT;
    v_slot_date  DATE;
    v_start_time TIME;
BEGIN
    IF NEW.status = 'cancelled' AND OLD.status != 'cancelled' THEN
        SELECT tutor_id, slot_date, start_time
        INTO v_tutor_id, v_slot_date, v_start_time
        FROM time_slots WHERE id = NEW.slot_id;

        -- Разблокируем только если нет других активных броней на это время
        IF NOT EXISTS (
            SELECT 1 FROM bookings b
            JOIN time_slots ts ON ts.id = b.slot_id
            WHERE ts.tutor_id = v_tutor_id
              AND ts.slot_date = v_slot_date
              AND ts.start_time = v_start_time
              AND b.status IN ('pending', 'confirmed')
              AND b.id != NEW.id
        ) THEN
            UPDATE time_slots SET is_available = TRUE
            WHERE tutor_id = v_tutor_id
              AND slot_date = v_slot_date
              AND start_time = v_start_time;
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- Заблокируем слоты для уже существующих активных броней
UPDATE time_slots ts SET is_available = FALSE
WHERE EXISTS (
    SELECT 1 FROM bookings b
    JOIN time_slots booked ON booked.id = b.slot_id
    WHERE booked.tutor_id = ts.tutor_id
      AND booked.slot_date = ts.slot_date
      AND booked.start_time = ts.start_time
      AND b.status IN ('pending', 'confirmed')
);

-- +goose Down

-- Вернуть старые триггерные функции (блокировка только одного слота)
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION mark_slot_unavailable()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE time_slots SET is_available = FALSE WHERE id = NEW.slot_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

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
