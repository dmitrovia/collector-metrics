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

var errServer = errors.New("proc return errors")

type testData struct {
	sFileStoragePath       string // f
	sDefKeyHashSha256      string // k
	sDatabaseDSN           string // d
	sRestore               string // r
	sPORT                  string // a
	sDefSavingIntervalDisk string // i
}

func getTestData() *[]testData {
	return &[]testData{
		{
			sFileStoragePath:  "../../internal/temp/metrics.json",
			sDefKeyHashSha256: "defkey",
			sDatabaseDSN: "postgres://postgres:postgres" +
				"@localhost" +
				":5432/praktikum?sslmode=disable",
			sRestore:               "false",
			sPORT:                  "localhost:8090",
			sDefSavingIntervalDisk: "60",
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
				"-a=" + test.sPORT,
				"-d=" + test.sDatabaseDSN,
				"-f=" + test.sFileStoragePath,
				"-i=" + test.sDefSavingIntervalDisk,
				"-k=" + test.sDefKeyHashSha256,
				"-r=" + test.sRestore,
			}

			cmd := exec.CommandContext(ctx, "./server", args...)
			out, err := cmd.CombinedOutput()
			sout := string(out)

			if err != nil &&
				!strings.Contains(err.Error(), "signal: killed") {
				t.Errorf("%v", err)
			}

			if strings.Contains(sout, "->") {
				fmt.Println(sout)
				t.Errorf("%v", errServer)
			}

			fmt.Println(sout)
		})
	}
}
