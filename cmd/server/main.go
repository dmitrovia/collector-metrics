// Main server application package.
package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"sync"

	"github.com/dmitrovia/collector-metrics/internal/logger"
	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
	si "github.com/dmitrovia/collector-metrics/internal/serverimplement"
	"github.com/dmitrovia/collector-metrics/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

//nolint:gochecknoglobals
var buildVersion,
	buildDate,
	buildCommit string = "N/A", "N/A", "N/A"

func main() {
	var (
		dataService *service.DS
		conn        *pgxpool.Pool
	)

	server := &http.Server{}
	params := &bizmodels.InitParams{}
	waitGroup := &sync.WaitGroup{}

	zlog, err := si.Initiate(params)
	if err != nil {
		fmt.Println("main->initiate: %w", err)

		return
	}

	logger.DoInfoLog("Build version: "+buildVersion, zlog)
	logger.DoInfoLog("Build date: "+buildDate, zlog)
	logger.DoInfoLog("Build commit: "+buildCommit, zlog)

	ctx, cancel := context.WithTimeout(
		context.Background(), params.WaitSecRespDB)

	conn, dataService, err = si.InitStorage(ctx, params)
	if err != nil {
		fmt.Println("main->initStorage: %w", err)

		return
	}

	if conn != nil {
		defer conn.Close()
	}

	defer cancel()

	err = si.InitiateServer(params, dataService, server, zlog)
	if err != nil {
		fmt.Println("main->InitiateServer: %w", err)

		return
	}

	go si.RunServer(server)
	go si.SaveMetrics(dataService, params, waitGroup)

	waitGroup.Add(1)
	waitGroup.Wait()

	err = server.Shutdown(ctx)
	if err != nil {
		fmt.Println("main->Shutdown: %w", err)

		return
	}

	err = dataService.SaveInFile(params.FileStoragePath)
	if err != nil {
		fmt.Println("main->SaveInFile: %w", err)
	}
}
