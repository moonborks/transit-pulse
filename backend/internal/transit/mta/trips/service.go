package trips

import "context"

type TripService struct {
	tripRepo *TripRepo
}

func NewTripService(tr *TripRepo) *TripService {
	return &TripService{tripRepo: tr}
}

func (s *TripService) GetAll(ctx context.Context) ([]*Trip, error) {
	return s.tripRepo.GetAll(ctx)
}

func (s *TripService) GetTrip(ctx context.Context, id string) (*Trip, error) {
	return s.tripRepo.GetTrip(ctx, id)
}
