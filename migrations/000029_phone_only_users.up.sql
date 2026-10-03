-- Phone-only (OTP) registration: allow accounts created with only a phone number.
-- Make email and password_hash nullable so a mobile user can sign up with just a phone.
--
-- The existing unique constraints are intentionally kept:
--   users_email_role_unique UNIQUE(email, role)
--   users_phone_role_unique UNIQUE(phone, role)
-- Postgres treats NULLs as distinct in a UNIQUE index, so multiple phone-only rows
-- (NULL email) coexist without violating users_email_role_unique. Do NOT add any
-- constraint that rejects multiple NULL emails -- that is the desired behavior.
-- Phone uniqueness per role is still enforced by users_phone_role_unique.

ALTER TABLE users ALTER COLUMN email DROP NOT NULL;
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;
