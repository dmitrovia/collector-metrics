package sendmetricendpoint

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

const timeout = 60

func SendMetricEndpoint(
	endpoint string,
	client *http.Client,
) {
	const contentTypeSendMetric string = "text/plain"

	ctx, cancel := context.WithTimeout(
		context.Background(), timeout*time.Second)

	defer cancel()

	req, _ := http.NewRequestWithContext(
		ctx, http.MethodPost,
		endpoint,
		nil)
	req.Header.Set("Content-Type", contentTypeSendMetric)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("SendMetricEndpoint->client.Do: %w", err)
	}

	defer resp.Body.Close()
}
