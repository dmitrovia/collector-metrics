package memoryrepository

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/dmitrovia/collector-metrics/internal/models/apimodels"
	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
)

var errGetStringValueMetric = errors.New("value by name not found")

var errGetValueMetric = errors.New("value by name not found")

const fmode os.FileMode = 0o666

type MemoryRepository struct {
	gauges   map[string]bizmodels.Gauge
	counters map[string]bizmodels.Counter
}

func (m *MemoryRepository) Init() {
	m.gauges = make(map[string]bizmodels.Gauge)
	m.counters = make(map[string]bizmodels.Counter)
}

func (m *MemoryRepository) getAllGauges() *map[string]bizmodels.Gauge {
	return &m.gauges
}

func (m *MemoryRepository) getAllCounters() *map[string]bizmodels.Counter {
	return &m.counters
}

func (m *MemoryRepository) SaveInFile(path string) error {
	var reqMetric apimodels.Metrics

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, fmode)
	if err != nil {
		return fmt.Errorf("MemoryRepository->SaveInFile: %w", err)
	}

	defer file.Close()

	for _, counter := range *m.getAllCounters() {
		reqMetric = apimodels.Metrics{}
		reqMetric.ID = counter.Name
		reqMetric.MType = "counter"
		reqMetric.Delta = &counter.Value

		data, err := json.Marshal(&reqMetric)
		if err != nil {
			return err
		}

		data = append(data, '\n')

		_, err = file.Write(data)
		if err != nil {
			return fmt.Errorf("MemoryRepository->Write: %w", err)
		}
	}

	for _, counter := range *m.getAllGauges() {
		reqMetric = apimodels.Metrics{}
		reqMetric.ID = counter.Name
		reqMetric.MType = "gauge"
		reqMetric.Value = &counter.Value

		data, err := json.Marshal(&reqMetric)
		if err != nil {
			return err
		}

		data = append(data, '\n')

		_, err = file.Write(data)
		if err != nil {
			return fmt.Errorf("MemoryRepository->SaveInFile: %w", err)
		}
	}

	return nil
}

func (m *MemoryRepository) LoadFromFile(path string) error {
	file, err := os.OpenFile(path, os.O_RDONLY|os.O_EXCL, fmode)
	if err != nil {
		return fmt.Errorf("LoadFromFile->OpenFile: %w", err)
	}

	defer file.Close()
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		data := scanner.Bytes()
		metric := apimodels.Metrics{}

		err = json.Unmarshal(data, &metric)
		if err != nil {
			return err
		}

		if metric.MType == "gauge" {
			m.AddGauge(&bizmodels.Gauge{Name: metric.ID, Value: *metric.Value})
		} else if metric.MType == "counter" {
			m.AddCounter(&bizmodels.Counter{Name: metric.ID, Value: *metric.Delta})
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("LoadFromFile->Err: %w", err)
		}
	}

	return nil
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
