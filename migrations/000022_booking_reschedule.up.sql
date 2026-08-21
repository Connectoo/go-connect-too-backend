ALTER TABLE bookings
    ADD COLUMN rescheduled_from_id UUID REFERENCES bookings (id) ON DELETE SET NULL;

CREATE INDEX idx_bookings_rescheduled_from_id ON bookings (rescheduled_from_id);
