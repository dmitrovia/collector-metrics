// Package sendmetricsjsonendpoint provides handler
// to send metrics to the server.
package sendmetricsjsonendpoint

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
	pb "github.com/dmitrovia/collector-metrics/pkg/microservice/v1"
	"google.golang.org/grpc/metadata"
)

const timeout = 60

// SendMJSONEndpoint - main endpoint method.
func SendMJSONEndpoint(
	epSettings *bizmodels.EndpointSettings,
) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(
		context.Background(), timeout*time.Second)

	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost, epSettings.URL, epSettings.SendData)
	if err != nil {
		return nil,
			fmt.Errorf("SendMJSONEndpoint->http.NewReq: %w", err)
	}

	req.Header.Set("X-Real-IP", epSettings.RealIPHeader)
	req.Header.Set("Content-Encoding", epSettings.Encoding)
	req.Header.Set("Accept-Encoding", epSettings.Encoding)
	req.Header.Set("Content-Type", epSettings.ContentType)

	if epSettings.Hash != "" {
		req.Header.Set("Hashsha256", epSettings.Hash)
	}

	resp, err := epSettings.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("SendMJSONEndpoint->Do: %w", err)
	}

	return resp, nil
}

// SendMJSONEndpointGRPC - main endpoint method.
func SendMJSONEndpointGRPC(
	epSettings *bizmodels.EndpointSettings,
) (*pb.SenderResponse, error) {
	ctx, cancel := context.WithTimeout(
		context.Background(), timeout*time.Second)

	defer cancel()

	metd := metadata.New(map[string]string{
		"X-Real-IP":        epSettings.RealIPHeader,
		"Content-Encoding": epSettings.Encoding,
		"Accept-Encoding":  epSettings.Encoding,
		"Content-Type":     epSettings.ContentType,
	})

	if epSettings.Hash != "" {
		metdH := metadata.New(map[string]string{
			"Hashsha256": epSettings.Hash,
		})

		metd = metadata.Join(metd, metdH)
	}

	ctx1 := metadata.NewOutgoingContext(ctx, metd)

	resp, err := epSettings.MicroServiceClient.Sender(
		ctx1, &pb.SenderRequest{
			Metrics: *epSettings.MetricsGRPC,
		})
	if err != nil {
		return nil, fmt.Errorf("SendMJSONEndGRPC->Sende: %w", err)
	}

	return resp, nil
}
