package service

import (
	"fmt"

	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
	"github.com/dmitrovia/collector-metrics/internal/storage"
)

type Service interface {
	GetMapStringsAllMetrics() *map[string]string
	AddGauge(mname string, mvalue float64)
	AddCounter(mname string, mvalue int64) *bizmodels.Counter
	GetStringValueGaugeMetric(mname string) (string, error)
	GetStringValueCounterMetric(mname string) (string, error)
	GetValueGaugeMetric(mname string) (float64, error)
	GetValueCounterMetric(mname string) (int64, error)
}

type MemoryService struct {
	repository storage.Repository
}

func (s *MemoryService) GetMapStringsAllMetrics() *map[string]string {
	return s.repository.GetMapStringsAllMetrics()
}

func (s *MemoryService) AddGauge(mname string, mvalue float64) {
	s.repository.AddGauge(&bizmodels.Gauge{Name: mname, Value: mvalue})
}

func (s *MemoryService) AddCounter(mname string, mvalue int64) *bizmodels.Counter {
	return s.repository.AddCounter(&bizmodels.Counter{Name: mname, Value: mvalue})
}

func (s *MemoryService) GetStringValueGaugeMetric(mname string) (string, error) {
	val, err := s.repository.GetStringValueGaugeMetric(mname)
	if err != nil {
		return val, fmt.Errorf("GetStringValueGaugeMetric: %w", err)
	}

	return val, nil
}

func (s *MemoryService) GetStringValueCounterMetric(mname string) (string, error) {
	val, err := s.repository.GetStringValueCounterMetric(mname)
	if err != nil {
		return val, fmt.Errorf("GetStringValueCounterMetric: %w", err)
	}

	return val, nil
}

func (s *MemoryService) GetValueGaugeMetric(mname string) (float64, error) {
	val, err := s.repository.GetValueGaugeMetric(mname)
	if err != nil {
		return val, fmt.Errorf("GetValueGaugeMetric: %w", err)
	}

	return val, nil
}

func (s *MemoryService) GetValueCounterMetric(mname string) (int64, error) {
	val, err := s.repository.GetValueCounterMetric(mname)
	if err != nil {
		return val, fmt.Errorf("GetValueCounterMetric: %w", err)
	}

	return val, nil
}

func NewMemoryService(repository storage.Repository) *MemoryService {
	return &MemoryService{repository: repository}
}
