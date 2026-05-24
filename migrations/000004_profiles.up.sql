CREATE TABLE customer_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE employee_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_customer_profiles_user_id ON customer_profiles (user_id);
CREATE INDEX idx_employee_profiles_user_id ON employee_profiles (user_id);

INSERT INTO customer_profiles (id, user_id, created_at, updated_at)
SELECT gen_random_uuid(), id, created_at, updated_at
FROM users
WHERE role = 'customer'
ON CONFLICT (user_id) DO NOTHING;

INSERT INTO employee_profiles (id, user_id, created_at, updated_at)
SELECT gen_random_uuid(), id, created_at, updated_at
FROM users
WHERE role = 'employee'
ON CONFLICT (user_id) DO NOTHING;
