package shapes

import "context"

type ShapeService struct {
	shapeRepo *ShapeRepo
}

func NewShapeService(sr *ShapeRepo) *ShapeService {
	return &ShapeService{shapeRepo: sr}
}

func (h *ShapeService) GetAll(ctx context.Context) ([]Shape, error) {
	return h.shapeRepo.GetAll(ctx)
}

func (h *ShapeService) GetShape(ctx context.Context, id string) ([]Shape, error) {
	return h.shapeRepo.GetShape(ctx, id)
}
