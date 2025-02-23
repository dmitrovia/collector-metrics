package getmetrichandler_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/dmitrovia/collector-metrics/internal/handlers/getmetrichandler"
	"github.com/dmitrovia/collector-metrics/internal/logger"
	"github.com/dmitrovia/collector-metrics/internal/middleware/loggermiddleware"
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

const tmpstr1 string = "111111111111111120000000000000000"

const get string = "GET"

var errRuntimeCaller = errors.New("errRuntimeCaller")

const defSavePathFile string = "/internal/temp/test1.txt"

const tmpstr string = "111111111111111111111111111111111111"

type testData struct {
	tn     string
	mt     string
	mn     string
	mv     string
	expcod int
	exbody string
	meth   string
}

func getTestData() *[]testData {
	return &[]testData{
		{
			meth: get, tn: "1", mt: bizmodels.GaugeName,
			mn: "Name1q", mv: "1", expcod: stok, exbody: "",
		},
		{
			meth: get, tn: "2", mt: bizmodels.CounterName,
			mn: "Name2q", mv: "1", expcod: stok, exbody: "",
		},
		{
			meth: get, tn: "3", mt: bizmodels.CounterName,
			mn: "Name3q", mv: "1", expcod: stok, exbody: "",
		},
		{
			meth: get, tn: "4", mt: "counter_new", mn: "Name4",
			mv: "1", expcod: bdreq, exbody: "",
		},
		{
			meth: get, tn: "5", mt: bizmodels.CounterName,
			mn: "Name5", mv: tmpstr, expcod: nfnd, exbody: "",
		},
		{
			meth: get, tn: "6", mt: bizmodels.CounterName,
			mn: "Name4q", mv: "-1", expcod: stok, exbody: "",
		},
		{
			meth: get, tn: "7", mt: bizmodels.CounterName,
			mn: "Name7", mv: "-1.1", expcod: nfnd, exbody: "",
		},
		{
			meth: get, tn: "8", mt: bizmodels.GaugeName,
			mn: "Name5q", mv: tmpstr1, expcod: stok, exbody: "",
		},
		{
			meth: get, tn: "10", mt: bizmodels.GaugeName,
			mn: "Name6q",
			mv: "-1.5", expcod: stok, exbody: "",
		},
		{
			meth: get, tn: "11", mt: bizmodels.GaugeName,
			mn: "Name7q", mv: "-1", expcod: stok, exbody: "",
		},
		{
			meth: get, tn: "12", mt: bizmodels.GaugeName,
			mn: "Name8q", mv: "5", expcod: stok, exbody: "",
		},
		{
			meth: get, tn: "13",
			mt: bizmodels.CounterName, mn: "_Name123_",
			mv: "1", expcod: nfnd, exbody: "",
		},
		{
			meth: get, tn: "15", mt: bizmodels.GaugeName,
			mn: "Name13", mv: "ASD", expcod: nfnd, exbody: "",
		},
	}
}

func initiate(
	mux *mux.Router,
	service *service.DS,
	memStorage *memoryrepository.MemoryRepository,
) error {
	memStorage.Init()

	hJSONGet := getmetrichandler.NewGetMetricHandler(
		service)

	zapLogger, err := logger.Initialize("info")
	if err != nil {
		return fmt.Errorf("initiate: %w", err)
	}

	getMMux := mux.Methods(http.MethodGet).Subrouter()
	getMMux.HandleFunc(
		"/value/{metric_type}/{metric_name}",
		hJSONGet.GetMetricHandler)
	getMMux.Use(loggermiddleware.RequestLogger(zapLogger))

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

func TestGetMetricHandler(t *testing.T) {
	t.Helper()
	t.Parallel()

	memStorage := &memoryrepository.MemoryRepository{}
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
		t.Run(http.MethodGet, func(tobj *testing.T) {
			tobj.Parallel()

			req, err := http.NewRequestWithContext(
				context.Background(),
				test.meth,
				url+"/value/"+test.mt+"/"+test.mn, nil)
			if err != nil {
				tobj.Fatal(err)
			}

			req.Header.Set("Content-Type", "text/plain")

			newr := httptest.NewRecorder()
			mux.ServeHTTP(newr, req)
			status := newr.Code
			body, _ := io.ReadAll(newr.Body)

			eql := assert.Equal(tobj,
				test.expcod,
				status, test.tn+": Response code didn't match expected")

			if !eql {
				return
			}

			if status == http.StatusOK {
				assert.Equal(tobj, test.mv,
					string(body),
					test.tn+": Response value didn't match expected")
			}
		})
	}
}
