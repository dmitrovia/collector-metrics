// Package defaulthandler provides handler
// displays a list of all metrics in html format.
package defaulthandler

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/dmitrovia/collector-metrics/internal/service"
)

//go:embed web
var metricsTemplate embed.FS

// DefaultHandler - describing the handler.
type DefaultHandler struct {
	serv service.Service
}

// NewDefaultHandler - to create an instance
// of a handler object.
func NewDefaultHandler(s service.Service) *DefaultHandler {
	return &DefaultHandler{serv: s}
}

// ViewData - object for mapping
// metrics into a template.
type ViewData struct {
	Metrics map[string]string
}

// DefaultHandler - main handler method.
func (h *DefaultHandler) DefaultHandler(
	writer http.ResponseWriter, _ *http.Request,
) {
	mapMetrics := make(map[string]string)

	counters, err := h.serv.GetAllCounters()
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)

		return
	}

	// comment - need to add context
	gauges, err := h.serv.GetAllGauges()
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)

		return
	}

	for key, value := range counters {
		mapMetrics[key] = strconv.FormatInt(value.Value, 10)
	}

	for key, value := range gauges {
		mapMetrics[key] = strconv.FormatFloat(
			value.Value, 'f', -1, 64)
	}

	data := ViewData{
		Metrics: mapMetrics,
	}

	tmpl, err := template.ParseFS(metricsTemplate,
		"web/template/allMetricsTemplate.html")
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		fmt.Println("DefaultHandler->template.ParseF: %w", err)
	} else {
		writer.Header().Set("Content-Type", "text/html")
		writer.WriteHeader(http.StatusOK)

		err = tmpl.Execute(writer, data)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			fmt.Println("DefaultHandler->tmpl.Execute: %w", err)
		}
	}
}
