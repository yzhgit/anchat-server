-- Create messages table (monthly sharding strategy, this table is the template)
CREATE TABLE IF NOT EXISTS messages (
    id BIGSERIAL PRIMARY KEY,
    message_id VARCHAR(64) NOT NULL UNIQUE,
    conversation_id VARCHAR(64) NOT NULL,
    conversation_type SMALLINT NOT NULL,  -- 1-single/2-group
    sender_id VARCHAR(36) NOT NULL,
    content_type SMALLINT NOT NULL,  -- 1-text/2-image/3-video/4-audio/5-file/6-location/7-card
    content JSONB NOT NULL,
    sequence BIGINT NOT NULL,  -- Incrementing sequence number within conversation
    reply_to VARCHAR(64),  -- Message ID being replied to
    at_users TEXT[],  -- List of user IDs being @ed
    status SMALLINT DEFAULT 0,  -- 0-normal, 1-recalled, 2-deleted
    burn_after_reading_seconds INT NOT NULL DEFAULT 0,  -- Burn-after-reading duration in seconds, 0 means disabled
    auto_delete_expire_time TIMESTAMPTZ,  -- Expiration time calculated from auto-delete policy
    burn_after_reading_expire_time TIMESTAMPTZ,  -- Expiration time calculated from burn-after-reading policy
    expire_time TIMESTAMPTZ,  -- Message expiration time, NULL means never expires
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_conversation_sequence UNIQUE (conversation_id, sequence)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_messages_conversation_sequence ON messages(conversation_id, sequence DESC);
CREATE INDEX IF NOT EXISTS idx_messages_sender_time ON messages(sender_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_messages_message_id ON messages(message_id);
CREATE INDEX IF NOT EXISTS idx_messages_status ON messages(status) WHERE status = 0;
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_messages_expire_time ON messages(expire_time) WHERE expire_time IS NOT NULL;

-- Create read receipts table
CREATE TABLE IF NOT EXISTS message_read_receipts (
    id BIGSERIAL PRIMARY KEY,
    conversation_id VARCHAR(64) NOT NULL,
    conversation_type SMALLINT NOT NULL,  -- 1-single/2-group
    user_id VARCHAR(36) NOT NULL,
    last_read_seq BIGINT NOT NULL,
    last_read_message_id VARCHAR(64),
    read_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_conversation_user UNIQUE (conversation_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_read_receipts_conversation ON message_read_receipts(conversation_id);
CREATE INDEX IF NOT EXISTS idx_read_receipts_user ON message_read_receipts(user_id);

-- Create conversation sequence table (for generating incrementing sequence numbers)
CREATE TABLE IF NOT EXISTS conversation_sequences (
    id BIGSERIAL PRIMARY KEY,
    conversation_id VARCHAR(64) NOT NULL UNIQUE,
    current_seq BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_conversation_sequences_conversation ON conversation_sequences(conversation_id);

-- Create message references table (for recording reference relationships between messages)
CREATE TABLE IF NOT EXISTS message_references (
    id BIGSERIAL PRIMARY KEY,
    message_id VARCHAR(64) NOT NULL,
    reply_to_message_id VARCHAR(64) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_message_reply UNIQUE (message_id, reply_to_message_id)
);

CREATE INDEX IF NOT EXISTS idx_message_references_message ON message_references(message_id);
CREATE INDEX IF NOT EXISTS idx_message_references_reply_to ON message_references(reply_to_message_id);

-- Conversations table
CREATE TABLE conversations (
    conversation_id      VARCHAR(100) PRIMARY KEY,
    conversation_type    SMALLINT     NOT NULL,                -- 1-single/2-group/3-system
    user_id         VARCHAR(100) NOT NULL,
    target_id       VARCHAR(100) NOT NULL,                -- For private chat: peer user ID, for group chat: group ID
    last_message_id VARCHAR(100),
    last_message_content TEXT,
    last_message_time    TIMESTAMPTZ,
    unread_count    INT          NOT NULL DEFAULT 0,
    is_pinned       BOOLEAN      NOT NULL DEFAULT FALSE,
    is_muted        BOOLEAN      NOT NULL DEFAULT FALSE,
    pin_time        TIMESTAMPTZ,
    burn_after_reading INT       NOT NULL DEFAULT 0,       -- Burn-after-reading duration in seconds, 0 means disabled
    auto_delete_duration INT     NOT NULL DEFAULT 0,       -- Auto-delete duration in seconds, 0 means disabled
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX uk_conversation_user_target ON conversations (user_id, conversation_type, target_id);
CREATE INDEX IF NOT EXISTS idx_conversations_user_id      ON conversations (user_id);
CREATE INDEX IF NOT EXISTS idx_conversations_updated_at   ON conversations (updated_at);

-- Message send idempotency table
CREATE TABLE IF NOT EXISTS message_send_idempotency (
    id BIGSERIAL PRIMARY KEY,
    sender_id VARCHAR(36) NOT NULL,
    conversation_id VARCHAR(64) NOT NULL,
    local_id VARCHAR(128) NOT NULL,
    message_id VARCHAR(64) NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_sender_conversation_local UNIQUE (sender_id, conversation_id, local_id)
);

CREATE INDEX IF NOT EXISTS idx_message_idempotency_message_id ON message_send_idempotency(message_id);
