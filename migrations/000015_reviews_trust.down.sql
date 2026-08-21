DROP TABLE IF EXISTS badges;
DROP TABLE IF EXISTS reports;
DROP TABLE IF EXISTS review_replies;
DROP TABLE IF EXISTS reviews;

ALTER TABLE employee_profiles
    DROP COLUMN IF EXISTS average_rating,
    DROP COLUMN IF EXISTS total_reviews;
