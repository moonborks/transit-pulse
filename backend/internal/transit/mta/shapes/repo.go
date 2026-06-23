package shapes

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ShapeRepo struct {
	db *pgxpool.Pool
}

func NewShapeRepo(db *pgxpool.Pool) *ShapeRepo {
	return &ShapeRepo{db: db}
}

func (r *ShapeRepo) GetAll(ctx context.Context) ([]Shape, error) {
	stmt := `
		SELECT
			id, sequence, lat, lon
		FROM
			shapes
	`

	rows, err := r.db.Query(ctx, stmt)
	if err != nil {
		slog.Error("query shapes table", "err", err)
		return nil, err
	}
	defer rows.Close()

	shapes := []Shape{}

	for rows.Next() {
		var shape Shape
		if err := rows.Scan(
			&shape.ID,
			&shape.Sequence,
			&shape.Lat,
			&shape.Lon,
		); err != nil {
			slog.Error("retrieving particular row from shapes table", "err", err)
		}
		shapes = append(shapes, shape)
	}

	return shapes, nil
}

func (r *ShapeRepo) GetShape(ctx context.Context, id string) ([]Shape, error) {
	stmt := `
		SELECT
			id, sequence, lat, lon
		FROM
			shapes
		WHERE
			id = $1
	`

	rows, err := r.db.Query(ctx, stmt, id)
	if err != nil {
		slog.Error("query shapes table", "err", err)
		return nil, err
	}
	defer rows.Close()

	shapes := []Shape{}

	for rows.Next() {
		var shape Shape
		if err := rows.Scan(
			&shape.ID,
			&shape.Sequence,
			&shape.Lat,
			&shape.Lon,
		); err != nil {
			slog.Error("retrieving particular row from shapes table", "err", err)
		}
		shapes = append(shapes, shape)
	}

	return shapes, nil
}

func (r *ShapeRepo) GetAllGroupedByShapeID(ctx context.Context) (map[string][]Shape, error) {
	rows, err := r.db.Query(ctx, `
        SELECT id, sequence, lat, lon
        FROM shapes
        ORDER BY id, sequence
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	grouped := make(map[string][]Shape)
	for rows.Next() {
		var s Shape
		if err := rows.Scan(&s.ID, &s.Sequence, &s.Lat, &s.Lon); err != nil {
			return nil, err
		}
		grouped[s.ID] = append(grouped[s.ID], s)
	}

	return grouped, rows.Err()
}

func (r *ShapeRepo) GetShapesGivenTargetShapeKey(
	ctx context.Context,
	tripTargetSequenceMap map[string]TargetShapeKey,
) (map[TargetShapeKey]Shape, error) {
	coordMap := make(map[TargetShapeKey]Shape)
	if len(tripTargetSequenceMap) == 0 {
		return coordMap, nil
	}

	shapeIDs := make([]string, 0, len(tripTargetSequenceMap)*2)
	sequences := make([]int32, 0, len(tripTargetSequenceMap)*2)

	for _, targetKey := range tripTargetSequenceMap {

		shapeIDs = append(shapeIDs, targetKey.ID)
		sequences = append(sequences, int32(targetKey.Sequence))

		prevSeq := targetKey.Sequence - 1
		if prevSeq < 1 {
			prevSeq = 1
		}
		shapeIDs = append(shapeIDs, targetKey.ID)
		sequences = append(sequences, int32(prevSeq))
	}

	query := `
		SELECT id, sequence, lat, lon
		FROM shapes
		WHERE (id, sequence) IN (
			SELECT * FROM UNNEST($1::text[], $2::int[])
		);
	`

	rows, err := r.db.Query(ctx, query, shapeIDs, sequences)
	if err != nil {
		slog.Error("pgx unnest query failed for exact shape rows", "err", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var shape Shape
		if err := rows.Scan(&shape.ID, &shape.Sequence, &shape.Lat, &shape.Lon); err != nil {
			slog.Error("scanning exact shape coordinate row failed", "err", err)
			return nil, err
		}

		mapKey := TargetShapeKey{ID: shape.ID, Sequence: shape.Sequence}
		coordMap[mapKey] = shape
	}

	return coordMap, nil
}
