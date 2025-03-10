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

func mainBody() {
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

/*
var errAgent = errors.New("proc return errors")

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

			fmt.Println(sout)*/
