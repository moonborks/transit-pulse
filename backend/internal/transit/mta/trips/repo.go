package trips

import "github.com/jackc/pgx/v5/pgxpool"

type TripRepo struct {
	db *pgxpool.Pool
}

func NewTripRepo(db *pgxpool.Pool) *TripRepo {
	return &TripRepo{db: db}
}
