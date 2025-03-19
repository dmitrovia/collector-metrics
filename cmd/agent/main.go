// Main agent application package.
package main

import (
	"fmt"

	"github.com/dmitrovia/collector-metrics/internal/agentimplement"
)

func main() {
	err := agentimplement.AgentProcess()
	if err != nil {
		fmt.Println("AgentProcess %w", err)

		return
	}
}
