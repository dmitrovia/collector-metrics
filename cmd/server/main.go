package main

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/dmitrovia/collector-metrics/internal/functions/validate"
	"github.com/dmitrovia/collector-metrics/internal/handlers/defaulthandler"
	"github.com/dmitrovia/collector-metrics/internal/handlers/getmetrichandler"
	"github.com/dmitrovia/collector-metrics/internal/handlers/getmetricjsonhandler"
	"github.com/dmitrovia/collector-metrics/internal/handlers/notallowedhandler"
	"github.com/dmitrovia/collector-metrics/internal/handlers/pinghandler"
	"github.com/dmitrovia/collector-metrics/internal/handlers/setmetrichandler"
	"github.com/dmitrovia/collector-metrics/internal/handlers/setmetricjsonhandler"
	"github.com/dmitrovia/collector-metrics/internal/logger"
	"github.com/dmitrovia/collector-metrics/internal/middleware/gzipcompressmiddleware"
	"github.com/dmitrovia/collector-metrics/internal/middleware/loggermiddleware"
	"github.com/dmitrovia/collector-metrics/internal/migrator"
	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
	"github.com/dmitrovia/collector-metrics/internal/service"
	dbrep "github.com/dmitrovia/collector-metrics/internal/storage/dbrepository"
	memrep "github.com/dmitrovia/collector-metrics/internal/storage/memoryrepository"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

const rTimeout = 15

const wTimeout = 15

const iTimeout = 60

const defPORT string = "localhost:8080"

const defSavePathFile string = "/internal/temp/metrics.json"

const defSavingIntervalDisk = 300

const defWaitSecRespDB = 10

var errParseFlags = errors.New("addr is not valid")

var errPath = errors.New("path is not valid")

const migrationsDir = "db/migrations"

const zapLogLevel = "info"

// const defPostgreConnURL = "postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable"

const defPostgreConnURL = ""

//go:embed db/migrations/*.sql
var MigrationsFS embed.FS

func main() {
	var (
		params      *bizmodels.InitParams
		server      *http.Server
		dataService *service.DataService
		conn        *pgx.Conn
	)

	server = new(http.Server)
	params = new(bizmodels.InitParams)
	waitGroup := new(sync.WaitGroup)

	zapLogger, err := initiate(params)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), params.WaitSecRespDB)

	conn, dataService, err = initStorage(ctx, params)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	initiateServer(params, dataService, server, zapLogger)

	err = useMigrations(params)
	if err != nil {
		if conn != nil {
			conn.Close(ctx)
		}

		cancel()
		os.Exit(1)
	}

	go runServer(server)

	waitGroup.Add(1)

	go saveMetrics(dataService, params, waitGroup)
	waitGroup.Wait()

	if conn != nil {
		conn.Close(ctx)
	}

	err = server.Shutdown(ctx)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = dataService.SaveInFile(params.FileStoragePath)
	if err != nil {
		fmt.Println("Error writing metrics to file: %w", err)
	}
}

func initStorage(ctx context.Context, par *bizmodels.InitParams) (*pgx.Conn, *service.DataService, error) {
	var (
		memStorage *memrep.MemoryRepository
		DBStorage  *dbrep.DBepository
	)

	DBStorage = new(dbrep.DBepository)
	memStorage = new(memrep.MemoryRepository)

	if par.DatabaseDSN != "" {
		datas := service.NewMemoryService(DBStorage)

		dbConn, err := pgx.Connect(ctx, par.DatabaseDSN)
		if err != nil {
			return nil, nil, fmt.Errorf("initStorage->pgx.Connect %w", err)
		}

		DBStorage.Initiate(par.DatabaseDSN, par.WaitSecRespDB, dbConn)

		return dbConn, datas, nil
	}

	datas := service.NewMemoryService(memStorage)
	memStorage.Init()

	return nil, datas, nil
}

func useMigrations(par *bizmodels.InitParams) error {
	if par.DatabaseDSN == "" {
		return nil
	}

	migrator, err := migrator.MustGetNewMigrator(MigrationsFS, migrationsDir)
	if err != nil {
		fmt.Println(err)

		return fmt.Errorf("useMigrations->migrator.MustGetNewMigrator %w", err)
	}

	conn, err := sql.Open("postgres", par.DatabaseDSN)
	if err != nil {
		fmt.Println(err)

		return fmt.Errorf("useMigrations->sql.Open %w", err)
	}

	defer conn.Close()

	err = migrator.ApplyMigrations(conn)
	if err != nil {
		fmt.Println(err)

		return fmt.Errorf("useMigrations->migrator.ApplyMigrations %w", err)
	}

	return nil
}

