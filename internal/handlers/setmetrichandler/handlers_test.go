package setmetrichandler_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dmitrovia/collector-metrics/internal/handlers/setmetrichandler"
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

const tmpstr string = "111111111111111111111111111111111111"

const tmpstr1 string = "111111111111111111111111111111111.0"

const cou string = "counter"

const gaug string = "gauge"

const name string = "Name"

const post string = "POST"

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
			meth: post, tn: "1", mt: gaug, mn: name,
			mv: "1.0", expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "2", mt: cou, mn: name,
			mv: "1", expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "3", mt: cou, mn: name,
			mv: "1", expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "4", mt: "counter_new", mn: name,
			mv: "1", expcod: bdreq, exbody: "",
		},
		{
			meth: post, tn: "5", mt: cou, mn: name,
			mv: tmpstr, expcod: bdreq, exbody: "",
		},
		{
			meth: post, tn: "6", mt: cou, mn: name,
			mv: "-1", expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "7", mt: cou, mn: name,
			mv: "-1.1", expcod: bdreq, exbody: "",
		},
		{
			meth: post, tn: "8", mt: gaug, mn: name,
			mv: tmpstr1, expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "10", mt: gaug, mn: name,
			mv: "-1.5", expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "11", mt: gaug, mn: name,
			mv: "-1", expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "12", mt: gaug, mn: name,
			mv: "5", expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "13", mt: cou, mn: "_Name123_",
			mv: "1", expcod: nfnd, exbody: "",
		},
		{
			meth: "PATCH", tn: "14", mt: cou, mn: name,
			mv: "1", expcod: nallwd, exbody: "",
		},
		{
			meth: post, tn: "15", mt: gaug, mn: name,
			mv: "ASD", expcod: bdreq, exbody: "",
		},
	}
}

func SetMetricHandler(t *testing.T) {
	t.Helper()

	memStorage := new(memoryrepository.MemoryRepository)

	testCases := getTestData()

	MemoryService := service.NewMemoryService(memStorage)
	memStorage.Init()

	handler := setmetrichandler.NewSetMetricHandler(
		MemoryService)

	for _, test := range *testCases {
		t.Run(http.MethodPost, func(t *testing.T) {
			req, err := http.NewRequestWithContext(
				context.Background(),
				test.meth,
				url+"/update/"+test.mt+"/"+test.mn+"/"+test.mv, nil)
			if err != nil {
				t.Fatal(err)
			}

			req.Header.Set("Content-Type", "text/plain")

			newr := httptest.NewRecorder()
			router := mux.NewRouter()
			router.HandleFunc("/update/{mt}/{mn}/{mv}",
				handler.SetMetricHandler)
			router.ServeHTTP(newr, req)
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
