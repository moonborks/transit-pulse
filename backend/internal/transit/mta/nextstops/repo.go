package nextstop

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
	iter := rc.rdb.Scan(ctx, 0, "*_*", 0).Iterator()

	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return []NextStop{}, err
	}

	vals, err := rc.rdb.MGet(ctx, keys...).Result()
	if err != nil {
		return []NextStop{}, err
	}

	spots := make([]NextStop, 0, len(vals))

	for _, v := range vals {
		if v == nil {
			continue
		}

		var s NextStop
		err := json.Unmarshal([]byte(v.(string)), &s)
		if err != nil {
			continue
		}

		spots = append(spots, s)
	}

	return spots, nil
}
