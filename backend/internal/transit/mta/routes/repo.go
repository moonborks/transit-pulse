package routes

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RouteRepo struct {
	db *pgxpool.Pool
}

func NewRouteRepo(db *pgxpool.Pool) *RouteRepo {
	return &RouteRepo{db: db}
}

func (r *RouteRepo) GetAll(ctx context.Context) ([]Route, error) {
	stmt := `
		SELECT
			id, short_name, long_name, type, color
		FROM
			routes
	`

	rows, err := r.db.Query(ctx, stmt)
	if err != nil {
		slog.Error("query routes table", "err", err)
		return nil, err
	}
	defer rows.Close()

	routes := []Route{}

	for rows.Next() {
		var route Route
		if err := rows.Scan(
			&route.ID,
			&route.ShortName,
			&route.LongName,
			&route.Type,
			&route.Color,
		); err != nil {
			slog.Error("retrieving particular row", "err", err)
		}
		routes = append(routes, route)
	}

	return routes, nil
}

func (r *RouteRepo) GetRoute(ctx context.Context, id string) (Route, error) {
	stmt := `
		SELECT
			id, short_name, long_name, type, color
		FROM
			routes
		WHERE
			id = $1
	`

	var route Route

	row := r.db.QueryRow(ctx, stmt, id)
	err := row.Scan(&route.ID, &route.ShortName, &route.LongName, &route.Type, &route.Color)
	if err != nil {
		slog.Error("retrieving row with specific id", "id", id, "err", err)
		return Route{}, err
	}

	return route, nil
}
