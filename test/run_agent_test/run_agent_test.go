package runagenttestgo_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/dmitrovia/collector-metrics/internal/agentimplement"
)

//nolint:usetesting
func TestMain(t *testing.T) {
	t.Helper()
	t.Parallel()

	os.Setenv("ADDRESS", "localhost:8093")

	go func() {
		err1 := agentimplement.AgentProcess()
		if err1 != nil {
			fmt.Println("AgentProcess %w", err1)

			return
		}
	}()

	<-time.After(time.Duration(15) * time.Second)
}
