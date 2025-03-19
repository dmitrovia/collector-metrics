package rungrpctestgo_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/dmitrovia/collector-metrics/internal/grpcimplement"
)

//nolint:usetesting
func TestMain(t *testing.T) {
	t.Helper()
	t.Parallel()

	os.Setenv("ADDRESS", "localhost:8094")

	go func() {
		err := grpcimplement.ServerProcess()
		if err != nil {
			fmt.Println("ServerProcessGRPC %w", err)

			return
		}
	}()

	<-time.After(time.Duration(15) * time.Second)
}
