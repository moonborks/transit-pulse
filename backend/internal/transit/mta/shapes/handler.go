package shapes

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"

	"github.com/moonborks/transit-pulse/internal/web"
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
	simplifyStr := r.URL.Query().Get("simplify")
	if simplifyStr == "" {
		simplifyStr = "false"
	}
	simplify, err := strconv.ParseBool(simplifyStr)
	if err != nil {
		web.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid simplify param")
		return
	}

	var shapes []Shape
	if simplify {
		shapes, err = h.shapeService.GetSimplifiedShapes(r.Context())
	} else {
		shapes, err = h.shapeService.GetAll(r.Context())
	}
	if err != nil {
		web.WriteError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"unable to retrieve shapes from table",
		)
		slog.Error("get all shapes", "err", err)
		return
	}

	if err := web.WriteJson(w, http.StatusOK, shapes); err != nil {
		slog.Error("writing response json for get all shapes")
	}
}

func (h *ShapeHandler) GetShape(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	shape, err := h.shapeService.GetShape(r.Context(), id)
	if err != nil {
		web.WriteError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"unable to retrieve all of the specified shape's sequences",
		)
		return
	}

	if err := web.WriteJson(w, http.StatusOK, shape); err != nil {
		slog.Error("writing response json")
	}
}
