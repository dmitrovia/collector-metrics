// Package apimodels
// describes the server and client(agent) models
package bizmodels

import (
	"bytes"
	"net/http"
	"sync"
	"time"
)

const GaugeName string = "gauge"

const CounterName string = "counter"

const MetricsPattern = "gauge|counter"

// Gauge - type of gauge metric.
type Gauge struct {
	Name  string
	Value float64
}

// Counter - type of gauge metric.
type Counter struct {
	Name  string
	Value int64
}

// Monitor - for storing runtime metrics.
type (
	Monitor struct {
		Alloc           Gauge
		TotalAlloc      Gauge
		BuckHashSys     Gauge
		Frees           Gauge
		Mallocs         Gauge
		Sys             Gauge
		GCCPUFraction   Gauge
		GCSys           Gauge
		HeapAlloc       Gauge
		HeapIdle        Gauge
		HeapInuse       Gauge
		HeapObjects     Gauge
		HeapReleased    Gauge
		HeapSys         Gauge
		LastGC          Gauge
		Lookups         Gauge
		MCacheInuse     Gauge
		MCacheSys       Gauge
		MSpanInuse      Gauge
		MSpanSys        Gauge
		NextGC          Gauge
		NumForcedGC     Gauge
		NumGC           Gauge
		OtherSys        Gauge
		PauseTotalNs    Gauge
		StackInuse      Gauge
		StackSys        Gauge
		TotalMemory     Gauge
		FreeMemory      Gauge
		CPUutilization1 Gauge

		PollCount   Counter
		RandomValue Gauge
	}
)

// Init - monitor initialization method.
func (m *Monitor) Init() {
	m.Alloc = Gauge{Name: "TotalMemory", Value: 0}
	m.BuckHashSys = Gauge{Name: "FreeMemory", Value: 0}
	m.Frees = Gauge{Name: "CPUutilization1", Value: 0}

	m.Alloc = Gauge{Name: "Alloc", Value: 0}
	m.BuckHashSys = Gauge{Name: "BuckHashSys", Value: 0}
	m.Frees = Gauge{Name: "Frees", Value: 0}
	m.GCCPUFraction = Gauge{Name: "GCCPUFraction", Value: 0}
	m.GCSys = Gauge{Name: "GCSys", Value: 0}
	m.HeapAlloc = Gauge{Name: "HeapAlloc", Value: 0}
	m.HeapIdle = Gauge{Name: "HeapIdle", Value: 0}
	m.HeapInuse = Gauge{Name: "HeapInuse", Value: 0}
	m.HeapObjects = Gauge{Name: "HeapObjects", Value: 0}
	m.HeapReleased = Gauge{Name: "HeapReleased", Value: 0}
	m.HeapSys = Gauge{Name: "HeapSys", Value: 0}
	m.LastGC = Gauge{Name: "LastGC", Value: 0}
	m.Lookups = Gauge{Name: "Lookups", Value: 0}
	m.MCacheInuse = Gauge{Name: "MCacheInuse", Value: 0}
	m.MCacheSys = Gauge{Name: "MCacheSys", Value: 0}
	m.MSpanInuse = Gauge{Name: "MSpanInuse", Value: 0}
	m.MSpanSys = Gauge{Name: "MSpanSys", Value: 0}
	m.Mallocs = Gauge{Name: "Mallocs", Value: 0}
	m.NextGC = Gauge{Name: "NextGC", Value: 0}
	m.NumForcedGC = Gauge{Name: "NumForcedGC", Value: 0}
	m.NumGC = Gauge{Name: "NumGC", Value: 0}
	m.OtherSys = Gauge{Name: "OtherSys", Value: 0}
	m.PauseTotalNs = Gauge{Name: "PauseTotalNs", Value: 0}
	m.StackInuse = Gauge{Name: "StackInuse", Value: 0}
	m.StackSys = Gauge{Name: "StackSys", Value: 0}
	m.Sys = Gauge{Name: "Sys", Value: 0}
	m.TotalAlloc = Gauge{Name: "TotalAlloc", Value: 0}

	m.PollCount = Counter{Name: "PollCount", Value: 0}
	m.RandomValue = Gauge{Name: "RandomValue", Value: 0}
}

// InitParams - store server configuration.
type InitParams struct {
	ConfigPath           string
	PORT                 string
	ValidateAddrPattern  string
	FileStoragePath      string
	DatabaseDSN          string
	Key                  string
	CryptoPrivateKeyPath string
	StoreInterval        int
	Restore              bool
	WaitSecRespDB        time.Duration
}

// InitParamsAgent - store agent configuration.
type InitParamsAgent struct {
	ConfigPath          string
	URL                 string
	PORT                string
	Key                 string
	CryptoPublicKeyPath string
	ReportInterval      int
	PollInterval        int
	ReqInternal         int
	StartReqInterval    int
	CountReqRetries     int
	RateLimit           int
	RepeatedReq         bool
}

// EndpointSettings - store endpoint configuration.
type EndpointSettings struct {
	SendData    *bytes.Reader
	Client      *http.Client
	URL         string
	Hash        string
	Encoding    string
	ContentType string
}

// JobData - store data for the worker.
type JobData struct {
	Mutex  *sync.Mutex
	Par    *InitParamsAgent
	Client *http.Client
	Mon    *Monitor
	Event  string
}
