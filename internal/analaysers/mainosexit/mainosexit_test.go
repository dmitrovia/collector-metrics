package mainosexit_test

import (
	"testing"

	"github.com/dmitrovia/collector-metrics/internal/analaysers/mainosexit"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestMyAnalyzer(
	t *testing.T,
) {
	t.Parallel()
	analysistest.Run(
		t, analysistest.TestData(),
		mainosexit.NewCheckAnalayser(), "./...")
}
