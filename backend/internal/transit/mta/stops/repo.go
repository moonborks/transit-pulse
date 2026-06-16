package stops

import "github.com/jackc/pgx/v5/pgxpool"

type StopRepo struct {
	db *pgxpool.Pool
}

func NewStopRepo(db *pgxpool.Pool) *StopRepo {
	return &StopRepo{db: db}
}
