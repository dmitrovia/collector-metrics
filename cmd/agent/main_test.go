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
var buildVersion,
	buildDate,
	buildCommit string = "N/A", "N/A", "N/A"

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

			addFlags(&test)
			mainBody()
		})
	}
}

func addFlags(test *testData) {
	os.Args = append(os.Args, "-k="+test.cDefKeyHashSha256)
	os.Args = append(os.Args, "-a="+test.cPORT)
	os.Args = append(os.Args, "-p="+test.cPollInterval)
	os.Args = append(os.Args, "-l="+test.cDefCountJobs)
	os.Args = append(os.Args, "-r="+test.cDefReportInterval)
}

func setEnv() error {
	err := os.Setenv("REPORT_INTERVAL", "10")
	if err != nil {
		return fmt.Errorf("REPORT_INTERVAL: %w", err)
	}

	err = os.Setenv("POLL_INTERVAL", "2")
	if err != nil {
		return fmt.Errorf("POLL_INTERVAL: %w", err)
	}

	err = os.Setenv("KEY", "defkey")
	if err != nil {
		return fmt.Errorf("KEY: %w", err)
	}

	err = os.Setenv("ADDRESS", "localhost:8090")
	if err != nil {
		return fmt.Errorf("ADDRESS: %w", err)
	}

	err = os.Setenv("RATE_LIMIT", "5")
	if err != nil {
		return fmt.Errorf("RATE_LIMIT: %w", err)
	}

	err = os.Setenv("CRYPTO_KEY_AGENT",
		"/internal/asymcrypto/keys/public.pem")
	if err != nil {
		return fmt.Errorf("CRYPTO_KEY_AGENT: %w", err)
	}

	err = os.Setenv("CONFIG_SERVER",
		"/internal/config/agent.json")
	if err != nil {
		return fmt.Errorf("CONFIG_SERVER: %w", err)
	}

	return nil
}

func mainBody() {
	err := setEnv()
	if err != nil {
		fmt.Println("main->setEnv: %w", err)

		return
	}

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

	logger.DoInfoLog("Build version: "+buildVersion, zlog)
	logger.DoInfoLog("Build date: "+buildDate, zlog)
	logger.DoInfoLog("Build commit: "+buildCommit, zlog)

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

	go exit(&channelCancel, &channelCancel1)

	waitGroup.Wait()
}

func exit(
	chc *chan os.Signal,
	chc1 *chan os.Signal,
) {
	<-time.After(time.Duration(30) * time.Second)

	*chc <- syscall.SIGTERM
	*chc1 <- syscall.SIGTERM
}
