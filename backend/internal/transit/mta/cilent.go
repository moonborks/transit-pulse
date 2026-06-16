package mta

import (
	"archive/zip"
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var mtaGTFS = "https://rrgtfsfeeds.s3.amazonaws.com/gtfs_supplemented.zip"

func RetrieveGTFS(ctx context.Context, pool *pgxpool.Pool, gtfsURL string) {
	resp, err := http.Get(mtaGTFS)
	if err != nil {
		slog.Error("GET from URL", "url", gtfsURL, "err", err)
		return
	}
	defer resp.Body.Close()

	tempFile, err := os.CreateTemp("", "temp.zip")
	if err != nil {
		slog.Error("creating a temp file", "err", err)
		return
	}
	defer os.Remove(tempFile.Name())

	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		slog.Error("copying response body into a temp file", "err", err)
		return
	}

	zr, err := zip.OpenReader(tempFile.Name())
	if err != nil {
		slog.Error("unzip file", "err", err)
		return
	}
	defer zr.Close()

	tx, err := pool.Begin(ctx)
	if err != nil {
		slog.Error("retrieve transaction connection to db", "err", err)
		return
	}
	defer tx.Rollback(ctx)

	err = createStagingTables(ctx, tx)
	if err != nil {
		slog.Error("create temp staging tables", "err", err)
		return
	}

	queries := map[string]string{
		"routes.txt": "COPY routes_staging FROM STDIN CSV HEADER",
		"shapes.txt": "COPY shapes_staging FROM STDIN CSV HEADER",
		"stops.txt":  "COPY stops_staging FROM STDIN CSV HEADER",
		"trips.txt":  "COPY trips_staging FROM STDIN CSV HEADER",
	}

	for _, file := range zr.File {
		query, ok := queries[file.Name]
		if !ok {
			continue
		}

		rc, err := file.Open()
		if err != nil {
			slog.Error("open file", "filename", file.Name, "err", err)
			return
		}

		_, err = tx.Conn().PgConn().CopyFrom(ctx, rc, query)
		rc.Close()
		if err != nil {
			slog.Error("copy data from csv", "filename", file.Name, "err", err)
			return
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		slog.Error("transaction commit", "err", err)
	}
}

func createStagingTables(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(
		ctx,
		`
			CREATE TEMP TABLE routes_staging (
				agency_id TEXT
				, route_id TEXT
				, route_short_name TEXT
				, route_long_name TEXT
				, route_type SMALLINT
				, route_desc TEXT
				, route_url TEXT
				, route_color TEXT
				, route_text_color TEXT
				, route_sort_order TEXT
			) ON COMMIT DROP;

			CREATE TEMP TABLE shapes_staging (
				shape_id TEXT
				, shape_pt_sequence INT
				, shape_pt_lat DOUBLE PRECISION
				, shape_pt_lon DOUBLE PRECISION
			) ON COMMIT DROP;

			CREATE TEMP TABLE stops_staging (
				stop_id TEXT
				, stop_name TEXT
				, stop_lat DOUBLE PRECISION
				, stop_lon DOUBLE PRECISION
				, location_type SMALLINT NULL
				, parent_station TEXT
			) ON COMMIT DROP;

			CREATE TEMP TABLE trips_staging (
				route_id TEXT
				, trip_id TEXT
				, service_id TEXT
				, trip_headsign TEXT
				, direction_id SMALLINT
				, shape_id TEXT
			) ON COMMIT DROP;
		`,
	)
	return err
}
