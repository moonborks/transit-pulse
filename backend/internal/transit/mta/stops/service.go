package stops

import (
	"context"

	nextstop "github.com/moonborks/transit-pulse/internal/transit/mta/nextstops"
)

type StopService struct {
	stopRepo      *StopRepo
	nextStopCache *nextstop.NextStopRepo
}

func NewStopService(sr *StopRepo, nsc *nextstop.NextStopRepo) *StopService {
	return &StopService{stopRepo: sr, nextStopCache: nsc}
}

func (s *StopService) GetAll(ctx context.Context) ([]Stop, error) {
	return s.stopRepo.GetAll(ctx)
}

func (s *StopService) GetStop(ctx context.Context, id string) (Stop, error) {
	return s.stopRepo.GetStop(ctx, id)
}

func (s *StopService) GetAllNextSpots(ctx context.Context) ([]nextstop.NextStop, error) {
	return s.nextStopCache.GetAllNextStops(ctx)
}
