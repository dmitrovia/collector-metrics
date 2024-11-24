package getmetricjsonhandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/dmitrovia/collector-metrics/internal/functions/validate"
	"github.com/dmitrovia/collector-metrics/internal/models/apimodels"
	"github.com/dmitrovia/collector-metrics/internal/service"
)

const metrics string = "gauge|counter"

var errGetReqDataJSON = errors.New("data is empty")

type validMetric struct {
	mtype string
	mname string
}

type ansData struct {
	mvalueFloat float64
	mvalueInt   int64
}

type GetMetricJSONHandler struct {
	serv service.Service
}

func NewGetMetricJSONHandler(s service.Service) *GetMetricJSONHandler {
	return &GetMetricJSONHandler{serv: s}
}

func (h *GetMetricJSONHandler) GetMetricJSONHandler(writer http.ResponseWriter, req *http.Request) {
	var valMetr *validMetric

	var answerData *ansData

	valMetr = new(validMetric)

	writer.Header().Set("Content-Type", "application/json")

	err := getReqDataJSON(req, valMetr)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)

		return
	}

	isValid, status := isValidMetric(req, valMetr)
	if !isValid {
		writer.WriteHeader(status)

		return
	}

	answerData = new(ansData)
	isSetAnsData, status := setAnswerDataForJSON(valMetr, answerData, h)

	if isSetAnsData {
		writer.WriteHeader(status)

		dataMarshal := apimodels.Metrics{}
		dataMarshal.ID = valMetr.mname
		dataMarshal.MType = valMetr.mtype

		if valMetr.mtype == "counter" {
			dataMarshal.Delta = &answerData.mvalueInt
		}

		if valMetr.mtype == "gauge" {
			dataMarshal.Value = &answerData.mvalueFloat
		}

		metricMarshall, err := json.Marshal(dataMarshal)
		if err != nil {
			fmt.Println("GetMetricJSONHandler->json.Marshal: %w", err)
			writer.WriteHeader(http.StatusBadRequest)

			return
		}

		_, err = writer.Write(metricMarshall)
		if err != nil {
			fmt.Println("GetMetricJSONHandler->writer.Write: %w", err)
			writer.WriteHeader(http.StatusBadRequest)

			return
		}

		return
	}

	writer.WriteHeader(status)
}

func getReqDataJSON(req *http.Request, metric *validMetric) error {
	var result apimodels.Metrics

	bodyD, err := io.ReadAll(req.Body)
	if err != nil {
		defer req.Body.Close()

		return fmt.Errorf("getReqDataJSON: %w", err)
	}

	defer req.Body.Close()

	if len(bodyD) == 0 {
		return fmt.Errorf("getReqDataJSON: %w", errGetReqDataJSON)
	}

	err = json.Unmarshal(bodyD, &result)
	if err != nil {
		return err
	}

	metric.mname = result.ID
	metric.mtype = result.MType

	return nil
}

func isValidMetric(r *http.Request, metric *validMetric) (bool, int) {
	if !validate.IsMethodPost(r.Method) {
		return false, http.StatusMethodNotAllowed
	}

	var pattern string
	pattern = "^[0-9a-zA-Z/ ]{1,40}$"
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

func setAnswerDataForJSON(metric *validMetric, ansd *ansData, h *GetMetricJSONHandler) (bool, int) {
	if metric.mtype == "gauge" {
		return setGaugeValueByType(metric, ansd, h)
	} else if metric.mtype == "counter" {
		return setCounterValueByType(metric, ansd, h)
	}

	return false, http.StatusNotFound
}

func setGaugeValueByType(metric *validMetric, ansd *ansData, h *GetMetricJSONHandler) (bool, int) {
	metricValue, err := h.serv.GetValueGaugeMetric(metric.mname)
	if err != nil {
		return false, http.StatusNotFound
	}

	ansd.mvalueFloat = metricValue

	return true, http.StatusOK
}

func setCounterValueByType(metric *validMetric, ansd *ansData, h *GetMetricJSONHandler) (bool, int) {
	metricValue, err := h.serv.GetValueCounterMetric(metric.mname)
	if err != nil {
		return false, http.StatusNotFound
	}

	ansd.mvalueInt = metricValue

	return true, http.StatusOK
}
