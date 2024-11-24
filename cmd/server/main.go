package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
	si "github.com/dmitrovia/collector-metrics/internal/serverimplement"
	"github.com/dmitrovia/collector-metrics/internal/service"
	"github.com/jackc/pgx/v5"
)

func main() {
	var (
		dataService *service.DS
		conn        *pgx.Conn
	)

	server := new(http.Server)
	params := new(bizmodels.InitParams)
	waitGroup := new(sync.WaitGroup)

	zapLogger, err := si.Initiate(params)
	if err != nil {
		fmt.Println("main->initiate: %w", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(), params.WaitSecRespDB)

	conn, dataService, err = si.InitStorage(ctx, params)
	if err != nil {
		fmt.Println("main->initStorage: %w", err)
		os.Exit(1)
	}

	si.InitiateServer(params, dataService, server, zapLogger)

	err = si.UseMigrations(params)
	if err != nil {
		if conn != nil {
			conn.Close(ctx)
		}

		cancel()
		os.Exit(1)
	}

	go si.RunServer(server)

	waitGroup.Add(1)

	go si.SaveMetrics(dataService, params, waitGroup)
	waitGroup.Wait()

	if conn != nil {
		conn.Close(ctx)
	}

	err = server.Shutdown(ctx)
	if err != nil {
		fmt.Println("main->Shutdown: %w", err)
		os.Exit(1)
	}

	err = dataService.SaveInFile(params.FileStoragePath)
	if err != nil {
		fmt.Println("main->SaveInFile: %w", err)
	}
}
