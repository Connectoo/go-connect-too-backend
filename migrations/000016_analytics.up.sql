CREATE TABLE employee_profile_views (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employee_profiles (id) ON DELETE CASCADE,
    viewed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_employee_profile_views_employee_viewed_at
    ON employee_profile_views (employee_id, viewed_at);

CREATE INDEX idx_bookings_created_at ON bookings (created_at);
CREATE INDEX idx_bookings_employee_created_at ON bookings (employee_id, created_at);
CREATE INDEX idx_reviews_employee_created_at ON reviews (employee_id, created_at);
CREATE INDEX idx_payments_created_at ON payments (created_at);
CREATE INDEX idx_payments_status_created_at ON payments (status, created_at);
