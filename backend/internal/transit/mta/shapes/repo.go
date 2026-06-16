package shapes

import "github.com/jackc/pgx/v5/pgxpool"

type ShapeRepo struct {
	db *pgxpool.Pool
}

func NewShapeRepo(db *pgxpool.Pool) *ShapeRepo {
	return &ShapeRepo{db: db}
}
