package trips

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TripRepo struct {
	db *pgxpool.Pool
}

func NewTripRepo(db *pgxpool.Pool) *TripRepo {
	return &TripRepo{db: db}
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

func (r *TripRepo) getCurrentFreqDay() FreqDay {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		slog.Error("loading location for trips in repository", "err", err)
		return Weekday
	}

	nycTime := time.Now().In(loc)
	weekday := nycTime.Weekday()

	return getFreqDayFromWeekday(weekday)
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

func (r *TripRepo) GetStopSequences(ctx context.Context, tripStopKeys []TripStopKey) (map[TripStopKey]int64, error) {
	if len(tripStopKeys) == 0 {
		return make(map[TripStopKey]int64), nil
	}
	currentFreqDay := r.getCurrentFreqDay()
	var (
		shortTripIDs []string
		nextStopIDs  []string
	)
	for _, key := range tripStopKeys {
		shortTripIDs = append(shortTripIDs, key.ShortTripID)
		nextStopIDs = append(nextStopIDs, key.StopID)
	}

	query := `
		SELECT 
			t.short_trip_id,
			t.stop_id,
			t.stop_sequence 
		FROM 
			times t
		JOIN 
			UNNEST($1::text[], $2::text[]) AS k(short_trip_id, stop_id)
			ON  t.short_trip_id = k.short_trip_id 
			AND t.stop_id = k.stop_id
		WHERE 
			(t.day_of_week = $3::freq_day OR t.day_of_week = 'everyday'::freq_day);
	`
	rows, err := r.db.Query(ctx, query, shortTripIDs, nextStopIDs, string(currentFreqDay))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resultMap := make(map[TripStopKey]int64, len(tripStopKeys))

	_, err = pgx.ForEachRow(rows,
		[]any{new(string), new(string), new(int64)},
		func() error {
			var shortTripID, nextStopID string
			var sequence int64

			if err := rows.Scan(&shortTripID, &nextStopID, &sequence); err != nil {
				return err
			}
			key := TripStopKey{ShortTripID: shortTripID, StopID: nextStopID}
			resultMap[key] = sequence
			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return resultMap, nil
}

func (r *TripRepo) GetPrevStopInfo(ctx context.Context, tripSequenceKeys []TripSequenceKey) (map[TripStopKey]PrevStopInfo, error) {
	resultMap := make(map[TripStopKey]PrevStopInfo, len(tripSequenceKeys))
	if len(tripSequenceKeys) == 0 {
		return resultMap, nil
	}

	currentFreqDay := r.getCurrentFreqDay()

	tripIDs := make([]string, len(tripSequenceKeys))
	sequences := make([]int32, len(tripSequenceKeys))

	for i, key := range tripSequenceKeys {
		tripIDs[i] = key.ShortTripID
		sequences[i] = int32(key.Sequence)
	}

	query := `
		SELECT 
			t.short_trip_id, 
			t.stop_id, 
			t.departure_time, 
			t.stop_sequence
		FROM 
			times t
		JOIN 
			UNNEST($1::text[], $2::int[]) 
			AS 
				k(short_trip_id, stop_sequence)
			ON  
				t.short_trip_id = k.short_trip_id 
			AND 
				t.stop_sequence = k.stop_sequence
		WHERE 
			(t.day_of_week = $3::freq_day OR t.day_of_week = 'everyday'::freq_day);
	`

	rows, err := r.db.Query(ctx, query, tripIDs, sequences, string(currentFreqDay))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	_, err = pgx.ForEachRow(rows,
		[]any{new(string), new(string), new(string), new(int32)},
		func() error {
			var shortTripID, stopID, departureTime string
			var stopSequence int32

			if err := rows.Scan(&shortTripID, &stopID, &departureTime, &stopSequence); err != nil {
				return err
			}

			key := TripStopKey{
				ShortTripID: shortTripID,
			}

			resultMap[key] = PrevStopInfo{
				PrevStopID:              stopID,
				PrevStationStopSequence: int64(stopSequence),
				PrevDepartureTime:       departureTime,
			}
			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return resultMap, nil
}

func (r *TripRepo) GetShapeSequences(ctx context.Context, contexts []TrainContext) (map[TripStopKey]ShapeRange, error) {
	resultMap := make(map[TripStopKey]ShapeRange, len(contexts))
	if len(contexts) == 0 {
		return resultMap, nil
	}

	shortTripIDs := make([]string, len(contexts))
	shapeIDs := make([]string, len(contexts))
	prevStopIDs := make([]string, len(contexts))
	nextStopIDs := make([]string, len(contexts))

	for i, c := range contexts {
		shortTripIDs[i] = c.ShortTripID
		nextStopIDs[i] = c.NextStopID
		prevStopIDs[i] = c.PrevStopID

		parts := strings.Split(c.ShortTripID, "_")
		if len(parts) > 1 {
			shapeIDs[i] = parts[1]
		} else {
			shapeIDs[i] = ""
		}
	}

	query := `
		SELECT 
			k.short_trip_id,
			k.next_stop_id,
			p_shape.sequence AS prev_seq,
			n_shape.sequence AS next_seq
		FROM 
			UNNEST($1::text[], $2::text[], $3::text[], $4::text[]) 
		AS 
			k(short_trip_id, shape_id, prev_stop_id, next_stop_id)
		JOIN 
			stops p_stop ON p_stop.id = k.prev_stop_id
		JOIN 
			stops n_stop ON n_stop.id = k.next_stop_id
		JOIN 
			shapes p_shape 
			ON 
				p_shape.id = k.shape_id 
			AND 
				p_shape.lat = p_stop.lat 
			AND 
				p_shape.lon = p_stop.lon
		JOIN 
			shapes n_shape ON n_shape.id = k.shape_id 
			AND 
				n_shape.lat = n_stop.lat 
			AND 
				n_shape.lon = n_stop.lon;
	`

	rows, err := r.db.Query(ctx, query, shortTripIDs, shapeIDs, prevStopIDs, nextStopIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var shortTripID, nextStopID string
		var prevSeq, nextSeq int64

		if err := rows.Scan(&shortTripID, &nextStopID, &prevSeq, &nextSeq); err != nil {
			return nil, err
		}

		mapKey := TripStopKey{
			ShortTripID: shortTripID,
			StopID:      nextStopID,
		}

		resultMap[mapKey] = ShapeRange{
			PrevShapeSequence: prevSeq,
			NextShapeSequence: nextSeq,
		}
	}

	return resultMap, rows.Err()
}

func (r *TripRepo) GetCoordinatesByShapeSequence(ctx context.Context, contexts []TrainContext) (map[TripStopKey]TrainCoordinates, error) {
	resultMap := make(map[TripStopKey]TrainCoordinates, len(contexts))
	if len(contexts) == 0 {
		return resultMap, nil
	}

	shortTripIDs := make([]string, len(contexts))
	nextStopIDs := make([]string, len(contexts))
	shapeIDs := make([]string, len(contexts))
	sequences := make([]int32, len(contexts))

	for i, c := range contexts {
		shortTripIDs[i] = c.ShortTripID
		nextStopIDs[i] = c.NextStopID
		sequences[i] = int32(c.CurrentShapeSequence)

		parts := strings.Split(c.ShortTripID, "_")
		if len(parts) > 1 {
			shapeIDs[i] = parts[1]
		} else {
			shapeIDs[i] = ""
		}
	}

	query := `
		SELECT 
			k.short_trip_id,
			k.next_stop_id,
			s.lat,
			s.lon
		FROM 
			UNNEST($1::text[], $2::text[], $3::text[], $4::int[]) AS k(short_trip_id, next_stop_id, shape_id, sequence)
		JOIN 
			shapes s ON s.id = k.shape_id AND s.sequence = k.sequence;
	`

	rows, err := r.db.Query(ctx, query, shortTripIDs, nextStopIDs, shapeIDs, sequences)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var shortTripID, nextStopID string
		var coords TrainCoordinates

		if err := rows.Scan(&shortTripID, &nextStopID, &coords.Lat, &coords.Lon); err != nil {
			return nil, err
		}

		mapKey := TripStopKey{
			ShortTripID: shortTripID,
			StopID:      nextStopID,
		}

		resultMap[mapKey] = coords
	}

	return resultMap, rows.Err()
}
