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

type validMetric struct {
	mtype string
	mname string
}

type ansData struct {
	mvalue string
}

type GetMetricHandler struct {
	serv service.Service
}

func NewGetMetricHandler(
	s service.Service,
) *GetMetricHandler {
	return &GetMetricHandler{serv: s}
}

func (h *GetMetricHandler) GetMetricHandler(
	writer http.ResponseWriter,
	req *http.Request,
) {
	var valMetr *validMetric

	var answerData *ansData

	valMetr = new(validMetric)

	getReqData(req, valMetr)

	isValid, status := isValidMetric(req, valMetr)
	if !isValid {
		writer.WriteHeader(status)

		return
	}

	answerData = new(ansData)
	isSetAnsData, status := setAnswerData(
		valMetr,
		answerData,
		h)

	if isSetAnsData {
		writer.WriteHeader(status)

		Body := answerData.mvalue
		fmt.Fprintf(writer, "%s", Body)

		return
	}

	writer.WriteHeader(status)
}

func getReqData(r *http.Request, metric *validMetric) {
	metric.mname = mux.Vars(r)["metric_name"]
	metric.mtype = mux.Vars(r)["metric_type"]
}

func isValidMetric(
	r *http.Request,
	metric *validMetric,
) (bool, int) {
	if !validate.IsMethodGet(r.Method) {
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

	return true, http.StatusOK
}

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
