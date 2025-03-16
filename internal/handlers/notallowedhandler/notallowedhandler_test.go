package notallowedhandler_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dmitrovia/collector-metrics/internal/handlers/notallowedhandler"
	"github.com/dmitrovia/collector-metrics/internal/storage/memoryrepository"
	"github.com/gorilla/mux"
)

const url string = "http://localhost:8080"

func initiate(
	mux *mux.Router,
	memStorage *memoryrepository.MemoryRepository,
) {
	memStorage.Init()

	hNotAllowed := notallowedhandler.NotAllowedHandler{}
	mux.MethodNotAllowedHandler = hNotAllowed
}

func TestNotAllowedHandler(t *testing.T) {
	t.Helper()
	t.Parallel()

	memStorage := &memoryrepository.MemoryRepository{}
	mux := mux.NewRouter()

	initiate(mux, memStorage)

	req, err := http.NewRequestWithContext(
		context.Background(),
		"NULL", url, nil)
	if err != nil {
		t.Error(err)
	}

	req.Header.Set("Content-Type", "text/plain")

	newr := httptest.NewRecorder()
	mux.ServeHTTP(newr, req)
	status := newr.Code

	fmt.Println(status)

	if status != http.StatusMovedPermanently {
		t.Error("status not 301")
	}
}
