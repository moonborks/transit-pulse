package trips

import (
	"context"
	"fmt"
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

	currentFreqDay := string(r.getCurrentFreqDay())
	resultMap := make(map[TripStopKey]int64, len(tripStopKeys))

	// --- PASS 1: Exact Match (Tier 1) ---
	if err := r.queryTierBatch(ctx, "exact", tripStopKeys, currentFreqDay, resultMap); err != nil {
		return nil, err
	}

	if len(resultMap) == len(tripStopKeys) {
		return resultMap, nil
	}

	// --- PASS 2: Suffix Match (Tier 2) ---
	missingKeys := getMissingKeysBatch(tripStopKeys, resultMap)
	if err := r.queryTierBatch(ctx, "suffix", missingKeys, currentFreqDay, resultMap); err != nil {
		return nil, err
	}

	if len(resultMap) == len(tripStopKeys) {
		return resultMap, nil
	}

	// --- PASS 3: Base Prefix Match (Tier 3) ---
	missingKeys = getMissingKeysBatch(tripStopKeys, resultMap)
	if err := r.queryTierBatch(ctx, "prefix", missingKeys, currentFreqDay, resultMap); err != nil {
		return nil, err
	}

	for _, key := range tripStopKeys {
		if _, exists := resultMap[key]; !exists {
			resultMap[key] = -1
		}
	}

	return resultMap, nil
}

func (r *TripRepo) queryTierBatch(ctx context.Context, tierType string, keys []TripStopKey, day string, resultMap map[TripStopKey]int64) error {
	if len(keys) == 0 {
		return nil
	}

	shortTripIDs := make([]string, 0, len(keys))
	nextStopIDs := make([]string, 0, len(keys))
	extractedValues := make([]string, 0, len(keys))

	for _, key := range keys {
		shortTripIDs = append(shortTripIDs, key.ShortTripID)
		nextStopIDs = append(nextStopIDs, key.StopID)

		parts := strings.Split(key.ShortTripID, "_")
		suffix := ""
		if len(parts) > 1 {
			suffix = parts[1]
		}

		switch tierType {
		case "suffix":
			extractedValues = append(extractedValues, suffix)
		case "prefix":
			prefix := suffix
			if len(suffix) > 4 {
				prefix = suffix[:4]
			}
			extractedValues = append(extractedValues, prefix)
		}
	}

	var query string
	var args []any

	switch tierType {
	case "exact":
		query = `
			SELECT ik.short_trip_id, ik.stop_id, t.stop_sequence
			FROM UNNEST($1::text[], $2::text[]) AS ik(short_trip_id, stop_id)
			JOIN times t ON t.stop_id = ik.stop_id AND t.short_trip_id = ik.short_trip_id
			WHERE t.day_of_week IN ($3::freq_day, 'everyday'::freq_day);`
		args = []any{shortTripIDs, nextStopIDs, day}
	case "suffix":
		query = `
			SELECT DISTINCT ON (ik.short_trip_id, ik.stop_id) ik.short_trip_id, ik.stop_id, t.stop_sequence
			FROM UNNEST($1::text[], $2::text[], $3::text[]) AS ik(short_trip_id, stop_id, suffix)
			JOIN times t ON t.stop_id = ik.stop_id AND t.trip_suffix = ik.suffix
			WHERE t.day_of_week IN ($4::freq_day, 'everyday'::freq_day);`
		args = []any{shortTripIDs, nextStopIDs, extractedValues, day}
	case "prefix":
		query = `
			SELECT DISTINCT ON (ik.short_trip_id, ik.stop_id) ik.short_trip_id, ik.stop_id, t.stop_sequence
			FROM UNNEST($1::text[], $2::text[], $3::text[]) AS ik(short_trip_id, stop_id, prefix)
			JOIN times t ON t.stop_id = ik.stop_id AND t.trip_base_prefix = ik.prefix
			WHERE t.day_of_week IN ($4::freq_day, 'everyday'::freq_day);`
		args = []any{shortTripIDs, nextStopIDs, extractedValues, day}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("batch query tier %s failed: %w", tierType, err)
	}
	defer rows.Close()

	var sID, stopID string
	var seq int64
	for rows.Next() {
		if err := rows.Scan(&sID, &stopID, &seq); err != nil {
			return err
		}
		resultMap[TripStopKey{ShortTripID: sID, StopID: stopID}] = seq
	}
	return rows.Err()
}

