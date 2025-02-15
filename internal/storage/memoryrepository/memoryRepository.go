package memoryrepository

import (
	"context"
	"errors"
	"fmt"

	"github.com/dmitrovia/collector-metrics/internal/models/apimodels"
	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
)

var errGetValueMetric = errors.New(
	"value by name not found")

var errMethod = errors.New(
	"method not implemented")

type MemoryRepository struct {
	gauges   map[string]bizmodels.Gauge
	counters map[string]bizmodels.Counter
}

func (m *MemoryRepository) AddMetrics(
	ctx *context.Context,
	gauges map[string]bizmodels.Gauge,
	counters map[string]bizmodels.Counter,
) error {
	for _, gauge := range gauges {
		err := m.AddGauge(ctx, &gauge)
		if err != nil {
			return fmt.Errorf("AddMetrics->m.AddGauge: %w", err)
		}
	}

	for _, counter := range counters {
		_, err := m.AddCounter(ctx, &counter)
		if err != nil {
			return fmt.Errorf("AddMetrics->m.AddCounter: %w", err)
		}
	}

	return nil
}

func (m *MemoryRepository) Init() {
	m.gauges = make(map[string]bizmodels.Gauge)
	m.counters = make(map[string]bizmodels.Counter)
}

func (m *MemoryRepository) GetAllGauges(
	_ *context.Context) (
	map[string]bizmodels.Gauge, error,
) {
	return m.gauges, nil
}

func (m *MemoryRepository) GetAllCounters(
	_ *context.Context) (
	map[string]bizmodels.Counter, error,
) {
	return m.counters, nil
}

func (m *MemoryRepository) GetGaugeMetric(
	_ *context.Context,
	name string,
) (*bizmodels.Gauge, error) {
	val, ok := m.gauges[name]
	if ok {
		return &val, nil
	}

	return nil, errGetValueMetric
}

func (m *MemoryRepository) GetCounterMetric(
	_ *context.Context,
	name string,
) (*bizmodels.Counter, error) {
	val, ok := m.counters[name]
	if ok {
		return &val, nil
	}

	return nil, errGetValueMetric
}

func (m *MemoryRepository) AddGauge(
	_ *context.Context,
	gauge *bizmodels.Gauge,
) error {
	m.gauges[gauge.Name] = *gauge

	return nil
}

func (m *MemoryRepository) AddCounter(
	_ *context.Context,
	counter *bizmodels.Counter,
) (*bizmodels.Counter, error) {
	val, ok := m.counters[counter.Name]

	var temp *bizmodels.Counter

	if ok {
		temp = new(bizmodels.Counter)
		temp.Name = val.Name
		temp.Value = val.Value + counter.Value
		m.counters[counter.Name] = *temp

		return temp, nil
	}

	m.counters[counter.Name] = *counter

	return counter, nil
}

func (m *MemoryRepository) GetAllMetricsAPI(
	_ *context.Context,
) (*apimodels.ArrMetrics, error) {
	return nil, errMethod
}
