package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Migrate(ctx context.Context, conn *pgxpool.Pool) error {
	_, err := conn.Exec(
		ctx,
		`CREATE TABLE IF NOT EXISTS schema_version (version INTEGER PRIMARY KEY)`,
	)
	if err != nil {
		return err
	}

	migrations := []struct {
		version int
		sql     string
	}{
		{1, `
			CREATE TABLE IF NOT EXISTS routes (
				id          TEXT PRIMARY KEY,
				short_name  TEXT,
				long_name   TEXT,
				type        SMALLINT,
				color       TEXT
			);

			CREATE TABLE IF NOT EXISTS stops (
				id              TEXT PRIMARY KEY,
				name            TEXT,
				lat             DOUBLE PRECISION,
				lon             DOUBLE PRECISION,
				location_type   SMALLINT NULL,
				parent_station  TEXT NULL
			);

			CREATE TABLE IF NOT EXISTS shapes (
				id          TEXT,
				sequence    INT,
				lat         DOUBLE PRECISION,
				lon         DOUBLE PRECISION,
				PRIMARY KEY (id, sequence)
			);

			ALTER TABLE shapes
			ADD CONSTRAINT unique_lat_lon UNIQUE(lat, lon);

			CREATE TYPE freq_day AS ENUM (
				'everyday'
				, 'weekday'
				, 'saturday'
				, 'sunday'
			);

			CREATE TABLE IF NOT EXISTS trips (
				id              TEXT PRIMARY KEY,
				day_of_week     freq_day NOT NULL DEFAULT 'everyday',
				short_trip_id   TEXT,
				route_id        TEXT REFERENCES routes(id),
				service_id      TEXT,
				headsign        TEXT,
				direction_id    SMALLINT,
				shape_id        TEXT NULL
			);

			CREATE TABLE IF NOT EXISTS times (
				day_of_week    freq_day NOT NULL DEFAULT 'everyday',
				short_trip_id  TEXT,
				trip_id        TEXT REFERENCES trips(id),
				stop_id        TEXT REFERENCES stops(id),
				arrival_time   TEXT,
				departure_time TEXT,
				stop_sequence  INT,
				PRIMARY KEY (trip_id, stop_id)
			);
		`},
	}

	for _, m := range migrations {
		var count int
		conn.QueryRow(ctx, "SELECT COUNT(*) FROM schema_version WHERE version = $1", m.version).
			Scan(&count)
		if count > 0 {
			continue
		}
		if _, err := conn.Exec(ctx, m.sql); err != nil {
			return fmt.Errorf("migration %d: %w", m.version, err)
		}
		conn.Exec(ctx, "INSERT INTO schema_version (version) VALUES ($1)", m.version)
	}
	return nil
}
