package setmetrichandler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dmitrovia/collector-metrics/internal/handlers/setmetrichandler"
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

const tmpstr string = "111111111111111111111111111111111111"

const tmpstr1 string = "111111111111111111111111111111111.0"

const post string = "POST"

type testData struct {
	tn     string
	mt     string
	mn     string
	mv     string
	exbody string
	meth   string
	expcod int
}

func getTestData() *[]testData {
	return &[]testData{
		{
			meth: post, tn: "1", mt: bizmodels.GaugeName,
			mn: "Name1", mv: "1.0", expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "2", mt: bizmodels.CounterName,
			mn: "Name2", mv: "1", expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "3", mt: bizmodels.CounterName,
			mn: "Name3", mv: "1", expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "4", mt: "counter_new", mn: "Name4",
			mv: "1", expcod: bdreq, exbody: "",
		},
		{
			meth: post, tn: "5", mt: bizmodels.CounterName,
			mn: "Name5", mv: tmpstr, expcod: bdreq, exbody: "",
		},
		{
			meth: post, tn: "6", mt: bizmodels.CounterName,
			mn: "Name6", mv: "-1", expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "7", mt: bizmodels.CounterName,
			mn: "Name7", mv: "-1.1", expcod: bdreq, exbody: "",
		},
		{
			meth: post, tn: "8", mt: bizmodels.GaugeName,
			mn: "Name8", mv: tmpstr1, expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "10", mt: bizmodels.GaugeName,
			mn: "Name9",
			mv: "-1.5", expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "11", mt: bizmodels.GaugeName,
			mn: "Name10", mv: "-1", expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "12", mt: bizmodels.GaugeName,
			mn: "Name11", mv: "5", expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "13",
			mt: bizmodels.CounterName, mn: "_Name123_",
			mv: "1", expcod: nfnd, exbody: "",
		},
		{
			meth: post, tn: "15", mt: bizmodels.GaugeName,
			mn: "Name13", mv: "ASD", expcod: bdreq, exbody: "",
		},
	}
}

func initiate(router *mux.Router,
	memStorage *memoryrepository.MemoryRepository,
) {
	memStorage.Init()
	MemoryService := service.NewMemoryService(memStorage,
		time.Duration(5))
	handler := setmetrichandler.NewSetMetricHandler(
		MemoryService)

	router.HandleFunc(
		"/update/{metric_type}/{metric_name}/{metric_value}",
		handler.SetMetricHandler)
}

func TestSetMetricHandler(t *testing.T) {
	t.Helper()
	t.Parallel()

	memStorage := &memoryrepository.MemoryRepository{}
	router := mux.NewRouter()

	initiate(router, memStorage)

	testCases := getTestData()

	for _, test := range *testCases {
		t.Run(http.MethodPost, func(tobj *testing.T) {
			tobj.Parallel()

			req, err := http.NewRequestWithContext(
				context.Background(),
				test.meth,
				url+"/update/"+test.mt+"/"+test.mn+"/"+test.mv, nil)
			if err != nil {
				tobj.Fatal(err)
			}

			req.Header.Set("Content-Type", "text/plain")

			newr := httptest.NewRecorder()
			router.ServeHTTP(newr, req)
			status := newr.Code

			assert.Equal(tobj,
				test.expcod,
				status, test.tn+": Response code didn't match expected")
		})
	}
}
