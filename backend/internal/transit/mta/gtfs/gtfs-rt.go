package gtfs

import (
	"io"
	"log/slog"
	"net/http"

	"google.golang.org/protobuf/proto"

	pb "github.com/moonborks/transit-pulse/internal/transit/mta/gtfs/proto"
)

func FetchRealtimeFeed(feedLink string) (*pb.FeedMessage, error) {
	resp, err := http.Get(feedLink)
	if err != nil {
		slog.Error("retrieving gtfs-rt data from mta endpoint", "err", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("reading gtfs-rt data from mta endpoint", "err", err)
		return nil, err
	}

	var feed pb.FeedMessage
	err = proto.Unmarshal(body, &feed)

	return &feed, err
}
