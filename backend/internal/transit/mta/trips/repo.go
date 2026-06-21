package trips

import (
	"context"
	"log/slog"
	"time"

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
			id
			, day_of_week
			, short_trip_id
			, route_id
			, service_id
			, headsign
			, direction_id
			, shape_id
		FROM
			trips
	`

	rows, err := r.db.Query(ctx, stmt)
	if err != nil {
		slog.Error("query trips table", "err", err)
		return nil, err
	}
	defer rows.Close()

	trips := []Trip{}

	for rows.Next() {
		var trip Trip
		if err := rows.Scan(
			&trip.ID,
			&trip.DayOfWeek,
			&trip.ShortTripID,
			&trip.RouteID,
			&trip.ServiceID,
			&trip.Headsign,
			&trip.DirectionID,
			&trip.ShapeID,
		); err != nil {
			slog.Error("retrieving particular row", "err", err)
			return []Trip{}, err
		}
		trips = append(trips, trip)
	}

	return trips, nil
}

func (r *TripRepo) GetTrip(ctx context.Context, id string) (Trip, error) {
	stmt := `
		SELECT
			id
			, day_of_week
			, short_trip_id
			, route_id
			, service_id
			, headsign
			, direction_id
			, shape_id
		FROM
			trips
		WHERE
			id = $1
	`

	var trip Trip

	row := r.db.QueryRow(ctx, stmt, id)
	err := row.Scan(
		&trip.ID,
		&trip.ShortTripID,
		&trip.RouteID,
		&trip.ServiceID,
		&trip.Headsign,
		&trip.DirectionID,
		&trip.ShapeID,
	)
	if err != nil {
		slog.Error("retrieving row with specific id", "id", id, "err", err)
		return Trip{}, err
	}

	return trip, nil
}

func (r *TripRepo) GetTripsForToday(ctx context.Context) ([]Trip, error) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		slog.Error("loading location for trips", "err", err)
		return []Trip{}, err
	}
	nycTime := time.Now().In(loc)
	weekday := nycTime.Weekday()
	freqDay := getFreqDayFromWeekday(weekday)

	query := `
		SELECT
			id
			, day_of_week
			, short_trip_id
			, route_id
			, service_id
			, headsign
			, direction_id
			, shape_id
		FROM 
			trips
		WHERE 
			day_of_week = $1
		OR 
			day_of_week = $2
	`
	rows, err := r.db.Query(ctx, query, freqDay, Everyday)
	if err != nil {
		slog.Error("get trips for today", "err", err)
		return []Trip{}, err
	}
	defer rows.Close()

	trips := []Trip{}

	for rows.Next() {
		var trip Trip
		if err := rows.Scan(
			&trip.ID,
			&trip.DayOfWeek,
			&trip.ShortTripID,
			&trip.RouteID,
			&trip.ServiceID,
			&trip.Headsign,
			&trip.DirectionID,
			&trip.ShapeID,
		); err != nil {
			slog.Error("retrieving particular row", "err", err)
			return []Trip{}, err
		}
		trips = append(trips, trip)
	}

	return trips, nil
}

func getFreqDayFromWeekday(w time.Weekday) FreqDay {
	switch w {
	case time.Saturday:
		return Saturday
	case time.Sunday:
		return Sunday
	default:
		return Weekday
	}
}
