package storage

import "github.com/dmitrovia/collector-metrics/internal/models"

type Repository interface {
	Init()
	GetStringValueGaugeMetric(name string) (string, error)
	GetStringValueCounterMetric(name string) (string, error)
	GetMapStringsAllMetrics() *map[string]string
	AddGauge(gauge *models.Gauge)
	AddCounter(counter *models.Counter)
}
