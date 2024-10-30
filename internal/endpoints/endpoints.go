package endpoints

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dmitrovia/collector-metrics/internal/models/apimodels"
)

func SendMetricEndpoint(ctx context.Context, endpoint string, httpC *http.Client) {
	const contentTypeSendMetric string = "text/plain"

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	req.Header.Set("Content-Type", contentTypeSendMetric)

	resp, err := httpC.Do(req)
	if err != nil {
		fmt.Println("SendMetricEndpoint: %w", err)
	}

	defer resp.Body.Close()
}

func SendMetricJSONEndpoint(ctx context.Context, endpoint string, data apimodels.Metrics, httpC *http.Client) error {
	const contentTypeSendMetric string = "application/json"

	metricMarshall, err := json.Marshal(data)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(metricMarshall)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, reader)
	if err != nil {
		return fmt.Errorf("SendMetricJSONEndpoint: %w", err)
	}

	req.Header.Set("Content-Type", contentTypeSendMetric)

	resp, err := httpC.Do(req)
	if err != nil {
		return fmt.Errorf("SendMetricJSONEndpoint: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
