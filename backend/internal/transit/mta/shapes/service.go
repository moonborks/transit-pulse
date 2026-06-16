package shapes

type ShapeService struct {
	shapeRepo *ShapeRepo
}

func NewShapeService(sr *ShapeRepo) *ShapeService {
	return &ShapeService{shapeRepo: sr}
}

func (h *ShapeService) GetAll() {
}

func (h *ShapeService) GetShape() {
}
