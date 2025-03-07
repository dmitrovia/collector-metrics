// Main agent application package.
package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/dmitrovia/collector-metrics/internal/agentimplement"
	"github.com/dmitrovia/collector-metrics/internal/logger"
	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
)

//nolint:gochecknoglobals
var buildVersion,
	buildDate,
	buildCommit string = "N/A", "N/A", "N/A"

func main() {
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
		&channelCancel,
		params,
		waitGroup,
		wgEndWork,
		client,
		monitor,
		jobs)

	waitGroup.Wait()
}
