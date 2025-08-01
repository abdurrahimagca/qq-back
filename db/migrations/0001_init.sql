--- up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE auth_provider AS ENUM ('email_otp', 'google_oauth');
CREATE TYPE privacy_level AS ENUM ('public', 'private', 'full_private');

CREATE TABLE IF NOT EXISTS auth (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    provider auth_provider NOT NULL,
    provider_id VARCHAR(255),
    is_suspended BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(provider, provider_id)
);

CREATE TABLE IF NOT EXISTS auth_otp_codes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    auth_id UUID NOT NULL REFERENCES auth(id) ON DELETE CASCADE,
    code TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP + INTERVAL '3 minutes'),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    privacy_level privacy_level NOT NULL DEFAULT 'public',
    auth_id UUID NOT NULL UNIQUE REFERENCES auth(id) ON DELETE CASCADE,
    username VARCHAR(512) NOT NULL UNIQUE,
    display_name VARCHAR(512) NOT NULL,
    avatar_url VARCHAR(512),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_auth_email ON auth(email);
CREATE INDEX idx_auth_provider ON auth(provider, provider_id);
CREATE INDEX idx_auth_otp_auth_id ON auth_otp_codes(auth_id);
CREATE INDEX idx_auth_otp_expires ON auth_otp_codes(expires_at);
CREATE INDEX idx_users_auth_id ON users(auth_id);
CREATE INDEX idx_users_username ON users(username);

--- down
DROP TABLE IF EXISTS auth_otp_codes;
DROP TABLE IF EXISTS auth;
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS auth_provider;