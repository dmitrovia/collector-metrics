package setmetrichandler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/dmitrovia/collector-metrics/internal/functions/validate"
	"github.com/dmitrovia/collector-metrics/internal/service"
	"github.com/gorilla/mux"
)

type SetMetricHandler struct {
	serv service.Service
}

const metrics string = "gauge|counter"

type validMetric struct {
	mtype       string
	mname       string
	mvalue      string
	mvalueFloat float64
	mvalueInt   int64
}

func NewSetMetricHandler(serv service.Service) *SetMetricHandler {
	return &SetMetricHandler{serv: serv}
}

func (h *SetMetricHandler) SetMetricHandler(writer http.ResponseWriter, req *http.Request) {
	var valm *validMetric

	var Body string

	valm = new(validMetric)

	getReqData(req, valm)

	isValid, status := isValidMetric(req, valm)
	if !isValid {
		writer.WriteHeader(status)

		return
	}

	addMetricToMemStore(h, valm)
	writer.WriteHeader(status)

	Body = "OK\n"
	fmt.Fprintf(writer, "%s", Body)
}

func getReqData(r *http.Request, m *validMetric) {
	m.mtype = mux.Vars(r)["metric_type"]
	m.mname = mux.Vars(r)["metric_name"]
	m.mvalue = mux.Vars(r)["metric_value"]
}

func addMetricToMemStore(h *SetMetricHandler, m *validMetric) {
	if m.mtype == "gauge" {
		h.serv.AddGauge(m.mname, m.mvalueFloat)
	} else if m.mtype == "counter" {
		h.serv.AddCounter(m.mname, m.mvalueInt)
	}
}

func isValidMetric(r *http.Request, metric *validMetric) (bool, int) {
	if !validate.IsMethodPost(r.Method) {
		return false, http.StatusMethodNotAllowed
	}

	var pattern string

	pattern = "^[0-9a-zA-Z/ ]{1,20}$"
	res, _ := validate.IsMatchesTemplate(metric.mname, pattern)

	if !res {
		return false, http.StatusNotFound
	}

	pattern = "^" + metrics + "$"
	res, _ = validate.IsMatchesTemplate(metric.mtype, pattern)

	if !res {
		return false, http.StatusBadRequest
	}

	if !isValidMeticValue(metric) {
		return false, http.StatusBadRequest
	}

	return true, http.StatusOK
}

func isValidMeticValue(m *validMetric) bool {
	if m.mtype == "gauge" {
		return isValidGaugeValue(m)
	} else if m.mtype == "counter" {
		return isValidCounterValue(m)
	}

	return false
}

func isValidGaugeValue(m *validMetric) bool {
	value, err := strconv.ParseFloat(m.mvalue, 64)
	if err == nil {
		m.mvalueFloat = value

		return true
	}

	return false
}

func isValidCounterValue(m *validMetric) bool {
	value, err := strconv.ParseInt(m.mvalue, 10, 64)
	if err == nil {
		m.mvalueInt = value

		return true
	}

	return false
}
