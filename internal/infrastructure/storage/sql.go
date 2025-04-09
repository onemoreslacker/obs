package storage

import "github.com/jackc/pgx/v5/pgxpool"

type LinksSQLService struct {
	pool *pgxpool.Pool
}

func NewLinksSQLService(pool *pgxpool.Pool) *LinksSQLService {
	return &LinksSQLService{
		pool: pool,
	}
}
