package app

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/moonborks/transit-pulse/internal/database"
	"github.com/moonborks/transit-pulse/internal/server"
	"github.com/moonborks/transit-pulse/internal/transit/mta/routes"
	"github.com/moonborks/transit-pulse/internal/transit/mta/shapes"
	"github.com/moonborks/transit-pulse/internal/transit/mta/stops"
	"github.com/moonborks/transit-pulse/internal/transit/mta/trips"
)

type App struct {
	DB     *pgxpool.Pool
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

	db, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Connect to database:", err)
		panic(err)
	}
	config := db.Config()
	slog.Info("Connected to database",
		"host", config.ConnConfig.Host,
		"port", config.ConnConfig.Port,
		"database", config.ConnConfig.Database,
	)
	db_table_err := database.Migrate(ctx, db)
	if db_table_err != nil {
		panic(db_table_err)
	}

	routeRepo := routes.NewRouteRepo(db)
	shapeRepo := shapes.NewShapeRepo(db)
	stopRepo := stops.NewStopRepo(db)
	tripRepo := trips.NewTripRepo(db)

	routeService := routes.NewRouteService(routeRepo)
	shapeService := shapes.NewShapeService(shapeRepo)
	stopService := stops.NewStopService(stopRepo)
	tripService := trips.NewTripService(tripRepo)

	routeHandler := routes.NewRouteHandler(routeService)
	shapeHandler := shapes.NewShapeHandler(shapeService)
	stopHandler := stops.NewStopHandler(stopService)
	tripHandler := trips.NewTripHandler(tripService)

	handlers := server.Handlers{
		Route: routeHandler,
		Shape: shapeHandler,
		Stop:  stopHandler,
		Trip:  tripHandler,
	}

	router := server.MainRouter(&handlers)

	return &App{
		DB:     db,
		Router: router,
	}
}
