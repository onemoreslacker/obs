-- +goose Up
-- +goose StatementBegin
BEGIN;

CREATE TABLE IF NOT EXISTS subs(
    chat_id BIGINT REFERENCES chats(id) ON DELETE CASCADE,
    link_id BIGINT REFERENCES links(id) ON DELETE CASCADE,
    PRIMARY KEY (chat_id, link_id)
    );

CREATE INDEX IF NOT EXISTS idx_subs_chat_link ON subs (chat_id, link_id);

END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
BEGIN;

DROP INDEX IF EXISTS idx_subs_chat_link;
DROP TABLE IF EXISTS subs;

END;
-- +goose StatementEnd
