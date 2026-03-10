-- migrations/009_fix_unique_booking_constraint.sql
-- +goose Up
-- Удаляем старое ограничение
ALTER TABLE bookings DROP CONSTRAINT IF EXISTS unique_slot_booking;

-- Создаём частичный уникальный индекс (не считает отменённые брони)
CREATE UNIQUE INDEX unique_slot_booking
ON bookings(slot_id)
WHERE status != 'cancelled';

-- +goose Down
-- Удаляем индекс
DROP INDEX IF EXISTS unique_slot_booking;
