package sender_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/dmitrovia/collector-metrics/internal/functions/compress"
	"github.com/dmitrovia/collector-metrics/internal/functions/hash"
	"github.com/dmitrovia/collector-metrics/internal/handlers/sender"
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

// const stbr int = http.StatusBadRequest

const defSavePathFile string = "/internal/temp/metrics.json"

var errPath = errors.New("path is not valid")

type testData struct {
	tn       string
	expcod   int
	exbody   string
	counters []bizmodels.Counter
	gauges   []bizmodels.Gauge
	key      string
	hash     string
}

func getTestData() *[]testData {
	tempC := []bizmodels.Counter{
		{Name: "Name1", Value: 1},
		{Name: "Name1", Value: 1},
		{Name: "Name__22", Value: 1},
		{Name: "Name__22****1", Value: 1},
		{Name: "Name2", Value: 999999999},
		{Name: "Name2", Value: -999999999},
		{Name: "Name4", Value: 0},
		{Name: "Name5", Value: 7456},
		{Name: "Name6", Value: -1},
		{Name: "Name343", Value: 555},
		{Name: randomString(5), Value: 0},
	}

	tempC1 := []bizmodels.Counter{
		{Name: "Name1", Value: 1},
	}

	tempG := []bizmodels.Gauge{
		{Name: "Name1", Value: 1.0},
		{Name: "Name1", Value: 1.0},
		{Name: "Name__22", Value: 1.0},
		{Name: "Name__22****1", Value: 1.0},
		{Name: "Name2", Value: 999999999.62},
		{Name: "Name3", Value: -999999999.38},
		{Name: "Name4", Value: 0},
		{Name: "Name5", Value: 7456.3231},
		{Name: "Name6", Value: -1.0},
		{Name: "Name343", Value: 555},
		{Name: randomString(5), Value: 0},
	}

	tempG1 := []bizmodels.Gauge{
		{Name: "Name1", Value: 1.0},
	}

	tempC2 := []bizmodels.Counter{}

	tempG2 := []bizmodels.Gauge{}

	return &[]testData{
		{
			counters: tempC, gauges: tempG,
			tn: "1", expcod: stok, exbody: "", key: "defaultKey",
		},
		{
			counters: tempC1, gauges: tempG1,
			tn: "2", expcod: stok, exbody: "", key: "samekey",
		},
		{
			counters: tempC2, gauges: tempG2,
			tn: "3", expcod: stok, exbody: "", key: "",
		},
		{
			counters: tempC2, gauges: tempG2,
			tn: "3", expcod: stok, exbody: "", key: "", hash: "123",
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
	settings *bizmodels.EndpointSettings,
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

	hJSONSets := sender.NewSenderHandler(
		dse, params)

	zapLogger, err := logger.Initialize("info")
	if err != nil {
		return fmt.Errorf("initiate: %w", err)
	}

	setMsJSONMux := mux.Methods(http.MethodPost).Subrouter()
	setMsJSONMux.HandleFunc(
		"/updates/",
		hJSONSets.SenderHandler)
	setMsJSONMux.Use(
		gzipcompressmiddleware.GzipMiddleware(),
		loggermiddleware.RequestLogger(zapLogger))

	settings.ContentType = "application/json"
	settings.Encoding = "gzip"
	settings.URL = url + "/updates/"

	return nil
}

func ExampleSender() {
	params := new(bizmodels.InitParams)
	settings := new(bizmodels.EndpointSettings)
	mux := mux.NewRouter()

	err := initiate(mux, params, settings)
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

	// Output: [
	// {
	//	"ID": "counter6611",
	//  "type" : "counter",
	//  "delta" 55,
	// },
	// {
	//	"ID": "gauge45",
	//	"type" : "gauge",
	//	"value" : 24.5,
	// ]
}

func BenchmarkSender(tobj *testing.B) {
	tobj.Helper()

	params := new(bizmodels.InitParams)
	settings := new(bizmodels.EndpointSettings)
	mux := mux.NewRouter()

	err := initiate(mux, params, settings)
	if err != nil {
		fmt.Println(err)

		return
	}

	testCases := getTestData()

	for _, test := range *testCases {
		tobj.Run(http.MethodPost, func(tobj *testing.B) {
			reqData, err := initReqData(settings, params, &test)
			if err != nil {
				fmt.Println(err)

				return
			}

			req, err := http.NewRequestWithContext(
				context.Background(),
				http.MethodPost, settings.URL, reqData)
			if err != nil {
				tobj.Fatal(err)
			}

			req.Header.Set("Content-Encoding", settings.Encoding)
			req.Header.Set("Accept-Encoding", settings.Encoding)
			req.Header.Set("Content-Type", settings.ContentType)

			if test.hash != "" {
				req.Header.Set("Hashsha256", test.hash)
			}

			if test.key != "" {
				req.Header.Set("Hashsha256", settings.Hash)
			}

			newr := httptest.NewRecorder()
			mux.ServeHTTP(newr, req)
			status := newr.Code

			assert.Equal(tobj,
				test.expcod,
				status, test.tn+": Response code didn't match expected")
		})
	}
}

func initReqData(settings *bizmodels.EndpointSettings,
	params *bizmodels.InitParams,
	testD *testData,
) (*bytes.Reader, error) {
	dataMarshal := getDataSend(testD)

	metricMarshall, err := json.Marshal(dataMarshal)
	if err != nil {
		return nil, err
	}

	metricCompress, err := compress.DeflateCompress(
		metricMarshall)
	if err != nil {
		return nil, fmt.Errorf("initReqData->DeflateCompress: %w",
			err)
	}

	if params.Key != "" {
		tHash, err := hash.MakeHashSHA256(&metricMarshall,
			testD.key)
		if err != nil {
			return nil, fmt.Errorf("initReqData->MakeHashSHA256: %w",
				err)
		}

		encodedStr := hex.EncodeToString(tHash)

		settings.Hash = encodedStr
	}

	return bytes.NewReader(metricCompress), nil
}

func getDataSend(testD *testData,
) *apimodels.ArrMetrics {
	var reqMetric apimodels.Metrics

	data := make(apimodels.ArrMetrics,
		0,
		len(testD.gauges)+len(testD.counters))

	for _, metric := range testD.counters {
		reqMetric = apimodels.Metrics{}
		reqMetric.ID = metric.Name
		reqMetric.MType = bizmodels.CounterName
		reqMetric.Delta = &metric.Value
		data = append(data, reqMetric)
	}

	for _, metric := range testD.gauges {
		reqMetric = apimodels.Metrics{}
		reqMetric.ID = metric.Name
		reqMetric.MType = bizmodels.GaugeName
		reqMetric.Value = &metric.Value
		data = append(data, reqMetric)
	}

	return &data
}

func randomString(n int) string {
	letters := []rune(
		"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}
