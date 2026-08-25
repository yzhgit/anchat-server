-- Create push logs table
CREATE TABLE IF NOT EXISTS push_logs (
    id          BIGSERIAL PRIMARY KEY,
    user_id     VARCHAR(36)  NOT NULL,
    push_type   SMALLINT     NOT NULL DEFAULT 0,  -- 0-unspecified/1-message_new/2-message_mention/3-friend_request/4-group_invited/5-call_invite
    title       VARCHAR(200),
    content     VARCHAR(1000),
    target_count INT         NOT NULL DEFAULT 0,
    success_count INT        NOT NULL DEFAULT 0,
    failure_count INT        NOT NULL DEFAULT 0,
    jpush_msg_id VARCHAR(100),
    status      SMALLINT     NOT NULL DEFAULT 1,  -- 1-pending/2-sent/3-failed
    error_msg   TEXT,
    created_at  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_push_logs_user_id    ON push_logs (user_id);
CREATE INDEX IF NOT EXISTS idx_push_logs_created_at ON push_logs (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_push_logs_status     ON push_logs (status);
