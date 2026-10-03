-- Rollback phone-only registration: restore NOT NULL on email and password_hash.
--
-- WARNING: any phone-only rows created while this migration was applied have NULL
-- email / password_hash and MUST be backfilled before this down migration can run,
-- otherwise SET NOT NULL fails on the existing NULLs. For a dev/local rollback where
-- no phone-only accounts exist this is a straightforward re-add.

ALTER TABLE users ALTER COLUMN email SET NOT NULL;
ALTER TABLE users ALTER COLUMN password_hash SET NOT NULL;
