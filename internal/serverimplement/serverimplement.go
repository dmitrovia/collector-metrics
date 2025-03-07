// Package for implementing server methods.
package serverimplement

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/dmitrovia/collector-metrics/internal/functions/config"
	"github.com/dmitrovia/collector-metrics/internal/functions/validate"
	"github.com/dmitrovia/collector-metrics/internal/handlers/defaulthandler"
	"github.com/dmitrovia/collector-metrics/internal/handlers/getmetrichandler"
	"github.com/dmitrovia/collector-metrics/internal/handlers/getmetricjsonhandler"
	"github.com/dmitrovia/collector-metrics/internal/handlers/notallowedhandler"
	"github.com/dmitrovia/collector-metrics/internal/handlers/pinghandler"
	"github.com/dmitrovia/collector-metrics/internal/handlers/sender"
	"github.com/dmitrovia/collector-metrics/internal/handlers/setmetrichandler"
	"github.com/dmitrovia/collector-metrics/internal/handlers/setmetricjsonhandler"
	"github.com/dmitrovia/collector-metrics/internal/logger"
	"github.com/dmitrovia/collector-metrics/internal/middleware/decryptmid"
	"github.com/dmitrovia/collector-metrics/internal/middleware/gzipcompressmiddleware"
	"github.com/dmitrovia/collector-metrics/internal/middleware/loggermiddleware"
	"github.com/dmitrovia/collector-metrics/internal/migrator"
	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
	"github.com/dmitrovia/collector-metrics/internal/service"
	"github.com/dmitrovia/collector-metrics/internal/storage/dbrepository"
	"github.com/dmitrovia/collector-metrics/internal/storage/memoryrepository"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const rTimeout = 60

const wTimeout = 60

const iTimeout = 60

const defPORT string = ""

const defSavePathFile string = ""

const defCryptoKeyPath string = ""

const defConfigPath string = "/internal/config/server.json"

const defSavingIntervalDisk = 0

const defWaitSecRespDB = 10

var errParseFlags = errors.New("addr is not valid")

var errPath = errors.New("path is not valid")

const migrationsDir = "db/migrations"

const zapLogLevel = "info"

const defPostgreConnURL = ""

const defKeyHashSha256 = ""

//go:embed db/migrations/*.sql
var MigrationsFS embed.FS

// InitStorage - initializes storage, database, or RAM.
func InitStorage(
	ctx context.Context, par *bizmodels.InitParams,
) (*pgxpool.Pool, *service.DS, error) {
	var (
		memStorage *memoryrepository.MemoryRepository
		DBStorage  *dbrepository.DBepository
	)

	DBStorage = &dbrepository.DBepository{}
	memStorage = &memoryrepository.MemoryRepository{}

	if par.DatabaseDSN != "" {
		datas := service.NewMemoryService(DBStorage,
			par.WaitSecRespDB)

		dbConn, err := pgxpool.New(ctx, par.DatabaseDSN)
		if err != nil {
			return nil, nil,
				fmt.Errorf("initStorage->pgx.Connect %w",
					err)
		}

		DBStorage.Initiate(par.DatabaseDSN, dbConn)

		return dbConn, datas, nil
	}

	datas := service.NewMemoryService(memStorage,
		par.WaitSecRespDB)

	memStorage.Init()

	return nil, datas, nil
}

// UseMigrations - starts working with migrations.
func UseMigrations(par *bizmodels.InitParams) error {
	if par.DatabaseDSN == "" {
		return nil
	}

	migrator, err := migrator.MustGetNewMigrator(
		MigrationsFS, migrationsDir)
	if err != nil {
		return fmt.Errorf("useMigrations->MustGetNewMigrator %w",
			err)
	}

	conn, err := sql.Open("postgres", par.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("useMigrations->sql.Open %w", err)
	}

	defer conn.Close()

	err = migrator.ApplyMigrations(conn)
	if err != nil {
		return fmt.Errorf("useMigrations->ApplyMigrations %w",
			err)
	}

	return nil
}

