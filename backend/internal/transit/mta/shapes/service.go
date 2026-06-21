package shapes

import (
	"context"
	"math"
)

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

func (h *ShapeService) GetSimplifiedShapes(ctx context.Context) ([]Shape, error) {
	shapeIdMap, err := h.shapeRepo.GetAllGroupedByShapeID(ctx)
	if err != nil {
		return []Shape{}, err
	}

	shapesNum := 0
	for shapeId, points := range shapeIdMap {
		simplifiedShapes := simplifyTransitPoints(points, 0.7, 0.00055)
		shapesNum += len(simplifiedShapes)
		shapeIdMap[shapeId] = simplifiedShapes
	}
	result := make([]Shape, 0, shapesNum)
	for _, points := range shapeIdMap {
		result = append(result, points...)
	}
	return result, nil
}

func distanceSq(a, b Shape) float64 {
	dy := b.Lat - a.Lat
	dx := b.Lon - a.Lon
	return (dx * dx) + (dy * dy)
}

func simplifyTransitPoints(points []Shape, thresholdDegrees float64, minDistDegrees float64) []Shape {
	if len(points) < 3 {
		return points
	}

	result := []Shape{points[0]}
	lastKept := points[0]
	minDistSq := minDistDegrees * minDistDegrees

	for i := 1; i < len(points)-1; i++ {
		// 1. Is this point part of a tight cluster? If too close to the last saved point, skip it.
		if distanceSq(lastKept, points[i]) < minDistSq {
			continue
		}

		// 2. Is it a sharp turn? Check incoming vs outgoing heading
		bearingIn := bearing(points[i-1], points[i])
		bearingOut := bearing(points[i], points[i+1])

		if angleDiff(bearingIn, bearingOut) > thresholdDegrees {
			result = append(result, points[i])
			lastKept = points[i] // Update our spatial anchor
		}
	}

	// Always preserve the final destination point
	result = append(result, points[len(points)-1])
	return result
}

func bearing(a, b Shape) float64 {
	dy := b.Lat - a.Lat
	dx := b.Lon - a.Lon
	return math.Atan2(dy, dx) * 180 / math.Pi
}

func angleDiff(a, b float64) float64 {
	d := math.Abs(a - b)
	if d > 180 {
		d = 360 - d
	}
	return d
}
