ALTER TABLE bookings
    ADD COLUMN source_booking_id UUID REFERENCES bookings (id) ON DELETE SET NULL;

CREATE INDEX idx_bookings_source_booking_id ON bookings (source_booking_id);
