ALTER TABLE employee_subscriptions
    ADD COLUMN auto_renew BOOLEAN NOT NULL DEFAULT true,
    ADD COLUMN cancelled_at TIMESTAMPTZ,
    ADD COLUMN cancellation_reason TEXT;

CREATE TABLE subscription_changes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_id UUID NOT NULL REFERENCES employee_subscriptions (id) ON DELETE CASCADE,
    employee_id UUID NOT NULL REFERENCES employee_profiles (id) ON DELETE CASCADE,
    change_type VARCHAR(50) NOT NULL,
    old_plan_id UUID REFERENCES subscription_plans (id) ON DELETE SET NULL,
    new_plan_id UUID REFERENCES subscription_plans (id) ON DELETE SET NULL,
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_subscription_changes_subscription_id ON subscription_changes (subscription_id);
CREATE INDEX idx_subscription_changes_employee_id ON subscription_changes (employee_id);
