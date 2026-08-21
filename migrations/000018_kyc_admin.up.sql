ALTER TABLE employee_kyc
    ADD COLUMN reviewed_by UUID REFERENCES users (id) ON DELETE SET NULL,
    ADD COLUMN reviewed_at TIMESTAMPTZ;

CREATE INDEX idx_employee_kyc_reviewed_by ON employee_kyc (reviewed_by);
