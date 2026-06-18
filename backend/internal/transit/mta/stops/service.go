package stops

import "context"

type StopService struct {
	stopRepo *StopRepo
}

func NewStopService(sr *StopRepo) *StopService {
	return &StopService{stopRepo: sr}
}

func (s *StopService) GetAll(ctx context.Context) ([]Stop, error) {
	return s.stopRepo.GetAll(ctx)
}

func (s *StopService) GetStop(ctx context.Context, id string) (Stop, error) {
	return s.stopRepo.GetStop(ctx, id)
}
