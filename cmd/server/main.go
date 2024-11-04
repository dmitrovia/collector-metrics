package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/dmitrovia/collector-metrics/internal/functions/validate"
	"github.com/dmitrovia/collector-metrics/internal/handlers/defaulthandler"
	"github.com/dmitrovia/collector-metrics/internal/handlers/getmetrichandler"
	"github.com/dmitrovia/collector-metrics/internal/handlers/getmetricjsonhandler"
	"github.com/dmitrovia/collector-metrics/internal/handlers/notallowedhandler"
	"github.com/dmitrovia/collector-metrics/internal/handlers/setmetrichandler"
	"github.com/dmitrovia/collector-metrics/internal/handlers/setmetricjsonhandler"
	"github.com/dmitrovia/collector-metrics/internal/logger"
	"github.com/dmitrovia/collector-metrics/internal/middleware/gzipcompressmiddleware"
	"github.com/dmitrovia/collector-metrics/internal/middleware/loggermiddleware"
	"github.com/dmitrovia/collector-metrics/internal/service"
	"github.com/dmitrovia/collector-metrics/internal/storage/memoryrepository"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

const rTimeout = 15

const wTimeout = 15

const iTimeout = 60

const defPORT string = "localhost:8080"

const defSavePathFile string = "../../internal/temp/metrics.json"

const defSavingIntervalDisk = 300

var errParseFlags = errors.New("addr is not valid")

type initParams struct {
	PORT                string
	validateAddrPattern string
	storeInterval       int
	FileStoragePath     string
	restore             bool
}

func main() {
	var (
		memStorage  *memoryrepository.MemoryRepository
		params      *initParams
		server      *http.Server
		zapLogLevel string
	)

	memStorage = new(memoryrepository.MemoryRepository)
	MemoryService := service.NewMemoryService(memStorage)
	memStorage.Init()

	waitGroup := new(sync.WaitGroup)

	server = new(http.Server)

	params = new(initParams)
	params.validateAddrPattern = "^[a-zA-Z/ ]{1,100}:[0-9]{1,10}$"

	zapLogLevel = "info"

	zapLogger, err := logger.Initialize(zapLogLevel)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = initiate(params, MemoryService, server, zapLogger)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if params.restore {
		err := MemoryService.LoadFromFile(params.FileStoragePath)
		if err != nil {
			fmt.Println("Error reading metrics from file: %w", err)
		}
	}

	go runServer(server)

	waitGroup.Add(1)

	go saveMetrics(MemoryService, params, waitGroup)
	waitGroup.Wait()

	err = server.Shutdown(context.TODO())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = MemoryService.SaveInFile(params.FileStoragePath)
	if err != nil {
		fmt.Println("Error writing metrics to file: %w", err)
	}
}

func saveMetrics(mser *service.MemoryService, par *initParams, wg *sync.WaitGroup) {
	defer wg.Done()

	channelCancel := make(chan os.Signal, 1)
	signal.Notify(channelCancel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	if par.storeInterval == 0 {
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
			case <-time.After(time.Duration(par.storeInterval) * time.Second):
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

func initiate(par *initParams, mser *service.MemoryService, server *http.Server, zapLogger *zap.Logger) error {
	err := setInitParams(par)
	if err != nil {
		return err
	}

	mux := mux.NewRouter()

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

	return err
}

func setInitParams(params *initParams) error {
	var err error

	envRA := os.Getenv("ADDRESS")
	envSI := os.Getenv("STORE_INTERVAL")
	envFSP := os.Getenv("FILE_STORAGE_PATH")
	envRestore := os.Getenv("RESTORE")

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

		params.storeInterval = value
	} else {
		flag.IntVar(&params.storeInterval, "i", defSavingIntervalDisk, "Metrics saving interval.")
	}

	if envFSP != "" {
		params.FileStoragePath = envFSP
	} else {
		flag.StringVar(&params.FileStoragePath, "f", defSavePathFile, "Directory for saving metrics.")
	}

	if envRestore != "" {
		value, err := strconv.ParseBool(envRestore)
		if err != nil {
			return fmt.Errorf("setInitParams->ParseBool %w", err)
		}

		params.restore = value
	} else {
		flag.BoolVar(&params.restore, "r", true, "Loading metrics at server startup.")
	}

	flag.Parse()

	res, err := validate.IsMatchesTemplate(params.PORT, params.validateAddrPattern)
	if err != nil {
		return fmt.Errorf("setInitParams->IsMatchesTemplate: %w", err)
	}

	if !res {
		return errParseFlags
	}

	return nil
}
