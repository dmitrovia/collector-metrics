package agentimplement

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/dmitrovia/collector-metrics/internal/endpoints/sendmetricsjsonendpoint"
	"github.com/dmitrovia/collector-metrics/internal/functions/compress"
	"github.com/dmitrovia/collector-metrics/internal/functions/random"
	"github.com/dmitrovia/collector-metrics/internal/functions/validate"
	"github.com/dmitrovia/collector-metrics/internal/models/apimodels"
	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
)

const defPORT string = "localhost:8080"

const defPollInterval int = 2

const defReportInterval int = 10

const metricGaugeCount int = 27

var errGetENV = errors.New(
	"REPORT_INTERVAL failed converting to int")

var errGetENV1 = errors.New(
	"POLL_INTERVAL failed converting to int")

var errParseFlags = errors.New("addr is not valid")

var errResponse = errors.New("error response")

func Collect(mod *bizmodels.Monitor,
	par *bizmodels.InitParamsAgent,
	wg *sync.WaitGroup,
	gauges *[]bizmodels.Gauge,
	cnts *map[string]bizmodels.Counter,
) {
	defer wg.Done()

	channelCancel := make(chan os.Signal, 1)
	signal.Notify(channelCancel,
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT)

	for {
		select {
		case <-channelCancel:
			return
		case <-time.After(
			time.Duration(par.PollInterval) * time.Second):
			setValuesMonitor(mod, gauges, cnts)
		}
	}
}

func Send(par *bizmodels.InitParamsAgent,
	wg *sync.WaitGroup,
	client *http.Client,
	gauges *[]bizmodels.Gauge,
	counters *map[string]bizmodels.Counter,
) {
	defer wg.Done()

	channelCancel := make(chan os.Signal, 1)
	signal.Notify(channelCancel,
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT)

	for {
		select {
		case <-channelCancel:
			return
		case <-time.After(
			time.Duration(par.ReportInterval) * time.Second):
			if !par.RepeatedReq {
				go reqMetricsJSON(par, client, gauges, counters)
			}
		}
	}
}

func reqMetricsJSON(par *bizmodels.InitParamsAgent,
	client *http.Client,
	gauges *[]bizmodels.Gauge,
	counters *map[string]bizmodels.Counter,
) {
	req, err := initReqData(gauges, counters)
	if err != nil {
		fmt.Println("reqMetricsJSON->initReqData: %w", err)

		return
	}

	sInterval := par.StartReqInterval

	for iter := 1; iter <= par.CountReqRetries; iter++ {
		resp, err := sendmetricsjsonendpoint.SendMJSONEndpoint(
			req, par.URL+"/updates/",
			client)
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

func initReqData(gauges *[]bizmodels.Gauge,
	counters *map[string]bizmodels.Counter,
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

	return bytes.NewReader(metricCompress), nil
}

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

func getDataSend(gauges *[]bizmodels.Gauge,
	counters *map[string]bizmodels.Counter,
) *apimodels.ArrMetrics {
	var reqMetric apimodels.Metrics

	data := make(apimodels.ArrMetrics,
		0,
		len(*gauges)+len(*counters))

	for _, metric := range *counters {
		reqMetric = apimodels.Metrics{}
		reqMetric.ID = metric.Name
		reqMetric.MType = "counter"
		reqMetric.Delta = &metric.Value
		data = append(data, reqMetric)
	}

	for _, metric := range *gauges {
		reqMetric = apimodels.Metrics{}
		reqMetric.ID = metric.Name
		reqMetric.MType = "gauge"
		reqMetric.Value = &metric.Value
		data = append(data, reqMetric)
	}

	return &data
}

func Initialization(params *bizmodels.InitParamsAgent,
	client *http.Client,
	mon *bizmodels.Monitor,
) error {
	var err error

	*client = http.Client{}

	params.URL = "http://"
	params.ReportInterval = 10
	params.PollInterval = 2
	params.ReqInternal = 2
	params.StartReqInterval = 1
	params.CountReqRetries = 3
	params.RepeatedReq = false
	params.ValidAddrPattern = "^[a-zA-Z/ ]{1,100}:[0-9]{1,10}$"

	err = parseFlags(params)
	if err != nil {
		return err
	}

	err = getENV(params)
	if err != nil {
		return err
	}

	params.URL += params.PORT

	mon.Init()

	return err
}

func getENV(params *bizmodels.InitParamsAgent) error {
	var err error

	envRunAddr := os.Getenv("ADDRESS")

	if envRunAddr != "" {
		res, err := validate.IsMatchesTemplate(envRunAddr,
			params.ValidAddrPattern)
		if err != nil {
			return fmt.Errorf("getENV: %w", err)
		}

		if !res {
			return errParseFlags
		}

		params.PORT = envRunAddr
	}

	envReportInterval := os.Getenv("REPORT_INTERVAL")

	if envReportInterval != "" {
		value, err := strconv.Atoi(envReportInterval)
		if err != nil {
			return errGetENV
		}

		params.ReportInterval = value
	}

	envPollInterval := os.Getenv("POLL_INTERVAL")
	if envPollInterval != "" {
		value, err := strconv.Atoi(envPollInterval)
		if err != nil {
			return errGetENV1
		}

		params.PollInterval = value
	}

	return err
}

func parseFlags(params *bizmodels.InitParamsAgent) error {
	var err error

	flag.StringVar(&params.PORT,
		"a",
		defPORT,
		"Port to listen on.")
	flag.IntVar(&params.PollInterval,
		"p",
		defPollInterval,
		"Frequency of sending metrics to the server.")
	flag.IntVar(&params.ReportInterval,
		"r", defReportInterval,
		"Frequency of polling metrics from the runtime package.")
	flag.Parse()

	res, err := validate.IsMatchesTemplate(params.PORT,
		params.ValidAddrPattern)

	if err != nil && !res {
		return errParseFlags
	}

	return nil
}

func setValuesMonitor(mon *bizmodels.Monitor,
	gauges *[]bizmodels.Gauge,
	counters *map[string]bizmodels.Counter,
) {
	const maxRandomValue int64 = 1000

	writeFromMemory(mon)

	mon.PollCount.Value++

	tmpCounters := make(map[string]bizmodels.Counter, 1)
	tmpCounters["PollCount"] = mon.PollCount

	tmpGauges := make([]bizmodels.Gauge, 0, metricGaugeCount)

	mon.RandomValue.Value = random.RandF64(maxRandomValue)

	tmpGauges = append(
		tmpGauges,
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

	*gauges = tmpGauges
	*counters = tmpCounters
}

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
