CREATE TABLE subscription_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    price BIGINT NOT NULL CHECK (price >= 0),
    currency CHAR(3) NOT NULL DEFAULT 'INR',
    duration_days INTEGER NOT NULL CHECK (duration_days > 0),
    service_limit INTEGER NOT NULL CHECK (service_limit >= -1),
    is_featured_allowed BOOLEAN NOT NULL DEFAULT false,
    is_priority_allowed BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE employee_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employee_profiles (id) ON DELETE CASCADE,
    plan_id UUID NOT NULL REFERENCES subscription_plans (id) ON DELETE RESTRICT,
    status VARCHAR(30) NOT NULL,
    starts_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT employee_subscriptions_status_check CHECK (status IN ('pending', 'active', 'expired', 'cancelled')),
    CONSTRAINT employee_subscriptions_active_dates_check CHECK (
        status <> 'active' OR (starts_at IS NOT NULL AND expires_at IS NOT NULL AND expires_at > starts_at)
    )
);

CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employee_profiles (id) ON DELETE CASCADE,
    subscription_id UUID NOT NULL REFERENCES employee_subscriptions (id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    provider_order_id VARCHAR(255) NOT NULL,
    provider_payment_id VARCHAR(255),
    amount BIGINT NOT NULL CHECK (amount >= 0),
    currency CHAR(3) NOT NULL,
    status VARCHAR(30) NOT NULL,
    raw_response JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT payments_status_check CHECK (status IN ('pending', 'success', 'failed'))
);

CREATE TABLE payment_webhook_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider VARCHAR(50) NOT NULL,
    event_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (provider, event_id)
);

CREATE INDEX idx_subscription_plans_active ON subscription_plans (is_active);
CREATE INDEX idx_employee_subscriptions_employee_status ON employee_subscriptions (employee_id, status);
CREATE INDEX idx_employee_subscriptions_expires_at ON employee_subscriptions (expires_at);
CREATE INDEX idx_payments_employee_id ON payments (employee_id);
CREATE UNIQUE INDEX idx_payments_provider_order ON payments (provider, provider_order_id);
CREATE INDEX idx_payments_status ON payments (status);
CREATE INDEX idx_payment_webhook_events_provider_type ON payment_webhook_events (provider, event_type);

INSERT INTO subscription_plans (
    id, name, price, currency, duration_days, service_limit,
    is_featured_allowed, is_priority_allowed, is_active, created_at, updated_at
) VALUES
    ('00000000-0000-0000-0000-000000000601', 'Free Trial', 0, 'INR', 14, 1, false, false, true, NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000602', 'Starter', 49900, 'INR', 30, 3, false, false, true, NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000603', 'Professional', 149900, 'INR', 30, 10, false, true, true, NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000604', 'Premium', 299900, 'INR', 30, -1, true, true, true, NOW(), NOW())
ON CONFLICT (name) DO NOTHING;
