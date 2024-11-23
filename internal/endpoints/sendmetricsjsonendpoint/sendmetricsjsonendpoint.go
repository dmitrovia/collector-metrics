package sendmetricsjsonendpoint

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
)

func SendMetricsJSONEndpoint(sendData *bytes.Reader, endp string, httpC *http.Client) (*http.Response, error) {
	const contentTypeSendMetric string = "application/json"

	const encoding = "gzip"

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, endp, sendData)
	if err != nil {
		return nil, fmt.Errorf("SendMetricsJSONEndpoint->http.NewRequestWithContext: %w", err)
	}

	req.Header.Set("Content-Encoding", encoding)
	req.Header.Set("Accept-Encoding", encoding)
	req.Header.Set("Content-Type", contentTypeSendMetric)

	resp, err := httpC.Do(req)
	if err != nil {
		return nil, fmt.Errorf("SendMetricsJSONEndpoint->httpC.Do: %w", err)
	}

	return resp, nil
}
