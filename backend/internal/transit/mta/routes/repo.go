package routes

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type RouteRepo struct {
	db *pgxpool.Pool
}

func NewRouteRepo(db *pgxpool.Pool) *RouteRepo {
	return &RouteRepo{db: db}
}

func (r *RouteRepo) GetAll() {
}

func (r *RouteRepo) GetRoute() {
}
