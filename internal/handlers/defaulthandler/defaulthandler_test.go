package defaulthandler_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/dmitrovia/collector-metrics/internal/handlers/defaulthandler"
	"github.com/dmitrovia/collector-metrics/internal/logger"
	"github.com/dmitrovia/collector-metrics/internal/middleware/gzipcompressmiddleware"
	"github.com/dmitrovia/collector-metrics/internal/middleware/loggermiddleware"
	"github.com/dmitrovia/collector-metrics/internal/service"
	"github.com/dmitrovia/collector-metrics/internal/storage/memoryrepository"
	"github.com/gorilla/mux"
)

const url string = "http://localhost:8080/"

var errRuntimeCaller = errors.New("errRuntimeCaller")

const defSavePathFile string = "/internal/temp/test1.txt"

func initiate(
	mux *mux.Router,
	service *service.DS,
	memStorage *memoryrepository.MemoryRepository,
) error {
	memStorage.Init()

	hDefault := defaulthandler.NewDefaultHandler(service)

	zapLogger, err := logger.Initialize("info")
	if err != nil {
		return fmt.Errorf("initiate: %w", err)
	}

	defaultMux := mux.Methods(http.MethodGet).Subrouter()
	defaultMux.HandleFunc("/", hDefault.DefaultHandler)
	defaultMux.Use(gzipcompressmiddleware.GzipMiddleware(),
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

func TestGetMetricHandler(t *testing.T) {
	t.Helper()
	t.Parallel()

	memStorage := &memoryrepository.MemoryRepository{}
	MemoryService := service.NewMemoryService(memStorage,
		time.Duration(5))
	mux := mux.NewRouter()

	err := initiate(mux, MemoryService, memStorage)
	if err != nil {
		fmt.Println(err)

		return
	}

	LoadFile(MemoryService)

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
		t.Error("status not ok")
	}
}
