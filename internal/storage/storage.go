// Package storage provides describe interface methods.
package storage

import (
	"context"

	"github.com/dmitrovia/collector-metrics/internal/models/apimodels"
	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
)

// Repository - for working with storage metrics.
type Repository interface {
	Init()
	GetGaugeMetric(ctx *context.Context,
		mname string) (*bizmodels.Gauge, error)
	GetCounterMetric(ctx *context.Context,
		mname string) (*bizmodels.Counter, error)
	AddGauge(ctx *context.Context,
		gauge *bizmodels.Gauge) error
	AddCounter(
		ctx *context.Context,
		counter *bizmodels.Counter,
		isNew bool) (*bizmodels.Counter, error)
	GetAllGauges(
		ctx *context.Context) (map[string]bizmodels.Gauge, error)
	GetAllCounters(
		ctx *context.Context) (map[string]bizmodels.Counter,
		error)
	AddMetrics(ctx *context.Context,
		gauges map[string]bizmodels.Gauge,
		counters map[string]bizmodels.Counter) error
	GetAllMetricsAPI(ctx *context.Context) (
		*apimodels.ArrMetrics,
		error)
}
