// Package getmetrichandler provides handler
// to get metric value by name and type.
package getmetrichandler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/dmitrovia/collector-metrics/internal/functions/validate"
	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
	"github.com/dmitrovia/collector-metrics/internal/service"
	"github.com/gorilla/mux"
)

// validMetric - object for storing the received metric.
type validMetric struct {
	mtype string
	mname string
}

// validMetric - object for store the response.
type ansData struct {
	mvalue string
}

// GetMetricHandler - describing the handler.
type GetMetricHandler struct {
	serv service.Service
}

// NewGetMetricHandler - to create an instance
// of a handler object.
func NewGetMetricHandler(
	s service.Service,
) *GetMetricHandler {
	return &GetMetricHandler{serv: s}
}

// GetMetricHandler - main handler method.
func (h *GetMetricHandler) GetMetricHandler(
	writer http.ResponseWriter,
	req *http.Request,
) {
	var valMetr *validMetric

	var answerData *ansData

	valMetr = &validMetric{}
	getReqData(req, valMetr)

	isValid := isValidMetric(valMetr, writer)
	if !isValid {
		return
	}

	answerData = &ansData{}
	isSetAnsData := setAnswerData(
		valMetr,
		answerData,
		h)

	if isSetAnsData {
		writer.WriteHeader(http.StatusOK)

		Body := answerData.mvalue
		fmt.Fprintf(writer, "%s", Body)

		return
	}

	writer.WriteHeader(http.StatusNotFound)
}

// getReqData - receives metrics
// from the request.
func getReqData(r *http.Request, metric *validMetric) {
	metric.mname = mux.Vars(r)["metric_name"]
	metric.mtype = mux.Vars(r)["metric_type"]
}

// isValidMetric - for metric validation.
func isValidMetric(
	metric *validMetric,
	writer http.ResponseWriter,
) bool {
	var pattern string
	pattern = "^[0-9a-zA-Z/ ]{1,40}$"
	res, _ := validate.IsMatchesTemplate(metric.mname, pattern)

	if !res {
		writer.WriteHeader(http.StatusNotFound)

		return false
	}

	pattern = "^" + bizmodels.MetricsPattern + "$"
	res, _ = validate.IsMatchesTemplate(metric.mtype, pattern)

	if !res {
		writer.WriteHeader(http.StatusBadRequest)

		return false
	}

	return true
}

// setAnswerData - record the response data.
func setAnswerData(
	metric *validMetric,
	ansd *ansData,
	h *GetMetricHandler,
) bool {
	if metric.mtype == bizmodels.GaugeName {
		return GetStringValueGaugeMetric(ansd, h, metric.mname)
	} else if metric.mtype == bizmodels.CounterName {
		return GetStringValueCounterMetric(ansd, h, metric.mname)
	}

	return false
}

// GetStringValueGaugeMetric - get
// gauge metric from service.
func GetStringValueGaugeMetric(
	ansd *ansData,
	h *GetMetricHandler,
	mname string,
) bool {
	val, err := h.serv.GetValueGM(mname)
	if err != nil {
		return false
	}

	ansd.mvalue = strconv.FormatFloat(val, 'f', -1, 64)

	return true
}

// GetStringValueCounterMetric - get
// counter metric from service.
func GetStringValueCounterMetric(
	ansd *ansData,
	h *GetMetricHandler,
	mname string,
) bool {
	val, err := h.serv.GetValueCM(mname)
	if err != nil {
		return false
	}

	ansd.mvalue = strconv.FormatInt(val, 10)

	return true
}
