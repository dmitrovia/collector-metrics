package service

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/dmitrovia/collector-metrics/internal/models/apimodels"
	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
	"github.com/dmitrovia/collector-metrics/internal/storage"
)

const fmd os.FileMode = 0o666

type Service interface {
	AddGauge(mname string, mvalue float64) error
	AddCounter(
		mname string,
		mvalue int64) (*bizmodels.Counter, error)
	GetValueGM(mname string) (float64, error)
	GetValueCM(mname string) (int64, error)
	SaveInFile(pth string) error
	LoadFromFile(pth string) error
	AddMetrics(
		gms map[string]bizmodels.Gauge,
		cms map[string]bizmodels.Counter) error
	GetAllGauges() (*map[string]bizmodels.Gauge, error)
	GetAllCounters() (*map[string]bizmodels.Counter, error)
}

type DS struct {
	repository  storage.Repository
	ctxDuration time.Duration
}

func (s *DS) GetAllGauges() (
	*map[string]bizmodels.Gauge, error,
) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		s.ctxDuration)
	defer cancel()

	gauges, err := s.repository.GetAllGauges(&ctx)
	if err != nil {
		return nil, fmt.Errorf("GetAllGauges->GetAllGauges: %w",
			err)
	}

	return gauges, nil
}

func (s *DS) GetAllCounters() (
	*map[string]bizmodels.Counter, error,
) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		s.ctxDuration)
	defer cancel()

	counters, err := s.repository.GetAllCounters(&ctx)
	if err != nil {
		return nil, fmt.Errorf("GetAllGauges->GetAllCounters: %w",
			err)
	}

	return counters, nil
}

func (s *DS) AddMetrics(
	gms map[string]bizmodels.Gauge,
	cms map[string]bizmodels.Counter,
) error {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		s.ctxDuration)
	defer cancel()

	err := s.repository.AddMetrics(&ctx, gms, cms)
	if err != nil {
		return fmt.Errorf("DataService->AddMetrics: %w", err)
	}

	return nil
}

func (s *DS) SaveInFile(pth string) error {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		s.ctxDuration)
	defer cancel()

	file, err := os.OpenFile(pth, os.O_WRONLY|os.O_CREATE, fmd)
	if err != nil {
		return fmt.Errorf("SaveInFile->os.OpenFile: %w", err)
	}

	defer file.Close()

	counters, err := s.repository.GetAllCounters(&ctx)
	if err != nil {
		return fmt.Errorf("SaveInFile->GetAllCounters: %w",
			err)
	}

	err = saveCounters(file, counters)
	if err != nil {
		return fmt.Errorf("SaveInFile->saveCounters: %w",
			err)
	}

	gauges, err := s.repository.GetAllGauges(&ctx)
	if err != nil {
		return fmt.Errorf("SaveInFile->GetAllGauges: %w",
			err)
	}

	err = saveGauges(file, gauges)
	if err != nil {
		return fmt.Errorf("SaveInFile->saveGauges: %w",
			err)
	}

	return nil
}

func saveCounters(file *os.File,
	counters *map[string]bizmodels.Counter,
) error {
	var reqMetric apimodels.Metrics

	for _, counter := range *counters {
		reqMetric = apimodels.Metrics{}
		reqMetric.ID = counter.Name
		reqMetric.MType = bizmodels.CounterName
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

	return nil
}

func saveGauges(file *os.File,
	gauges *map[string]bizmodels.Gauge,
) error {
	var reqMetric apimodels.Metrics

	for _, gauge := range *gauges {
		reqMetric = apimodels.Metrics{}
		reqMetric.ID = gauge.Name
		reqMetric.MType = bizmodels.GaugeName
		reqMetric.Value = &gauge.Value

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

func (s *DS) LoadFromFile(pth string) error {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		s.ctxDuration)
	defer cancel()

	file, err := os.OpenFile(pth, os.O_RDONLY|os.O_EXCL, fmd)
	if err != nil {
		return fmt.Errorf("LoadFromFile->os.OpenFile: %w", err)
	}

	defer file.Close()
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		data := scanner.Bytes()
		tmpm := apimodels.Metrics{}

		err = json.Unmarshal(data, &tmpm)
		if err != nil {
			return err
		}

		if tmpm.MType == bizmodels.GaugeName {
			gauge := bizmodels.Gauge{
				Name:  tmpm.ID,
				Value: *tmpm.Value,
			}

			err := s.repository.AddGauge(&ctx, &gauge)
			if err != nil {
				return fmt.Errorf("LoadFromFile->AddGauge: %w", err)
			}
		} else if tmpm.MType == bizmodels.CounterName {
			counter := bizmodels.Counter{
				Name:  tmpm.ID,
				Value: *tmpm.Delta,
			}

			_, err := s.repository.AddCounter(&ctx, &counter)
			if err != nil {
				return fmt.Errorf("LoadFromFile->AddCounter: %w", err)
			}
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("LoadFromFile->Err: %w", err)
		}
	}

	return nil
}

func (s *DS) AddGauge(mname string, mvalue float64) error {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		s.ctxDuration)
	defer cancel()

	gauge := bizmodels.Gauge{Name: mname, Value: mvalue}

	err := s.repository.AddGauge(&ctx, &gauge)
	if err != nil {
		return fmt.Errorf("AddGauge->AddGauge %w", err)
	}

	return nil
}

func (s *DS) AddCounter(
	name string,
	value int64,
) (*bizmodels.Counter, error) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		s.ctxDuration)
	defer cancel()

	counter := bizmodels.Counter{Name: name, Value: value}

	res, err := s.repository.AddCounter(&ctx, &counter)
	if err != nil {
		return nil, fmt.Errorf("AddCounter->AddCounter %w", err)
	}

	return res, nil
}

func (s *DS) GetValueGM(mname string) (float64, error) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		s.ctxDuration)
	defer cancel()

	val, err := s.repository.GetGaugeMetric(&ctx, mname)
	if err != nil {
		return 0, fmt.Errorf("GetValueGM: %w", err)
	}

	return val.Value, nil
}

func (s *DS) GetValueCM(mname string) (int64, error) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		s.ctxDuration)
	defer cancel()

	val, err := s.repository.GetCounterMetric(&ctx, mname)
	if err != nil {
		return 0, fmt.Errorf("GetValueCM: %w", err)
	}

	return val.Value, nil
}

func NewMemoryService(repository storage.Repository,
	ctxDur time.Duration,
) *DS {
	return &DS{repository: repository, ctxDuration: ctxDur}
}
