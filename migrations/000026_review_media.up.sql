ALTER TABLE reviews
    ADD COLUMN review_images JSONB NOT NULL DEFAULT '[]';
