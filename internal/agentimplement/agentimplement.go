// Package for implementing agent methods.
package agentimplement

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/dmitrovia/collector-metrics/internal/endpoints/sendmetricsjsonendpoint"
	"github.com/dmitrovia/collector-metrics/internal/functions/asymcrypto"
	"github.com/dmitrovia/collector-metrics/internal/functions/compress"
	"github.com/dmitrovia/collector-metrics/internal/functions/config"
	"github.com/dmitrovia/collector-metrics/internal/functions/hash"
	"github.com/dmitrovia/collector-metrics/internal/functions/ip"
	"github.com/dmitrovia/collector-metrics/internal/functions/random"
	"github.com/dmitrovia/collector-metrics/internal/functions/validate"
	"github.com/dmitrovia/collector-metrics/internal/logger"
	"github.com/dmitrovia/collector-metrics/internal/models/apimodels"
	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"go.uber.org/zap"
)

const zapLogLevel = "info"

const defPORT string = ""

const validAddrPattern = "^[a-zA-Z/ ]{1,100}:[0-9]{1,10}$"

const defKeyHashSha256 = ""

const defPollInterval int = 0

const defReportInterval int = 0

const metricGaugeCount int = 30

const defCountJobs int = 3

var errGetENV = errors.New(
	"REPORT_INTERVAL failed converting to int")

var errGetENV1 = errors.New(
	"POLL_INTERVAL failed converting to int")

var errGetENV2 = errors.New(
	"RATE_LIMIT failed converting to int")

var errParseFlags = errors.New("addr is not valid")

var errResponse = errors.New("error response")

const defCryptoKeyPath string = ""

const defConfigPath string = "/internal/config/agent.json"

func worker(jobs <-chan bizmodels.JobData,
	wgew *sync.WaitGroup,
) {
	for event := range jobs {
		wgew.Add(1)

		switch event.Event {
		case "setValuesMonitor":
			setValuesMonitor(event.Mon, event.Mutex)
		case "setMonitorFromGoPsUtil":
			setMonitorFromGoPsUtil(event.Mon, event.Mutex)
		case "reqMetricsJSON":
			reqMetricsJSON(event.Par, event.Client, event.Mon)
		}

		wgew.Done()
	}
}

// Collect - collects device metrics
// every PollInterval seconds using workers.
func Collect(
	chc *chan os.Signal,
	par *bizmodels.InitParamsAgent,
	wg *sync.WaitGroup,
	wgEndWork *sync.WaitGroup,
	mon *bizmodels.Monitor,
	jobs chan bizmodels.JobData,
) {
	defer wg.Done()

	var mutex sync.Mutex

	signal.Notify(*chc,
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT)

	for {
		select {
		case <-*chc:
			*chc <- syscall.SIGTERM

			wgEndWork.Wait()

			return
		case <-time.After(
			time.Duration(par.PollInterval) * time.Second):
			dataChan := &bizmodels.JobData{}
			dataChan.Event = "setValuesMonitor"
			dataChan.Mutex = &mutex
			dataChan.Mon = mon

			jobs <- *dataChan

			go worker(jobs, wgEndWork)

			dataChan1 := &bizmodels.JobData{}
			dataChan1.Event = "setMonitorFromGoPsUtil"
			dataChan1.Mutex = &mutex
			dataChan1.Mon = mon

			jobs <- *dataChan1

			go worker(jobs, wgEndWork)
		}
	}
}

// Send - sends metrics to the server
// every ReportInterval seconds using workers.
func Send(
	chc *chan os.Signal,
	par *bizmodels.InitParamsAgent,
	wgMain *sync.WaitGroup,
	wgEndWork *sync.WaitGroup,
	client *http.Client,
	mon *bizmodels.Monitor,
	jobs chan bizmodels.JobData,
) {
	defer wgMain.Done()

	signal.Notify(*chc,
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT)

	for {
		select {
		case <-*chc:
			*chc <- syscall.SIGTERM

			wgEndWork.Wait()

			return
		case <-time.After(
			time.Duration(par.ReportInterval) * time.Second):
			if !par.RepeatedReq {
				dataChan := &bizmodels.JobData{}
				dataChan.Event = "reqMetricsJSON"
				dataChan.Mon = mon
				dataChan.Par = par
				dataChan.Client = client

				jobs <- *dataChan

				go worker(jobs, wgEndWork)
			}
		}
	}
}

