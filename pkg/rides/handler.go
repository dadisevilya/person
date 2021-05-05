package rides

import (
	"github.com/gorilla/mux"
	skeleton "github.com/gtforge/go-skeleton-draft/core"
	"github.com/unrolled/render"
	"net/http"
)

type handler struct {
	skeleton.BaseHTTPHandler
	generator Generator
}

func NewHandler(render *render.Render, generator Generator) skeleton.HTTPHandler {
	return &handler{
		BaseHTTPHandler: skeleton.BaseHTTPHandler{
			Render: render,
		},
		generator: generator,
	}
}

func (h handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/rides", h.GetRide).Methods(http.MethodGet)
	router.HandleFunc("/ride/{id:[0-9]+}", h.GetRideByID).Methods(http.MethodGet)
}

func (h handler) GetRide(w http.ResponseWriter, req *http.Request) {
	output := h.generator.Create()

	h.JSON(w, http.StatusCreated, output)
}

func (h handler) GetRideByID(w http.ResponseWriter, req *http.Request) {
	output := h.generator.Create()

	h.JSON(w, http.StatusCreated, output)
}
