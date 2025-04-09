--liquibase formatted sql

-- changeset onemoreslacker:00-schema-users-table
CREATE TABLE IF NOT EXISTS chats (
    id BIGINT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
-- rollback DROP TABLE IF EXISTS chats;

-- changeset onemoreslacker:00-schema-links-table
CREATE TABLE IF NOT EXISTS links (
    id BIGINT PRIMARY KEY,
    url TEXT NOT NULL,
    tags TEXT ARRAY NOT NULL,
    filters TEXT ARRAY NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
-- rollback DROP TABLE IF EXISTS links;

-- changeset onemoreslacker:00-schema-tracking-links-table
CREATE TABLE IF NOT EXISTS tracking_links (
    chatID BIGINT REFERENCES chats(id),
    linkID BIGINT REFERENCES links(id) ON DELETE CASCADE,
    PRIMARY KEY (chatID, linkID)
);
-- rollback DROP TABLE IF EXISTS tracking_links;

-- changeset onemoreslacker:00-schema-index-links-table
CREATE INDEX IF NOT EXISTS idx_links_url ON links USING HASH(url);
-- rollback DROP INDEX IF EXISTS idx_links_url;