// fillMetrics - fills the
// array and map with metrics from monitor.
func fillMetrics(mon *bizmodels.Monitor,
	gauges *[]bizmodels.Gauge,
	counters map[string]bizmodels.Counter,
) {
	counters["PollCount"] = mon.PollCount

	*gauges = append(
		*gauges,
		mon.TotalMemory,
		mon.FreeMemory,
		mon.CPUutilization1,
		mon.Alloc,
		mon.BuckHashSys,
		mon.Frees,
		mon.GCCPUFraction,
		mon.GCSys,
		mon.HeapAlloc,
		mon.HeapIdle,
		mon.HeapInuse,
		mon.HeapObjects,
		mon.HeapReleased,
		mon.HeapSys,
		mon.LastGC,
		mon.Lookups,
		mon.MCacheInuse,
		mon.MCacheInuse,
		mon.MCacheSys,
		mon.MSpanInuse,
		mon.MSpanSys,
		mon.Mallocs,
		mon.NextGC,
		mon.NumForcedGC,
		mon.NumGC,
		mon.OtherSys,
		mon.PauseTotalNs,
		mon.StackInuse,
		mon.StackSys,
		mon.Sys,
		mon.TotalAlloc,
		mon.RandomValue)
}

func getSettings(client *http.Client,
	par *bizmodels.InitParamsAgent,
	mon *bizmodels.Monitor,
) (*bizmodels.EndpointSettings, error) {
	gauges := make([]bizmodels.Gauge, 0, metricGaugeCount)
	counters := make(map[string]bizmodels.Counter, 1)

	fillMetrics(mon, &gauges, counters)

	settings := &bizmodels.EndpointSettings{}
	settings.Client = client
	settings.ContentType = "application/json"
	settings.Encoding = "gzip"
	settings.URL = par.URL + "/updates/"

	ips, err := ip.GetLocalIPs()
	if err != nil {
		return nil, fmt.Errorf("getSettings->GetLocalIPs: %w",
			err)
	}

	settings.RealIPHeader = ips[0]

	req, err := initReqData(&gauges, counters, settings, par)
	if err != nil {
		return nil, fmt.Errorf("getSettings->initReqData: %w",
			err)
	}

	settings.SendData = req

	return settings, nil
}

// reqMetricsJSON - prepares data
// for the request and sends the request to the server.
func reqMetricsJSON(par *bizmodels.InitParamsAgent,
	client *http.Client,
	mon *bizmodels.Monitor,
) {
	settings, err := getSettings(client, par, mon)
	if err != nil {
		fmt.Println("reqMetricsJSON->getSettings: %w", err)

		return
	}

	sInterval := par.StartReqInterval

	for iter := 1; iter <= par.CountReqRetries; iter++ {
		resp, err := sendmetricsjsonendpoint.SendMJSONEndpoint(
			settings)
		if err != nil {
			par.RepeatedReq = true

			fmt.Println("reqMetricsJSON->SendMJSONEndpoint: %w",
				err)

			time.Sleep(time.Duration(sInterval) * time.Second)

			sInterval += par.ReqInternal

			continue
		}

		_, err = parseResponse(resp)
		if err != nil {
			par.RepeatedReq = true

			fmt.Println("reqMetricsJSON->parseResponse: %w", err)

			time.Sleep(time.Duration(sInterval) * time.Second)

			sInterval += par.ReqInternal

			continue
		}

		break
	}

	par.RepeatedReq = false
}

// initReqData - prepares the body
// for the encrypted request.
func initReqData(gauges *[]bizmodels.Gauge,
	counters map[string]bizmodels.Counter,
	settings *bizmodels.EndpointSettings,
	params *bizmodels.InitParamsAgent,
) (*bytes.Reader, error) {
	dataMarshal := getDataSend(gauges, counters)

	metricMarshall, err := json.Marshal(dataMarshal)
	if err != nil {
		return nil, err
	}

	metricCompress, err := compress.DeflateCompress(
		metricMarshall)
	if err != nil {
		return nil, fmt.Errorf("initReqData->DeflateCompress: %w",
			err)
	}

	key, err := os.ReadFile(params.CryptoPublicKeyPath)
	if err != nil {
		return nil, fmt.Errorf(
			"initReqData->ReadFile: %w", err)
	}

	encr, err := asymcrypto.Encrypt(&metricCompress, &key)
	if err != nil {
		return nil, fmt.Errorf(
			"initReqData->Encrypt: %w", err)
	}

	if params.Key != "" {
		tHash, err := hash.MakeHashSHA256(&metricMarshall,
			params.Key)
		if err != nil {
			return nil, fmt.Errorf("initReqData->MakeHashSHA256: %w",
				err)
		}

		encodedStr := hex.EncodeToString(tHash)

		settings.Hash = encodedStr
	}

	return bytes.NewReader(*encr), nil
}

// parseResponse - parses the response from the server.
func parseResponse(
	response *http.Response,
) (*[]byte, error) {
	defer response.Body.Close()

	if response.StatusCode == http.StatusOK {
		out, err := compress.DeflateDecompress(response.Body)
		if err != nil {
			fmt.Println("parseResponse->DeflateDecompress: %w", err)
		}

		return &out, nil
	}

	fmt.Printf("anscode: %d\n", response.StatusCode)

	return nil, errResponse
}

