package trips

import (
	"context"
	"log/slog"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/moonborks/transit-pulse/internal/transit/mta/nextstops"
	"github.com/moonborks/transit-pulse/internal/transit/mta/shapes"
)

type TripService struct {
	tripRepo     *TripRepo
	nextStopRepo *nextstops.NextStopRepo
	shapeRepo    *shapes.ShapeRepo
}

func NewTripService(
	tr *TripRepo,
	nsr *nextstops.NextStopRepo,
	sr *shapes.ShapeRepo,
) *TripService {
	return &TripService{
		tripRepo:     tr,
		nextStopRepo: nsr,
		shapeRepo:    sr,
	}
}

func (s *TripService) GetAll(ctx context.Context) ([]Trip, error) {
	return s.tripRepo.GetAll(ctx)
}

func (s *TripService) GetTrip(ctx context.Context, id string) (Trip, error) {
	return s.tripRepo.GetTrip(ctx, id)
}

func (s *TripService) GetTripsForToday(ctx context.Context) ([]TripAPI, error) {
	trips, err := s.tripRepo.GetTripsForToday(ctx)
	if err != nil {
		return []TripAPI{}, err
	}
	apiPayload := make([]TripAPI, len(trips))
	for i, trip := range trips {
		apiPayload[i] = TripAPI{
			RouteID:  trip.RouteID,
			Headsign: trip.Headsign,
			ShapeID:  trip.ShapeID,
		}
	}

	return apiPayload, nil
}

