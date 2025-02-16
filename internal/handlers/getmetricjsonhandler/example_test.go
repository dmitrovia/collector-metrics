package getmetricjsonhandler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	"github.com/dmitrovia/collector-metrics/internal/storage/dbrepository"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
)

const url string = "http://localhost:8080"

const stok int = http.StatusOK

const nfnd int = http.StatusNotFound

const post string = "POST"

const defSavePathFile string = "/internal/temp/metrics.json"

var errPath = errors.New("path is not valid")

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

func getTestData() []testData {
	return []testData{
		{
			meth: post, tn: "1", mt: bizmodels.GaugeName,
			mn: "Name1", value: 1, expcod: stok,
		},
		{
			meth: post, tn: "2", mt: bizmodels.CounterName,
			mn: "Name1", delta: 1, expcod: stok,
		},
		{
			meth: post, tn: "3", mt: bizmodels.CounterName,
			mn: "Name__22", delta: 1, expcod: nfnd,
		},
		{
			meth: post, tn: "4", mt: bizmodels.CounterName,
			mn: "Name__22****1", delta: 1, expcod: nfnd,
		},
		{
			meth: post, tn: "5", mt: bizmodels.CounterName,
			mn: "Name2", delta: 999999999, expcod: stok,
		},
		{
			meth: post, tn: "6", mt: bizmodels.CounterName,
			mn: "Name2", delta: -999999999, expcod: stok,
		},
		{
			meth: post, tn: "7", mt: bizmodels.CounterName,
			mn: "Name4", delta: 0, expcod: stok,
		},
		{
			meth: post, tn: "8", mt: bizmodels.CounterName,
			mn: "Name5", delta: 7456, expcod: stok,
		},
		{
			meth: post, tn: "9", mt: bizmodels.CounterName,
			mn: "Name6", delta: -1, expcod: stok,
		},
		{
			meth: post, tn: "10", mt: bizmodels.CounterName,
			mn: "Name343", delta: 555, expcod: stok,
		},
		{
			meth: post, tn: "101", mt: bizmodels.CounterName,
			mn: "Name58888", delta: 555, expcod: nfnd,
		},
	}
}

func getTD2() []testData {
	return []testData{
		{
			meth: post, tn: "11", mt: bizmodels.GaugeName,
			mn: "Name1", value: 1.0, expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "12", mt: bizmodels.GaugeName,
			mn: "Name1", value: 1.0, expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "13", mt: bizmodels.GaugeName,
			mn: "Name__22", value: 1.0, expcod: nfnd, exbody: "",
		},
		{
			meth: post, tn: "14", mt: bizmodels.GaugeName,
			mn: "Name__22****1", value: 1.0, expcod: nfnd,
			exbody: "",
		},
		{
			meth: post, tn: "15", mt: bizmodels.GaugeName,
			mn: "Name2", value: 999999999.62,
			expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "16", mt: bizmodels.GaugeName,
			mn: "Name2", value: -999999999.38,
			expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "17", mt: bizmodels.GaugeName,
			mn: "Name4", value: 0, expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "18", mt: bizmodels.GaugeName,
			mn: "Name5", value: 7456.3231, expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "19", mt: bizmodels.GaugeName,
			mn: "Name6", value: -1, expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "20", mt: bizmodels.GaugeName,
			mn: "Name343", value: 555, expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "21", mt: bizmodels.GaugeName,
			mn: "Name58888", value: 555, expcod: nfnd, exbody: "",
		},
	}
}

func setHandlerParams(params *bizmodels.InitParams) error {
	params.
		ValidateAddrPattern = "^[a-zA-Z/ ]{1,100}:[0-9]{1,10}$"
	params.DatabaseDSN = "postgres://postgres:postgres" +
		"@localhost" +
		":5432/praktikum?sslmode=disable"
	_, path, _, ok := runtime.Caller(0)

	if !ok {
		return fmt.Errorf("setInitParams->runtime.Caller( %w",
			errPath)
	}

	Root := filepath.Join(filepath.Dir(path), "../..")
	params.FileStoragePath = Root + defSavePathFile
	params.Key = "defaultKey"
	params.Restore = true
	params.ValidateAddrPattern = ""
	params.WaitSecRespDB = 10 * time.Second

	return nil
}

func initiate(
	mux *mux.Router,
	params *bizmodels.InitParams,
) error {
	err := setHandlerParams(params)
	if err != nil {
		return fmt.Errorf("initStorage->pgx.Connect %w",
			err)
	}

	storage := new(dbrepository.DBepository)

	ctx, cancel := context.WithTimeout(
		context.Background(), params.WaitSecRespDB)

	defer cancel()

	dse := service.NewMemoryService(storage,
		params.WaitSecRespDB)

	dbConn, err := pgxpool.New(ctx, params.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("initStorage->pgx.Connect %w",
			err)
	}

	storage.Initiate(params.DatabaseDSN, dbConn)

	hJSONGet := getmetricjsonhandler.NewGetMJSONHandler(
		dse)

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

func ExampleGetMetricJSONHandler() {
	params := new(bizmodels.InitParams)

	test := testData{
		mt: "gauge", mn: "gauge45",
	}

	mux := mux.NewRouter()

	err := initiate(mux, params)
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

	// Output: {
	//	"ID": "gauge45",
	//  "type" : "gauge",
	//  "delta" 24.5,
	// },
}

func BenchmarkGetMetricJSONHandler(tobj *testing.B) {
	tobj.Helper()

	params := new(bizmodels.InitParams)

	result := make([]testData, 0)
	result = append(result, getTestData()...)
	result = append(result, getTD2()...)

	mux := mux.NewRouter()

	err := initiate(mux, params)
	if err != nil {
		fmt.Println(err)

		return
	}

	for _, test := range result {
		tobj.Run(http.MethodPost, func(tobj *testing.B) {
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
