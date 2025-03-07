package getmetricjsonhandler_test

import (
	"bytes"
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/dmitrovia/collector-metrics/internal/handlers/getmetricjsonhandler"
	"github.com/dmitrovia/collector-metrics/internal/logger"
	"github.com/dmitrovia/collector-metrics/internal/middleware/gzipcompressmiddleware"
	"github.com/dmitrovia/collector-metrics/internal/middleware/loggermiddleware"
	"github.com/dmitrovia/collector-metrics/internal/migrator"
	"github.com/dmitrovia/collector-metrics/internal/models/apimodels"
	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
	"github.com/dmitrovia/collector-metrics/internal/service"
	"github.com/dmitrovia/collector-metrics/internal/storage/dbrepository"
	"github.com/dmitrovia/collector-metrics/internal/storage/memoryrepository"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
)

const url string = "http://localhost:8080"

const stok int = http.StatusOK

const nfnd int = http.StatusNotFound

const post string = "POST"

const defSavePathFile string = "/internal/temp/"

const defSavePathFile1 string = "/internal/temp/met.json"

var errRuntimeCaller = errors.New("errRuntimeCaller")

var errPath = errors.New("path is not valid")

const migrationsDir = "db/migrations"

//go:embed db/migrations/*.sql
var MigrationsFS embed.FS

type testData struct {
	tn     string
	mt     string
	mn     string
	exbody string
	meth   string
	delta  int64
	expcod int
	value  float64
}

func getTestData() []testData {
	return []testData{
		{
			meth: post, tn: "1", mt: bizmodels.GaugeName,
			mn: "Name1w", value: 1, expcod: stok,
		},
		{
			meth: post, tn: "2", mt: bizmodels.CounterName,
			mn: "Name2w", delta: 1, expcod: stok,
		},
		{
			meth: post, tn: "3", mt: bizmodels.CounterName,
			mn: "Name__22", delta: 1, expcod: nfnd,
		},
		{
			meth: post, tn: "4", mt: bizmodels.CounterName,
			mn: "Name__22****1", delta: 1, expcod: nfnd,
		},
		{
			meth: post, tn: "5", mt: bizmodels.CounterName,
			mn: "Name3w", delta: 999999999, expcod: stok,
		},
		{
			meth: post, tn: "6", mt: bizmodels.CounterName,
			mn: "Name4w", delta: -999999999, expcod: stok,
		},
		{
			meth: post, tn: "7", mt: bizmodels.CounterName,
			mn: "Name5w", delta: 0, expcod: stok,
		},
		{
			meth: post, tn: "8", mt: bizmodels.CounterName,
			mn: "Name6w", delta: 7456, expcod: stok,
		},
		{
			meth: post, tn: "9", mt: bizmodels.CounterName,
			mn: "Name7w", delta: -1, expcod: stok,
		},
		{
			meth: post, tn: "10", mt: bizmodels.CounterName,
			mn: "Name8w", delta: 555, expcod: stok,
		},
		{
			meth: post, tn: "101", mt: bizmodels.CounterName,
			mn: "Name58888", delta: 555, expcod: nfnd,
		},
	}
}

func getTD2() []testData {
	return []testData{
		{
			meth: post, tn: "11", mt: bizmodels.GaugeName,
			mn: "Name9w", value: 1.0, expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "12", mt: bizmodels.GaugeName,
			mn: "Name9w", value: 1.0, expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "13", mt: bizmodels.GaugeName,
			mn: "Name__22", value: 1.0, expcod: nfnd, exbody: "",
		},
		{
			meth: post, tn: "14", mt: bizmodels.GaugeName,
			mn: "Name__22****1", value: 1.0, expcod: nfnd,
			exbody: "",
		},
		{
			meth: post, tn: "15", mt: bizmodels.GaugeName,
			mn: "Name10w", value: 999999999.62,
			expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "16", mt: bizmodels.GaugeName,
			mn: "Name11w", value: -999999999.38,
			expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "17", mt: bizmodels.GaugeName,
			mn: "Name13w", value: 0, expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "18", mt: bizmodels.GaugeName,
			mn: "Name13w", value: 7456.3231, expcod: stok,
			exbody: "",
		},
		{
			meth: post, tn: "19", mt: bizmodels.GaugeName,
			mn: "Name14w", value: -1, expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "20", mt: bizmodels.GaugeName,
			mn: "Name15w", value: 555, expcod: stok, exbody: "",
		},
		{
			meth: post, tn: "21", mt: bizmodels.GaugeName,
			mn: "Name58888", value: 555, expcod: nfnd, exbody: "",
		},
	}
}

