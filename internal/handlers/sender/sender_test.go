package sender_test

import (
	"bytes"
	"context"
	"database/sql"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/dmitrovia/collector-metrics/internal/functions/asymcrypto"
	"github.com/dmitrovia/collector-metrics/internal/functions/compress"
	"github.com/dmitrovia/collector-metrics/internal/functions/hash"
	"github.com/dmitrovia/collector-metrics/internal/handlers/sender"
	"github.com/dmitrovia/collector-metrics/internal/logger"
	"github.com/dmitrovia/collector-metrics/internal/middleware/decryptmid"
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

const bdreq int = http.StatusBadRequest

const defSavePathFile string = "/internal/temp/metrics.json"

var errPath = errors.New("path is not valid")

var errResponse = errors.New("error response")

const migrationsDir = "db/migrations"

//go:embed db/migrations/*.sql
var MigrationsFS embed.FS

type testData struct {
	tn       string
	exbody   string
	key      string
	hash     string
	counters []bizmodels.Counter
	gauges   []bizmodels.Gauge
	expcod   int
}

type viewData struct {
	id    string
	mtype string
	delta int64
	value float64
}

func getTestData() *[]testData {
	tempC := []bizmodels.Counter{
		{Name: "Name1", Value: 1},
		{Name: "Name1", Value: 1},
		{Name: "Name__22", Value: 1},
		{Name: "Name__22****1", Value: 1},
		{Name: "Name2", Value: 999999999},
		{Name: "Name2", Value: -999999999},
		{Name: "Name4", Value: 0},
		{Name: "Name5", Value: 7456},
		{Name: "Name6", Value: -1},
		{Name: "Name343", Value: 555},
		{Name: randomString(5), Value: 0},
	}

	tempC1 := []bizmodels.Counter{
		{Name: "Name1", Value: 1},
	}

	tempG := []bizmodels.Gauge{
		{Name: "Name1", Value: 1.0},
		{Name: "Name1", Value: 1.0},
		{Name: "Name__22", Value: 1.0},
		{Name: "Name__22****1", Value: 1.0},
		{Name: "Name2", Value: 999999999.62},
		{Name: "Name3", Value: -999999999.38},
		{Name: "Name4", Value: 0},
		{Name: "Name5", Value: 7456.3231},
		{Name: "Name6", Value: -1.0},
		{Name: "Name343", Value: 555},
		{Name: randomString(5), Value: 0},
	}

	tempG1 := []bizmodels.Gauge{
		{Name: "Name1", Value: 1.0},
	}

	tempC2 := []bizmodels.Counter{}

	tempG2 := []bizmodels.Gauge{}

	return &[]testData{
		{
			counters: tempC, gauges: tempG,
			tn: "1", expcod: stok, exbody: "", key: "defaultKey",
		},
		{
			counters: tempC1, gauges: tempG1,
			tn: "2", expcod: stok, exbody: "", key: "samekey",
		},
		{
			counters: tempC2, gauges: tempG2,
			tn: "3", expcod: bdreq, exbody: "", key: "",
		},
		{
			counters: tempC2, gauges: tempG2,
			tn: "3", expcod: bdreq, exbody: "", key: "", hash: "123",
		},
	}
}

func UseMigrations(par *bizmodels.InitParams) error {
	if par.DatabaseDSN == "" {
		return nil
	}

	migrator, err := migrator.MustGetNewMigrator(
		MigrationsFS, migrationsDir)
	if err != nil {
		return fmt.Errorf("useMigrations->MustGetNewMi: %w", err)
	}

	conn, err := sql.Open("postgres", par.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("useMigrations->sql.Open: %w", err)
	}

	defer conn.Close()

	err = migrator.ApplyMigrations(conn)
	if err != nil {
		return fmt.Errorf("useMigrations->ApplyMigra: %w", err)
	}

	return nil
}

func setHandlerParams(params *bizmodels.InitParams) error {
	params.
		ValidateAddrPattern = "^[a-zA-Z/ ]{1,100}:[0-9]{1,10}$"
	params.DatabaseDSN = "postgres://postgres:postgres" +
		"@postgres" +
		":5432/praktikum?sslmode=disable"
	_, path, _, ok := runtime.Caller(0)

	if !ok {
		return fmt.Errorf("setHandlerParams->Caller: %w", errPath)
	}

	Root := filepath.Join(filepath.Dir(path), "../../../")
	params.CryptoPrivateKeyPath = Root +
		"/internal/asymcrypto/keys/private.pem"

	params.FileStoragePath = Root + defSavePathFile
	params.Key = "defaultKey"
	params.Restore = true
	params.ValidateAddrPattern = ""
	params.WaitSecRespDB = 10 * time.Second
	params.TrustedSubnet = "0.0.0.0/0"

	return nil
}

//nolint:funlen
func initiate(
	mux *mux.Router,
	params *bizmodels.InitParams,
	settings *bizmodels.EndpointSettings,
	isMemRepo bool,
) error {
	err := setHandlerParams(params)
	if err != nil {
		return fmt.Errorf("initiate->setHandlerParams: %w", err)
	}

	storage := &dbrepository.DBepository{}
	mem := &memoryrepository.MemoryRepository{}
	ctx, cancel := context.WithTimeout(
		context.Background(), params.WaitSecRespDB)

	defer cancel()

	var dse *service.DS

	if isMemRepo {
		dse = service.NewMemoryService(mem, params.WaitSecRespDB)
	} else {
		dse = service.NewMemoryService(
			storage, params.WaitSecRespDB)
	}

	dbConn, err := pgxpool.New(ctx, params.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("initiate->pgxpool.New: %w", err)
	}

	if isMemRepo {
		mem.Init()
	} else {
		storage.Initiate(params.DatabaseDSN, dbConn)
	}

	hJSONSets := sender.NewSenderHandler(
		dse, params)

	zapLogger, err := logger.Initialize("info")
	if err != nil {
		return fmt.Errorf("Initialize: %w", err)
	}

	setMsJSONMux := mux.Methods(http.MethodPost).Subrouter()
	setMsJSONMux.HandleFunc(
		"/updates/",
		hJSONSets.SenderHandler)
	setMsJSONMux.Use(
		decryptmid.DecryptMiddleware(*params),
		gzipcompressmiddleware.GzipMiddleware(),
		loggermiddleware.RequestLogger(zapLogger))

	settings.ContentType = "application/json"
	settings.Encoding = "gzip"
	settings.URL = url + "/updates/"

	_ = UseMigrations(params)

	if !isMemRepo {
		// for coverage
		counters := make(map[string]bizmodels.Counter, 1)
		gauges := make(map[string]bizmodels.Gauge, 1)
		tobj := &bizmodels.Counter{}
		tobj.Name = "counter3333"
		tobj.Value = 3
		counters[tobj.Name] = *tobj

		tobj1 := &bizmodels.Gauge{}
		tobj1.Name = "gauge333"
		tobj1.Value = 3.3
		gauges[tobj.Name] = *tobj1

		err = dse.AddMetrics(gauges, counters)
		if err != nil {
			return fmt.Errorf("AddMetrics: %w", err)
		}
	} // for coverage

	return nil
}

func parseResponse(
	buf *bytes.Buffer,
) ([]viewData, error) {
	var result apimodels.ArrMetrics

	vdata := make([]viewData, 0)

	out, err := compress.DeflateDecompress(buf)
	if err != nil {
		fmt.Println("parseResponse->DeflateDecompress: %w", err)

		return nil, errResponse
	}

	err = json.Unmarshal(out, &result)
	if err != nil {
		return nil, err
	}

	for _, metric := range result {
		temp := &viewData{}

		temp.id = metric.ID
		temp.mtype = metric.MType

		if temp.mtype == bizmodels.CounterName {
			temp.delta = *metric.Delta
		} else if temp.mtype == bizmodels.GaugeName {
			temp.value = *metric.Value
		}

		vdata = append(vdata, *temp)
	}

	return vdata, nil
}

func TestSender(t *testing.T) {
	t.Helper()
	t.Parallel()

	params := &bizmodels.InitParams{}
	settings := &bizmodels.EndpointSettings{}
	mux := mux.NewRouter()

	err := initiate(mux, params, settings, false)
	if err != nil {
		fmt.Println(err)

		return
	}

	testCases := getTestData()

	for _, test := range *testCases {
		t.Run(http.MethodPost, func(tobj *testing.T) {
			tobj.Parallel()

			reqData, err := initReqData(params, &test)
			if err != nil {
				fmt.Println(err)

				return
			}

			req, err := http.NewRequestWithContext(
				context.Background(),
				http.MethodPost, settings.URL, reqData)
			if err != nil {
				tobj.Fatal(err)
			}

			req.Header.Set("Content-Encoding", settings.Encoding)
			req.Header.Set("Accept-Encoding", settings.Encoding)
			req.Header.Set("Content-Type", settings.ContentType)

			if test.hash != "" {
				req.Header.Set("Hashsha256", test.hash)
			}

			if test.key != "" {
				req.Header.Set("Hashsha256", settings.Hash)
			}

			newr := httptest.NewRecorder()
			mux.ServeHTTP(newr, req)
			status := newr.Code

			assert.Equal(tobj,
				test.expcod,
				status, test.tn+": Response code didn't match expected")
		})
	}
}

func initReqData(params *bizmodels.InitParams,
	testD *testData,
) (*bytes.Reader, error) {
	dataMarshal := getDataSend(testD)

	metricMarshall, err := json.Marshal(dataMarshal)
	if err != nil {
		return nil, err
	}

	metricCompress, err := compress.DeflateCompress(
		metricMarshall)
	if err != nil {
		return nil, fmt.Errorf("initReqData->DeflateCom: %w", err)
	}

	_, path, _, ok := runtime.Caller(0)

	if !ok {
		return nil, fmt.Errorf("initReqData->Caller: %w", err)
	}

	Root := filepath.Join(filepath.Dir(path), "../../../")
	pathPubicKey := Root +
		"/internal/asymcrypto/keys/public.pem"

	key, err := os.ReadFile(pathPubicKey)
	if err != nil {
		return nil, fmt.Errorf("initReqData->ReadFile: %w", err)
	}

	encr, err := asymcrypto.Encrypt(&metricCompress, &key)
	if err != nil {
		return nil, fmt.Errorf("initReqData->Encrypt: %w", err)
	}

	if params.Key != "" {
		tHash, err := hash.MakeHashSHA256(&metricMarshall,
			testD.key)
		if err != nil {
			return nil, fmt.Errorf("initReqData->MakeHash: %w", err)
		}

		encodedStr := hex.EncodeToString(tHash)

		testD.hash = encodedStr
	}

	return bytes.NewReader(*encr), nil
}

func getDataSend(testD *testData,
) *apimodels.ArrMetrics {
	var reqMetric apimodels.Metrics

	data := make(apimodels.ArrMetrics,
		0,
		len(testD.gauges)+len(testD.counters))

	for _, metric := range testD.counters {
		reqMetric = apimodels.Metrics{}
		reqMetric.ID = metric.Name
		reqMetric.MType = bizmodels.CounterName
		reqMetric.Delta = &metric.Value
		data = append(data, reqMetric)
	}

	for _, metric := range testD.gauges {
		reqMetric = apimodels.Metrics{}
		reqMetric.ID = metric.Name
		reqMetric.MType = bizmodels.GaugeName
		reqMetric.Value = &metric.Value
		data = append(data, reqMetric)
	}

	return &data
}

func randomString(n int) string {
	letters := []rune(
		"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}
