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

type testData struct {
	sFileStoragePath       string // f
	sDefKeyHashSha256      string // k
	sDatabaseDSN           string // d
	sRestore               string // r
	sPORT                  string // a
	sDefSavingIntervalDisk string // i
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
			sPORT:                  "localhost:8090",
			sDefSavingIntervalDisk: "60",
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
}

func setEnv() error {
	err := os.Setenv("DATABASE_DSN",
		"postgres://postgres:postgres"+
			"@postgres"+
			":5432/praktikum?sslmode=disable")
	if err != nil {
		return fmt.Errorf("DATABASE_DSN: %w", err)
	}

	err = os.Setenv("RESTORE", "true")
	if err != nil {
		return fmt.Errorf("RESTORE: %w", err)
	}

	err = os.Setenv("ADDRESS", "localhost:8090")
	if err != nil {
		return fmt.Errorf("ADDRESS: %w", err)
	}

	err = os.Setenv("STORE_INTERVAL", "60")
	if err != nil {
		return fmt.Errorf("STORE_INTERVAL: %w", err)
	}

	err = os.Setenv("KEY", "defkey")
	if err != nil {
		return fmt.Errorf("KEY: %w", err)
	}

	err = os.Setenv("CRYPTO_KEY_SERVER",
		"/internal/asymcrypto/keys/private.pem")
	if err != nil {
		return fmt.Errorf("CRYPTO_KEY_SERVER: %w", err)
	}

	err = os.Setenv("CONFIG_SERVER",
		"/internal/config/server.json")
	if err != nil {
		return fmt.Errorf("CONFIG_SERVER: %w", err)
	}

	err = os.Setenv("TRUSTED_SUBNET", "0.0.0.0/0")
	if err != nil {
		return fmt.Errorf("TRUSTED_SUBNET: %w", err)
	}

	return nil
}

//nolint:funlen
func mainBody() {
	err := setEnv()
	if err != nil {
		fmt.Println("main->setEnv: %w", err)

		return
	}

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

	defer cancel()

	conn, dataService, err = si.InitStorage(ctx, params)
	if err != nil {
		fmt.Println("main->initStorage: %w", err)

		return
	}

	if conn != nil {
		defer conn.Close()
	}

	err = si.InitiateServer(params, dataService, server, zlog)
	if err != nil {
		fmt.Println("main->InitiateServer: %w", err)

		return
	}

	time.Sleep(time.Duration(20) * time.Second)

	go si.RunServer(server)

	waitGroup.Add(1)

	channelCancel := make(chan os.Signal, 1)

	go si.SaveMetrics(&channelCancel,
		dataService, params, waitGroup)

	go exit(ctx, &channelCancel, server)
	waitGroup.Wait()

	err = dataService.SaveInFile(params.FileStoragePath)
	if err != nil {
		fmt.Println("main->SaveInFile: %w", err)
	}
}

func exit(
	ctx context.Context,
	chc *chan os.Signal,
	server *http.Server,
) {
	<-time.After(time.Duration(30) * time.Second)

	err := server.Shutdown(ctx)
	if err != nil {
		fmt.Println("main->Shutdown: %w", err)

		return
	}

	*chc <- syscall.SIGTERM
}

/*
var errServer = errors.New("proc return errors")

ctx, cancel := context.WithTimeout(
	context.Background(), 60*time.Second)
defer cancel()

args := []string{
	"-a=" + test.sPORT,
	"-d=" + test.sDatabaseDSN,
	"-f=" + test.sFileStoragePath,
	"-i=" + test.sDefSavingIntervalDisk,
	"-k=" + test.sDefKeyHashSha256,
	"-r=" + test.sRestore,
}

cmd := exec.CommandContext(ctx, "./server", args...)
out, err := cmd.CombinedOutput()
sout := string(out)

if err != nil &&
	!strings.Contains(err.Error(), "signal: killed") {
	t.Errorf("%v", err)
}

if strings.Contains(sout, "->") {
	fmt.Println(sout)
	t.Errorf("%v", errServer)
}

fmt.Println(sout) */
