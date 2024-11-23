package service

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/dmitrovia/collector-metrics/internal/models/apimodels"
	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
	"github.com/dmitrovia/collector-metrics/internal/storage"
)

const fmode os.FileMode = 0o666

var errGetStringValueMetric = errors.New("value by name not found")

type Service interface {
	GetMapStringsAllMetrics() *map[string]string
	AddGauge(mname string, mvalue float64)
	AddCounter(mname string, mvalue int64) *bizmodels.Counter
	GetStringValueGaugeMetric(mname string) (string, error)
	GetStringValueCounterMetric(mname string) (string, error)
	GetValueGaugeMetric(mname string) (float64, error)
	GetValueCounterMetric(mname string) (int64, error)
	SaveInFile(path string) error
	LoadFromFile(path string) error
}

type DataService struct {
	repository storage.Repository
}

func (s *DataService) SaveInFile(path string) error {
	var reqMetric apimodels.Metrics

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, fmode)
	if err != nil {
		return fmt.Errorf("SaveInFile->os.OpenFile: %w", err)
	}

	defer file.Close()

	for _, counter := range *s.repository.GetAllCounters() {
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
			return fmt.Errorf("SaveInFile->Write: %w", err)
		}
	}

	for _, counter := range *s.repository.GetAllGauges() {
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
			return fmt.Errorf("SaveInFile->Write: %w", err)
		}
	}

	return nil
}

func (s *DataService) LoadFromFile(path string) error {
	file, err := os.OpenFile(path, os.O_RDONLY|os.O_EXCL, fmode)
	if err != nil {
		return fmt.Errorf("LoadFromFile->os.OpenFile: %w", err)
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
			s.repository.AddGauge(&bizmodels.Gauge{Name: metric.ID, Value: *metric.Value})
		} else if metric.MType == "counter" {
			s.repository.AddCounter(&bizmodels.Counter{Name: metric.ID, Value: *metric.Delta})
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("LoadFromFile->Err: %w", err)
		}
	}

	return nil
}

func (s *DataService) GetMapStringsAllMetrics() *map[string]string {
	mapMetrics := make(map[string]string)

	for key, value := range *s.repository.GetAllCounters() {
		mapMetrics[key] = strconv.FormatInt(value.Value, 10)
	}

	for key, value := range *s.repository.GetAllGauges() {
		mapMetrics[key] = strconv.FormatFloat(value.Value, 'f', -1, 64)
	}

	return &mapMetrics
}

func (s *DataService) AddGauge(mname string, mvalue float64) {
	s.repository.AddGauge(&bizmodels.Gauge{Name: mname, Value: mvalue})
}

func (s *DataService) AddCounter(mname string, mvalue int64) *bizmodels.Counter {
	return s.repository.AddCounter(&bizmodels.Counter{Name: mname, Value: mvalue})
}

func (s *DataService) GetStringValueGaugeMetric(mname string) (string, error) {
	val, err := s.repository.GetGaugeMetric(mname)
	if err != nil {
		return "0", errGetStringValueMetric
	}

	return strconv.FormatFloat(val.Value, 'f', -1, 64), nil
}

func (s *DataService) GetStringValueCounterMetric(mname string) (string, error) {
	val, err := s.repository.GetCounterMetric(mname)
	if err != nil {
		return "0", errGetStringValueMetric
	}

	return strconv.FormatInt(val.Value, 10), nil
}

func (s *DataService) GetValueGaugeMetric(mname string) (float64, error) {
	val, err := s.repository.GetGaugeMetric(mname)
	if err != nil {
		return 0, fmt.Errorf("GetValueGaugeMetric: %w", err)
	}

	return val.Value, nil
}

func (s *DataService) GetValueCounterMetric(mname string) (int64, error) {
	val, err := s.repository.GetCounterMetric(mname)
	if err != nil {
		return 0, fmt.Errorf("GetValueCounterMetric: %w", err)
	}

	return val.Value, nil
}

func NewMemoryService(repository storage.Repository) *DataService {
	return &DataService{repository: repository}
}
