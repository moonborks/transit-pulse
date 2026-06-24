package jobs

import (
	"context"
	"log/slog"
	"sync"
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

func RunRealTimeGTFSJob(ctx context.Context, rdb *redis.Client, gtfsSSE *gtfs.SSE) {
	gtfsRT := []string{
		"https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-ace",
		"https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-bdfm",
		"https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-g",
		"https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-jz",
		"https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-nqrw",
		"https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-l",
		"https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs",
		"https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-si",
	}

	var wg sync.WaitGroup
	for _, feed := range gtfsRT {
		wg.Add(1)

		go func(feed string) {
			defer wg.Done()
			gtfs.FetchRealtimeFeed(ctx, rdb, feed)
		}(feed)
	}

	wg.Wait()
	gtfsSSE.TripChannel <- time.Now().String()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			slog.Info("retrieving realtime gtfs retrieval job")

			var wg sync.WaitGroup
			for _, feed := range gtfsRT {
				wg.Add(1)

				go func(feed string) {
					defer wg.Done()
					gtfs.FetchRealtimeFeed(ctx, rdb, feed)
				}(feed)
			}

			wg.Wait()
			gtfsSSE.TripChannel <- time.Now().String()
		}
	}
}
