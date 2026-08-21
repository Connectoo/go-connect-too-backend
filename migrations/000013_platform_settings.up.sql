CREATE TABLE platform_settings (
    key VARCHAR(100) PRIMARY KEY,
    value JSONB NOT NULL DEFAULT '{}',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO platform_settings (key, value)
VALUES (
    'general',
    '{"site_name":"Go Connect","support_email":"support@example.com","maintenance_mode":false}'::jsonb
);
