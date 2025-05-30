-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS filters (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    filter_value TEXT NOT NULL,
    link_id BIGINT REFERENCES links (id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS filters;
-- +goose StatementEnd
