package defaulthandler

import (
	"fmt"
	"html/template"
	"net/http"

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
	writer.WriteHeader(http.StatusNotFound)

	data := ViewData{
		Metrics: *h.serv.GetMapStringsAllMetrics(),
	}

	tmpl, err := template.ParseFiles("../../internal/html/allMetricsTemplate.html")
	if err != nil {
		fmt.Println(err)
	} else {
		err = tmpl.Execute(writer, data)
		if err != nil {
			fmt.Println(err)
		}
	}
}
