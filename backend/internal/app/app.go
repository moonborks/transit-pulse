package app

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moonborks/transit-pulse/internal/database"
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

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Route("/api", func(api chi.Router) {

	})

	return &App{
		DB:     db,
		Router: router,
	}
}
