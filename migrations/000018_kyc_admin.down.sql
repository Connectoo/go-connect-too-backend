DROP INDEX IF EXISTS idx_employee_kyc_reviewed_by;

ALTER TABLE employee_kyc
    DROP COLUMN IF EXISTS reviewed_by,
    DROP COLUMN IF EXISTS reviewed_at;