func getMissingKeysBatch(original []TripStopKey, found map[TripStopKey]int64) []TripStopKey {
	var missing []TripStopKey
	for _, key := range original {
		if _, exists := found[key]; !exists {
			missing = append(missing, key)
		}
	}
	return missing
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
		SELECT DISTINCT ON (ik.short_trip_id)
			ik.short_trip_id,
			t.stop_id, 
			t.departure_time, 
			t.stop_sequence
		FROM 
			UNNEST($1::text[], $2::int[]) AS ik(short_trip_id, stop_sequence)
		JOIN 
			times t 
			ON  t.stop_id = t.stop_id
			AND t.trip_suffix LIKE SPLIT_PART(ik.short_trip_id, '_', 2) || '%'
			AND t.stop_sequence = ik.stop_sequence
		WHERE 
			(t.day_of_week = $3::freq_day OR t.day_of_week = 'everyday'::freq_day)
		ORDER BY 
			ik.short_trip_id, 
			t.stop_sequence DESC;
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
		SELECT DISTINCT ON (k.short_trip_id, k.next_stop_id)
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
				p_shape.id LIKE k.shape_id || '%' 
			AND 
				p_shape.lat = p_stop.lat 
			AND 
				p_shape.lon = p_stop.lon
		JOIN 
			shapes n_shape 
			ON 
				n_shape.id LIKE k.shape_id || '%' 
			AND 
				n_shape.lat = n_stop.lat 
			AND 
				n_shape.lon = n_stop.lon
		ORDER BY 
			k.short_trip_id, k.next_stop_id, p_shape.sequence ASC;
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
		SELECT DISTINCT ON (k.short_trip_id, k.next_stop_id)
			k.short_trip_id,
			k.next_stop_id,
			s.lat,
			s.lon
		FROM 
			UNNEST($1::text[], $2::text[], $3::text[], $4::int[]) AS k(short_trip_id, next_stop_id, shape_id, sequence)
		JOIN 
			shapes s ON s.id LIKE k.shape_id || '%' AND s.sequence = k.sequence
		ORDER BY 
			k.short_trip_id, k.next_stop_id;
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

func (r *TripRepo) GetPositionsWithHistory(ctx context.Context, contexts []TrainContext) (map[TripStopKey]TrainCoordinates, map[TripStopKey]TrainCoordinates, error) {
	currentMap := make(map[TripStopKey]TrainCoordinates, len(contexts))
	prevMap := make(map[TripStopKey]TrainCoordinates, len(contexts))
	if len(contexts) == 0 {
		return currentMap, prevMap, nil
	}

	batchSize := len(contexts) * 2
	shortTripIDs := make([]string, 0, batchSize)
	nextStopIDs := make([]string, 0, batchSize)
	shapeIDs := make([]string, 0, batchSize)
	sequences := make([]int32, 0, batchSize)
	isPrevMarker := make([]bool, 0, batchSize)

	for _, c := range contexts {
		parts := strings.Split(c.ShortTripID, "_")
		shapeID := ""
		if len(parts) > 1 {
			shapeID = parts[1]
		}

		prevSeq := int32(c.CurrentShapeSequence - 1)
		if prevSeq < 1 {
			prevSeq = 1
		}

		shortTripIDs = append(shortTripIDs, c.ShortTripID)
		nextStopIDs = append(nextStopIDs, c.NextStopID)
		shapeIDs = append(shapeIDs, shapeID)
		sequences = append(sequences, int32(c.CurrentShapeSequence))
		isPrevMarker = append(isPrevMarker, false)

		shortTripIDs = append(shortTripIDs, c.ShortTripID)
		nextStopIDs = append(nextStopIDs, c.NextStopID)
		shapeIDs = append(shapeIDs, shapeID)
		sequences = append(sequences, prevSeq)
		isPrevMarker = append(isPrevMarker, true)
	}

	query := `
		SELECT DISTINCT ON (k.short_trip_id, k.next_stop_id, k.is_prev)
			k.short_trip_id,
			k.next_stop_id,
			k.is_prev,
			s.lat,
			s.lon
		FROM 
			UNNEST($1::text[], $2::text[], $3::text[], $4::int[], $5::boolean[]) 
		AS 
			k(short_trip_id, next_stop_id, shape_id, sequence, is_prev)
		JOIN 
			shapes s ON s.id LIKE k.shape_id || '%' AND s.sequence = k.sequence
		ORDER BY 
			k.short_trip_id, k.next_stop_id, k.is_prev;
	`

	rows, err := r.db.Query(ctx, query, shortTripIDs, nextStopIDs, shapeIDs, sequences, isPrevMarker)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var shortTripID, nextStopID string
		var isPrev bool
		var coords TrainCoordinates

		if err := rows.Scan(&shortTripID, &nextStopID, &isPrev, &coords.Lat, &coords.Lon); err != nil {
			return nil, nil, err
		}

		mapKey := TripStopKey{
			ShortTripID: shortTripID,
			StopID:      nextStopID,
		}

		if isPrev {
			prevMap[mapKey] = coords
		} else {
			currentMap[mapKey] = coords
		}
	}

	return currentMap, prevMap, rows.Err()
}
