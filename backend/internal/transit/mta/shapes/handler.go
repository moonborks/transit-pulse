package shapes

import (
	"net/http"

	"github.com/go-chi/chi"
)

type ShapeHandler struct {
	shapeService *ShapeService
}

func NewShapeHandler(ss *ShapeService) *ShapeHandler {
	return &ShapeHandler{shapeService: ss}
}

func ShapeRoutes(h *ShapeHandler) http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.GetAll)
	r.Get("/{id}", h.GetShape)
	return r
}

func (h *ShapeHandler) GetAll(w http.ResponseWriter, r *http.Request) {
}

func (h *ShapeHandler) GetShape(w http.ResponseWriter, r *http.Request) {
}
