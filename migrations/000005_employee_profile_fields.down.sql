DROP INDEX IF EXISTS idx_employee_profiles_verification_status;

ALTER TABLE employee_profiles
    DROP COLUMN IF EXISTS verification_status,
    DROP COLUMN IF EXISTS skills,
    DROP COLUMN IF EXISTS languages,
    DROP COLUMN IF EXISTS service_area_radius_km,
    DROP COLUMN IF EXISTS longitude,
    DROP COLUMN IF EXISTS latitude,
    DROP COLUMN IF EXISTS location_text,
    DROP COLUMN IF EXISTS profile_photo_url,
    DROP COLUMN IF EXISTS experience_years,
    DROP COLUMN IF EXISTS bio,
    DROP COLUMN IF EXISTS phone,
    DROP COLUMN IF EXISTS display_name;
