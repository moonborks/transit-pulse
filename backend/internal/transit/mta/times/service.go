package times

import "context"

type TimeService struct {
	timeRepo *TimeRepo
}

func NewTimeService(tr *TimeRepo) *TimeService {
	return &TimeService{timeRepo: tr}
}

func (s *TimeService) GetAll(ctx context.Context) ([]Time, error) {
	return s.timeRepo.GetAll(ctx)
}
