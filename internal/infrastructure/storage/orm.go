package storage

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LinksOrmService struct {
	pool *pgxpool.Pool
	sb   sq.StatementBuilderType
}

func NewLinksOrmService(pool *pgxpool.Pool) LinksRepository {
	return &LinksOrmService{
		pool: pool,
		sb:   sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}
