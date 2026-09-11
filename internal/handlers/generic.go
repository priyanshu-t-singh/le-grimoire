package handlers

import (
	"html/template"
	"le-grimoire/templates"
	"net/http"
)

func (h *Handler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	h.RespondWithData(w, "OK")
}

type SimulatorPageData struct {
	BaseURL string
}

var tmpl = template.Must(template.New("simulator").Parse(templates.ClientSimulatorHTML))

func (h *Handler) SimulatorHandler(w http.ResponseWriter, r *http.Request) {
	data := SimulatorPageData{
		BaseURL: "http://" + r.Host,
	}

	err := tmpl.Execute(w, data)
	if err != nil {
		h.RespondWithError(w, err)
		return
	}
}
