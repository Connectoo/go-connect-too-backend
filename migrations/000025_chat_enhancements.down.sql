ALTER TABLE chat_messages
    DROP COLUMN IF EXISTS content_type,
    DROP COLUMN IF EXISTS attachment_url;
