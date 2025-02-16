// Package sendmetricsjsonendpoint provides handler
// to send metrics to the server.
package sendmetricsjsonendpoint

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
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
			fmt.Errorf(
				"SendMJSONEndpoint->http.NewRequestWithContext: %w",
				err)
	}

	req.Header.Set("Content-Encoding", epSettings.Encoding)
	req.Header.Set("Accept-Encoding", epSettings.Encoding)
	req.Header.Set("Content-Type", epSettings.ContentType)

	if epSettings.Hash != "" {
		req.Header.Set("Hashsha256", epSettings.Hash)
	}

	resp, err := epSettings.Client.Do(req)
	if err != nil {
		return nil,
			fmt.Errorf("SendMJSONEndpoint->client.Do: %w",
				err)
	}

	return resp, nil
}
