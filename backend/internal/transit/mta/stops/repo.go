package stops

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type StopRepo struct {
	db *pgxpool.Pool
}

func NewStopRepo(db *pgxpool.Pool) *StopRepo {
	return &StopRepo{db: db}
}

func (r *StopRepo) GetAll(ctx context.Context) ([]*Stop, error) {
	stmt := `
		SELECT
			id, name, lat, lon
		FROM
			stops
	`

	rows, err := r.db.Query(ctx, stmt)
	if err != nil {
		slog.Error("query stops table", "err", err)
		return nil, err
	}
	defer rows.Close()

	stops := []*Stop{}

	for rows.Next() {
		var stop Stop
		if err := rows.Scan(
			&stop.ID,
			&stop.Name,
			&stop.Lat,
			&stop.Lon,
		); err != nil {
			slog.Error("retrieving particular row", "err", err)
		}
		stops = append(stops, &stop)
	}

	return stops, nil
}

func (r *StopRepo) GetStop(ctx context.Context, id string) (*Stop, error) {
	stmt := `
		SELECT
			id, name, lat, lon
		FROM
			stops
		WHERE
			id = $1
	`

	var stop Stop

	row := r.db.QueryRow(ctx, stmt, id)
	err := row.Scan(&stop.ID, &stop.Name, &stop.Lat, &stop.Lon)
	if err != nil {
		slog.Error("retrieving row with specific id", "id", id, "err", err)
		return nil, err
	}

	return &stop, nil
}
