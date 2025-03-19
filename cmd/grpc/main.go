package main

import (
	"fmt"
	_ "net/http/pprof"

	"github.com/dmitrovia/collector-metrics/internal/grpcimplement"
)

func main() {
	err := grpcimplement.ServerProcess()
	if err != nil {
		fmt.Println("ServerProcessGRPC %w", err)

		return
	}
}
