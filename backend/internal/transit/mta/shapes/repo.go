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

func (r *ShapeRepo) GetAll(ctx context.Context) ([]*Shape, error) {
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

	shapes := []*Shape{}

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
		shapes = append(shapes, &shape)
	}

	return shapes, nil
}

func (r *ShapeRepo) GetShape(ctx context.Context, id string) ([]*Shape, error) {
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

	shapes := []*Shape{}

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
		shapes = append(shapes, &shape)
	}

	return shapes, nil
}
