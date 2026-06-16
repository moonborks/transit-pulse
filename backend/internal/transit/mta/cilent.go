package mta

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
)

var mtaGTFS = "https://rrgtfsfeeds.s3.amazonaws.com/gtfs_supplemented.zip"

func RetrieveGTFS(gtfsURL string) {
	resp, err := http.Get(mtaGTFS)
	if err != nil {
		slog.Error("unable to GET from URL", "url", gtfsURL, "err", err)
		return
	}
	defer resp.Body.Close()

	tempFile, err := os.CreateTemp("", "temp.zip")
	if err != nil {
		slog.Error("error creating a temp file", "err", err)
		return
	}
	defer os.Remove(tempFile.Name())

	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		slog.Error("error copying response body into a temp file", "err", err)
		return
	}

	err = unzipGTFS(tempFile.Name())
	if err != nil {
		slog.Error("failed to unzip file", "err", err)
		return
	}
}

func unzipGTFS(filename string) error {
	zr, err := zip.OpenReader(filename)
	if err != nil {
		return fmt.Errorf("failed to create zip reader: %w", err)
	}
	defer zr.Close()

	for _, file := range zr.File {
		switch file.Name {
		case "routes.txt":
			rc, err := file.Open()
			if err != nil {
				slog.Error("unable to read routes.txt", "err", err)
			}
			parseRoutes(rc)
			rc.Close()
		case "shapes.txt":
			rc, err := file.Open()
			if err != nil {
				slog.Error("unable to read shapes.txt", "err", err)
			}
			parseShapes(rc)
			rc.Close()
		case "stops.txt":
			rc, err := file.Open()
			if err != nil {
				slog.Error("unable to read stops.txt", "err", err)
			}
			parseStops(rc)
			rc.Close()
		case "trips.txt":
			rc, err := file.Open()
			if err != nil {
				slog.Error("unable to read trips.txt", "err", err)
			}
			parseTrips(rc)
			rc.Close()
		default:
			slog.Info("file is unused", "filename", file.Name)
		}
	}

	return nil
}

func parseRoutes(rc io.ReadCloser) {
	reader := csv.NewReader(rc)
	_, err := reader.Read()
	if err != nil {
		slog.Error("unable to read from routes.txt", "err", err)
		return
	}

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			slog.Error("failed to read a line from routes.txt", "err", err)
		}

		fmt.Println(row)
	}
}

func parseShapes(rc io.ReadCloser) {
	reader := csv.NewReader(rc)
	_, err := reader.Read()
	if err != nil {
		slog.Error("unable to read from shapes.txt", "err", err)
		return
	}

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			slog.Error("failed to read a line from shapes.txt", "err", err)
		}

		fmt.Println(row)
	}
}

func parseStops(rc io.ReadCloser) {
	reader := csv.NewReader(rc)
	_, err := reader.Read()
	if err != nil {
		slog.Error("unable to read from stops.txt", "err", err)
		return
	}

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			slog.Error("failed to read a line from stops.txt", "err", err)
		}

		fmt.Println(row)
	}
}

func parseTrips(rc io.ReadCloser) {
	reader := csv.NewReader(rc)
	_, err := reader.Read()
	if err != nil {
		slog.Error("unable to read from trips.txt", "err", err)
		return
	}

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			slog.Error("failed to read a line from trips.txt", "err", err)
		}

		fmt.Println(row)
	}
}
