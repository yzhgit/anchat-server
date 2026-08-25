-- Create friendships table
CREATE TABLE IF NOT EXISTS friendships (
    id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    friend_id VARCHAR(36) NOT NULL,
    remark VARCHAR(50),
    status SMALLINT DEFAULT 1,  -- 0-deleted, 1-normal
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_user_friend UNIQUE (user_id, friend_id)
);

CREATE INDEX IF NOT EXISTS idx_friendships_user_id ON friendships(user_id) WHERE status = 1;
CREATE INDEX IF NOT EXISTS idx_friendships_friend_id ON friendships(friend_id) WHERE status = 1;
CREATE INDEX IF NOT EXISTS idx_friendships_updated_at ON friendships(updated_at);

-- Create friend requests table
CREATE TABLE IF NOT EXISTS friend_requests (
    id BIGSERIAL PRIMARY KEY,
    from_user_id VARCHAR(36) NOT NULL,
    to_user_id VARCHAR(36) NOT NULL,
    message VARCHAR(200),
    source SMALLINT DEFAULT 1,  -- 1-search, 2-qrcode, 3-group, 4-contacts
    status SMALLINT DEFAULT 1,  -- 1-pending, 2-accepted, 3-rejected, 4-expired
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_friend_requests_to_user ON friend_requests(to_user_id, status);
CREATE INDEX IF NOT EXISTS idx_friend_requests_from_user ON friend_requests(from_user_id);
CREATE INDEX IF NOT EXISTS idx_friend_requests_created_at ON friend_requests(created_at);

-- Create blacklist table
CREATE TABLE IF NOT EXISTS blacklists (
    id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    blocked_user_id VARCHAR(36) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_user_blocked UNIQUE (user_id, blocked_user_id)
);

CREATE INDEX IF NOT EXISTS idx_blacklists_user_id ON blacklists(user_id);
CREATE INDEX IF NOT EXISTS idx_blacklists_blocked_user ON blacklists(blocked_user_id);
