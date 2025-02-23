// Package memoryrepository provides
// working with memory
package memoryrepository

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/dmitrovia/collector-metrics/internal/models/apimodels"
	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
)

var errGetValueMetric = errors.New(
	"value by name not found")

// MemoryRepository - describing the storage.
type MemoryRepository struct {
	gauges   map[string]bizmodels.Gauge
	counters map[string]bizmodels.Counter
	mutexG   *sync.Mutex
	mutexC   *sync.Mutex
}

// AddMetrics - adds metrics to the memory.
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

// Init - initialization of initial parameters.
func (m *MemoryRepository) Init() {
	m.gauges = make(map[string]bizmodels.Gauge)
	m.counters = make(map[string]bizmodels.Counter)
	m.mutexG = &sync.Mutex{}
	m.mutexC = &sync.Mutex{}
}

// GetAllGauges - get all gauges metrics from memory.
func (m *MemoryRepository) GetAllGauges(
	_ *context.Context) (
	map[string]bizmodels.Gauge, error,
) {
	return m.gauges, nil
}

// GetAllCounters - get all counters metrics from memory.
func (m *MemoryRepository) GetAllCounters(
	_ *context.Context) (
	map[string]bizmodels.Counter, error,
) {
	return m.counters, nil
}

// GetGaugeMetric - get gauge metric by name from memory.
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

// GetCounterMetric - get counter
// metric by name from memory.
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

// AddGauge - add the gauge metric to the memory.
func (m *MemoryRepository) AddGauge(
	_ *context.Context,
	gauge *bizmodels.Gauge,
) error {
	m.mutexG.Lock()
	defer m.mutexG.Unlock()
	m.gauges[gauge.Name] = *gauge

	return nil
}

// AddCounter - add the coutner metric to the memory.
func (m *MemoryRepository) AddCounter(
	_ *context.Context,
	counter *bizmodels.Counter,
) (*bizmodels.Counter, error) {
	m.mutexC.Lock()
	defer m.mutexC.Unlock()
	val, ok := m.counters[counter.Name]

	var temp *bizmodels.Counter

	if ok {
		temp = &bizmodels.Counter{}
		temp.Name = val.Name
		temp.Value = val.Value + counter.Value
		m.counters[counter.Name] = *temp

		return temp, nil
	}

	m.counters[counter.Name] = *counter

	return counter, nil
}

// GetAllMetricsAPI - get all metrics in API format.
func (m *MemoryRepository) GetAllMetricsAPI(
	ctx *context.Context,
) (*apimodels.ArrMetrics, error) {
	arr1, err := m.GetAllGaugesAPI(ctx)
	if err != nil {
		return nil, fmt.Errorf(
			"GetAllMetricsAPI->m.GetAllGaugesAPI %w", err)
	}

	arr2, err := m.GetAllCountersAPI(ctx)
	if err != nil {
		return nil, fmt.Errorf(
			"GetAllMetricsAPI->m.GetAllCountersAPI %w", err)
	}

	result := make(apimodels.ArrMetrics, 0)
	result = append(result, arr1...)
	result = append(result, arr2...)

	return &result, nil
}

// GetAllGaugesAPI - get all gauge metrics in API format.
func (m *MemoryRepository) GetAllGaugesAPI(
	ctx *context.Context) (
	apimodels.ArrMetrics,
	error,
) {
	apigauges := make(apimodels.ArrMetrics, 0)
	gauges, _ := m.GetAllGauges(ctx)

	for _, gauge := range gauges {
		temp := &apimodels.Metrics{}
		temp.ID = gauge.Name
		temp.Value = &gauge.Value
		temp.MType = bizmodels.GaugeName

		apigauges = append(apigauges, *temp)
	}

	return apigauges, nil
}

// GetAllCountersAPI - get all
// counter metrics in API format.
func (m *MemoryRepository) GetAllCountersAPI(
	ctx *context.Context) (
	apimodels.ArrMetrics,
	error,
) {
	apicounters := make(apimodels.ArrMetrics, 0)
	counters, _ := m.GetAllCounters(ctx)

	for _, counter := range counters {
		temp := &apimodels.Metrics{}
		temp.ID = counter.Name
		temp.Delta = &counter.Value
		temp.MType = bizmodels.CounterName

		apicounters = append(apicounters, *temp)
	}

	return apicounters, nil
}
