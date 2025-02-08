package sender

import (
	"crypto/hmac"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/dmitrovia/collector-metrics/internal/functions/hash"
	"github.com/dmitrovia/collector-metrics/internal/functions/validate"
	"github.com/dmitrovia/collector-metrics/internal/models/apimodels"
	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
	"github.com/dmitrovia/collector-metrics/internal/service"
)

type SetMetricJSONHandler struct {
	serv   service.Service
	params *bizmodels.InitParams
}

var errGetReqDataJSON = errors.New("data is empty")

var errHashDoesNotMatch = errors.New("hash does not match")

func NewSetMsJSONHandler(
	serv service.Service,
	par *bizmodels.InitParams,
) *SetMetricJSONHandler {
	return &SetMetricJSONHandler{serv: serv, params: par}
}

func (h *SetMetricJSONHandler) SetMetricsJSONHandler(
	writer http.ResponseWriter, req *http.Request,
) {
	writer.Header().Set("Content-Type", "application/json")

	gauges := make(map[string]bizmodels.Gauge)
	counters := make(map[string]bizmodels.Counter)

	writer.WriteHeader(http.StatusOK)

	err := getReqJSONData(req, gauges, counters, h.params)
	if err != nil {
		fmt.Println("SetMetricsJSONHandler->getReqJSONData: %w",
			err)
		writer.WriteHeader(http.StatusBadRequest)

		return
	}

	err = addMetricToMemStore(h, gauges, counters)
	if err != nil {
		fmt.Println(
			"SetMetricsJSONHandler->addMetricToMemStore: %w",
			err)
		writer.WriteHeader(http.StatusBadRequest)

		return
	}

	marshal, err := formResponeBody(h)
	if err != nil {
		fmt.Println("SetMetricsJSONHandler->formResponeBody: %w",
			err)
		writer.WriteHeader(http.StatusBadRequest)

		return
	}

	err = addHash(writer, marshal, h.params.Key)
	if err != nil {
		fmt.Println("SetMetricsJSONHandler->addHash: %w",
			err)
		writer.WriteHeader(http.StatusBadRequest)

		return
	}

	_, err = writer.Write(*marshal)
	if err != nil {
		fmt.Println("SetMetricsJSONHandler->writer.Write: %w",
			err)
		writer.WriteHeader(http.StatusBadRequest)

		return
	}
}

func formResponeBody(
	handler *SetMetricJSONHandler,
) (*[]byte, error) {
	tmpGauges, err := handler.serv.GetAllGauges()
	if err != nil {
		return nil,
			fmt.Errorf("formResponeBody->GetAllGauges: %w",
				err)
	}

	tmpCounters, err := handler.serv.GetAllCounters()
	if err != nil {
		return nil,
			fmt.Errorf("formResponeBody->GetAllCounters: %w",
				err)
	}

	marshal := make(apimodels.ArrMetrics,
		0,
		len(*tmpGauges)+len(*tmpCounters))

	for _, vmr := range *tmpGauges {
		tmp := apimodels.Metrics{}
		tmp.ID = vmr.Name
		tmp.MType = bizmodels.GaugeName
		tmp.Value = &vmr.Value

		marshal = append(marshal, tmp)
	}

	for _, vmr := range *tmpCounters {
		tmp := apimodels.Metrics{}
		tmp.ID = vmr.Name
		tmp.MType = bizmodels.CounterName
		tmp.Delta = &vmr.Value

		marshal = append(marshal, tmp)
	}

	metricsMarshall, err := json.Marshal(marshal)
	if err != nil {
		return nil,
			fmt.Errorf("formResponeBody->Marshal: %w",
				err)
	}

	return &metricsMarshall, nil
}

func getReqJSONData(req *http.Request,
	gauges map[string]bizmodels.Gauge,
	counters map[string]bizmodels.Counter,
	params *bizmodels.InitParams,
) error {
	var results apimodels.ArrMetrics

	bodyD, err := io.ReadAll(req.Body)
	if err != nil {
		defer req.Body.Close()

		return fmt.Errorf("getReqJSONData: %w", err)
	}

	defer req.Body.Close()

	err = checkHash(&bodyD,
		req.Header.Get("Hashsha256"), params.Key)
	if err != nil {
		return fmt.Errorf("getReqJSONData->checkHash: %w",
			err)
	}

	if len(bodyD) == 0 {
		return fmt.Errorf("getReqJSONData: %w", errGetReqDataJSON)
	}

	err = json.Unmarshal(bodyD, &results)
	if err != nil {
		return fmt.Errorf("getReqJSONData->json.Unmarshal: %w",
			err)
	}

	for _, res := range results {
		isValid := isValidJSONMetric(&res)
		if isValid {
			addValidMetric(&res, gauges, counters)
		}
	}

	return nil
}

func checkHash(dataReq *[]byte,
	hashReq string,
	key string,
) error {
	if key == "" || hashReq == "" {
		return nil
	}

	tHash, err := hash.MakeHashSHA256(dataReq,
		key)
	if err != nil {
		return fmt.Errorf("checkHash->MakeHashSHA256: %w",
			err)
	}

	decoded, err := hex.DecodeString(hashReq)
	if err != nil {
		return fmt.Errorf("checkHash->DecodeString: %w",
			err)
	}

	if !hmac.Equal(tHash, decoded) {
		return errHashDoesNotMatch
	}

	return nil
}

func addHash(writer http.ResponseWriter,
	dataResp *[]byte, key string,
) error {
	if key == "" {
		return nil
	}

	tHash, err := hash.MakeHashSHA256(dataResp,
		key)
	if err != nil {
		return fmt.Errorf("addHash->MakeHashSHA256: %w",
			err)
	}

	writer.Header().Set("Hashsha256", string(tHash))

	return nil
}

func addValidMetric(res *apimodels.Metrics,
	gauges map[string]bizmodels.Gauge,
	counters map[string]bizmodels.Counter,
) {
	if res.MType == bizmodels.GaugeName {
		gauge := new(bizmodels.Gauge)

		gauge.Name = res.ID
		gauge.Value = *res.Value
		gauges[res.ID] = *gauge
	} else if res.MType == bizmodels.CounterName {
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

func addMetricToMemStore(
	handler *SetMetricJSONHandler,
	gauges map[string]bizmodels.Gauge,
	counters map[string]bizmodels.Counter,
) error {
	err := handler.serv.AddMetrics(gauges, counters)
	if err != nil {
		return fmt.Errorf("addMetricToMemStore->AddMetrics: %w",
			err)
	}

	return nil
}

func isValidJSONMetric(metric *apimodels.Metrics,
) bool {
	var pattern string

	pattern = "^[0-9a-zA-Z/ ]{1,40}$"
	res, _ := validate.IsMatchesTemplate(metric.ID, pattern)

	if !res {
		return false
	}

	pattern = "^" + bizmodels.MetricsPattern + "$"
	res, _ = validate.IsMatchesTemplate(metric.MType, pattern)

	return res
}
