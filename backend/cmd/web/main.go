package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/moonborks/transit-pulse/internal/app"
)

func main() {
	if os.Getenv("APP_ENV") != "production" {
		err := godotenv.Load(".env")
		if err != nil {
			panic(err)
		}
	}
	app := app.NewApp()
	var port = "8888"
	fmt.Println("Server running on :" + port)
	server_err := http.ListenAndServe(":"+port, app.Router)
	if server_err != nil {
		slog.Error("Server Start", "err", server_err)
	}
}
