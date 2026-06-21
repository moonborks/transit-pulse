package gtfs

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"

	pb "github.com/moonborks/transit-pulse/internal/transit/mta/gtfs/proto"
)

func FetchRealtimeFeed(ctx context.Context, rdb *redis.Client, feedLink string) {
	resp, err := http.Get(feedLink)
	if err != nil {
		slog.Error("retrieving gtfs-rt data from mta endpoint", "err", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("reading gtfs-rt data from mta endpoint", "err", err)
		return
	}

	var feed pb.FeedMessage
	err = proto.Unmarshal(body, &feed)

	writeProtoMsgToCache(ctx, rdb, &feed)
}

type Train struct {
	StopID      *string   `json:"stop_id"`
	TripID      *string   `json:"trip_id"`
	RouteID     *string   `json:"route_id"`
	ArrivalTime time.Time `json:"arrival_time"`
}

func writeProtoMsgToCache(ctx context.Context, rdb *redis.Client, feed *pb.FeedMessage) error {
	for _, entity := range feed.Entity {

		if entity.TripUpdate == nil || len(entity.TripUpdate.StopTimeUpdate) == 0 {
			continue
		}

		stu := entity.TripUpdate.StopTimeUpdate[0]

		var arrival time.Time
		if stu.Arrival != nil && stu.Arrival.Time != nil {
			arrival = time.Unix(*stu.Arrival.Time, 0)
		} else {
			continue
		}

		train := Train{
			// stop id -> future stop coordinates in stops table parent_station (from stops.txt)
			// and the direction the train is moving to (‘N’ or ‘S’). For example, a northbound trip Hunts Point Ave stop is 613N.
			StopID:      stu.StopId,
			TripID:      entity.TripUpdate.Trip.TripId,
			RouteID:     entity.TripUpdate.Trip.RouteId,
			ArrivalTime: arrival,
		}

		data, err := json.Marshal(train)
		if err != nil {
			slog.Error("marshal train", "err", err)
			continue
		}

		var route, trip string
		if train.RouteID != nil {
			route = *train.RouteID
		}
		if train.TripID != nil {
			trip = *train.TripID
		}

		key := route + "_" + trip

		err = rdb.Set(ctx, key, data, 30*time.Second).Err()
		if err != nil {
			slog.Error("adding train data to valkey cache", "err", err)
		}
	}
	return nil
}
