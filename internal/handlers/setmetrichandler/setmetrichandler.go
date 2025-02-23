// Package setmetrichandler provides handler
// to receive one metric.
package setmetrichandler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/dmitrovia/collector-metrics/internal/functions/validate"
	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
	"github.com/dmitrovia/collector-metrics/internal/service"
	"github.com/gorilla/mux"
)

// SetMetricHandler - describing the handler.
type SetMetricHandler struct {
	serv service.Service
}

// validMetric - object for storing the received metric.
type validMetric struct {
	mtype       string
	mname       string
	mvalue      string
	mvalueFloat float64
	mvalueInt   int64
}

// NewSetMetricHandler - to create an instance
// of a handler object.
func NewSetMetricHandler(
	serv service.Service,
) *SetMetricHandler {
	return &SetMetricHandler{serv: serv}
}

// SetMetricHandler - main handler method.
func (h *SetMetricHandler) SetMetricHandler(
	writer http.ResponseWriter, req *http.Request,
) {
	var valm *validMetric

	var Body string

	valm = &validMetric{}

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

// getReqData - receives data
// from the request.
func getReqData(r *http.Request, m *validMetric) {
	m.mtype = mux.Vars(r)["metric_type"]
	m.mname = mux.Vars(r)["metric_name"]
	m.mvalue = mux.Vars(r)["metric_value"]
}

// addMetricToMemStore - adds the validated
// metric to the memory.
func addMetricToMemStore(
	handler *SetMetricHandler, metr *validMetric,
) {
	if metr.mtype == bizmodels.GaugeName {
		err := handler.serv.AddGauge(metr.mname, metr.mvalueFloat)
		if err != nil {
			fmt.Println(
				"addMetricToMemStore->AddGauge: %w",
				err)
		}
	} else if metr.mtype == bizmodels.CounterName {
		res, err := handler.serv.AddCounter(
			metr.mname, metr.mvalueInt)
		if err != nil {
			fmt.Println("addMetricToMemStore->AddCounter: %w",
				err)
		}

		metr.mvalueInt = res.Value
	}
}

// isValidMetric - for metric validation.
func isValidMetric(
	r *http.Request,
	metric *validMetric,
) (bool, int) {
	if !validate.IsMethodPost(r.Method) {
		return false, http.StatusMethodNotAllowed
	}

	var pattern string

	pattern = "^[0-9a-zA-Z/ ]{1,40}$"
	res, _ := validate.IsMatchesTemplate(metric.mname, pattern)

	if !res {
		return false, http.StatusNotFound
	}

	pattern = "^" + bizmodels.MetricsPattern + "$"
	res, _ = validate.IsMatchesTemplate(metric.mtype, pattern)

	if !res {
		return false, http.StatusBadRequest
	}

	if !isValidMeticValue(metric) {
		return false, http.StatusBadRequest
	}

	return true, http.StatusOK
}

// isValidGaugeValue - for metric validation
// using parsing.
func isValidMeticValue(m *validMetric) bool {
	if m.mtype == bizmodels.GaugeName {
		return isValidGaugeValue(m)
	} else if m.mtype == bizmodels.CounterName {
		return isValidCounterValue(m)
	}

	return false
}

// isValidGaugeValue - for gauge metric validation
// and to write the parsed value.
func isValidGaugeValue(m *validMetric) bool {
	value, err := strconv.ParseFloat(m.mvalue, 64)
	if err == nil {
		m.mvalueFloat = value

		return true
	}

	return false
}

// isValidCounterValue - for counter metric validation
// and to write the parsed value.
func isValidCounterValue(m *validMetric) bool {
	value, err := strconv.ParseInt(m.mvalue, 10, 64)
	if err == nil {
		m.mvalueInt = value

		return true
	}

	return false
}
