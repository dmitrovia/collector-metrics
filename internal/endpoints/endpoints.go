package endpoints

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dmitrovia/collector-metrics/internal/functions/compress"
	"github.com/dmitrovia/collector-metrics/internal/models/apimodels"
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

func SendMetricsJSONEndpoint(ctx context.Context, endpoint string, metrics *apimodels.ArrMetrics, httpC *http.Client) error {
	const contentTypeSendMetric string = "application/json"

	const encoding = "gzip"

	metricMarshall, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	metricCompress, err := compress.DeflateCompress(metricMarshall)
	if err != nil {
		return fmt.Errorf("SendMetricJSONEndpoint->GzipCompress: %w", err)
	}

	reader := bytes.NewReader(metricCompress)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, reader)
	if err != nil {
		return fmt.Errorf("SendMetricJSONEndpoint->NewRequestWithContext: %w", err)
	}

	req.Header.Set("Content-Type", contentTypeSendMetric)
	req.Header.Set("Content-Encoding", encoding)
	req.Header.Set("Accept-Encoding", encoding)

	resp, err := httpC.Do(req)
	if err != nil {
		return fmt.Errorf("SendMetricJSONEndpoint->Do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		out, err := compress.DeflateDecompress(resp.Body)
		if err != nil {
			fmt.Println(out)

			return fmt.Errorf("SendMetricJSONEndpoint->DeflateDecompress: %w", err)
		}
	} else {
		fmt.Printf("anscode: %d\n", resp.StatusCode)
	}

	return nil
}

/*func SendMetricJSONEndpoint(ctx context.Context, endpoint string, data apimodels.Metrics, httpC *http.Client) error {
	const contentTypeSendMetric string = "application/json"

	const encoding = "gzip"

	metricMarshall, err := json.Marshal(data)
	if err != nil {
		return err
	}

	metricMarshall, err = compress.DeflateCompress(metricMarshall)
	if err != nil {
		return fmt.Errorf("SendMetricJSONEndpoint->GzipCompress: %w", err)
	}

	reader := bytes.NewReader(metricMarshall)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, reader)
	if err != nil {
		return fmt.Errorf("SendMetricJSONEndpoint->NewRequestWithContext: %w", err)
	}

	req.Header.Set("Content-Type", contentTypeSendMetric)
	req.Header.Set("Content-Encoding", encoding)
	req.Header.Set("Accept-Encoding", encoding)

	resp, err := httpC.Do(req)
	if err != nil {
		return fmt.Errorf("SendMetricJSONEndpoint->Do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		out, err := compress.DeflateDecompress(resp.Body)
		if err != nil {
			fmt.Println(out)

			return fmt.Errorf("SendMetricJSONEndpoint->DeflateDecompress: %w", err)
		}
	} else {
		fmt.Printf("anscode: %d\n", resp.StatusCode)
	}

	return nil
}*/
