package trips

type TripService struct {
	tripRepo *TripRepo
}

func NewTripService(tr *TripRepo) *TripService {
	return &TripService{tripRepo: tr}
}
