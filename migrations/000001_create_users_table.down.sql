-- Drop users table
-- First drop child tables that depend on users, then drop users to avoid foreign key errors
DROP TABLE IF EXISTS user_profiles;
DROP TABLE IF EXISTS user_settings;
DROP TABLE IF EXISTS user_qrcodes;
DROP TABLE IF EXISTS user_push_tokens;
DROP TABLE IF EXISTS user_sessions;
DROP TABLE IF EXISTS user_devices;
DROP TABLE IF EXISTS users;