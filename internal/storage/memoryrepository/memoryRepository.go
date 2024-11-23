package memoryrepository

import (
	"errors"
	"fmt"

	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
)

var errGetValueMetric = errors.New("value by name not found")

type MemoryRepository struct {
	gauges   map[string]bizmodels.Gauge
	counters map[string]bizmodels.Counter
}

func (m *MemoryRepository) AddMetrics(gauges map[string]bizmodels.Gauge, counters map[string]bizmodels.Counter) error {
	for _, gauge := range gauges {
		err := m.AddGauge(&gauge)
		if err != nil {
			fmt.Println(err)
		}
	}

	for _, counter := range counters {
		_, err := m.AddCounter(&counter)
		if err != nil {
			fmt.Println(err)
		}
	}

	return nil
}

func (m *MemoryRepository) Init() {
	m.gauges = make(map[string]bizmodels.Gauge)
	m.counters = make(map[string]bizmodels.Counter)
}

func (m *MemoryRepository) GetAllGauges() *map[string]bizmodels.Gauge {
	return &m.gauges
}

func (m *MemoryRepository) GetAllCounters() *map[string]bizmodels.Counter {
	return &m.counters
}

func (m *MemoryRepository) GetGaugeMetric(name string) (*bizmodels.Gauge, error) {
	val, ok := m.gauges[name]
	if ok {
		return &val, nil
	}

	return nil, errGetValueMetric
}

func (m *MemoryRepository) GetCounterMetric(name string) (*bizmodels.Counter, error) {
	val, ok := m.counters[name]
	if ok {
		return &val, nil
	}

	return nil, errGetValueMetric
}

func (m *MemoryRepository) AddGauge(gauge *bizmodels.Gauge) error {
	m.gauges[gauge.Name] = *gauge

	return nil
}

func (m *MemoryRepository) AddCounter(counter *bizmodels.Counter) (*bizmodels.Counter, error) {
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
