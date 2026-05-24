ALTER TABLE employee_profiles
    ADD COLUMN display_name VARCHAR(255),
    ADD COLUMN phone VARCHAR(30),
    ADD COLUMN bio TEXT,
    ADD COLUMN experience_years INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN profile_photo_url TEXT,
    ADD COLUMN location_text VARCHAR(500),
    ADD COLUMN latitude DOUBLE PRECISION,
    ADD COLUMN longitude DOUBLE PRECISION,
    ADD COLUMN service_area_radius_km DOUBLE PRECISION,
    ADD COLUMN languages TEXT[] NOT NULL DEFAULT '{}',
    ADD COLUMN skills TEXT[] NOT NULL DEFAULT '{}',
    ADD COLUMN verification_status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (verification_status IN ('pending', 'approved', 'rejected'));

CREATE INDEX idx_employee_profiles_verification_status ON employee_profiles (verification_status);
