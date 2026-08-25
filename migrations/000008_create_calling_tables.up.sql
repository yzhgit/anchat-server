-- Call sessions table
CREATE TABLE IF NOT EXISTS call_sessions (
    id           BIGSERIAL    PRIMARY KEY,
    call_id      VARCHAR(36)  NOT NULL UNIQUE,
    caller_id    VARCHAR(36)  NOT NULL,
    callee_id    VARCHAR(36)  NOT NULL,
    call_type    SMALLINT     NOT NULL DEFAULT 0,  -- 0-audio/1-video
    status       SMALLINT     NOT NULL DEFAULT 0,  -- 0-ringing/1-connected/2-ended/3-rejected/4-missed/5-cancelled
    room_name    VARCHAR(100) NOT NULL,
    started_at   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    connected_at TIMESTAMP,
    ended_at     TIMESTAMP,
    duration     INT          NOT NULL DEFAULT 0,  -- Call duration in seconds
    created_at   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_call_sessions_call_id   ON call_sessions (call_id);
CREATE INDEX IF NOT EXISTS idx_call_sessions_caller_id ON call_sessions (caller_id);
CREATE INDEX IF NOT EXISTS idx_call_sessions_callee_id ON call_sessions (callee_id);
CREATE INDEX IF NOT EXISTS idx_call_sessions_created_at ON call_sessions (created_at DESC);

-- Meeting rooms table
CREATE TABLE IF NOT EXISTS meeting_rooms (
    id               BIGSERIAL    PRIMARY KEY,
    room_id          VARCHAR(36)  NOT NULL UNIQUE,
    creator_id       VARCHAR(36)  NOT NULL,
    title            VARCHAR(200) NOT NULL,
    room_name        VARCHAR(100) NOT NULL UNIQUE,  -- LiveKit Room name
    password_hash    VARCHAR(200),                  -- Optional password hash
    max_participants INT          NOT NULL DEFAULT 0,
    status           SMALLINT     NOT NULL DEFAULT 0,  -- 0-active/1-ended
    started_at       TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ended_at         TIMESTAMP,
    created_at       TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_meeting_rooms_room_id    ON meeting_rooms (room_id);
CREATE INDEX IF NOT EXISTS idx_meeting_rooms_creator_id ON meeting_rooms (creator_id);
CREATE INDEX IF NOT EXISTS idx_meeting_rooms_status     ON meeting_rooms (status);
CREATE INDEX IF NOT EXISTS idx_meeting_rooms_created_at ON meeting_rooms (created_at DESC);
