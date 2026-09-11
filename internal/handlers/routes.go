package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"le-grimoire/internal/models"
)

type Handler struct {
	Books    BookStore
	Renderer PageRenderer
	Devices  DeviceStore
	States   StateMachiner
	Cache    FrameStore
	Log      Logger
}

func InitRoutes(h *Handler, router *http.ServeMux) {
	router.HandleFunc("GET /health", h.HealthCheckHandler)
	router.HandleFunc("GET /simulator", h.SimulatorHandler)

	v1 := http.NewServeMux()
	v1.HandleFunc("GET /current_page", h.CurrentPageHandler)
	v1.HandleFunc("POST /push_button", h.PushButtonHandler)

	router.Handle("/api/v1/", http.StripPrefix("/api/v1", v1))
}

func (h *Handler) JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			http.Error(w, "Failed to marshal JSON", http.StatusInternalServerError)
			return
		}
		_, err = w.Write(jsonData)
		if err != nil {
			http.Error(w, "Failed to write JSON", http.StatusInternalServerError)
			return
		}
	}
}

func (h *Handler) RespondWithData(w http.ResponseWriter, data any) {
	h.JSON(w, http.StatusOK, models.NewDataResponse(data))
}

func (h *Handler) RespondWithError(w http.ResponseWriter, err error) {
	h.JSON(w, statusCodeForError(err), models.NewErrorResponse(err))
}

func (h *Handler) RespondWithStatusError(w http.ResponseWriter, code int, err error) {
	h.JSON(w, code, models.NewErrorResponse(err))
}

func statusCodeForError(err error) int {
	if err == nil {
		return http.StatusInternalServerError
	}

	if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
		return http.StatusRequestEntityTooLarge
	}

	if strings.EqualFold(strings.TrimSpace(err.Error()), "UNAUTHENTICATED") {
		return http.StatusUnauthorized
	}

	return http.StatusInternalServerError
}
