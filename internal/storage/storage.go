package storage

import "github.com/dmitrovia/collector-metrics/internal/models/bizmodels"

type Repository interface {
	Init()
	GetStringValueGaugeMetric(name string) (string, error)
	GetStringValueCounterMetric(name string) (string, error)
	GetValueGaugeMetric(mname string) (float64, error)
	GetValueCounterMetric(mname string) (int64, error)
	GetMapStringsAllMetrics() *map[string]string
	AddGauge(gauge *bizmodels.Gauge)
	AddCounter(counter *bizmodels.Counter) *bizmodels.Counter
}
