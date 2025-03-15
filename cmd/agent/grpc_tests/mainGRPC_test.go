package main_test

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/dmitrovia/collector-metrics/internal/agentimplement"
	"github.com/dmitrovia/collector-metrics/internal/logger"
	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
)

//nolint:gochecknoglobals
var buildVersion1,
	buildDate1,
	buildCommit1 string = "N/A", "N/A", "N/A"

type testData1 struct {
	cDefKeyHashSha256  string // k
	cPORT              string // a
	cPollInterval      string // p
	cDefCountJobs      string // l
	cDefReportInterval string // r
	cUpdateURL         string // update-url
	cUseGRPC           string // use-grpc
}

func getTestData1() *[]testData1 {
	return &[]testData1{
		{
			cDefKeyHashSha256:  "defkey",
			cPORT:              "localhost:8091",
			cPollInterval:      "3",
			cDefCountJobs:      "5",
			cDefReportInterval: "10",
			cUseGRPC:           "true",
			cUpdateURL:         "http://localhost:8091/v1/updates",
		},
	}
}

func TestMain1(t *testing.T) {
	testCases := getTestData1()

	t.Helper()
	t.Parallel()

	for _, test := range *testCases {
		t.Run("server", func(tobj *testing.T) {
			tobj.Parallel()

			addFlags1(&test)
			mainBody1()
		})
	}
}

func addFlags1(test *testData1) {
	os.Args = append(os.Args, "-k="+test.cDefKeyHashSha256)
	os.Args = append(os.Args, "-a="+test.cPORT)
	os.Args = append(os.Args, "-p="+test.cPollInterval)
	os.Args = append(os.Args, "-l="+test.cDefCountJobs)
	os.Args = append(os.Args, "-r="+test.cDefReportInterval)
	os.Args = append(os.Args, "-update-url="+test.cUpdateURL)
	os.Args = append(os.Args, "-use-grpc="+test.cUseGRPC)
}

func mainBody1() {
	waitGroup := &sync.WaitGroup{}
	monitor := &bizmodels.Monitor{}
	client := &http.Client{}
	params := &bizmodels.InitParamsAgent{}

	zlog, err := agentimplement.Initialization(params,
		monitor)
	if err != nil {
		fmt.Println("main->initialization: %w", err)

		return
	}

	logger.DoInfoLog("Build version: "+buildVersion1, zlog)
	logger.DoInfoLog("Build date: "+buildDate1, zlog)
	logger.DoInfoLog("Build commit: "+buildCommit1, zlog)

	jobs := make(chan bizmodels.JobData, params.RateLimit)
	channelCancel := make(chan os.Signal, 1)
	channelCancel1 := make(chan os.Signal, 1)
	wgEndWork := &sync.WaitGroup{}

	defer close(jobs)

	waitGroup.Add(1)

	go agentimplement.Collect(
		&channelCancel,
		params,
		waitGroup,
		wgEndWork,
		monitor,
		jobs)

	waitGroup.Add(1)

	go agentimplement.Send(
		&channelCancel1,
		params,
		waitGroup,
		wgEndWork,
		client,
		monitor,
		jobs)

	go exit1(&channelCancel, &channelCancel1)

	waitGroup.Wait()
}

func exit1(
	chc *chan os.Signal,
	chc1 *chan os.Signal,
) {
	<-time.After(time.Duration(30) * time.Second)

	*chc <- syscall.SIGTERM
	*chc1 <- syscall.SIGTERM
}
