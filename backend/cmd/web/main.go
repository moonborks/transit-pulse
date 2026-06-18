package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/moonborks/transit-pulse/internal/app"
	"github.com/moonborks/transit-pulse/internal/transit/mta/gtfs"
)

func main() {
	if os.Getenv("APP_ENV") != "production" {
		err := godotenv.Load(".env")
		if err != nil {
			panic(err)
		}
	}
	app := app.NewApp()
	gtfs.RetrieveStaticGTFS(context.Background(), app.DB, "")
	port := "8888"
	fmt.Println("Server running on :" + port)
	server_err := http.ListenAndServe(":"+port, app.Router)
	if server_err != nil {
		slog.Error("Server Start", "err", server_err)
	}
}
