package endpoints

import (
	"context"
	"fmt"
	"net/http"
)

func SendMetricEndpoint(ctx context.Context, endpoint string, httpC *http.Client) {
	const contentTypeSendMetric string = "text/plain"

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	req.Header.Set("Content-Type", contentTypeSendMetric)

	resp, err := httpC.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()
}