func (s *TripService) GetTripPositions(ctx context.Context) ([]TripTrainLocationAPI, error) {
	slog.Info("Starting GetTripPositions execution")

	var nextStops []nextstops.NextStop
	nextStops, err := s.nextStopRepo.GetAllNextStops(ctx)
	if err != nil {
		slog.Error("Failed to fetch active next stops", "err", err)
		return []TripTrainLocationAPI{}, err
	}
	numOfTrains := len(nextStops)

	slog.Debug("Successfully fetched active next stops", "count", numOfTrains)

	tripStopKeys := make([]TripStopKey, 0, 2*numOfTrains)
	for _, nextStop := range nextStops {
		tripStopKey := TripStopKey{ShortTripID: nextStop.ShortTripID, StopID: nextStop.StopID}
		tripStopKeys = append(tripStopKeys, tripStopKey)
	}

	tripStopKeySequenceMap := make(map[TripStopKey]int64, numOfTrains)
	tripStopKeySequenceMap, err = s.tripRepo.GetStopSequences(ctx, tripStopKeys)
	if err != nil {
		slog.Error("Failed to get stop sequences", "err", err)
		return []TripTrainLocationAPI{}, err
	}
	slog.Debug("Successfully retrieved stop sequences", "mapped_count", len(tripStopKeySequenceMap))

	tripSequenceKeys := make([]TripSequenceKey, 0, numOfTrains)

	slog.Debug("checking tripStopKeySequenceMap initialization",
		"total_keys", len(tripStopKeySequenceMap),
	)

	for tripStopKey, sequence := range tripStopKeySequenceMap {
		calculatedSeq := sequence - 1
		if calculatedSeq > 0 {
			tripSequenceKey := TripSequenceKey{
				ShortTripID: tripStopKey.ShortTripID,
				Sequence:    calculatedSeq,
			}
			tripSequenceKeys = append(tripSequenceKeys, tripSequenceKey)
		}
	}

	slog.Debug("finished building tripSequenceKeys",
		"input_map_size", len(tripStopKeySequenceMap),
		"output_slice_size", len(tripSequenceKeys),
	)
	tripStopKeyToPrevStopInfoMap := make(map[TripStopKey]PrevStopInfo, len(tripSequenceKeys))
	tripStopKeyToPrevStopInfoMap, err = s.tripRepo.GetPrevStopInfo(ctx, tripSequenceKeys)
	if err != nil {
		slog.Error("Failed to get previous stop details", "err", err)
		return []TripTrainLocationAPI{}, err
	}
	slog.Debug(
		"Successfully retrieved previous stop details",
		"mapped_count",
		len(tripStopKeyToPrevStopInfoMap),
	)

	slog.Debug("Assembling initial in-memory train contexts")
	trainContexts := buildTrainContexts(
		nextStops,
		tripStopKeySequenceMap,
		tripStopKeyToPrevStopInfoMap,
		time.Now(),
	)
	slog.Info("Train contexts compiled", "active_contexts_count", len(trainContexts))

	slog.Debug(
		"Fetching shape sequences bounding ranges from repository",
		"contexts_count",
		len(trainContexts),
	)
	shapeSequences, err := s.tripRepo.GetShapeSequences(ctx, trainContexts)
	if err != nil {
		slog.Error("Failed to get shape sequences", "err", err)
		return []TripTrainLocationAPI{}, err
	}
	slog.Debug("Successfully retrieved shape sequences bounds", "mapped_count", len(shapeSequences))

	slog.Debug("Interpolating current shape sequence numbers using time progress")
	calculateCurrentShapeSequences(trainContexts, shapeSequences)

	slog.Debug("Batch looking up target latitude/longitude coordinates from repository")
	tripStopTrainCoordinates := make(map[TripStopKey]TrainCoordinates, len(trainContexts))
	tripStopTrainCoordinates, err = s.tripRepo.GetCoordinatesByShapeSequence(ctx, trainContexts)
	if err != nil {
		slog.Error("Failed to look up shape path coordinates", "err", err)
		return []TripTrainLocationAPI{}, err
	}
	slog.Debug(
		"Successfully retrieved exact physical track coordinates",
		"mapped_count",
		len(tripStopTrainCoordinates),
	)

	slog.Debug("Building final API payload and calculating travel bearings")
	trainLocations := make([]TripTrainLocationAPI, 0, len(trainContexts))
	skippedCoordsCount := 0

	currentCoordsMap, prevCoordsMap, err := s.tripRepo.GetPositionsWithHistory(ctx, trainContexts)
	if err != nil {
		return nil, err
	}

	for _, train := range trainContexts {
		mapKey := TripStopKey{
			ShortTripID: train.ShortTripID,
			StopID:      train.NextStopID,
		}

		currentCoords, hasCurrent := currentCoordsMap[mapKey]
		if !hasCurrent {
			skippedCoordsCount++
			slog.Debug("Skipping train: missing track coordinates",
				"short_trip_id", train.ShortTripID,
				"next_stop_id", train.NextStopID,
			)
			continue
		}
		prevCoords, hasPrev := prevCoordsMap[mapKey]
		bearing := 0.0

		if train.CurrentShapeSequence <= 1 {
			if strings.HasSuffix(train.NextStopID, "S") ||
				strings.Contains(train.ShortTripID, "..S") {
				bearing = 180.0
			} else {
				bearing = 0.0
			}
		} else if hasPrev && (currentCoords.Lat != prevCoords.Lat || currentCoords.Lon != prevCoords.Lon) {
			deltaLat := currentCoords.Lat - prevCoords.Lat
			deltaLon := currentCoords.Lon - prevCoords.Lon

			radians := math.Atan2(deltaLon, deltaLat)
			bearing = radians * (180.0 / math.Pi)

			if bearing < 0 {
				bearing += 360.0
			}
		}

		trainLocations = append(trainLocations, TripTrainLocationAPI{
			TripID:     train.ShortTripID,
			RouteID:    train.RouteID,
			Lat:        currentCoords.Lat,
			Lon:        currentCoords.Lon,
			Bearing:    bearing,
			NextStopID: train.NextStopID,
		})
	}

	if skippedCoordsCount > 0 {
		slog.Warn(
			"Some trains were omitted due to missing track coordinate points",
			"omitted_count",
			skippedCoordsCount,
		)
	}

	slog.Info("GetTripPositions execution complete", "final_payload_count", len(trainLocations))

	return trainLocations, nil
}

