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

	writer.WriteHeader(http.StatusOK)

	err := getReqData(h, req)
	if err != nil {
		fmt.Println("SetMetricsJSONHandler->getReqJSONData: %w",
			err)
		writer.WriteHeader(http.StatusBadRequest)

		return
	}

	err = writeResp(writer, h)
	if err != nil {
		fmt.Println("SetMetricsJSONHandler->formRespone: %w",
			err)
		writer.WriteHeader(http.StatusBadRequest)

		return
	}
}

func writeResp(
	writer http.ResponseWriter,
	handler *SetMetricJSONHandler,
) error {
	arr, err := handler.serv.GetAllMetricsAPI()
	if err != nil {
		return fmt.Errorf("writeResp->GetAllMetricsAPI: %w",
			err)
	}

	marshal, err := json.Marshal(arr)
	if err != nil {
		return fmt.Errorf("writeResp->Marshal: %w",
			err)
	}

	if handler.params.Key != "" {
		tHash, err := hash.MakeHashSHA256(&marshal,
			handler.params.Key)
		if err != nil {
			return fmt.Errorf("writeResp->MakeHashSHA256: %w",
				err)
		}

		writer.Header().Set("Hashsha256", string(tHash))
	}

	_, err = writer.Write(marshal)
	if err != nil {
		return fmt.Errorf("writeResp->Write: %w",
			err)
	}

	return nil
}

func getReqData(
	handler *SetMetricJSONHandler,
	req *http.Request,
) error {
	var results apimodels.ArrMetrics

	bodyD, err := io.ReadAll(req.Body)
	if err != nil {
		defer req.Body.Close()

		return fmt.Errorf("getReqData: %w", err)
	}

	defer req.Body.Close()

	if len(bodyD) == 0 {
		return fmt.Errorf("getReqData: %w", errGetReqDataJSON)
	}

	err = checkHash(&bodyD,
		req.Header.Get("Hashsha256"), handler.params.Key)
	if err != nil {
		return fmt.Errorf("getReqData->checkHash: %w",
			err)
	}

	err = json.Unmarshal(bodyD, &results)
	if err != nil {
		return fmt.Errorf("getReqData->json.Unmarshal: %w",
			err)
	}

	for _, res := range results {
		isValid := isValidJSONMetric(&res)
		if isValid {
			err = addValidMetric(&res, handler)
			if err != nil {
				fmt.Println(err)
			}
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

func addValidMetric(res *apimodels.Metrics,
	handler *SetMetricJSONHandler,
) error {
	if res.MType == bizmodels.GaugeName {
		err := handler.serv.AddGauge(res.ID, *res.Value)
		if err != nil {
			return fmt.Errorf("addValidMetric->AddGauge: %w",
				err)
		}
	} else if res.MType == bizmodels.CounterName {
		_, err := handler.serv.AddCounter(res.ID, *res.Delta)
		if err != nil {
			return fmt.Errorf("addValidMetric->AddCounter: %w",
				err)
		}
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
