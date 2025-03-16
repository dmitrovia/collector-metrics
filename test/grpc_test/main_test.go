package main_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/dmitrovia/collector-metrics/internal/Interceptors/decompressinterceptor"
	"github.com/dmitrovia/collector-metrics/internal/Interceptors/decryptinterceptor"
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

type testData struct {
	sFileStoragePath       string // f
	sDefKeyHashSha256      string // k
	sDatabaseDSN           string // d
	sRestore               string // r
	sPORT                  string // a
	sDefSavingIntervalDisk string // i
	grpcPort               string
}

func getTestData() *[]testData {
	return &[]testData{
		{
			sFileStoragePath:  "../../internal/temp/metrics.json",
			sDefKeyHashSha256: "defkey",
			sDatabaseDSN: "postgres://postgres:postgres" +
				"@postgres" +
				":5432/praktikum?sslmode=disable",
			sRestore:               "false",
			sPORT:                  "localhost:8091",
			sDefSavingIntervalDisk: "60",
			grpcPort:               "50053",
		},
	}
}

func TestMain(t *testing.T) {
	testCases := getTestData()

	t.Helper()
	t.Parallel()

	for _, test := range *testCases {
		t.Run("server", func(tobj *testing.T) {
			tobj.Parallel()

			addFlags(&test)
			mainBody()
		})
	}
}

func addFlags(test *testData) {
	os.Args = append(os.Args, "-a="+test.sPORT)
	os.Args = append(os.Args, "-d="+test.sDatabaseDSN)
	os.Args = append(os.Args, "-f="+test.sFileStoragePath)
	os.Args = append(os.Args,
		"-i="+test.sDefSavingIntervalDisk)
	os.Args = append(os.Args, "-k="+test.sDefKeyHashSha256)
	os.Args = append(os.Args, "-r="+test.sRestore)
	os.Args = append(os.Args, "-grpcp="+test.grpcPort)
}

//nolint:funlen
func mainBody() {
	var (
		dataService *service.DS
		conn        *pgxpool.Pool
	)

	server := &http.Server{}
	params := &bizmodels.InitParams{}
	waitGroup := &sync.WaitGroup{}

	interceptors := make([]grpc.ServerOption, 0)

	interceptors = append(interceptors,
		grpc.ChainUnaryInterceptor(
			decryptinterceptor.DecryptInterceptor(params),
			decompressinterceptor.DecompressInterceptor(),
		))
	grpcServer := grpc.NewServer(interceptors...)

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

	go exit(ctx, &channelCancel, server, grpcServer)
	waitGroup.Wait()
}

func exit(
	ctx context.Context,
	chc *chan os.Signal,
	server *http.Server,
	grpcS *grpc.Server,
) {
	<-time.After(time.Duration(30) * time.Second)

	err := server.Shutdown(ctx)
	if err != nil {
		fmt.Println("main->Shutdown: %w", err)

		return
	}

	grpcS.Stop()

	*chc <- syscall.SIGTERM
}
