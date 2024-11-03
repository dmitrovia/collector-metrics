package defaulthandler

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"runtime"

	"github.com/dmitrovia/collector-metrics/internal/service"
)

type DefaultHandler struct {
	serv service.Service
}

func NewDefaultHandler(s service.Service) *DefaultHandler {
	return &DefaultHandler{serv: s}
}

type ViewData struct {
	Metrics map[string]string
}

func (h *DefaultHandler) DefaultHandler(writer http.ResponseWriter, _ *http.Request) {
	data := ViewData{
		Metrics: *h.serv.GetMapStringsAllMetrics(),
	}

	_, path, _, ok := runtime.Caller(0)

	if !ok {
		writer.WriteHeader(http.StatusBadRequest)

		return
	}

	Root := filepath.Join(filepath.Dir(path), "../..")

	tmpl, err := template.ParseFiles(Root + "/html/allMetricsTemplate.html")
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		fmt.Println(err)
	} else {
		writer.Header().Set("Content-Type", "text/html")
		writer.WriteHeader(http.StatusOK)

		err = tmpl.Execute(writer, data)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			fmt.Println(err)
		}
	}
}
