// Main agent application package.
package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/dmitrovia/collector-metrics/internal/agentimplement"
	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
)

func main() {
	waitGroup := new(sync.WaitGroup)
	monitor := new(bizmodels.Monitor)
	client := new(http.Client)
	params := new(bizmodels.InitParamsAgent)

	err := agentimplement.Initialization(
		params,
		monitor)
	if err != nil {
		fmt.Println("main->initialization: %w", err)

		return
	}

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
