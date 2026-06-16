package stops

type StopService struct {
	stopRepo *StopRepo
}

func NewStopService(sr *StopRepo) *StopService {
	return &StopService{stopRepo: sr}
}
