package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"sync"

	"github.com/dmitrovia/collector-metrics/internal/grpcimplement"
	"github.com/dmitrovia/collector-metrics/internal/logger"
	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
	si "github.com/dmitrovia/collector-metrics/internal/serverimplement"
	"github.com/dmitrovia/collector-metrics/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

//nolint:gochecknoglobals
var buildVersion,
	buildDate,
	buildCommit string = "N/A", "N/A", "N/A"

//nolint:funlen
func main() {
	var (
		dataService *service.DS
		conn        *pgxpool.Pool
	)

	server := &http.Server{}
	params := &bizmodels.InitParams{}
	waitGroup := &sync.WaitGroup{}
	grpcServer := grpc.NewServer()

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
	defer cancel()

	conn, dataService, err = si.InitStorage(ctx, params)
	if err != nil {
		fmt.Println("main->initStorage: %w", err)

		return
	}

	if conn != nil {
		defer conn.Close()
	}

	err = grpcimplement.InitiateServer(params,
		dataService, server)
	if err != nil {
		fmt.Println("main->InitiateServer: %w", err)

		return
	}

	channelCancel := make(chan os.Signal, 1)

	waitGroup.Add(1)

	go si.SaveMetrics(&channelCancel,
		dataService, params, waitGroup)

	go grpcimplement.RunGRPCServer(grpcServer,
		params, dataService)
	go si.RunServer(server)

	waitGroup.Wait()

	err = server.Shutdown(ctx)
	if err != nil {
		fmt.Println("main->Shutdown: %w", err)

		return
	}

	grpcServer.Stop()

	err = dataService.SaveInFile(params.FileStoragePath)
	if err != nil {
		fmt.Println("main->SaveInFile: %w", err)
	}
}
