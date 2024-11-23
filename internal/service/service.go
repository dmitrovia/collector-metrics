package service

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/dmitrovia/collector-metrics/internal/models/apimodels"
	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
	"github.com/dmitrovia/collector-metrics/internal/storage"
)

const fmode os.FileMode = 0o666

type Service interface {
	AddGauge(mname string, mvalue float64) error
	AddCounter(mname string, mvalue int64) (*bizmodels.Counter, error)
	GetValueGaugeMetric(mname string) (float64, error)
	GetValueCounterMetric(mname string) (int64, error)
	SaveInFile(path string) error
	LoadFromFile(path string) error
	AddMetrics(gauges map[string]bizmodels.Gauge, counters map[string]bizmodels.Counter) error
	GetAllGauges() *map[string]bizmodels.Gauge
	GetAllCounters() *map[string]bizmodels.Counter
}

type DataService struct {
	repository storage.Repository
}

func (s *DataService) GetAllGauges() *map[string]bizmodels.Gauge {
	return s.repository.GetAllGauges()
}

func (s *DataService) GetAllCounters() *map[string]bizmodels.Counter {
	return s.repository.GetAllCounters()
}

func (s *DataService) AddMetrics(gauges map[string]bizmodels.Gauge, counters map[string]bizmodels.Counter) error {
	err := s.repository.AddMetrics(gauges, counters)
	if err != nil {
		return fmt.Errorf("DataService->AddMetrics: %w", err)
	}
	return nil
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
			err := s.repository.AddGauge(&bizmodels.Gauge{Name: metric.ID, Value: *metric.Value})
			if err != nil {
				fmt.Println(err)
			}
		} else if metric.MType == "counter" {
			_, err := s.repository.AddCounter(&bizmodels.Counter{Name: metric.ID, Value: *metric.Delta})
			if err != nil {
				fmt.Println(err)
			}
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("LoadFromFile->Err: %w", err)
		}
	}

	return nil
}

func (s *DataService) AddGauge(mname string, mvalue float64) error {
	err := s.repository.AddGauge(&bizmodels.Gauge{Name: mname, Value: mvalue})
	if err != nil {
		return fmt.Errorf("AddGauge->s.repository.AddGauge %w", err)
	}

	return nil
}

func (s *DataService) AddCounter(mname string, mvalue int64) (*bizmodels.Counter, error) {
	res, err := s.repository.AddCounter(&bizmodels.Counter{Name: mname, Value: mvalue})
	if err != nil {
		return nil, fmt.Errorf("AddCounter->.repository.AddCounter %w", err)
	}

	return res, nil
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
