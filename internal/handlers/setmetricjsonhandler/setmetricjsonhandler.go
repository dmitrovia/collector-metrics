package setmetricjsonhandler

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

type SetMetricJSONHandler struct {
	serv service.Service
}

const metrics string = "gauge|counter"

var errGetReqDataJSON = errors.New("data is empty")

type validMetric struct {
	mtype       string
	mname       string
	mvalueFloat float64
	mvalueInt   int64
}

func NewSetMetricJSONHandler(serv service.Service) *SetMetricJSONHandler {
	return &SetMetricJSONHandler{serv: serv}
}

func (h *SetMetricJSONHandler) SetMetricJSONHandler(writer http.ResponseWriter, req *http.Request) {
	var valm *validMetric

	dataMarshal := apimodels.Metrics{}

	valm = new(validMetric)

	writer.Header().Set("Content-Type", "application/json")

	err := getReqJSONData(req, valm)
	if err != nil {
		fmt.Println(err)
		writer.WriteHeader(http.StatusBadRequest)

		return
	}

	isValid, status := isValidJSONMetric(req, valm)
	if !isValid {
		writer.WriteHeader(status)

		return
	}

	addMetricToMemStore(h, valm)
	writer.WriteHeader(status)

	dataMarshal.ID = valm.mname
	dataMarshal.MType = valm.mtype

	if valm.mtype == "counter" {
		dataMarshal.Delta = &valm.mvalueInt
	}

	if valm.mtype == "gauge" {
		dataMarshal.Value = &valm.mvalueFloat
	}

	metricMarshall, err := json.Marshal(dataMarshal)
	if err != nil {
		fmt.Println(err)
		writer.WriteHeader(http.StatusBadRequest)

		return
	}

	_, err = writer.Write(metricMarshall)
	if err != nil {
		fmt.Println(err)
		writer.WriteHeader(http.StatusBadRequest)

		return
	}
}

func getReqJSONData(req *http.Request, metric *validMetric) error {
	var result apimodels.Metrics

	bodyD, err := io.ReadAll(req.Body)
	if err != nil {
		defer req.Body.Close()

		return fmt.Errorf("getReqJSONData: %w", err)
	}

	defer req.Body.Close()

	if len(bodyD) == 0 {
		return fmt.Errorf("getReqJSONData: %w", errGetReqDataJSON)
	}

	err = json.Unmarshal(bodyD, &result)
	if err != nil {
		return err
	}

	metric.mname = result.ID
	metric.mtype = result.MType

	if result.Value != nil {
		metric.mvalueFloat = *result.Value
	}

	if result.Delta != nil {
		metric.mvalueInt = *result.Delta
	}

	return nil
}

func addMetricToMemStore(handler *SetMetricJSONHandler, vmet *validMetric) {
	if vmet.mtype == "gauge" {
		err := handler.serv.AddGauge(vmet.mname, vmet.mvalueFloat)
		if err != nil {
			fmt.Println(err)
		}
	} else if vmet.mtype == "counter" {
		res, err := handler.serv.AddCounter(vmet.mname, vmet.mvalueInt)
		if err != nil {
			fmt.Println(err)
		}

		vmet.mvalueInt = res.Value
	}
}

func isValidJSONMetric(r *http.Request, metric *validMetric) (bool, int) {
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
