ALTER TABLE users DROP CONSTRAINT users_email_unique;
ALTER TABLE users DROP CONSTRAINT users_phone_unique;

ALTER TABLE users ADD CONSTRAINT users_email_role_unique UNIQUE (email, role);
ALTER TABLE users ADD CONSTRAINT users_phone_role_unique UNIQUE (phone, role);
