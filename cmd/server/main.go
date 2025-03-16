// Main server application package.
package main

import (
	"fmt"
	_ "net/http/pprof"

	"github.com/dmitrovia/collector-metrics/internal/serverimplement"
)

func main() {
	err := serverimplement.ServerProcess()
	if err != nil {
		fmt.Println("ServerProcess %w", err)

		return
	}
}
