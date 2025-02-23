package setmetricjsonhandler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dmitrovia/collector-metrics/internal/handlers/setmetricjsonhandler"
	"github.com/dmitrovia/collector-metrics/internal/logger"
	"github.com/dmitrovia/collector-metrics/internal/middleware/gzipcompressmiddleware"
	"github.com/dmitrovia/collector-metrics/internal/middleware/loggermiddleware"
	"github.com/dmitrovia/collector-metrics/internal/models/apimodels"
	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
	"github.com/dmitrovia/collector-metrics/internal/service"
	"github.com/dmitrovia/collector-metrics/internal/storage/memoryrepository"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

const url string = "http://localhost:8080"

const stok int = http.StatusOK

const nallwd int = http.StatusMethodNotAllowed

const nfnd int = http.StatusNotFound

const bdreq int = http.StatusBadRequest

const tmpstr int64 = 999999999999999999

const tmpstr1 float64 = 111111111111111111111111111111111.0

const (
	post            string = "POST"
	defSavePathFile string = "/internal/temp/metrics.json"
)

type testData struct {
	tn     string
	mt     string
	mn     string
	delta  int64
	value  float64
	expcod int
	exbody string
	meth   string
}

func getTestData() *[]testData {
	return &[]testData{
		{
			meth: post, tn: "1", mt: bizmodels.GaugeName,
			mn: "Name1", value: 1.0, expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "2", mt: bizmodels.CounterName,
			mn: "Name2", delta: 1, expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "3", mt: bizmodels.CounterName,
			mn: "Name3", delta: 1, expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "4", mt: bizmodels.GaugeName,
			mn: "Name4", value: tmpstr1, expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "5", mt: bizmodels.CounterName,
			mn: "Name5", delta: tmpstr, expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "6", mt: "counter_new", mn: "Name6",
			value: 1, expcod: bdreq, exbody: "",
		},
		{
			meth: post, tn: "7", mt: bizmodels.CounterName,
			mn: "Name7", delta: -1, expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "8", mt: bizmodels.GaugeName,
			mn: "Name8", value: -1.5, expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "9", mt: bizmodels.GaugeName,
			mn: "Name9", value: -1, expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "10", mt: bizmodels.GaugeName,
			mn: "Name10", value: 5, expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "11",
			mt: bizmodels.CounterName, mn: "_Name123_",
			value: 1, expcod: nfnd, exbody: "",
		},
		{
			meth: "PATCH", tn: "12",
			mt: bizmodels.CounterName, mn: "Name11",
			value: 1, expcod: nallwd, exbody: "",
		},
		{
			meth: post, tn: "13",
			mt: bizmodels.CounterName, mn: "",
			value: 0, delta: 0, expcod: nfnd, exbody: "",
		},
	}
}

func initiate(
	memStorage *memoryrepository.MemoryRepository,
	mux *mux.Router,
) (*service.DS, error) {
	memStorage.Init()

	MemoryService := service.NewMemoryService(memStorage,
		time.Duration(5))

	hJSONSet := setmetricjsonhandler.NewSetMJH(MemoryService)

	zapLogger, err := logger.Initialize("info")
	if err != nil {
		return nil, fmt.Errorf("initiate: %w", err)
	}

	setMJSONMux := mux.Methods(http.MethodPost).Subrouter()
	setMJSONMux.HandleFunc(
		"/update/",
		hJSONSet.SetMJSONHandler)
	setMJSONMux.Use(gzipcompressmiddleware.GzipMiddleware(),
		loggermiddleware.RequestLogger(zapLogger))

	return MemoryService, nil
}

func TestSetMetricJSONHandler(t *testing.T) {
	t.Helper()
	t.Parallel()

	testCases := getTestData()

	memStorage := &memoryrepository.MemoryRepository{}
	mux := mux.NewRouter()

	_, err := initiate(memStorage, mux)
	if err != nil {
		fmt.Println(err)

		return
	}

	for _, test := range *testCases {
		t.Run(http.MethodPost, func(tobj *testing.T) {
			tobj.Parallel()

			reqData, err := formReqBody(&test)
			if err != nil {
				fmt.Println(err)

				return
			}

			req, err := http.NewRequestWithContext(
				context.Background(),
				test.meth,
				url+"/update/", reqData)
			if err != nil {
				tobj.Fatal(err)
			}

			req.Header.Set("Content-Type", "application/json")

			newr := httptest.NewRecorder()
			mux.ServeHTTP(newr, req)
			status := newr.Code

			assert.Equal(tobj,
				test.expcod,
				status, test.tn+": Response code didn't match expected")
		})
	}
}

func formReqBody(
	data *testData,
) (*bytes.Reader, error) {
	metr := &apimodels.Metrics{}
	metr.MType = data.mt
	metr.ID = data.mn
	metr.Delta = &data.delta
	metr.Value = &data.value

	marshall, err := json.Marshal(metr)
	if err != nil {
		return nil,
			fmt.Errorf("formReqBody->Marshal: %w",
				err)
	}

	return bytes.NewReader(marshall), nil
}