// getDataSend - receives data in API format.
func getDataSend(gauges *[]bizmodels.Gauge,
	counters map[string]bizmodels.Counter,
) *apimodels.ArrMetrics {
	var reqMetric apimodels.Metrics

	data := make(apimodels.ArrMetrics,
		0,
		len(*gauges)+len(counters))

	for _, metric := range counters {
		reqMetric = apimodels.Metrics{}
		reqMetric.ID = metric.Name
		reqMetric.MType = bizmodels.CounterName
		reqMetric.Delta = &metric.Value
		data = append(data, reqMetric)
	}

	for _, metric := range *gauges {
		reqMetric = apimodels.Metrics{}
		reqMetric.ID = metric.Name
		reqMetric.MType = bizmodels.GaugeName
		reqMetric.Value = &metric.Value
		data = append(data, reqMetric)
	}

	return &data
}

// Initialization - initializes
// data for the agent to operate.
func Initialization(params *bizmodels.InitParamsAgent,
	mon *bizmodels.Monitor,
) (*zap.Logger, error) {
	var err error

	params.URL = "http://"
	params.ReportInterval = 10
	params.PollInterval = 2
	params.ReqInternal = 2
	params.StartReqInterval = 1
	params.CountReqRetries = 3
	params.RepeatedReq = false

	err = parseFlags(params)
	if err != nil {
		return nil, err
	}

	err = getParamsFromCFG(params)
	if err != nil {
		return nil, err
	}

	err = getIntervalsEnv(params)
	if err != nil {
		return nil, err
	}

	err = getENV(params)
	if err != nil {
		return nil, err
	}

	params.URL += params.PORT

	mon.Init()

	zlog, err := logger.Initialize(zapLogLevel)
	if err != nil {
		return nil, fmt.Errorf(
			"Initialization->logger.Initialize %w",
			err)
	}

	return zlog, nil
}

// getIntervalsEnv - gets environment variables.
func getIntervalsEnv(
	params *bizmodels.InitParamsAgent,
) error {
	envReportInterval := os.Getenv("REPORT_INTERVAL")
	envPollInterval := os.Getenv("POLL_INTERVAL")

	if envReportInterval != "" {
		value, err := strconv.Atoi(envReportInterval)
		if err != nil {
			return errGetENV
		}

		params.ReportInterval = value
	}

	if envPollInterval != "" {
		value, err := strconv.Atoi(envPollInterval)
		if err != nil {
			return errGetENV1
		}

		params.PollInterval = value
	}

	return nil
}

// getENV - gets environment variables.
//
//nolint:cyclop
func getENV(params *bizmodels.InitParamsAgent) error {
	key := os.Getenv("KEY")
	envRunAddr := os.Getenv("ADDRESS")
	envRateLimit := os.Getenv("RATE_LIMIT")
	cryptoKey := os.Getenv("CRYPTO_KEY_AGENT")
	cfgServer := os.Getenv("CONFIG_SERVER")

	_, path, _, isok := runtime.Caller(0)
	Root := filepath.Join(filepath.Dir(path), "../..")

	if cfgServer != "" && isok {
		params.ConfigPath = Root + cfgServer
	}

	if cryptoKey != "" && isok {
		params.CryptoPublicKeyPath = Root + cryptoKey
	}

	if key != "" {
		params.Key = key
	}

	if envRunAddr != "" {
		res, err := validate.IsMatchesTemplate(envRunAddr,
			validAddrPattern)
		if err != nil {
			return fmt.Errorf("getENV: %w", err)
		}

		if !res {
			return errParseFlags
		}

		params.PORT = envRunAddr
	}

	if envRateLimit != "" {
		value, err := strconv.Atoi(envRateLimit)
		if err != nil {
			return errGetENV2
		}

		params.RateLimit = value
	}

	return nil
}

// parseFlags - parses passed flags into variables.
func parseFlags(params *bizmodels.InitParamsAgent) error {
	var err error

	flag.StringVar(&params.Key,
		"k", defKeyHashSha256,
		"key for signatures for the SHA256 algorithm.")
	flag.StringVar(&params.PORT,
		"a",
		defPORT,
		"Port to listen on.")
	flag.IntVar(&params.PollInterval,
		"p",
		defPollInterval,
		"Frequency of sending metrics to the server.")
	flag.IntVar(&params.RateLimit,
		"l",
		defCountJobs,
		"maximum number of goroutines.")
	flag.IntVar(&params.ReportInterval,
		"r", defReportInterval,
		"Frequency of polling metrics from the runtime package.")

	_, path, _, ok := runtime.Caller(0)

	defPath1 := defConfigPath

	if ok {
		Root := filepath.Join(filepath.Dir(path), "../..")
		defPath1 = Root + defConfigPath
	}

	flag.StringVar(&params.ConfigPath,
		"config", defPath1, "cfg server path.")
	flag.StringVar(&params.CryptoPublicKeyPath,
		"crypto-key", defCryptoKeyPath,
		"asymmetric encryption public key.")
	flag.Parse()

	res, err := validate.IsMatchesTemplate(params.PORT,
		validAddrPattern)

	if err != nil && !res {
		return errParseFlags
	}

	return nil
}

