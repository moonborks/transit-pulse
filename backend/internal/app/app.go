package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/moonborks/transit-pulse/internal/database"
	"github.com/moonborks/transit-pulse/internal/jobs"
	"github.com/moonborks/transit-pulse/internal/server"
	"github.com/moonborks/transit-pulse/internal/transit/mta/nextstops"
	"github.com/moonborks/transit-pulse/internal/transit/mta/routes"
	"github.com/moonborks/transit-pulse/internal/transit/mta/shapes"
	"github.com/moonborks/transit-pulse/internal/transit/mta/stops"
	"github.com/moonborks/transit-pulse/internal/transit/mta/times"
	"github.com/moonborks/transit-pulse/internal/transit/mta/trips"
)

var (
	mtaGTFS = "https://rrgtfsfeeds.s3.amazonaws.com/gtfs_supplemented.zip"
	gtfsRT  = []string{
		"https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-ace",
		"https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-bdfm",
		"https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-g",
		"https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-jz",
		"https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-nqrw",
		"https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-l",
		"https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs",
		"https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-si",
	}
)

type App struct {
	DB     *pgxpool.Pool
	RDB    *redis.Client
	Router http.Handler
}

func NewApp() *App {
	var handler slog.Handler
	if os.Getenv("APP_ENV") == "production" {
		handler = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}
	logger := slog.New(handler)
	slog.SetDefault(logger)

	ctx := context.Background()
	config, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		slog.Error("Parse database config:", "err", err)
	}
	config.MinConns = 5

	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		slog.Error("Connect to database:", "err", err)
	}

	slog.Info(
		"Connected to database",
		"host", config.ConnConfig.Host,
		"port", config.ConnConfig.Port,
		"database", config.ConnConfig.Database,
	)
	db_table_err := database.Migrate(ctx, db)
	if db_table_err != nil {
		panic(db_table_err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("VALKEY_URL"),
		Password: "",
		DB:       0,
	})

	go jobs.RunStaticGTFSJob(ctx, db, mtaGTFS)
	for _, rtFeed := range gtfsRT {
		go jobs.RunRealTimeGTFSJob(ctx, rdb, rtFeed)
	}

	routeRepo := routes.NewRouteRepo(db)
	shapeRepo := shapes.NewShapeRepo(db)
	stopRepo := stops.NewStopRepo(db)
	tripRepo := trips.NewTripRepo(db)
	timeRepo := times.NewTimeRepo(db, rdb)
	nextStopRepo := nextstops.NewNextStopRepo(rdb)

	routeService := routes.NewRouteService(routeRepo)
	shapeService := shapes.NewShapeService(shapeRepo)
	stopService := stops.NewStopService(stopRepo, nextStopRepo)
	tripService := trips.NewTripService(tripRepo, nextStopRepo, shapeRepo)
	timeService := times.NewTimeService(timeRepo)

	routeHandler := routes.NewRouteHandler(routeService, stopService)
	shapeHandler := shapes.NewShapeHandler(shapeService)
	stopHandler := stops.NewStopHandler(stopService)
	tripHandler := trips.NewTripHandler(tripService)
	timeHandler := times.NewTimeHandler(timeService)

	handlers := server.Handlers{
		Route: routeHandler,
		Shape: shapeHandler,
		Stop:  stopHandler,
		Trip:  tripHandler,
		Time:  timeHandler,
	}

	router := server.MainRouter(&handlers)

	return &App{
		DB:     db,
		RDB:    rdb,
		Router: router,
	}
}
