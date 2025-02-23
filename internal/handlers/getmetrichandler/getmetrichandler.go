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

	isValid, status := isValidMetric(valMetr)
	if !isValid {
		writer.WriteHeader(status)

		return
	}

	writer.WriteHeader(status)

	answerData = &ansData{}
	isSetAnsData, status := setAnswerData(
		valMetr,
		answerData,
		h)
	writer.WriteHeader(status)

	if isSetAnsData {
		Body := answerData.mvalue
		fmt.Fprintf(writer, "%s", Body)

		return
	}
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
) (bool, int) {
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

	return true, http.StatusOK
}

// setAnswerData - record the response data.
func setAnswerData(
	metric *validMetric,
	ansd *ansData,
	h *GetMetricHandler,
) (bool, int) {
	if metric.mtype == bizmodels.GaugeName {
		return GetStringValueGaugeMetric(ansd, h, metric.mname)
	} else if metric.mtype == bizmodels.CounterName {
		return GetStringValueCounterMetric(ansd, h, metric.mname)
	}

	return false, http.StatusNotFound
}

// GetStringValueGaugeMetric - get
// gauge metric from service.
func GetStringValueGaugeMetric(
	ansd *ansData,
	h *GetMetricHandler,
	mname string,
) (bool, int) {
	val, err := h.serv.GetValueGM(mname)
	if err != nil {
		return false, http.StatusNotFound
	}

	ansd.mvalue = strconv.FormatFloat(val, 'f', -1, 64)

	return true, http.StatusOK
}

// GetStringValueCounterMetric - get
// counter metric from service.
func GetStringValueCounterMetric(
	ansd *ansData,
	h *GetMetricHandler,
	mname string,
) (bool, int) {
	val, err := h.serv.GetValueCM(mname)
	if err != nil {
		return false, http.StatusNotFound
	}

	ansd.mvalue = strconv.FormatInt(val, 10)

	return true, http.StatusOK
}