// setMonitorFromGoPsUtil - gets CPU Metrics.
func setMonitorFromGoPsUtil(mon *bizmodels.Monitor,
	mutex *sync.Mutex,
) {
	virtMem, _ := mem.VirtualMemory()

	cores, _ := cpu.Info()

	mutex.Lock()
	mon.TotalMemory.Value = float64(virtMem.Total)
	mon.FreeMemory.Value = float64(virtMem.Free)

	for _, core := range cores {
		mon.CPUutilization1.Value += float64(core.CPU)
	}
	mutex.Unlock()
}

// setMonitorFromGoPsUtil - gets
// and sets CPU Metrics to monitor.
func setValuesMonitor(mon *bizmodels.Monitor,
	mutex *sync.Mutex,
) {
	const maxRandomValue int64 = 1000

	mutex.Lock()
	writeFromMemory(mon)

	mon.PollCount.Value++

	tmpCounters := make(map[string]bizmodels.Counter, 1)
	tmpCounters["PollCount"] = mon.PollCount

	rand, err := random.RandF64(maxRandomValue)
	if err != nil {
		fmt.Println("setValuesMonitor->random.RandF64: %w", err)
	} else {
		mon.RandomValue.Value = rand
	}

	mutex.Unlock()
}

// writeFromMemory - read runtime metrics.
func writeFromMemory(mon *bizmodels.Monitor) {
	var rtm runtime.MemStats

	runtime.ReadMemStats(&rtm)

	mon.Alloc.Value = float64(rtm.Alloc)
	mon.BuckHashSys.Value = float64(rtm.BuckHashSys)
	mon.Frees.Value = float64(rtm.Frees)
	mon.GCCPUFraction.Value = rtm.GCCPUFraction
	mon.GCSys.Value = float64(rtm.GCSys)
	mon.HeapAlloc.Value = float64(rtm.HeapAlloc)
	mon.HeapIdle.Value = float64(rtm.HeapIdle)
	mon.HeapInuse.Value = float64(rtm.HeapInuse)
	mon.HeapObjects.Value = float64(rtm.HeapObjects)
	mon.HeapReleased.Value = float64(rtm.HeapReleased)
	mon.HeapSys.Value = float64(rtm.HeapSys)
	mon.LastGC.Value = float64(rtm.LastGC)
	mon.Lookups.Value = float64(rtm.Lookups)
	mon.MCacheInuse.Value = float64(rtm.MCacheInuse)
	mon.MCacheSys.Value = float64(rtm.MCacheSys)
	mon.MSpanInuse.Value = float64(rtm.MSpanInuse)
	mon.MSpanSys.Value = float64(rtm.MSpanSys)
	mon.Mallocs.Value = float64(rtm.Mallocs)
	mon.NextGC.Value = float64(rtm.NextGC)
	mon.NumForcedGC.Value = float64(rtm.NumForcedGC)
	mon.NumGC.Value = float64(rtm.NumGC)
	mon.OtherSys.Value = float64(rtm.OtherSys)
	mon.PauseTotalNs.Value = float64(rtm.PauseTotalNs)
	mon.StackInuse.Value = float64(rtm.StackInuse)
	mon.StackSys.Value = float64(rtm.StackSys)
	mon.Sys.Value = float64(rtm.Sys)
	mon.TotalAlloc.Value = float64(rtm.TotalAlloc)
}

func getParamsFromCFG(
	par *bizmodels.InitParamsAgent,
) error {
	cfg, err := config.LoadConfigAgent(par.ConfigPath)
	if err != nil {
		return fmt.Errorf(
			"getParamsFromCFG->LoadConfigAgent: %w",
			err)
	}

	_, path, _, ok := runtime.Caller(0)

	if !ok {
		return fmt.Errorf(
			"getParamsFromCFG->Caller: %w",
			err)
	}

	Root := filepath.Join(filepath.Dir(path), "../..")

	if par.CryptoPublicKeyPath == "" {
		par.CryptoPublicKeyPath = Root + cfg.CryptoPublicKeyPath
	}

	if par.Key == "" {
		par.Key = cfg.Key
	}

	if par.PORT == "" {
		par.PORT = cfg.PORT
	}

	if par.PollInterval == 0 {
		par.PollInterval = cfg.PollInterval
	}

	if par.ReportInterval == 0 {
		par.ReportInterval = cfg.ReportInterval
	}

	return nil
}
