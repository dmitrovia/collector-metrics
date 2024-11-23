package setmetricsjsonhandler

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

type SetMetricJSONHandler struct {
	serv service.Service
}

const metrics string = "gauge|counter"

var errGetReqDataJSON = errors.New("data is empty")

func NewSetMetricsJSONHandler(serv service.Service) *SetMetricJSONHandler {
	return &SetMetricJSONHandler{serv: serv}
}

func DeSerialize(slice interface{}, r io.Reader) error {
	e := json.NewDecoder(r)

	return e.Decode(slice)
}

func (h *SetMetricJSONHandler) SetMetricsJSONHandler(writer http.ResponseWriter, req *http.Request) {
	var gauges map[string]bizmodels.Gauge

	var counters map[string]bizmodels.Counter

	writer.Header().Set("Content-Type", "application/json")

	gauges = make(map[string]bizmodels.Gauge)
	counters = make(map[string]bizmodels.Counter)

	err := getReqJSONData(req, gauges, counters)
	if err != nil {
		fmt.Println(err)
		writer.WriteHeader(http.StatusBadRequest)

		return
	}

	err = addMetricToMemStore(h, gauges, counters)
	if err != nil {
		fmt.Println(err)
		writer.WriteHeader(http.StatusBadRequest)

		return
	}

	marshal := formResponeBody(h)

	metricsMarshall, err := json.Marshal(marshal)
	if err != nil {
		fmt.Println(err)
		writer.WriteHeader(http.StatusBadRequest)

		return
	}

	writer.WriteHeader(http.StatusOK)

	_, err = writer.Write(metricsMarshall)
	if err != nil {
		fmt.Println(err)
		writer.WriteHeader(http.StatusBadRequest)

		return
	}
}

func formResponeBody(handler *SetMetricJSONHandler) *apimodels.ArrMetrics {
	tmpGauges := handler.serv.GetAllGauges()
	tmpCounters := handler.serv.GetAllGauges()
	marshal := make(apimodels.ArrMetrics, 0, len(*tmpGauges)+len(*tmpCounters))

	for _, vmr := range *tmpGauges {
		tmp := apimodels.Metrics{}
		tmp.ID = vmr.Name
		tmp.MType = "gauge"
		tmp.Value = &vmr.Value

		marshal = append(marshal, tmp)
	}

	for _, vmr := range *tmpCounters {
		tmp := apimodels.Metrics{}
		tmp.ID = vmr.Name
		tmp.MType = "counter"
		tmp.Value = &vmr.Value

		marshal = append(marshal, tmp)
	}

	return &marshal
}

func getReqJSONData(req *http.Request, gauges map[string]bizmodels.Gauge, counters map[string]bizmodels.Counter) error {
	var results apimodels.ArrMetrics

	/*err := DeSerialize(&result, req.Body)
	if err != nil {
		return fmt.Errorf("getReqJSONData->DeSerialize: %w", err)
	}
	defer req.Body.Close()*/

	bodyD, err := io.ReadAll(req.Body)
	if err != nil {
		defer req.Body.Close()

		return fmt.Errorf("getReqJSONData: %w", err)
	}

	defer req.Body.Close()

	if len(bodyD) == 0 {
		return fmt.Errorf("getReqJSONData: %w", errGetReqDataJSON)
	}

	err = json.Unmarshal(bodyD, &results)
	if err != nil {
		return fmt.Errorf("getReqJSONData->json.Unmarshal: %w", err)
	}

	for _, res := range results {
		isValid := isValidJSONMetric(req, &res)
		if isValid {
			addValidMetric(&res, gauges, counters)
		}
	}

	return nil
}

func addValidMetric(res *apimodels.Metrics, gauges map[string]bizmodels.Gauge, counters map[string]bizmodels.Counter) {
	if res.MType == "gauge" {
		gauge := new(bizmodels.Gauge)

		gauge.Name = res.ID
		gauge.Value = *res.Value
		gauges[res.ID] = *gauge
	} else if res.MType == "counter" {
		val, ok := counters[res.ID]

		var temp *bizmodels.Counter

		if ok {
			temp = new(bizmodels.Counter)
			temp.Name = val.Name
			temp.Value = val.Value + *res.Delta
			counters[res.ID] = *temp
		} else {
			counter := new(bizmodels.Counter)

			counter.Name = res.ID
			counter.Value = *res.Delta
			counters[res.ID] = *counter
		}
	}
}

func addMetricToMemStore(handler *SetMetricJSONHandler, gauges map[string]bizmodels.Gauge, counters map[string]bizmodels.Counter) error {
	err := handler.serv.AddMetrics(gauges, counters)
	if err != nil {
		return fmt.Errorf("addMetricToMemStore->handler.serv.AddMetrics: %w", err)
	}

	return nil
}

func isValidJSONMetric(r *http.Request, metric *apimodels.Metrics) bool {
	if !validate.IsMethodPost(r.Method) {
		return false
	}

	var pattern string

	pattern = "^[0-9a-zA-Z/ ]{1,40}$"
	res, _ := validate.IsMatchesTemplate(metric.ID, pattern)

	if !res {
		return false
	}

	pattern = "^" + metrics + "$"
	res, _ = validate.IsMatchesTemplate(metric.MType, pattern)

	return res
}
