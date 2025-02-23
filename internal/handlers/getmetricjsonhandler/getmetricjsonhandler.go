// Package getmetricjsonhandler provides handler
// to get metric by name and type in json format.
package getmetricjsonhandler

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

var errGetReqDataJSON = errors.New("data is empty")

// GetMetricJSONHandler - describing the handler.
type GetMetricJSONHandler struct {
	serv service.Service
}

// NewGetMJSONHandler - to create an instance
// of a handler object.
func NewGetMJSONHandler(
	s service.Service,
) *GetMetricJSONHandler {
	return &GetMetricJSONHandler{serv: s}
}

// GetMetricJSONHandler - main handler method.
func (h *GetMetricJSONHandler) GetMetricJSONHandler(
	writer http.ResponseWriter, req *http.Request,
) {
	writer.Header().Set("Content-Type", "application/json")

	met, err := getReqDataJSON(req)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)

		return
	}

	isValid, status := isValidMetric(met)
	if !isValid {
		writer.WriteHeader(status)

		return
	}

	status, err = writeAns(writer, met, h)
	if err != nil {
		fmt.Println("GetMetricJSONHandler->getAns: %w",
			err)
		writer.WriteHeader(status)

		return
	}

	writer.WriteHeader(status)
}

// getReqDataJSON - receives metrics
// from the request.
func getReqDataJSON(req *http.Request,
) (*apimodels.Metrics, error) {
	var result apimodels.Metrics

	bodyD, err := io.ReadAll(req.Body)
	if err != nil {
		defer req.Body.Close()

		return nil, fmt.Errorf("getReqDataJSON: %w", err)
	}

	defer req.Body.Close()

	if len(bodyD) == 0 {
		return nil, fmt.Errorf("getReqDataJSON: %w",
			errGetReqDataJSON)
	}

	err = json.Unmarshal(bodyD, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// isValidMetric - for metric validation.
func isValidMetric(metric *apimodels.Metrics,
) (bool, int) {
	var pattern string
	pattern = "^[0-9a-zA-Z/ ]{1,40}$"
	res, _ := validate.IsMatchesTemplate(metric.ID, pattern)

	if !res {
		return false, http.StatusNotFound
	}

	pattern = "^" + bizmodels.MetricsPattern + "$"
	res, _ = validate.IsMatchesTemplate(metric.MType, pattern)

	if !res {
		return false, http.StatusBadRequest
	}

	return true, http.StatusOK
}

// writeAns - writes the response
// in json format to the response body.
// First, the resulting validated metric
// is recorded in the service.
func writeAns(
	writer http.ResponseWriter,
	metric *apimodels.Metrics,
	hand *GetMetricJSONHandler,
) (int, error) {
	if metric.MType == bizmodels.CounterName {
		val, err := getCounterValueToAnswer(metric.ID, hand)
		if err != nil {
			return http.StatusNotFound, err
		}

		metric.Delta = val
	}

	if metric.MType == bizmodels.GaugeName {
		val, err := getGaugeValueToAnswer(metric.ID, hand)
		if err != nil {
			return http.StatusNotFound, err
		}

		metric.Value = val
	}

	metricMarshall, err := json.Marshal(metric)
	if err != nil {
		return http.StatusBadRequest, err
	}

	_, err = writer.Write(metricMarshall)
	if err != nil {
		return http.StatusBadRequest,
			fmt.Errorf("writeAns->Write %w", err)
	}

	return http.StatusOK, nil
}

// getGaugeValueToAnswer - get gauge metric from service.
func getGaugeValueToAnswer(
	metricID string,
	h *GetMetricJSONHandler,
) (*float64, error) {
	metricValue, err := h.serv.GetValueGM(metricID)
	if err != nil {
		return nil,
			fmt.Errorf("setGaugeValueToAnswer->GetValueGM %w", err)
	}

	return &metricValue, nil
}

// getGaugeValueToAnswer - get gauge metric from service.
func getCounterValueToAnswer(
	metricID string,
	h *GetMetricJSONHandler,
) (*int64, error) {
	metricValue, err := h.serv.GetValueCM(metricID)
	if err != nil {
		return nil,
			fmt.Errorf("setCounterValueToAnswer->GetValueCM %w",
				err)
	}

	return &metricValue, nil
}
