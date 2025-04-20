-- +goose Up
CREATE TABLE IF NOT EXISTS tracking_links (
    chat_id BIGINT REFERENCES chats(id) ON DELETE CASCADE,
    link_id BIGINT REFERENCES links(id) ON DELETE CASCADE,
    PRIMARY KEY (chat_id, link_id)
);

CREATE INDEX IF NOT EXISTS idx_tracking_links_chat ON tracking_links (chat_id);
CREATE INDEX IF NOT EXISTS idx_tracking_links_link ON tracking_links (link_id);