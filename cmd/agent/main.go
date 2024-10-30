package main

import (
	"context"
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

	"github.com/dmitrovia/collector-metrics/internal/endpoints"
	"github.com/dmitrovia/collector-metrics/internal/functions/random"
	"github.com/dmitrovia/collector-metrics/internal/functions/validate"
	"github.com/dmitrovia/collector-metrics/internal/models"
)

const defPORT string = "localhost:8080"

const defPollInterval int = 2

const defReportInterval int = 10

const metricGaugeCount int = 27

var errGetENV = errors.New("REPORT_INTERVAL failed converting to int")

var errGetENV1 = errors.New("POLL_INTERVAL failed converting to int")

var errParseFlags = errors.New("addr is not valid")

type initParams struct {
	url                 string
	PORT                string
	reportInterval      int
	pollInterval        int
	validateAddrPattern string
}

func main() {
	var waitGroup *sync.WaitGroup

	var monitor *models.Monitor

	var httpClient *http.Client

	var gauges *[]models.Gauge

	var counters *map[string]models.Counter

	var params *initParams

	waitGroup = new(sync.WaitGroup)
	monitor = new(models.Monitor)
	httpClient = new(http.Client)
	gauges = new([]models.Gauge)
	counters = new(map[string]models.Counter)

	params = new(initParams)
	params.url = "http://"
	params.reportInterval = 10
	params.pollInterval = 2
	params.validateAddrPattern = "^[a-zA-Z/ ]{1,100}:[0-9]{1,10}$"

	err := initialization(params, httpClient, monitor)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	waitGroup.Add(1)

	go collect(monitor, params, waitGroup, gauges, counters)

	waitGroup.Add(1)

	go send(params, waitGroup, httpClient, gauges, counters)
	waitGroup.Wait()
}

func collect(mod *models.Monitor, par *initParams, wg *sync.WaitGroup, gauges *[]models.Gauge, counters *map[string]models.Counter) {
	defer wg.Done()

	channelCancel := make(chan os.Signal, 1)
	signal.Notify(channelCancel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	for {
		select {
		case <-channelCancel:
			return
		case <-time.After(time.Duration(par.pollInterval) * time.Second):
			setValuesMonitor(mod, gauges, counters)
		}
	}
}

func send(par *initParams, wg *sync.WaitGroup, httpC *http.Client, gauges *[]models.Gauge, counters *map[string]models.Counter) {
	defer wg.Done()

	channelCancel := make(chan os.Signal, 1)
	signal.Notify(channelCancel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	for {
		select {
		case <-channelCancel:
			return
		case <-time.After(time.Duration(par.reportInterval) * time.Second):
			go doReqSendMetrics(par.url, httpC, gauges, counters)
		}
	}
}

func doReqSendMetrics(url string, httpC *http.Client, gauges *[]models.Gauge, counters *map[string]models.Counter) {
	tmpURL := url + "/update/" + "counter/"
	for _, metric := range *counters {
		endpoints.SendMetricEndpoint(context.Background(), tmpURL+metric.Name+"/"+strconv.FormatInt(metric.Value, 10), httpC)
	}

	tmpURL = url + "/update/" + "gauge/"
	for _, metric := range *gauges {
		endpoints.SendMetricEndpoint(context.Background(), tmpURL+metric.Name+"/"+strconv.FormatFloat(metric.Value, 'f', -1, 64), httpC)
	}
}

func initialization(params *initParams, httpC *http.Client, mon *models.Monitor) error {
	var err error

	*httpC = http.Client{}

	err = parseFlags(params)
	if err != nil {
		return err
	}

	err = getENV(params)
	if err != nil {
		return err
	}

	params.url += params.PORT

	mon.Init()

	return err
}

func getENV(params *initParams) error {
	var err error

	envRunAddr := os.Getenv("ADDRESS")

	if envRunAddr != "" {
		err = addrIsValid(envRunAddr, params)
		if err != nil {
			return err
		}
	}

	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		value, err := strconv.Atoi(envReportInterval)
		if err != nil {
			return errGetENV
		}

		params.reportInterval = value
	}

	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		value, err := strconv.Atoi(envPollInterval)
		if err != nil {
			return errGetENV1
		}

		params.pollInterval = value
	}

	return err
}

func addrIsValid(addr string, params *initParams) error {
	res, err := validate.IsMatchesTemplate(addr, params.validateAddrPattern)
	if err == nil {
		if res {
			params.PORT = addr
		} else {
			return errParseFlags
		}
	} else {
		return fmt.Errorf("addrIsValid: %w", err)
	}

	return nil
}

func parseFlags(params *initParams) error {
	var err error

	flag.StringVar(&params.PORT, "a", defPORT, "Port to listen on.")
	flag.IntVar(&params.pollInterval, "p", defPollInterval, "Frequency of sending metrics to the server.")
	flag.IntVar(&params.reportInterval, "r", defReportInterval, "Frequency of polling metrics from the runtime package.")
	flag.Parse()

	res, err := validate.IsMatchesTemplate(params.PORT, params.validateAddrPattern)

	if err != nil && !res {
		return errParseFlags
	}

	return nil
}

func setValuesMonitor(mon *models.Monitor, gauges *[]models.Gauge, counters *map[string]models.Counter) {
	const maxRandomValue int64 = 1000

	writeFromMemory(mon)

	mon.PollCount.Value++

	tmpCounters := make(map[string]models.Counter, 1)
	tmpCounters["PollCount"] = mon.PollCount

	tmpGauges := make([]models.Gauge, 0, metricGaugeCount)

	mon.RandomValue.Value = random.RandF64(maxRandomValue)

	tmpGauges = append(tmpGauges, mon.Alloc, mon.BuckHashSys, mon.Frees, mon.GCCPUFraction, mon.GCSys)
	tmpGauges = append(tmpGauges, mon.HeapAlloc, mon.HeapIdle, mon.HeapInuse, mon.HeapObjects, mon.HeapReleased)
	tmpGauges = append(tmpGauges, mon.HeapSys, mon.LastGC, mon.Lookups, mon.MCacheInuse, mon.MCacheInuse)
	tmpGauges = append(tmpGauges, mon.MCacheSys, mon.MSpanInuse, mon.MSpanSys, mon.Mallocs, mon.NextGC)
	tmpGauges = append(tmpGauges, mon.NumForcedGC, mon.NumGC, mon.OtherSys, mon.PauseTotalNs, mon.StackInuse)
	tmpGauges = append(tmpGauges, mon.StackSys, mon.Sys, mon.TotalAlloc, mon.RandomValue)

	*gauges = tmpGauges
	*counters = tmpCounters
}

func writeFromMemory(mon *models.Monitor) {
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
