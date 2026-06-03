DROP INDEX IF EXISTS idx_bookings_source_booking_id;

ALTER TABLE bookings
    DROP COLUMN IF EXISTS source_booking_id;
