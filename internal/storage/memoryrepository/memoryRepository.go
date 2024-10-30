package memoryrepository

import (
	"errors"
	"strconv"

	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
)

var errGetStringValueMetric = errors.New("value by name not found")

var errGetValueMetric = errors.New("value by name not found")

type MemoryRepository struct {
	gauges   map[string]bizmodels.Gauge
	counters map[string]bizmodels.Counter
}

func (m *MemoryRepository) Init() {
	m.gauges = make(map[string]bizmodels.Gauge)
	m.counters = make(map[string]bizmodels.Counter)
}

func (m *MemoryRepository) GetStringValueGaugeMetric(name string) (string, error) {
	val, ok := m.gauges[name]
	if ok {
		return strconv.FormatFloat(val.Value, 'f', -1, 64), nil
	}

	return "0", errGetStringValueMetric
}

func (m *MemoryRepository) GetStringValueCounterMetric(name string) (string, error) {
	val, ok := m.counters[name]
	if ok {
		return strconv.FormatInt(val.Value, 10), nil
	}

	return "0", errGetStringValueMetric
}

func (m *MemoryRepository) GetValueGaugeMetric(name string) (float64, error) {
	val, ok := m.gauges[name]
	if ok {
		return val.Value, nil
	}

	return 0, errGetValueMetric
}

func (m *MemoryRepository) GetValueCounterMetric(name string) (int64, error) {
	val, ok := m.counters[name]
	if ok {
		return val.Value, nil
	}

	return 0, errGetValueMetric
}

func (m *MemoryRepository) GetMapStringsAllMetrics() *map[string]string {
	mapMetrics := make(map[string]string)

	for key, value := range m.counters {
		mapMetrics[key] = strconv.FormatInt(value.Value, 10)
	}

	for key, value := range m.gauges {
		mapMetrics[key] = strconv.FormatFloat(value.Value, 'f', -1, 64)
	}

	return &mapMetrics
}

func (m *MemoryRepository) AddGauge(gauge *bizmodels.Gauge) {
	m.gauges[gauge.Name] = *gauge
}

func (m *MemoryRepository) AddCounter(counter *bizmodels.Counter) *bizmodels.Counter {
	val, ok := m.counters[counter.Name]

	var temp *bizmodels.Counter

	if ok {
		temp = new(bizmodels.Counter)
		temp.Name = val.Name
		temp.Value = val.Value + counter.Value
		m.counters[counter.Name] = *temp

		return temp
	}

	m.counters[counter.Name] = *counter

	return counter
}
