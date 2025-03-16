package funtest_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
	"github.com/dmitrovia/collector-metrics/internal/serverimplement"
)

func TestMain(t *testing.T) {
	t.Helper()
	t.Parallel()

	_, path, _, ok := runtime.Caller(0)

	if !ok {
		return
	}

	Root := filepath.Join(filepath.Dir(path), "../../..")

	par := bizmodels.InitParams{}
	par.ConfigPath = Root + "/internal/config/server.json"

	err := serverimplement.GetParamsFromCFG(&par)
	if err != nil {
		t.Errorf(`GetParamsFromCFG("") = %v, want "", error`, err)
	}
}
