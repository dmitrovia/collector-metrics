package pinghandler_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dmitrovia/collector-metrics/internal/handlers/pinghandler"
	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
	"github.com/dmitrovia/collector-metrics/internal/service"
	"github.com/dmitrovia/collector-metrics/internal/storage/memoryrepository"
	"github.com/gorilla/mux"
)

const url string = "http://localhost:8080/ping"

func initiate(
	mux *mux.Router,
	service *service.DS,
	memStorage *memoryrepository.MemoryRepository,
) {
	memStorage.Init()

	par := &bizmodels.InitParams{}
	par.DatabaseDSN = "postgres://postgres:postgres" +
		"@postgres" +
		":5432/praktikum?sslmode=disable"
	par.WaitSecRespDB = 10 * time.Second
	hPing := pinghandler.NewPingHandler(service, par)

	getPingBDMux := mux.Methods(http.MethodGet).Subrouter()
	getPingBDMux.HandleFunc("/ping", hPing.PingHandler)
}

func TestGetMetricHandler(t *testing.T) {
	t.Helper()
	t.Parallel()

	memStorage := &memoryrepository.MemoryRepository{}
	MemoryService := service.NewMemoryService(memStorage,
		time.Duration(5))
	mux := mux.NewRouter()

	initiate(mux, MemoryService, memStorage)

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet, url, nil)
	if err != nil {
		t.Error(err)
	}

	req.Header.Set("Content-Type", "text/plain")

	newr := httptest.NewRecorder()
	mux.ServeHTTP(newr, req)
	status := newr.Code

	fmt.Println(status)

	if status != http.StatusOK {
		t.Error("status not 200")
	}
}
