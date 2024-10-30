package getmetrichandler

import (
	"fmt"
	"net/http"

	"github.com/dmitrovia/collector-metrics/internal/functions/validate"
	"github.com/dmitrovia/collector-metrics/internal/service"
	"github.com/gorilla/mux"
)

const metrics string = "gauge|counter"

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

func NewGetMetricHandler(s service.Service) *GetMetricHandler {
	return &GetMetricHandler{serv: s}
}

func (h *GetMetricHandler) GetMetricHandler(writer http.ResponseWriter, req *http.Request) {
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
	isSetAnsData, status := setAnswerData(valMetr, answerData, h)

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

func isValidMetric(r *http.Request, metric *validMetric) (bool, int) {
	if !validate.IsMethodGet(r.Method) {
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

	return true, http.StatusOK
}

func setAnswerData(metric *validMetric, ansd *ansData, h *GetMetricHandler) (bool, int) {
	if metric.mtype == "gauge" {
		return setValueByType(metric, ansd, h.serv.GetStringValueGaugeMetric)
	} else if metric.mtype == "counter" {
		return setValueByType(metric, ansd, h.serv.GetStringValueCounterMetric)
	}

	return false, http.StatusNotFound
}

func setValueByType(metric *validMetric, ansd *ansData, getFunction func(string) (string, error)) (bool, int) {
	metricStringValue, err := getFunction(metric.mname)
	if err != nil {
		return false, http.StatusNotFound
	}

	ansd.mvalue = metricStringValue

	return true, http.StatusOK
}
