ALTER TABLE employee_profiles
    ADD COLUMN average_rating NUMERIC(3, 2),
    ADD COLUMN total_reviews INTEGER NOT NULL DEFAULT 0;

CREATE TABLE reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id UUID NOT NULL UNIQUE REFERENCES bookings (id) ON DELETE CASCADE,
    customer_id UUID NOT NULL REFERENCES customer_profiles (id) ON DELETE CASCADE,
    employee_id UUID NOT NULL REFERENCES employee_profiles (id) ON DELETE CASCADE,
    rating SMALLINT NOT NULL,
    comment TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT reviews_rating_check CHECK (rating >= 1 AND rating <= 5),
    CONSTRAINT reviews_status_check CHECK (status IN ('pending', 'approved', 'hidden'))
);

CREATE INDEX idx_reviews_employee_id ON reviews (employee_id);
CREATE INDEX idx_reviews_customer_id ON reviews (customer_id);
CREATE INDEX idx_reviews_status ON reviews (status);
CREATE INDEX idx_reviews_employee_status ON reviews (employee_id, status);

CREATE TABLE review_replies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    review_id UUID NOT NULL UNIQUE REFERENCES reviews (id) ON DELETE CASCADE,
    employee_id UUID NOT NULL REFERENCES employee_profiles (id) ON DELETE CASCADE,
    reply TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_review_replies_employee_id ON review_replies (employee_id);

CREATE TABLE reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reporter_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    reported_user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    booking_id UUID REFERENCES bookings (id) ON DELETE SET NULL,
    reason VARCHAR(100) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'open',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT reports_status_check CHECK (status IN ('open', 'resolved'))
);

CREATE INDEX idx_reports_reporter_id ON reports (reporter_id);
CREATE INDEX idx_reports_reported_user_id ON reports (reported_user_id);
CREATE INDEX idx_reports_status ON reports (status);

CREATE TABLE badges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employee_profiles (id) ON DELETE CASCADE,
    badge_type VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT badges_employee_type_unique UNIQUE (employee_id, badge_type)
);

CREATE INDEX idx_badges_employee_id ON badges (employee_id);
