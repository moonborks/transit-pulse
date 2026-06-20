package trips

import "context"

type TripService struct {
	tripRepo *TripRepo
}

func NewTripService(tr *TripRepo) *TripService {
	return &TripService{tripRepo: tr}
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
