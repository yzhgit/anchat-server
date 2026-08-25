-- Create groups table
CREATE TABLE IF NOT EXISTS groups (
    id BIGSERIAL PRIMARY KEY,
    group_id VARCHAR(36) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    avatar VARCHAR(255),
    announcement TEXT,
    owner_id VARCHAR(36) NOT NULL,
    member_count INT DEFAULT 0,
    max_members INT DEFAULT 500,
    is_muted BOOLEAN DEFAULT FALSE,
    description TEXT,
    status SMALLINT DEFAULT 1,  -- 0-dissolved, 1-normal
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_groups_owner_id ON groups(owner_id);
CREATE INDEX IF NOT EXISTS idx_groups_status ON groups(status);
CREATE INDEX IF NOT EXISTS idx_groups_created_at ON groups(created_at);

-- Create group members table
CREATE TABLE IF NOT EXISTS group_members (
    id BIGSERIAL PRIMARY KEY,
    group_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    group_nickname VARCHAR(50),
    group_remark VARCHAR(20),           -- Custom remark name for this group, visible only to the user
    role SMALLINT NOT NULL DEFAULT 3,  -- 1-owner, 2-admin, 3-member
    muted_until TIMESTAMP,
    joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_group_user UNIQUE (group_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_group_members_user_id ON group_members(user_id);
CREATE INDEX IF NOT EXISTS idx_group_members_role ON group_members(role);
CREATE INDEX IF NOT EXISTS idx_group_members_joined_at ON group_members(joined_at);

-- Create group settings table
CREATE TABLE IF NOT EXISTS group_settings (
    group_id VARCHAR(36) PRIMARY KEY,
    join_verify BOOLEAN DEFAULT TRUE,
    allow_member_invite BOOLEAN DEFAULT TRUE,
    allow_view_history BOOLEAN DEFAULT TRUE,
    allow_add_friend BOOLEAN DEFAULT TRUE,
    allow_member_modify BOOLEAN DEFAULT FALSE,
    show_member_nickname BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create group join requests table
CREATE TABLE IF NOT EXISTS group_join_requests (
    id BIGSERIAL PRIMARY KEY,
    group_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    inviter_id VARCHAR(36),  -- Inviter ID (NULL means active application)
    message VARCHAR(200),
    status SMALLINT NOT NULL DEFAULT 1,  -- 1-pending, 2-accepted, 3-rejected
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_group_join_requests_group_id ON group_join_requests(group_id);
CREATE INDEX IF NOT EXISTS idx_group_join_requests_user_id ON group_join_requests(user_id);
CREATE INDEX IF NOT EXISTS idx_group_join_requests_status ON group_join_requests(status);

-- Create group pinned messages table
CREATE TABLE IF NOT EXISTS group_pinned_messages (
    id BIGSERIAL PRIMARY KEY,
    group_id VARCHAR(36) NOT NULL,
    message_id VARCHAR(64) NOT NULL,
    message_seq BIGINT,
    pinned_by VARCHAR(36) NOT NULL,
    content TEXT,
    content_type SMALLINT NOT NULL DEFAULT 1, -- 1-text/2-image/3-video/4-audio/5-file/6-location/7-card
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_group_pinned_message UNIQUE (group_id, message_id)
);

CREATE INDEX IF NOT EXISTS idx_group_pinned_messages_group_id ON group_pinned_messages(group_id);
CREATE INDEX IF NOT EXISTS idx_group_pinned_messages_group_updated_at ON group_pinned_messages(group_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_group_pinned_messages_content_type ON group_pinned_messages(content_type);

-- Create group qrcode table
CREATE TABLE IF NOT EXISTS group_qrcodes (
    id         BIGSERIAL    PRIMARY KEY,
    group_id   VARCHAR(36)  NOT NULL,
    token      VARCHAR(64)  NOT NULL UNIQUE,
    created_by VARCHAR(36)  NOT NULL,
    expire_at  TIMESTAMP    NOT NULL,
    is_active  BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_group_qrcodes_group_id ON group_qrcodes(group_id);
CREATE INDEX IF NOT EXISTS idx_group_qrcodes_token    ON group_qrcodes(token);