// SaveMetrics - writes metrics to file
// once or every StoreInterval seconds.
func SaveMetrics(
	chc *chan os.Signal,
	mser *service.DS,
	par *bizmodels.InitParams, wg *sync.WaitGroup,
) {
	defer wg.Done()

	signal.Notify(*chc,
		os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	if par.StoreInterval == 0 {
		err := mser.SaveInFile(par.FileStoragePath)
		if err != nil {
			fmt.Println("Error writing metrics to file: %w", err)
		}

		sig := <-*chc
		log.Println("Quitting after signal_1:", sig)
	} else {
		for {
			select {
			case sig := <-*chc:
				log.Println("Quitting after signal_2:", sig)

				return
			case <-time.After(
				time.Duration(par.StoreInterval) * time.Second):
				err := mser.SaveInFile(par.FileStoragePath)
				if err != nil {
					fmt.Println("Error writing metrics to file: %w", err)
				}
			}
		}
	}
}

// RunServer - starts the server.
func RunServer(server *http.Server) {
	err := server.ListenAndServe()
	if err != nil {
		log.Printf("Error starting server: %s\n", err)

		return
	}
}

// Initialization - initializes
// data for the server to operate.
func Initiate(
	par *bizmodels.InitParams,
) (*zap.Logger, error) {
	err := initiateFlags(par)
	if err != nil {
		return nil, fmt.Errorf("initiate->initiateFlags %w", err)
	}

	err = getParamsFromCFG(par)
	if err != nil {
		return nil, fmt.Errorf(
			"initiate->getParamsFromCFG %w", err)
	}

	zlog, err := logger.Initialize(zapLogLevel)
	if err != nil {
		return nil, fmt.Errorf("initiate->logger.Initialize %w",
			err)
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

// initiateFlags - parses passed flags into variables.
func initiateFlags(par *bizmodels.InitParams) error {
	_, path, _, ok := runtime.Caller(0)

	if !ok {
		return fmt.Errorf("setInitParams->runtime.Caller( %w",
			errPath)
	}

	Root := filepath.Join(filepath.Dir(path), "../..")
	temp := Root + defSavePathFile

	flag.StringVar(&par.ConfigPath,
		"config", Root+defConfigPath, "cfg server path.")
	flag.StringVar(&par.PORT,
		"a", defPORT, "Port to listen on.")
	flag.StringVar(&par.DatabaseDSN,
		"d", defPostgreConnURL, "database connection address.")
	flag.StringVar(&par.FileStoragePath,
		"f", temp, "Directory for saving metrics.")
	flag.IntVar(&par.StoreInterval,
		"i", defSavingIntervalDisk, "Metrics saving interval.")
	flag.StringVar(&par.Key, "k", defKeyHashSha256,
		"key for signatures for the SHA256 algorithm.")
	flag.StringVar(&par.CryptoPrivateKeyPath,
		"crypto-key", Root+defCryptoKeyPath,
		"asymmetric encryption pivate key.")
	flag.BoolVar(&par.Restore,
		"r", true, "Loading metrics at server startup.")
	flag.Parse()

	return nil
}

// AttachProfiler - defining handlers for pprof.
func AttachProfiler(router *mux.Router) {
	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
}

// InitiateServer - initializes server data.
func InitiateServer(
	par *bizmodels.InitParams,
	mser *service.DS,
	server *http.Server,
	zapLogger *zap.Logger,
) error {
	mux := mux.NewRouter()
	AttachProfiler(mux)

	initPostMethods(mux, mser, zapLogger, par)
	initGetMethods(mux, mser, zapLogger, par)

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

	err := UseMigrations(par)
	if err != nil {
		fmt.Println("InitiateServer->UseMigrations: %w", err)

		return err
	}

	return nil
}

// initGetMethods - initializes get handlers.
func initGetMethods(
	mux *mux.Router,
	dse *service.DS,
	zapLogger *zap.Logger,
	par *bizmodels.InitParams,
) {
	hPing := pinghandler.NewPingHandler(dse, par)
	hGet := getmetrichandler.NewGetMetricHandler(dse)
	hDefault := defaulthandler.NewDefaultHandler(dse)
	hNotAllowed := notallowedhandler.NotAllowedHandler{}

	// mux.PathPrefix("/debug/").Handler(http.DefaultServeMux)

	getMMux := mux.Methods(http.MethodGet).Subrouter()
	getMMux.HandleFunc(
		"/value/{metric_type}/{metric_name}",
		hGet.GetMetricHandler)
	getMMux.Use(loggermiddleware.RequestLogger(zapLogger))

	getPingBDMux := mux.Methods(http.MethodGet).Subrouter()
	getPingBDMux.HandleFunc("/ping", hPing.PingHandler)
	getPingBDMux.Use(gzipcompressmiddleware.GzipMiddleware(),
		loggermiddleware.RequestLogger(zapLogger))

	mux.MethodNotAllowedHandler = hNotAllowed

	defaultMux := mux.Methods(http.MethodGet).Subrouter()
	defaultMux.HandleFunc("/", hDefault.DefaultHandler)
	defaultMux.Use(gzipcompressmiddleware.GzipMiddleware(),
		loggermiddleware.RequestLogger(zapLogger))
}

// initPostMethods - initializes post handlers.
func initPostMethods(
	mux *mux.Router,
	dse *service.DS,
	zapLogger *zap.Logger,
	par *bizmodels.InitParams,
) {
	hSet := setmetrichandler.NewSetMetricHandler(dse)
	hJSONSet := setmetricjsonhandler.NewSetMJH(dse)
	hJSONSets := sender.NewSenderHandler(
		dse, par)
	hJSONGet := getmetricjsonhandler.NewGetMJSONHandler(dse)

	setMMux := mux.Methods(http.MethodPost).Subrouter()
	setMMux.HandleFunc(
		"/update/{metric_type}/{metric_name}/{metric_value}",
		hSet.SetMetricHandler)
	setMMux.Use(loggermiddleware.RequestLogger(zapLogger))

	getMJSONMux := mux.Methods(http.MethodPost).Subrouter()
	getMJSONMux.HandleFunc(
		"/value/",
		hJSONGet.GetMetricJSONHandler)
	getMJSONMux.Use(gzipcompressmiddleware.GzipMiddleware(),
		loggermiddleware.RequestLogger(zapLogger))

	setMJSONMux := mux.Methods(http.MethodPost).Subrouter()
	setMJSONMux.HandleFunc(
		"/update/",
		hJSONSet.SetMJSONHandler)
	setMJSONMux.Use(gzipcompressmiddleware.GzipMiddleware(),
		loggermiddleware.RequestLogger(zapLogger))

	setMsJSONMux := mux.Methods(http.MethodPost).Subrouter()
	setMsJSONMux.HandleFunc(
		"/updates/",
		hJSONSets.SenderHandler)
	setMsJSONMux.Use(
		decryptmid.DecryptMiddleware(*par),
		gzipcompressmiddleware.GzipMiddleware(),
		loggermiddleware.RequestLogger(zapLogger))
}

// setInitParamsDB - gets environment variables.
func setInitParamsDB(params *bizmodels.InitParams) {
	params.WaitSecRespDB = defWaitSecRespDB * time.Second

	envDatabaseDSN := os.Getenv("DATABASE_DSN")

	if envDatabaseDSN != "" {
		params.DatabaseDSN = envDatabaseDSN
	}
}

// setInitParamsFileStorage - gets environment variables.
func setInitParamsFileStorage(
	params *bizmodels.InitParams,
) error {
	// envFSP := os.Getenv("FILE_STORAGE_PATH")
	envRestore := os.Getenv("RESTORE")

	// if envFSP != "" {
	//	params.FileStoragePath = envFSP
	//}

	if envRestore != "" {
		value, err := strconv.ParseBool(envRestore)
		if err != nil {
			return fmt.Errorf("setInitParams->ParseBool %w", err)
		}

		params.Restore = value
	}

	return nil
}

// setInitParams - gets environment variables.
//
//nolint:cyclop
func setInitParams(params *bizmodels.InitParams) error {
	envRA := os.Getenv("ADDRESS")
	envSI := os.Getenv("STORE_INTERVAL")
	key := os.Getenv("KEY")
	cryptoKey := os.Getenv("CRYPTO_KEY_SERVER")
	cfgServer := os.Getenv("CONFIG_SERVER")

	_, path, _, isok := runtime.Caller(0)
	Root := filepath.Join(filepath.Dir(path), "../..")

	if cfgServer != "" && isok {
		params.ConfigPath = Root + cfgServer
	}

	if cryptoKey != "" && isok {
		params.CryptoPrivateKeyPath = Root + cryptoKey
	}

	if key != "" {
		params.Key = key
	}

	if envRA != "" {
		params.PORT = envRA
	}

	if envSI != "" {
		value, err := strconv.Atoi(envSI)
		if err != nil {
			return fmt.Errorf("setInitParams->Atoi %w", err)
		}

		params.StoreInterval = value
	}

	res, err := validate.IsMatchesTemplate(
		params.PORT, params.ValidateAddrPattern)
	if err != nil {
		return fmt.Errorf("setInitParams->IsMatchesTemplate: %w",
			err)
	}

	if !res {
		return errParseFlags
	}

	return nil
}

func getParamsFromCFG(
	par *bizmodels.InitParams,
) error {
	cfg, err := config.LoadConfigServer(par.ConfigPath)
	if err != nil {
		return fmt.Errorf(
			"getParamsFromCFG->LoadConfigServer: %w",
			err)
	}

	if par.CryptoPrivateKeyPath == "" {
		par.CryptoPrivateKeyPath = cfg.CryptoPrivateKeyPath
	}

	if par.DatabaseDSN == "" {
		par.DatabaseDSN = cfg.DatabaseDSN
	}

	if par.FileStoragePath == "" {
		par.FileStoragePath = cfg.FileStoragePath
	}

	if par.Key == "" {
		par.Key = cfg.Key
	}

	if par.PORT == "" {
		par.PORT = cfg.PORT
	}

	if !par.Restore {
		par.Restore = cfg.Restore
	}

	if par.StoreInterval == 0 {
		par.StoreInterval = cfg.StoreInterval
	}

	return nil
}
