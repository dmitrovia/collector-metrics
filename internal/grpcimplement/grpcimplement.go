package grpcimplement

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/dmitrovia/collector-metrics/internal/grpchandlers"
	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
	"github.com/dmitrovia/collector-metrics/internal/serverimplement"
	"github.com/dmitrovia/collector-metrics/internal/service"
	pb "github.com/dmitrovia/collector-metrics/pkg/microservice/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

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
		return fmt.Errorf(
			"InitiateServer->RegisterMicroServiceHandlerServer %w",
			err)
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
		err := mser.LoadFromFile(par.FileStoragePath)
		if err != nil {
			fmt.Println("Error reading metrics from file: %w", err)
		}
	}

	err = serverimplement.UseMigrations(par)
	if err != nil {
		return fmt.Errorf("InitiateServer->UseMigrations %w", err)
	}

	return nil
}
