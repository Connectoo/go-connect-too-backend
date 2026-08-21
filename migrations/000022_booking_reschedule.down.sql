DROP INDEX IF EXISTS idx_bookings_rescheduled_from_id;

ALTER TABLE bookings
    DROP COLUMN IF EXISTS rescheduled_from_id;
