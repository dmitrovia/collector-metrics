// Main agent application package.
package main

import (
	"fmt"
	"net/http"
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

	waitGroup.Add(1)

	jobs := make(chan bizmodels.JobData, params.RateLimit)

	defer close(jobs)

	go agentimplement.Collect(
		params,
		waitGroup,
		monitor,
		jobs)

	waitGroup.Add(1)

	go agentimplement.Send(
		params,
		waitGroup,
		client,
		monitor,
		jobs)

	waitGroup.Wait()
}
