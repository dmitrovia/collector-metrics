package funtest_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/dmitrovia/collector-metrics/internal/agentimplement"
	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
)

func TestMain(t *testing.T) {
	t.Helper()
	t.Parallel()

	_, path, _, ok := runtime.Caller(0)

	if !ok {
		return
	}

	Root := filepath.Join(filepath.Dir(path), "../../..")

	par := bizmodels.InitParamsAgent{}
	par.ConfigPath = Root + "/internal/config/agent.json"

	err := agentimplement.GetParamsFromCFG(&par)
	if err != nil {
		t.Errorf(`GetParamsFromCFG("") = %v, want "", error`, err)
	}
}
