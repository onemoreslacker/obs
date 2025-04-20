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
    tags TEXT[] NOT NULL,
    filters TEXT[] NOT NULL,
    is_activity_recorded BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
-- rollback DROP TABLE IF EXISTS links;

-- changeset onemoreslacker:00-schema-index-links-updated-at
CREATE INDEX idx_links_updated_at ON links (updated_at);
-- rollback DROP INDEX IF EXISTS idx_links_updated_at;

-- changeset onemoreslacker:00-schema-index-links-url
CREATE INDEX idx_links_url ON links USING HASH (url);
-- rollback DROP INDEX IF EXISTS idx_links_url;

-- changeset onemoreslacker:00-schema-tracking-links-table
CREATE TABLE IF NOT EXISTS tracking_links (
    chat_id BIGINT REFERENCES chats(id) ON DELETE CASCADE,
    link_id BIGINT REFERENCES links(id) ON DELETE CASCADE,
    PRIMARY KEY (chat_id, link_id)
);
-- rollback DROP TABLE IF EXISTS tracking_links;

-- changeset onemoreslacker:00-schema-index-tracking-links-table-chat-id
CREATE INDEX idx_tracking_links_chat ON tracking_links (chat_id);
-- rollback DROP INDEX IF EXISTS idx_tracking_links_chat;

-- changeset onemoreslacker:00-schema-index-tracking-links-table-link-id
CREATE INDEX idx_tracking_links_link ON tracking_links (link_id);
-- rollback DROP INDEX IF EXISTS idx_tracking_links_link;

