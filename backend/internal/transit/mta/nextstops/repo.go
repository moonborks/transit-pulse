package nextstops

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

type NextStopRepo struct {
	rdb *redis.Client
}

func NewNextStopRepo(rdb *redis.Client) *NextStopRepo {
	return &NextStopRepo{rdb: rdb}
}

func (rc *NextStopRepo) GetAllNextStops(ctx context.Context) ([]NextStop, error) {
	keys, err := rc.rdb.Keys(ctx, "mta:active:stops:gtfs*").Result()
	if err != nil {
		return []NextStop{}, err
	}

	if len(keys) == 0 {
		return []NextStop{}, nil
	}

	var allVals []string

	for _, key := range keys {
		vals, err := rc.rdb.HVals(ctx, key).Result()
		if err != nil {
			continue
		}
		allVals = append(allVals, vals...)
	}

	stops := make([]NextStop, 0, len(allVals))
	for _, v := range allVals {
		var s NextStop
		if err := json.Unmarshal([]byte(v), &s); err != nil {
			continue
		}
		stops = append(stops, s)
	}

	return stops, nil
}
