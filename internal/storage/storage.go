package storage

import "github.com/dmitrovia/collector-metrics/internal/models/bizmodels"

type Repository interface {
	Init()
	GetGaugeMetric(mname string) (*bizmodels.Gauge, error)
	GetCounterMetric(mname string) (*bizmodels.Counter, error)
	AddGauge(gauge *bizmodels.Gauge) error
	AddCounter(counter *bizmodels.Counter) (*bizmodels.Counter, error)
	GetAllGauges() *map[string]bizmodels.Gauge
	GetAllCounters() *map[string]bizmodels.Counter
	AddMetrics(gauges map[string]bizmodels.Gauge, counters map[string]bizmodels.Counter) error
}
