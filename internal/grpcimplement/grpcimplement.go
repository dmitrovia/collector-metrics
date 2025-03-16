package grpcimplement

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/dmitrovia/collector-metrics/internal/Interceptors/decompressinterceptor"
	"github.com/dmitrovia/collector-metrics/internal/Interceptors/decryptinterceptor"
	"github.com/dmitrovia/collector-metrics/internal/grpchandlers"
	"github.com/dmitrovia/collector-metrics/internal/logger"
	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
	si "github.com/dmitrovia/collector-metrics/internal/serverimplement"
	"github.com/dmitrovia/collector-metrics/internal/service"
	pb "github.com/dmitrovia/collector-metrics/pkg/microservice/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

//nolint:gochecknoglobals
var buildVersion,
	buildDate,
	buildCommit string = "N/A", "N/A", "N/A"

const rTimeout = 60

const wTimeout = 60

const iTimeout = 60

const grpcPort string = ":50051"

// RunServer - starts the server.
func RunGRPCServer(grpcServer *grpc.Server,
	params *bizmodels.InitParams,
	dse *service.DS,
) {
	fmt.Println("gRPC server start" + grpcPort)

	listen, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Printf("RunGRPCServer->Listen: %s\n", err)

		return
	}

	reflection.Register(grpcServer)

	hand := &grpchandlers.MicroserviceServer{}
	hand.Params = params
	hand.Serv = dse

	pb.RegisterMicroServiceServer(grpcServer, hand)

	err = grpcServer.Serve(listen)
	if err != nil {
		log.Printf("RunGRPCServer->Serve: %s\n", err)

		return
	}
}

// InitiateServer - initializes server data.
func InitiateServer(
	par *bizmodels.InitParams,
	mser *service.DS,
	server *http.Server,
) error {
	mux := runtime.NewServeMux()

	hand := &grpchandlers.MicroserviceServer{}
	hand.Params = par
	hand.Serv = mser

	err := pb.RegisterMicroServiceHandlerServer(
		context.Background(),
		mux,
		hand)
	if err != nil {
		return fmt.Errorf("InitiateServer->Register: %w", err)
	}

	*server = http.Server{
		Addr:         par.PORT,
		Handler:      mux,
		ErrorLog:     nil,
		ReadTimeout:  rTimeout * time.Second,
		WriteTimeout: wTimeout * time.Second,
		IdleTimeout:  iTimeout * time.Second,
	}

	if par.Restore {
		err1 := mser.LoadFromFile(par.FileStoragePath)
		if err1 != nil {
			fmt.Println("Error reading metrics from file: %w", err1)
		}
	}

	err = si.UseMigrations(par)
	if err != nil {
		return fmt.Errorf("InitiateServer->UseMigration: %w", err)
	}

	return nil
}

//nolint:funlen
func ServerProcess() error {
	var (
		dataService *service.DS
		conn        *pgxpool.Pool
	)

	server := &http.Server{}
	params := &bizmodels.InitParams{}
	waitGroup := &sync.WaitGroup{}

	zlog, err := si.Initiate(params)
	if err != nil {
		return fmt.Errorf("si.Initiate: %w", err)
	}

	logger.DoInfoLog("Build version: "+buildVersion, zlog)
	logger.DoInfoLog("Build date: "+buildDate, zlog)
	logger.DoInfoLog("Build commit: "+buildCommit, zlog)

	interceptors := make([]grpc.ServerOption, 0)

	interceptors = append(interceptors,
		grpc.ChainUnaryInterceptor(
			decryptinterceptor.DecryptInterceptor(params),
			decompressinterceptor.DecompressInterceptor(),
		))
	grpcServer := grpc.NewServer(interceptors...)

	ctx, cancel := context.WithTimeout(
		context.Background(), params.WaitSecRespDB)
	defer cancel()

	conn, dataService, err = si.InitStorage(ctx, params)
	if err != nil {
		return fmt.Errorf("InitStorage: %w", err)
	}

	if conn != nil {
		defer conn.Close()
	}

	err = InitiateServer(params,
		dataService, server)
	if err != nil {
		return fmt.Errorf("InitiateServer: %w", err)
	}

	channelCancel := make(chan os.Signal, 1)

	waitGroup.Add(1)

	go si.SaveMetrics(&channelCancel,
		dataService, params, waitGroup)

	go RunGRPCServer(grpcServer,
		params, dataService)
	go si.RunServer(server)

	waitGroup.Wait()

	err = server.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("serverShutdown: %w", err)
	}

	grpcServer.Stop()

	err = dataService.SaveInFile(params.FileStoragePath)
	if err != nil {
		fmt.Println("main->SaveInFile: %w", err)
	}

	return nil
}
