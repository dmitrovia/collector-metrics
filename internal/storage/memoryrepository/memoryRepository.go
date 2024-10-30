package memoryrepository

import (
	"errors"
	"strconv"

	"github.com/dmitrovia/collector-metrics/internal/models"
)

var errGetStringValueCounterMetric = errors.New("addr is not valid")

type MemoryRepository struct {
	gauges   map[string]models.Gauge
	counters map[string]models.Counter
}

func (m *MemoryRepository) Init() {
	m.gauges = make(map[string]models.Gauge)
	m.counters = make(map[string]models.Counter)
}

func (m *MemoryRepository) GetStringValueGaugeMetric(name string) (string, error) {
	val, ok := m.gauges[name]
	if ok {
		return strconv.FormatFloat(val.Value, 'f', -1, 64), nil
	}

	return "", errGetStringValueCounterMetric
}

func (m *MemoryRepository) GetStringValueCounterMetric(name string) (string, error) {
	val, ok := m.counters[name]
	if ok {
		return strconv.FormatInt(val.Value, 10), nil
	}

	return "", errGetStringValueCounterMetric
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

func (m *MemoryRepository) AddGauge(gauge *models.Gauge) {
	m.gauges[gauge.Name] = *gauge
}

func (m *MemoryRepository) AddCounter(counter *models.Counter) {
	val, ok := m.counters[counter.Name]

	var temp *models.Counter

	if ok {
		temp = new(models.Counter)
		temp.Name = val.Name
		temp.Value = val.Value + counter.Value
		m.counters[counter.Name] = *temp
	} else {
		m.counters[counter.Name] = *counter
	}
}
