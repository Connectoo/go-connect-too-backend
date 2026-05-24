ALTER TABLE users DROP CONSTRAINT users_email_role_unique;
ALTER TABLE users DROP CONSTRAINT users_phone_role_unique;

ALTER TABLE users ADD CONSTRAINT users_email_unique UNIQUE (email);
ALTER TABLE users ADD CONSTRAINT users_phone_unique UNIQUE (phone);
