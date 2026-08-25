-- Admin users table
CREATE TABLE IF NOT EXISTS admin_users (
    id            VARCHAR(36) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    username      VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email         VARCHAR(100),
    role          SMALLINT NOT NULL DEFAULT 2,          -- 1=superadmin, 2=admin, 3=readonly
    status        SMALLINT NOT NULL DEFAULT 1,          -- 1=active, 0=disabled
    last_login_at TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Audit logs table
CREATE TABLE IF NOT EXISTS audit_logs (
    id            BIGSERIAL PRIMARY KEY,
    admin_id      VARCHAR(36),
    action        VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50),
    resource_id   VARCHAR(100),
    details       JSONB,
    ip_address    VARCHAR(50),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_admin_id   ON audit_logs(admin_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action     ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);

-- System configs table
CREATE TABLE IF NOT EXISTS system_configs (
    key         VARCHAR(100) PRIMARY KEY,
    value       TEXT NOT NULL DEFAULT '',
    description VARCHAR(255),
    updated_by  VARCHAR(36),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Initial admin account (username: admin, password: Admin@123456)
INSERT INTO admin_users (username, password_hash, role)
VALUES ('admin', '$2a$10$AN7p9wUwBMIHWhFsw6h8ouO41ZtoK7XG9ddCZSqQH6pcTZ12v1tPu', 1)
ON CONFLICT (username) DO NOTHING;

-- Initial system configs
INSERT INTO system_configs (key, value, description) VALUES
    ('site.name',           'AnyChat', 'Site name'),
    ('site.description',    'AnyChat instant messaging system', 'Site description'),
    ('user.max_friends',    '1000', 'Maximum number of friends'),
    ('group.max_members',   '500', 'Maximum group members'),
    ('message.max_recall',  '120', 'Message recall time limit (seconds)')
ON CONFLICT (key) DO NOTHING;
