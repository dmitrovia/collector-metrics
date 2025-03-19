package runservertestgo_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/dmitrovia/collector-metrics/internal/serverimplement"
)

//nolint:usetesting
func TestMain(t *testing.T) {
	t.Helper()
	t.Parallel()

	os.Setenv("ADDRESS", "localhost:8093")

	go func() {
		err := serverimplement.ServerProcess()
		if err != nil {
			fmt.Println("ServerProcessGRPC %w", err)

			return
		}
	}()

	<-time.After(time.Duration(15) * time.Second)
}
