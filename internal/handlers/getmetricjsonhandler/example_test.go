package getmetricjsonhandler_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
	"github.com/gorilla/mux"
)

func ExampleGetMetricJSONHandler() {
	params := &bizmodels.InitParams{}

	test := testData{
		mt: "gauge", mn: "gauge45",
	}

	mux := mux.NewRouter()

	err := initiate(mux, params, true)
	if err != nil {
		fmt.Println(err)

		return
	}

	reqData, err := formReqBody(&test)
	if err != nil {
		fmt.Println(err)

		return
	}

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"http://localhost:8080/value/", reqData)
	if err != nil {
		fmt.Println(err)
	}

	req.Header.Set("Content-Type", "application/json")

	newr := httptest.NewRecorder()
	mux.ServeHTTP(newr, req)

	fmt.Println(newr.Body)

	// Output:
	// "id":"gauge45","type":"gauge","delta":0,"value":24.5}
}
