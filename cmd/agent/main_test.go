package main_test

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"
)

var errAgent = errors.New("proc return errors")

type testData struct {
	cDefKeyHashSha256  string // k
	cPORT              string // a
	cPollInterval      string // p
	cDefCountJobs      string // l
	cDefReportInterval string // r
}

func getTestData() *[]testData {
	return &[]testData{
		{
			cDefKeyHashSha256:  "defkey",
			cPORT:              "localhost:8090",
			cPollInterval:      "3",
			cDefCountJobs:      "5",
			cDefReportInterval: "10",
		},
	}
}

func TestMain(t *testing.T) {
	testCases := getTestData()

	t.Helper()
	t.Parallel()

	for _, test := range *testCases {
		t.Run("server", func(tobj *testing.T) {
			tobj.Parallel()

			ctx, cancel := context.WithTimeout(
				context.Background(), 60*time.Second)
			defer cancel()

			args := []string{
				"-k=" + test.cDefKeyHashSha256,
				"-a=" + test.cPORT,
				"-p=" + test.cPollInterval,
				"-l=" + test.cDefCountJobs,
				"-r=" + test.cDefReportInterval,
			}

			time.Sleep(time.Duration(2) * time.Second)

			cmd := exec.CommandContext(ctx, "./agent", args...)
			out, err := cmd.CombinedOutput()
			sout := string(out)

			if err != nil &&
				!strings.Contains(err.Error(), "signal: killed") {
				t.Errorf("%v", err)
			}

			if strings.Contains(sout, "->") {
				fmt.Println(sout)
				t.Errorf("%v", errAgent)
			}

			fmt.Println(sout)
		})
	}
}
