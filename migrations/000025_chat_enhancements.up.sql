ALTER TABLE chat_messages
    ADD COLUMN attachment_url TEXT,
    ADD COLUMN content_type VARCHAR(100);
