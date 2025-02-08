package setmetricjsonhandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/dmitrovia/collector-metrics/internal/functions/validate"
	"github.com/dmitrovia/collector-metrics/internal/models/apimodels"
	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
	"github.com/dmitrovia/collector-metrics/internal/service"
)

type SetMJSONHandler struct {
	serv service.Service
}

var errGetReqDataJSON = errors.New("data is empty")

type validMetric struct {
	mtype       string
	mname       string
	mvalueFloat float64
	mvalueInt   int64
}

func NewSetMJH(s service.Service) *SetMJSONHandler {
	return &SetMJSONHandler{serv: s}
}

func (h *SetMJSONHandler) SetMJSONHandler(
	writer http.ResponseWriter,
	req *http.Request,
) {
	valm := new(validMetric)

	writer.Header().Set("Content-Type", "application/json")

	err := getReqJSONData(req, valm)
	if err != nil {
		fmt.Println("SetMJSONHandler->getReqJSONData: %w", err)
		writer.WriteHeader(http.StatusBadRequest)

		return
	}

	isValid, status := isValidM(valm)
	if !isValid {
		writer.WriteHeader(status)

		return
	}

	addMetricToMemStore(h, valm)

	writer.WriteHeader(status)

	dataMarshal := formResponeBody(valm)

	metricMarshall, err := json.Marshal(dataMarshal)
	if err != nil {
		fmt.Println("SetMJSONHandler->json.Marshal: %w", err)
		writer.WriteHeader(http.StatusBadRequest)

		return
	}

	_, err = writer.Write(metricMarshall)
	if err != nil {
		fmt.Println("SetMJSONHandler->writer.Write: %w", err)
		writer.WriteHeader(http.StatusBadRequest)

		return
	}
}

func formResponeBody(valm *validMetric) *apimodels.Metrics {
	dataMarshal := apimodels.Metrics{}

	dataMarshal.ID = valm.mname
	dataMarshal.MType = valm.mtype

	if valm.mtype == bizmodels.CounterName {
		dataMarshal.Delta = &valm.mvalueInt
	}

	if valm.mtype == bizmodels.GaugeName {
		dataMarshal.Value = &valm.mvalueFloat
	}

	return &dataMarshal
}

func getReqJSONData(
	req *http.Request,
	metric *validMetric,
) error {
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

func addMetricToMemStore(
	handler *SetMJSONHandler,
	vmet *validMetric,
) {
	if vmet.mtype == bizmodels.GaugeName {
		_ = handler.serv.AddGauge(vmet.mname, vmet.mvalueFloat)
	} else if vmet.mtype == bizmodels.CounterName {
		res, _ := handler.serv.AddCounter(
			vmet.mname, vmet.mvalueInt)

		vmet.mvalueInt = res.Value
	}
}

func isValidM(metric *validMetric,
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
