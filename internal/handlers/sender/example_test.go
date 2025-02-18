package sender_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
	"github.com/gorilla/mux"
)

func ExampleSender() {
	params := &bizmodels.InitParams{}
	settings := &bizmodels.EndpointSettings{}
	mux := mux.NewRouter()

	err := initiate(mux, params, settings, true)
	if err != nil {
		fmt.Println(err)

		return
	}

	tempC := []bizmodels.Counter{
		{Name: "counter6611", Value: 55},
	}

	tempG := []bizmodels.Gauge{
		{Name: "gauge45", Value: 24.5},
	}

	test := testData{
		counters: tempC, gauges: tempG,
		tn: "1", key: "defaultKey",
	}

	reqData, err := initReqData(settings, params, &test)
	if err != nil {
		fmt.Println(err)

		return
	}

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost, "http://localhost:8080/updates/",
		reqData)
	if err != nil {
		fmt.Println(err)
	}

	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("Hashsha256", settings.Hash)

	newr := httptest.NewRecorder()
	mux.ServeHTTP(newr, req)

	res, err := parseResponse(newr.Body)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(res)

	// Output:
	// [{gauge45 gauge 0 24.5} {counter6611 counter 55 0}]
}