func setHandlerParams(params *bizmodels.InitParams) error {
	params.
		ValidateAddrPattern = "^[a-zA-Z/ ]{1,100}:[0-9]{1,10}$"
	params.DatabaseDSN = "postgres://postgres:postgres" +
		"@postgres" +
		":5432/praktikum?sslmode=disable"
	_, path, _, ok := runtime.Caller(0)

	if !ok {
		return fmt.Errorf("setInitParams->runtime.Caller( %w",
			errPath)
	}

	Root := filepath.Join(filepath.Dir(path), "../..")
	params.FileStoragePath = Root + defSavePathFile
	params.Key = "defaultKey"
	params.Restore = true
	params.ValidateAddrPattern = ""
	params.WaitSecRespDB = 10 * time.Second

	return nil
}

func initiate(
	mux *mux.Router,
	params *bizmodels.InitParams,
	isMemRepo bool,
) error {
	err := setHandlerParams(params)
	if err != nil {
		return fmt.Errorf("initiate->psetHandlerParams %w", err)
	}

	storage := &dbrepository.DBepository{}
	memst := &memoryrepository.MemoryRepository{}
	tmp := params.WaitSecRespDB

	var dse *service.DS

	ctx, cancel := context.WithTimeout(
		context.Background(), tmp)
	defer cancel()

	if isMemRepo {
		dse = service.NewMemoryService(memst, tmp)
	} else {
		dse = service.NewMemoryService(storage, tmp)
	}

	dbConn, err := pgxpool.New(ctx, params.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("initStorage->pgxpool.New %w", err)
	}

	if isMemRepo {
		memst.Init()
	} else {
		storage.Initiate(params.DatabaseDSN, dbConn)
	}

	hJSONGet := getmetricjsonhandler.NewGetMJSONHandler(dse)

	zapLogger, err := logger.Initialize("info")
	if err != nil {
		return fmt.Errorf("initiate: %w", err)
	}

	getMJSONMux := mux.Methods(http.MethodPost).Subrouter()
	getMJSONMux.HandleFunc("/value/",
		hJSONGet.GetMetricJSONHandler)
	getMJSONMux.Use(gzipcompressmiddleware.GzipMiddleware(),
		loggermiddleware.RequestLogger(zapLogger))

	if isMemRepo {
		LoadFile(dse, "test3.txt")
	} else {
		_ = UseMigrations(params)

		LoadFile(dse, "test2.txt")
	}

	saveMetrics(dse)

	return nil
}

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

func TestGetMetricJSONHandler(t *testing.T) {
	t.Helper()
	t.Parallel()

	params := &bizmodels.InitParams{}

	result := make([]testData, 0)
	result = append(result, getTestData()...)
	result = append(result, getTD2()...)

	mux := mux.NewRouter()

	err := initiate(mux, params, false)
	if err != nil {
		fmt.Println(err)

		return
	}

	for _, test := range result {
		t.Run(http.MethodPost, func(tobj *testing.T) {
			tobj.Parallel()

			reqData, err := formReqBody(&test)
			if err != nil {
				fmt.Println(err)

				return
			}

			req, err := http.NewRequestWithContext(
				context.Background(),
				test.meth,
				url+"/value/", reqData)
			if err != nil {
				tobj.Fatal(err)
			}

			req.Header.Set("Content-Type", "application/json")

			newr := httptest.NewRecorder()
			mux.ServeHTTP(newr, req)
			status := newr.Code

			assert.Equal(tobj,
				test.expcod,
				status, test.tn+": Response code didn't match expected")
		})
	}
}

func saveMetrics(service *service.DS) {
	_, path, _, ok := runtime.Caller(0)

	if !ok {
		fmt.Println("err")

		return
	}

	Root := filepath.Join(filepath.Dir(path), "../../..")

	FileStoragePath := Root + defSavePathFile1

	err := service.SaveInFile(FileStoragePath)
	if err != nil {
		fmt.Println("err")
	}
}

func formReqBody(
	data *testData,
) (*bytes.Reader, error) {
	metr := &apimodels.Metrics{}
	metr.MType = data.mt
	metr.ID = data.mn
	metr.Delta = &data.delta
	metr.Value = &data.value

	marshall, err := json.Marshal(metr)
	if err != nil {
		return nil,
			fmt.Errorf("formReqBody->Marshal: %w",
				err)
	}

	return bytes.NewReader(marshall), nil
}

func LoadFile(mems *service.DS,
	filen string,
) {
	_, path, _, ok := runtime.Caller(0)

	if !ok {
		fmt.Println(errRuntimeCaller)
	}

	Root := filepath.Join(filepath.Dir(path), "../../..")
	temp := Root + defSavePathFile + filen

	err := mems.LoadFromFile(temp)
	if err != nil {
		fmt.Println("Error reading metrics from file: %w", err)
	}
}
