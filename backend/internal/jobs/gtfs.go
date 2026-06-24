package jobs

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/moonborks/transit-pulse/internal/transit/mta/gtfs"
)

func RunStaticGTFSJob(ctx context.Context, pool *pgxpool.Pool, gtfsURL string) {
	gtfs.RetrieveStaticGTFS(ctx, pool, gtfsURL)

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			slog.Info("running static gtfs retrieval job", "url", gtfsURL)

			gtfs.RetrieveStaticGTFS(ctx, pool, gtfsURL)
		}
	}
}

func RunRealTimeGTFSJob(ctx context.Context, rdb *redis.Client, mtaURL string) {
	gtfs.FetchRealtimeFeed(ctx, rdb, mtaURL)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// slog.Info("retrieving realtime gtfs retrieval job", "url", mtaURL)

			gtfs.FetchRealtimeFeed(ctx, rdb, mtaURL)
		}
	}
}
