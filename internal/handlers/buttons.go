package handlers

import (
	"encoding/json"
	"fmt"
	"le-grimoire/internal/middleware"
	"le-grimoire/internal/state"
	"net/http"
)

type buttonRequest struct {
	ButtonID string `json:"button_id"` // Expected values: "A", "B", "C", "D", "E", "F"
	Type     string `json:"type"`      // Expected values: "short_press", "long_press"
}

// TODO: Check if this struct still required after testing.
type pushButtonResponse struct {
	Action string            `json:"action"`
	State  state.DeviceState `json:"state"`
}

func (h *Handler) PushButtonHandler(w http.ResponseWriter, r *http.Request) {
	deviceID, ok := middleware.GetDeviceID(r.Context())
	if !ok {
		h.RespondWithStatusError(w, http.StatusUnauthorized, fmt.Errorf("device ID not found in headers"))
		return
	}

	var req buttonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondWithStatusError(w, http.StatusBadRequest, err)
		return
	}

	ds, err := h.App.DeviceRepository.GetDeviceState(deviceID)
	if err != nil {
		h.RespondWithError(w, err)
		return
	}

	action, err := h.App.StateMachine.ApplyButton(r.Context(), ds, req.ButtonID, req.Type)
	if err != nil {
		h.RespondWithError(w, err)
		return
	}

	// Invalidates cache if force refresh (Long-press E) was triggered on Reader page
	if (req.ButtonID == "E") && (req.Type == "long" || req.Type == "long_press") {
		top := ds.Top()
		if top.Type == state.PageReader {
			var chapterID int
			_, err := fmt.Sscanf(top.Params["chapter_id"], "%d", &chapterID)
			if err != nil {
				h.RespondWithError(w, err)
				return
			}

			if chapterID > 0 && h.App.FrameCache != nil {
				h.App.FrameCache.Invalidate(fmt.Sprintf("%d", chapterID))
			}
		}
	}

	if err := h.App.DeviceRepository.SaveDeviceState(ds); err != nil {
		h.RespondWithError(w, err)
		return
	}

	h.RespondWithData(w, pushButtonResponse{Action: action, State: *ds})
}
