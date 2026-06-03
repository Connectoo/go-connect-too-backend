CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    data JSONB NOT NULL DEFAULT '{}',
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_id ON notifications (user_id);
CREATE INDEX idx_notifications_user_id_created_at ON notifications (user_id, created_at DESC);

CREATE TABLE chat_conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customer_profiles (id) ON DELETE CASCADE,
    employee_id UUID NOT NULL REFERENCES employee_profiles (id) ON DELETE CASCADE,
    booking_id UUID REFERENCES bookings (id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_chat_conversations_booking_id ON chat_conversations (booking_id)
WHERE booking_id IS NOT NULL;

CREATE INDEX idx_chat_conversations_customer_id ON chat_conversations (customer_id);
CREATE INDEX idx_chat_conversations_employee_id ON chat_conversations (employee_id);

CREATE TABLE chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES chat_conversations (id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chat_messages_message_not_empty CHECK (length(trim(message)) > 0)
);

CREATE INDEX idx_chat_messages_conversation_id ON chat_messages (conversation_id, created_at);

CREATE TABLE device_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    platform VARCHAR(20) NOT NULL,
    token VARCHAR(512) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT device_tokens_platform_check CHECK (platform IN ('ios', 'android', 'web')),
    UNIQUE (user_id, token)
);

CREATE INDEX idx_device_tokens_user_id ON device_tokens (user_id);
