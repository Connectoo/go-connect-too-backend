CREATE TABLE bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customer_profiles (id) ON DELETE CASCADE,
    employee_id UUID NOT NULL REFERENCES employee_profiles (id) ON DELETE CASCADE,
    service_id UUID NOT NULL REFERENCES employee_services (id) ON DELETE RESTRICT,
    booking_date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    customer_notes TEXT,
    employee_notes TEXT,
    total_amount NUMERIC(12, 2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT bookings_time_order CHECK (start_time < end_time),
    CONSTRAINT bookings_status_check CHECK (
        status IN (
            'pending',
            'accepted',
            'rejected',
            'in_progress',
            'completed',
            'cancelled',
            'no_show'
        )
    )
);

CREATE INDEX idx_bookings_customer_id ON bookings (customer_id);
CREATE INDEX idx_bookings_employee_id ON bookings (employee_id);
CREATE INDEX idx_bookings_service_id ON bookings (service_id);
CREATE INDEX idx_bookings_employee_date ON bookings (employee_id, booking_date);
CREATE INDEX idx_bookings_employee_date_status ON bookings (employee_id, booking_date, status);

CREATE TABLE booking_status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id UUID NOT NULL REFERENCES bookings (id) ON DELETE CASCADE,
    old_status VARCHAR(20),
    new_status VARCHAR(20) NOT NULL,
    changed_by_user_id UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_booking_status_history_booking_id ON booking_status_history (booking_id);
