package sendmetricjsonendpoint

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dmitrovia/collector-metrics/internal/functions/compress"
	"github.com/dmitrovia/collector-metrics/internal/models/apimodels"
)

const timeout = 60

func SendMetricJSONEndpoint(
	endpoint string,
	data apimodels.Metrics,
	client *http.Client,
) error {
	const contentTypeSendMetric string = "application/json"

	const encoding = "gzip"

	ctx, cancel := context.WithTimeout(
		context.Background(), time.Duration(timeout))

	defer cancel()

	metricMarshall, err := json.Marshal(data)
	if err != nil {
		return err
	}

	metricMarshall, err = compress.DeflateCompress(
		metricMarshall)
	if err != nil {
		return fmt.Errorf(
			"SendMetricJSONEndpoint->GzipCompress: %w", err)
	}

	reader := bytes.NewReader(metricMarshall)

	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost,
		endpoint,
		reader)
	if err != nil {
		return fmt.Errorf(
			"SendMetricJSONEndpoint->NewRequestWithContext: %w", err)
	}

	req.Header.Set("Content-Type", contentTypeSendMetric)
	req.Header.Set("Content-Encoding", encoding)
	req.Header.Set("Accept-Encoding", encoding)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("SendMetricJSONEndpoint->Do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		out, err := compress.DeflateDecompress(resp.Body)
		if err != nil {
			fmt.Println(out)

			return fmt.Errorf(
				"SendMetricJSONEndpoint->DeflateDecompress: %w",
				err)
		}
	} else {
		fmt.Printf("anscode: %d\n", resp.StatusCode)
	}

	return nil
}
