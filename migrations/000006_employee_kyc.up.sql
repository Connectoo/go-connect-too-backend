CREATE TABLE employee_kyc (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL UNIQUE REFERENCES employee_profiles (id) ON DELETE CASCADE,
    id_proof_url TEXT NOT NULL,
    address_proof_url TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'approved', 'rejected')),
    rejection_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_employee_kyc_employee_id ON employee_kyc (employee_id);
CREATE INDEX idx_employee_kyc_status ON employee_kyc (status);