func parseScheduledDepartureTimeToTime(
	scheduledTimeStr string,
	realTimeBaseline time.Time,
) time.Time {
	fallbackTime := realTimeBaseline.Add(-2 * time.Minute)
	if scheduledTimeStr == "" {
		return fallbackTime
	}

	parts := strings.Split(scheduledTimeStr, ":")
	if len(parts) != 3 {
		return fallbackTime
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return fallbackTime
	}
	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return fallbackTime
	}
	seconds, err := strconv.Atoi(parts[2])
	if err != nil {
		return fallbackTime
	}

	loc := fallbackTime.Location()
	midnight := time.Date(
		fallbackTime.Year(),
		fallbackTime.Month(),
		fallbackTime.Day(),
		0, 0, 0, 0, loc,
	)

	durationOffset := time.Duration(
		hours,
	)*time.Hour + time.Duration(
		minutes,
	)*time.Minute + time.Duration(
		seconds,
	)*time.Second
	return midnight.Add(durationOffset)
}

func buildTrainContexts(
	nextStops []nextstops.NextStop,
	tripStopKeySequenceMap map[TripStopKey]int64,
	tripStopKeyToPrevStopInfoMap map[TripStopKey]PrevStopInfo,
	now time.Time,
) []TrainContext {
	trainContexts := make([]TrainContext, 0, len(nextStops))

	for _, nextStop := range nextStops {

		lookupKey := TripStopKey{
			ShortTripID: nextStop.ShortTripID,
			StopID:      nextStop.StopID,
		}

		prevLookupKey := TripStopKey{
			ShortTripID: nextStop.ShortTripID,
			StopID:      "",
		}
		nextSequence, hasSequence := tripStopKeySequenceMap[lookupKey]
		prevInfo, found := tripStopKeyToPrevStopInfoMap[prevLookupKey]
		if !found || !hasSequence {
			continue
		}

		parsedArrivalTime, err := time.Parse(time.RFC3339, nextStop.ArrivalTime)
		if err != nil {
			slog.Error("failed to parse arrival time string", "err", err, "val", nextStop.ArrivalTime)
			continue
		}

		parsedDepartureTime := parseScheduledDepartureTimeToTime(
			prevInfo.PrevDepartureTime,
			parsedArrivalTime,
		)
		totalTime := parsedArrivalTime.Sub(parsedDepartureTime)
		elapsedTime := now.Sub(parsedDepartureTime)

		progress := 0.0
		if totalTime.Seconds() > 0 {
			progress = elapsedTime.Seconds() / totalTime.Seconds()
		}

		if progress < 0.0 {
			progress = 0.0
		} else if progress > 1.0 {
			progress = 1.0
		}

		ctxRecord := TrainContext{
			ShortTripID:             nextStop.ShortTripID,
			RouteID:                 nextStop.RouteID,
			NextStopID:              nextStop.StopID,
			PrevStopID:              prevInfo.PrevStopID,
			NextStationStopSequence: nextSequence,
			PrevStationStopSequence: prevInfo.PrevStationStopSequence,
			NextArrivalTime:         nextStop.ArrivalTime,
			PrevDepartureTime:       prevInfo.PrevDepartureTime,
			ProgressPercentage:      progress * 100.0,
		}

		trainContexts = append(trainContexts, ctxRecord)
	}

	return trainContexts
}

func calculateCurrentShapeSequences(
	contexts []TrainContext,
	shapeRanges map[TripStopKey]ShapeRange,
) []TrainContext {
	for i, ctx := range contexts {
		mapKey := TripStopKey{
			ShortTripID: ctx.ShortTripID,
			StopID:      ctx.NextStopID,
		}

		ranges, found := shapeRanges[mapKey]
		if !found {
			continue
		}

		progressRatio := ctx.ProgressPercentage / 100.0
		sequenceDelta := float64(ranges.NextShapeSequence - ranges.PrevShapeSequence)
		interpolatedSequence := (sequenceDelta * progressRatio) + float64(ranges.PrevShapeSequence)

		contexts[i].CurrentShapeSequence = int64(math.Round(interpolatedSequence))
	}

	return contexts
}
