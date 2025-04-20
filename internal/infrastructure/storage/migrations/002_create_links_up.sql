-- +goose Up
CREATE TABLE IF NOT EXISTS links (
    id BIGINT PRIMARY KEY,
    url TEXT NOT NULL,
    tags TEXT[] NOT NULL,
    filters TEXT[] NOT NULL,
    is_activity_recorded BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_links_updated_at ON links (updated_at);
CREATE INDEX idx_links_url ON links USING HASH (url);