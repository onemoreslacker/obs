-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tags (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    tag TEXT NOT NULL,
    link_id BIGINT REFERENCES links (id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tags;
-- +goose StatementEnd
