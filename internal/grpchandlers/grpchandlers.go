package grpchandlers

import (
	"context"
	"crypto/hmac"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/dmitrovia/collector-metrics/internal/functions/hash"
	"github.com/dmitrovia/collector-metrics/internal/functions/validate"
	"github.com/dmitrovia/collector-metrics/internal/models/apimodels"
	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
	"github.com/dmitrovia/collector-metrics/internal/service"
	pb "github.com/dmitrovia/collector-metrics/pkg/microservice/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var errHashDoesNotMatch = errors.New("hash does not match")

type MicroserviceServer struct {
	pb.UnimplementedMicroServiceServer

	Params *bizmodels.InitParams
	Serv   service.Service
}

func (s *MicroserviceServer) Sender(
	ctx context.Context,
	req *pb.SenderRequest,
) (*pb.SenderResponse, error) {
	response := &pb.SenderResponse{}

	metad, _ := metadata.FromIncomingContext(ctx)

	fmt.Println(req.GetMetrics())
	fmt.Println(metad)

	err := getReqData(req, &metad, s.Params, s.Serv)
	if err != nil {
		fmt.Println("SetMetricsJSONHandler->getReqData: %w",
			err)

		return nil, status.Errorf(codes.Unknown, "Ошибка")
	}

	err = writeResp(s.Serv, response)
	if err != nil {
		fmt.Println("SetMetricsJSONHandler->writeResp: %w",
			err)

		return nil, status.Errorf(codes.Unknown, "Ошибка")
	}

	return nil, status.Errorf(codes.OK, "Success")
}

// writeResp - writes the response
// in json format to the response body.
// First, the metrics are obtained
// from the service and encrypted.
func writeResp(
	serv service.Service,
	response *pb.SenderResponse,
) error {
	arr, err := serv.GetAllMetricsAPI()
	if err != nil {
		return fmt.Errorf("writeResp->GetAllMetricsAPI: %w",
			err)
	}

	fmt.Println("123")

	marshal, err := json.Marshal(arr)
	if err != nil {
		return fmt.Errorf("writeResp->Marshal: %w",
			err)
	}

	response.Metrics = marshal

	return nil
}

// getReqData - receives metrics
// from the request and validates it.
func getReqData(
	req *pb.SenderRequest,
	metad *metadata.MD,
	params *bizmodels.InitParams,
	serv service.Service,
) error {
	var results apimodels.ArrMetrics

	Hashsha256 := ""
	arrh := metad.Get("Hashsha256")

	if arrh != nil {
		Hashsha256 = arrh[0]
	}

	err := checkHash(&req.Metrics,
		Hashsha256, params.Key)
	if err != nil {
		return fmt.Errorf("getReqData->checkHash: %w",
			err)
	}

	err = json.Unmarshal(req.GetMetrics(), &results)
	if err != nil {
		return fmt.Errorf("getReqData->json.Unmarshal: %w",
			err)
	}

	for _, res := range results {
		isValid := isValidJSONMetric(&res)
		if isValid {
			err = addValidMetric(&res, serv)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

	return nil
}

// checkHash - checks the received
// hash in the request with the current one.
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

// addValidMetric - adds the validated
// metric to the service.
func addValidMetric(res *apimodels.Metrics,
	serv service.Service,
) error {
	if res.MType == bizmodels.GaugeName {
		err := serv.AddGauge(res.ID, *res.Value)
		if err != nil {
			return fmt.Errorf("addValidMetric->AddGauge: %w",
				err)
		}
	} else if res.MType == bizmodels.CounterName {
		_, err := serv.AddCounter(res.ID, *res.Delta,
			false)
		if err != nil {
			return fmt.Errorf("addValidMetric->AddCounter: %w",
				err)
		}
	}

	return nil
}

// isValidJSONMetric - for metric validation.
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
