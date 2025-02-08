package getmetricjsonhandler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/dmitrovia/collector-metrics/internal/handlers/getmetricjsonhandler"
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

const nfnd int = http.StatusNotFound

const bdreq int = http.StatusBadRequest

const tmpstr1 float64 = 111111111111111111111111111111111.0

const post string = "POST"

var errRuntimeCaller = errors.New("errRuntimeCaller")

const defSavePathFile string = "/internal/temp/metrics.json"

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
			meth: post, tn: "4", mt: "counter_new", mn: "Name4",
			delta: 1, expcod: bdreq, exbody: "",
		},
		{
			meth: post, tn: "6", mt: bizmodels.CounterName,
			mn: "Name6", delta: -1, expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "8", mt: bizmodels.GaugeName,
			mn: "Name8", value: tmpstr1, expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "10", mt: bizmodels.GaugeName,
			mn:    "Name9",
			value: -1.5, expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "11", mt: bizmodels.GaugeName,
			mn: "Name10", value: -1, expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "12", mt: bizmodels.GaugeName,
			mn: "Name11", value: 5, expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "13",
			mt: bizmodels.CounterName, mn: "_Name123_",
			delta: 1, expcod: nfnd, exbody: "",
		},
		{
			meth: post, tn: "14",
			mt: bizmodels.CounterName, mn: "gggf4",
			delta: 1, expcod: nfnd, exbody: "",
		},
		{
			meth: post, tn: "15",
			mt: bizmodels.GaugeName, mn: "fghgf44",
			delta: 1, expcod: nfnd, exbody: "",
		},
	}
}

func initiate(
	mux *mux.Router,
	service *service.DS,
	memStorage *memoryrepository.MemoryRepository,
) error {
	memStorage.Init()

	hJSONGet := getmetricjsonhandler.NewGetMJSONHandler(
		service)

	zapLogger, err := logger.Initialize("info")
	if err != nil {
		return fmt.Errorf("initiate: %w", err)
	}

	getMJSONMux := mux.Methods(http.MethodPost).Subrouter()
	getMJSONMux.HandleFunc(
		"/value/",
		hJSONGet.GetMetricJSONHandler)
	getMJSONMux.Use(gzipcompressmiddleware.GzipMiddleware(),
		loggermiddleware.RequestLogger(zapLogger))

	return nil
}

func LoadFile(mems *service.DS) {
	_, path, _, ok := runtime.Caller(0)

	if !ok {
		fmt.Println(errRuntimeCaller)
	}

	Root := filepath.Join(filepath.Dir(path), "../../..")
	temp := Root + defSavePathFile

	err := mems.LoadFromFile(temp)
	if err != nil {
		fmt.Println("Error reading metrics from file: %w", err)
	}
}

func TestGetMetricJSONHandler(t *testing.T) {
	t.Helper()
	t.Parallel()

	memStorage := new(memoryrepository.MemoryRepository)
	testCases := getTestData()
	MemoryService := service.NewMemoryService(memStorage,
		time.Duration(5))
	mux := mux.NewRouter()

	err := initiate(mux, MemoryService, memStorage)
	if err != nil {
		fmt.Println(err)

		return
	}

	LoadFile(MemoryService)

	for _, test := range *testCases {
		t.Run(http.MethodPost, func(t *testing.T) {
			t.Parallel()

			reqData, err := formReqBody(&test)
			if err != nil {
				fmt.Println(err)

				return
			}

			req, err := http.NewRequestWithContext(
				context.Background(),
				test.meth,
				url+"/value/", reqData)
			if err != nil {
				t.Fatal(err)
			}

			req.Header.Set("Content-Type", "application/json")

			newr := httptest.NewRecorder()
			mux.ServeHTTP(newr, req)
			status := newr.Code
			body, _ := io.ReadAll(newr.Body)

			assert.Equal(t,
				test.expcod,
				status, test.tn+": Response code didn't match expected")

			if test.exbody != "" {
				assert.JSONEq(t, test.exbody, string(body))
			}
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