func saveMetrics(mser *service.DataService, par *bizmodels.InitParams, wg *sync.WaitGroup) {
	defer wg.Done()

	channelCancel := make(chan os.Signal, 1)
	signal.Notify(channelCancel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	if par.StoreInterval == 0 {
		err := mser.SaveInFile(par.FileStoragePath)
		if err != nil {
			fmt.Println("Error writing metrics to file: %w", err)
		}

		sig := <-channelCancel
		log.Println("Quitting after signal_1:", sig)
	} else {
		for {
			select {
			case sig := <-channelCancel:
				log.Println("Quitting after signal_2:", sig)

				return
			case <-time.After(time.Duration(par.StoreInterval) * time.Second):
				err := mser.SaveInFile(par.FileStoragePath)
				if err != nil {
					fmt.Println("Error writing metrics to file: %w", err)
				}
			}
		}
	}
}

func runServer(server *http.Server) {
	err := server.ListenAndServe()
	if err != nil {
		log.Printf("Error starting server: %s\n", err)

		return
	}
}

func initiate(par *bizmodels.InitParams) (*zap.Logger, error) {
	zlog, err := logger.Initialize(zapLogLevel)
	if err != nil {
		return nil, fmt.Errorf("initiate->logger.Initialize %w", err)
	}

	err = setInitParamsFileStorage(par)
	if err != nil {
		return nil, err
	}

	setInitParamsDB(par)
	par.ValidateAddrPattern = "^[a-zA-Z/ ]{1,100}:[0-9]{1,10}$"

	err = setInitParams(par)
	if err != nil {
		return nil, err
	}

	return zlog, nil
}

func initiateServer(par *bizmodels.InitParams, mser *service.DataService, server *http.Server, zapLogger *zap.Logger) {
	mux := mux.NewRouter()

	handlerPing := pinghandler.NewPingHandler(mser, par)
	handlerSet := setmetrichandler.NewSetMetricHandler(mser)
	handlerJSONSet := setmetricjsonhandler.NewSetMetricJSONHandler(mser)
	handlerJSONGet := getmetricjsonhandler.NewGetMetricJSONHandler(mser)
	handlerGet := getmetrichandler.NewGetMetricHandler(mser)
	handlerDefault := defaulthandler.NewDefaultHandler(mser)
	handlerNotAllowed := notallowedhandler.NotAllowedHandler{}

	setMetricMux := mux.Methods(http.MethodPost).Subrouter()
	setMetricMux.HandleFunc("/update/{metric_type}/{metric_name}/{metric_value}", handlerSet.SetMetricHandler)
	setMetricMux.Use(loggermiddleware.RequestLogger(zapLogger))

	getMEtricMux := mux.Methods(http.MethodGet).Subrouter()
	getMEtricMux.HandleFunc("/value/{metric_type}/{metric_name}", handlerGet.GetMetricHandler)
	getMEtricMux.Use(loggermiddleware.RequestLogger(zapLogger))

	getMEtricJSONMux := mux.Methods(http.MethodPost).Subrouter()
	getMEtricJSONMux.HandleFunc("/value/", handlerJSONGet.GetMetricJSONHandler)
	getMEtricJSONMux.Use(gzipcompressmiddleware.GzipMiddleware())
	getMEtricJSONMux.Use(loggermiddleware.RequestLogger(zapLogger))

	setMetricJSONMux := mux.Methods(http.MethodPost).Subrouter()
	setMetricJSONMux.HandleFunc("/update/", handlerJSONSet.SetMetricJSONHandler)
	setMetricJSONMux.Use(gzipcompressmiddleware.GzipMiddleware())
	setMetricJSONMux.Use(loggermiddleware.RequestLogger(zapLogger))

	getPingBDMux := mux.Methods(http.MethodGet).Subrouter()
	getPingBDMux.HandleFunc("/ping", handlerPing.PingHandler)
	getPingBDMux.Use(gzipcompressmiddleware.GzipMiddleware())
	getPingBDMux.Use(loggermiddleware.RequestLogger(zapLogger))

	mux.MethodNotAllowedHandler = handlerNotAllowed

	defaultMux := mux.Methods(http.MethodGet).Subrouter()
	defaultMux.HandleFunc("/", handlerDefault.DefaultHandler)
	defaultMux.Use(gzipcompressmiddleware.GzipMiddleware())
	defaultMux.Use(loggermiddleware.RequestLogger(zapLogger))

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
}

func setInitParamsDB(params *bizmodels.InitParams) {
	params.WaitSecRespDB = defWaitSecRespDB * time.Second

	envDatabaseDSN := os.Getenv("DATABASE_DSN")

	if envDatabaseDSN != "" {
		params.DatabaseDSN = envDatabaseDSN
	} else {
		flag.StringVar(&params.DatabaseDSN, "d", defPostgreConnURL, "database connection address.")
	}

	flag.Parse()
}

func setInitParamsFileStorage(params *bizmodels.InitParams) error {
	envFSP := os.Getenv("FILE_STORAGE_PATH")
	envRestore := os.Getenv("RESTORE")

	if envFSP != "" {
		params.FileStoragePath = envFSP
	} else {
		_, path, _, ok := runtime.Caller(0)

		if !ok {
			return fmt.Errorf("setInitParams->runtime.Caller( %w", errPath)
		}

		Root := filepath.Join(filepath.Dir(path), "../..")
		temp := Root + defSavePathFile
		fmt.Println(temp)
		flag.StringVar(&params.FileStoragePath, "f", temp, "Directory for saving metrics.")
	}

	if envRestore != "" {
		value, err := strconv.ParseBool(envRestore)
		if err != nil {
			return fmt.Errorf("setInitParams->ParseBool %w", err)
		}

		params.Restore = value
	} else {
		flag.BoolVar(&params.Restore, "r", true, "Loading metrics at server startup.")
	}

	flag.Parse()

	return nil
}

func setInitParams(params *bizmodels.InitParams) error {
	var err error

	envRA := os.Getenv("ADDRESS")
	envSI := os.Getenv("STORE_INTERVAL")

	if envRA != "" {
		params.PORT = envRA
	} else {
		flag.StringVar(&params.PORT, "a", defPORT, "Port to listen on.")
	}

	if envSI != "" {
		value, err := strconv.Atoi(envSI)
		if err != nil {
			return fmt.Errorf("setInitParams->Atoi %w", err)
		}

		params.StoreInterval = value
	} else {
		flag.IntVar(&params.StoreInterval, "i", defSavingIntervalDisk, "Metrics saving interval.")
	}

	flag.Parse()

	res, err := validate.IsMatchesTemplate(params.PORT, params.ValidateAddrPattern)
	if err != nil {
		return fmt.Errorf("setInitParams->IsMatchesTemplate: %w", err)
	}

	if !res {
		return errParseFlags
	}

	return nil
}
