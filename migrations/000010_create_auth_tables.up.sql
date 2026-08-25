-- Verification codes table
CREATE TABLE IF NOT EXISTS verification_codes (
    id              BIGSERIAL    PRIMARY KEY,
    code_id         VARCHAR(64)  NOT NULL UNIQUE,
    target          VARCHAR(128) NOT NULL,
    target_type     SMALLINT     NOT NULL,  -- 1-sms, 2-email
    code_hash       VARCHAR(128) NOT NULL,
    purpose         SMALLINT     NOT NULL,  -- 1-register,2-login,3-reset_password,4-bind_phone,5-change_phone,6-bind_email,7-change_email
    expires_at      TIMESTAMP    NOT NULL,
    verified_at     TIMESTAMP,
    status          SMALLINT     NOT NULL DEFAULT 1,  -- 1-pending,2-verified,3-expired,4-locked,5-cancelled
    send_ip         VARCHAR(64),
    send_device_id  VARCHAR(128),
    attempt_count   INT          NOT NULL DEFAULT 0,
    provider        VARCHAR(32),
    provider_message_id VARCHAR(128),
    created_at      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_verification_codes_target ON verification_codes (target, target_type, purpose);
CREATE INDEX IF NOT EXISTS idx_verification_codes_code_id ON verification_codes (code_id);
CREATE INDEX IF NOT EXISTS idx_verification_codes_expires_at ON verification_codes (expires_at);

-- Rate limits table
CREATE TABLE IF NOT EXISTS rate_limits (
    id              BIGSERIAL    PRIMARY KEY,
    target          VARCHAR(128) NOT NULL,
    target_type     SMALLINT     NOT NULL,  -- 1-sms, 2-email
    action          SMALLINT     NOT NULL,  -- 1-send_code
    count           INT          NOT NULL DEFAULT 1,
    window_start    TIMESTAMP    NOT NULL,
    window_end      TIMESTAMP    NOT NULL,
    created_at      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(target, target_type, action, window_start)
);

CREATE INDEX IF NOT EXISTS idx_rate_limits_target ON rate_limits (target, target_type, action);

-- Verification templates table
CREATE TABLE IF NOT EXISTS verification_templates (
    id              BIGSERIAL    PRIMARY KEY,
    purpose         SMALLINT     NOT NULL UNIQUE,
    name            VARCHAR(64)  NOT NULL,
    sms_template_id VARCHAR(128),
    sms_content     VARCHAR(512),
    email_subject   VARCHAR(128),
    email_content   TEXT,
    is_active       BOOLEAN      NOT NULL DEFAULT true,
    created_at      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Insert default templates
INSERT INTO verification_templates (purpose, name, sms_template_id, sms_content, email_subject, email_content) VALUES
(1, 'Register', 'SMS_123456', '【AnyChat】Your verification code is {code}, valid for 5 minutes. Do not share.', 'AnyChat Email Verification', '<!DOCTYPE html><html><head><meta charset="utf-8"></head><body><div style="max-width: 600px; margin: 0 auto; padding: 20px;"><h2 style="color: #333;">AnyChat Email Verification</h2><p>Hello,</p><p>Your email verification code is: <strong style="font-size: 24px; color: #1890ff;">{code}</strong></p><p>The code is valid for 5 minutes, please do not share with others.</p><p style="color: #999; font-size: 12px;">If this was not your operation, please ignore this email.</p></div></body></html>'),
(2, 'Login', 'SMS_123457', '【AnyChat】Your login verification code is {code}, valid for 5 minutes. Do not share.', 'AnyChat Login Verification', '<!DOCTYPE html><html><head><meta charset="utf-8"></head><body><div style="max-width: 600px; margin: 0 auto; padding: 20px;"><h2 style="color: #333;">AnyChat Login Verification</h2><p>Hello,</p><p>Your login verification code is: <strong style="font-size: 24px; color: #1890ff;">{code}</strong></p><p>The code is valid for 5 minutes, please do not share with others.</p></div></body></html>'),
(3, 'Reset Password', 'SMS_123458', '【AnyChat】You are resetting your password, verification code is {code}, valid for 5 minutes.', 'AnyChat Reset Password', '<!DOCTYPE html><html><head><meta charset="utf-8"></head><body><div style="max-width: 600px; margin: 0 auto; padding: 20px;"><h2 style="color: #333;">AnyChat Reset Password</h2><p>Hello,</p><p>You are resetting your password, verification code is: <strong style="font-size: 24px; color: #1890ff;">{code}</strong></p><p>The code is valid for 5 minutes, please do not share with others.</p></div></body></html>'),
(4, 'Bind Phone', 'SMS_123459', '【AnyChat】You are binding your phone number, verification code is {code}, valid for 5 minutes.', 'AnyChat Bind Phone', '<!DOCTYPE html><html><head><meta charset="utf-8"></head><body><div style="max-width: 600px; margin: 0 auto; padding: 20px;"><h2 style="color: #333;">AnyChat Bind Phone</h2><p>Hello,</p><p>You are binding your phone number, verification code is: <strong style="font-size: 24px; color: #1890ff;">{code}</strong></p><p>The code is valid for 5 minutes, please do not share with others.</p></div></body></html>'),
(5, 'Change Phone', 'SMS_123460', '【AnyChat】You are changing your phone number, verification code is {code}, valid for 5 minutes.', 'AnyChat Change Phone', '<!DOCTYPE html><html><head><meta charset="utf-8"></head><body><div style="max-width: 600px; margin: 0 auto; padding: 20px;"><h2 style="color: #333;">AnyChat Change Phone Number</h2><p>Hello,</p><p>You are changing your phone number, verification code is: <strong style="font-size: 24px; color: #1890ff;">{code}</strong></p><p>The code is valid for 5 minutes, please do not share with others.</p></div></body></html>'),
(6, 'Bind Email', 'SMS_123461', '【AnyChat】You are binding your email, verification code is {code}, valid for 5 minutes.', 'AnyChat Bind Email', '<!DOCTYPE html><html><head><meta charset="utf-8"></head><body><div style="max-width: 600px; margin: 0 auto; padding: 20px;"><h2 style="color: #333;">AnyChat Bind Email</h2><p>Hello,</p><p>You are binding your email, verification code is: <strong style="font-size: 24px; color: #1890ff;">{code}</strong></p><p>The code is valid for 5 minutes, please do not share with others.</p></div></body></html>'),
(7, 'Change Email', 'SMS_123462', '【AnyChat】You are changing your email, verification code is {code}, valid for 5 minutes.', 'AnyChat Change Email', '<!DOCTYPE html><html><head><meta charset="utf-8"></head><body><div style="max-width: 600px; margin: 0 auto; padding: 20px;"><h2 style="color: #333;">AnyChat Change Email</h2><p>Hello,</p><p>You are changing your email, verification code is: <strong style="font-size: 24px; color: #1890ff;">{code}</strong></p><p>The code is valid for 5 minutes, please do not share with others.</p></div></body></html>');
