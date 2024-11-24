package sendmetricsjsonendpoint

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"
)

const timeout = 60

func SendMJSONEndpoint(
	sendData *bytes.Reader, endp string, client *http.Client,
) (*http.Response, error) {
	const contentTypeSendMetric string = "application/json"

	const encoding = "gzip"

	ctx, cancel := context.WithTimeout(
		context.Background(), time.Duration(timeout))

	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost, endp, sendData)
	if err != nil {
		return nil,
			fmt.Errorf(
				"SendMJSONEndpoint->http.NewRequestWithContext: %w",
				err)
	}

	req.Header.Set("Content-Encoding", encoding)
	req.Header.Set("Accept-Encoding", encoding)
	req.Header.Set("Content-Type", contentTypeSendMetric)

	resp, err := client.Do(req)
	if err != nil {
		return nil,
			fmt.Errorf("SendMJSONEndpoint->client.Do: %w",
				err)
	}

	return resp, nil
}
