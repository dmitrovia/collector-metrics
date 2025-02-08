package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
	si "github.com/dmitrovia/collector-metrics/internal/serverimplement"
	"github.com/dmitrovia/collector-metrics/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	var (
		dataService *service.DS
		conn        *pgxpool.Pool
	)

	server := new(http.Server)
	params := new(bizmodels.InitParams)
	waitGroup := new(sync.WaitGroup)

	zapLogger, err := si.Initiate(params)
	if err != nil {
		fmt.Println("main->initiate: %w", err)

		return
	}

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

	si.InitiateServer(params, dataService, server, zapLogger)

	err = si.UseMigrations(params)
	if err != nil {
		fmt.Println("main->UseMigrations: %w", err)

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
