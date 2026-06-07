DROP TABLE IF EXISTS subscription_changes;

ALTER TABLE employee_subscriptions
    DROP COLUMN IF EXISTS cancellation_reason,
    DROP COLUMN IF EXISTS cancelled_at,
    DROP COLUMN IF EXISTS auto_renew;
