CREATE TABLE refunds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id UUID NOT NULL REFERENCES payments (id) ON DELETE RESTRICT,
    amount BIGINT NOT NULL CHECK (amount > 0),
    reason TEXT,
    status VARCHAR(30) NOT NULL DEFAULT 'pending',
    provider_refund_id VARCHAR(255),
    created_by UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT refunds_status_check CHECK (status IN ('pending', 'completed', 'failed'))
);

CREATE INDEX idx_refunds_payment_id ON refunds (payment_id);
CREATE INDEX idx_refunds_status ON refunds (status);
