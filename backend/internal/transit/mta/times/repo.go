package times

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type TimeRepo struct {
	db  *pgxpool.Pool
	rdb *redis.Client
}

func NewTimeRepo(db *pgxpool.Pool, rdb *redis.Client) *TimeRepo {
	return &TimeRepo{db: db, rdb: rdb}
}

func (r *TimeRepo) GetAll(ctx context.Context) ([]Time, error) {
	stmt := `
		SELECT
			trip_id, stop_id, arrival_time, departure_time, stop_sequence
		FROM
			times
	`

	times := []Time{}

	rows, err := r.db.Query(ctx, stmt)
	if err != nil {
		slog.Error("query times table", "err", err)
	}

	for rows.Next() {
		var time Time
		if err := rows.Scan(
			&time.TripID,
			&time.StopID,
			&time.ArrivalTime,
			&time.DepartureTime,
			&time.StopSequence,
		); err != nil {
			slog.Error("retrieving particular row in times table", "err", err)
		}
		times = append(times, time)
	}

	rows.Close()

	return times, nil
}
