package trips

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TripRepo struct {
	db *pgxpool.Pool
}

func NewTripRepo(db *pgxpool.Pool) *TripRepo {
	return &TripRepo{db: db}
}

func (r *TripRepo) GetAll(ctx context.Context) ([]Trip, error) {
	stmt := `
		SELECT
			id, route_id, service_id, headsign, direction_id, shape_id
		FROM
			trips
	`

	rows, err := r.db.Query(ctx, stmt)
	if err != nil {
		slog.Error("query trips table", "err", err)
		return nil, err
	}

	trips := []Trip{}

	for rows.Next() {
		var trip Trip
		if err := rows.Scan(
			&trip.ID,
			&trip.RouteID,
			&trip.ServiceID,
			&trip.HeadSign,
			&trip.DirectionID,
			&trip.ShapeID,
		); err != nil {
			slog.Error("retrieving particular row", "err", err)
		}
		trips = append(trips, trip)
	}

	return trips, nil
}

func (r *TripRepo) GetTrip(ctx context.Context, id string) (Trip, error) {
	stmt := `
		SELECT
			id, route_id, service_id, headsign, direction_id, shape_id
		FROM
			trips
		WHERE
			id = $1
	`

	var trip Trip

	row := r.db.QueryRow(ctx, stmt, id)
	err := row.Scan(
		&trip.ID,
		&trip.RouteID,
		&trip.ServiceID,
		&trip.HeadSign,
		&trip.DirectionID,
		&trip.ShapeID,
	)
	if err != nil {
		slog.Error("retrieving row with specific id", "id", id, "err", err)
		return Trip{}, err
	}

	return trip, nil
}
