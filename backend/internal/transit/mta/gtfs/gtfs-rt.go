package gtfs

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
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
	if err != nil {
		slog.Error("unmarshal feed", "err", err)
		return
	}

	writeProtoMsgToCache(ctx, rdb, &feed, feedLink)
}

type Train struct {
	StopID      *string   `json:"stop_id"`
	ShortTripID *string   `json:"short_trip_id"`
	RouteID     *string   `json:"route_id"`
	ArrivalTime time.Time `json:"arrival_time"`
}

func writeProtoMsgToCache(ctx context.Context, rdb *redis.Client, feed *pb.FeedMessage, feedLink string) error {
	pipe := rdb.Pipeline()

	unquotedLink, err := url.PathUnescape(feedLink)
	if err != nil {
		unquotedLink = feedLink
	}

	feedName := "gtfs"
	if idx := strings.LastIndex(unquotedLink, "gtfs"); idx != -1 {
		feedName = unquotedLink[idx:]
	}

	liveKey := "mta:active:stops:" + feedName
	shadowKey := "mta:active:stops:shadow:" + feedName
	hasUpdates := false

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
			ShortTripID: entity.TripUpdate.Trip.TripId,
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
		if train.ShortTripID != nil {
			trip = *train.ShortTripID
		}

		fieldKey := route + "_" + trip

		// Queue the writes onto the clean shadow collection
		pipe.HSet(ctx, shadowKey, fieldKey, data)
		hasUpdates = true
	}

	if hasUpdates {
		pipe.Expire(ctx, shadowKey, 45*time.Second)
		pipe.Rename(ctx, shadowKey, liveKey)

		_, err := pipe.Exec(ctx)
		if err != nil {
			slog.Error("flushing train pipeline to redis cache", "err", err)
			return err
		}
	}

	return nil
}
