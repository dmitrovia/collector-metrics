package dbrepository

import (
	"context"
	"fmt"
	"time"

	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
	"github.com/jackc/pgx/v5"
)

// var errGetValueMetric = errors.New("value by name not found")

type DBepository struct {
	databaseDSN   string
	waitSecRespDB time.Duration
	conn          *pgx.Conn
}

func (m *DBepository) Initiate(dsn string, waitSecRespDB time.Duration, conn *pgx.Conn) {
	m.waitSecRespDB = waitSecRespDB
	m.databaseDSN = dsn
	m.conn = conn
}

func (m *DBepository) Init() {
}

func (m *DBepository) GetAllGauges() *map[string]bizmodels.Gauge {
	var gauges map[string]bizmodels.Gauge

	ctx, cancel := context.WithTimeout(context.Background(), m.waitSecRespDB)
	defer cancel()

	gauges = make(map[string]bizmodels.Gauge)

	rows, err := m.conn.Query(ctx, "select name, value from gauges")
	if err != nil {
		fmt.Printf("Query error: %v", err)
	} else {
		defer rows.Close()

		for rows.Next() {
			var name string

			var value float64

			var temp *bizmodels.Gauge

			err = rows.Scan(&name, &value)
			if err != nil {
				fmt.Printf("Scan error: %v", err)
			} else {
				temp = new(bizmodels.Gauge)
				temp.Name = name
				temp.Value = value

				gauges[name] = *temp
			}
		}
	}

	return &gauges
}

func (m *DBepository) GetAllCounters() *map[string]bizmodels.Counter {
	var counters map[string]bizmodels.Counter

	ctx, cancel := context.WithTimeout(context.Background(), m.waitSecRespDB)
	defer cancel()

	counters = make(map[string]bizmodels.Counter)

	rows, err := m.conn.Query(ctx, "select name, value from counters")
	if err != nil {
		fmt.Printf("Query error: %v", err)
	} else {
		defer rows.Close()

		for rows.Next() {
			var name string

			var value int64

			var temp *bizmodels.Counter

			err = rows.Scan(&name, &value)
			if err != nil {
				fmt.Printf("Scan error: %v", err)
			} else {
				temp = new(bizmodels.Counter)
				temp.Name = name
				temp.Value = value

				counters[name] = *temp
			}
		}
	}

	return &counters
}

func (m *DBepository) GetGaugeMetric(name string) (*bizmodels.Gauge, error) {
	var temp *bizmodels.Gauge

	var nameMetric string

	var value float64

	ctx, cancel := context.WithTimeout(context.Background(), m.waitSecRespDB)
	defer cancel()

	temp = new(bizmodels.Gauge)

	err := m.conn.QueryRow(ctx, "select name, value from gauges where name=$1", name).Scan(&nameMetric, &value)
	if err != nil {
		return nil, fmt.Errorf("GetGaugeMetric->m.conn.QueryRow %w", err)
	}

	temp.Name = nameMetric
	temp.Value = value

	return temp, nil
}

func (m *DBepository) GetCounterMetric(name string) (*bizmodels.Counter, error) {
	var temp *bizmodels.Counter

	var nameMetric string

	var value int64

	ctx, cancel := context.WithTimeout(context.Background(), m.waitSecRespDB)
	defer cancel()

	temp = new(bizmodels.Counter)

	err := m.conn.QueryRow(ctx, "select name, value from counters where name=$1", name).Scan(&nameMetric, &value)
	if err != nil {
		return nil, fmt.Errorf("GetGaugeMetric->m.conn.QueryRow %w", err)
	}

	temp.Name = nameMetric
	temp.Value = value

	return temp, nil
}

func (m *DBepository) AddGauge(gauge *bizmodels.Gauge) {
	ctx, cancel := context.WithTimeout(context.Background(), m.waitSecRespDB)
	defer cancel()

	rows, err := m.conn.Exec(ctx, "UPDATE gauges SET value = $1 where name=$2", gauge.Value, gauge.Name)
	if err != nil {
		fmt.Printf("AddGauge->Update error: %v", err)
	}

	if rows.RowsAffected() == 0 {
		_, err := m.conn.Exec(ctx, "INSERT INTO gauges (name, value) VALUES ($1, $2)", gauge.Name, gauge.Value)
		if err != nil {
			fmt.Printf("AddGauge->INSERT INTO error: %v", err)
		}
	}
}

func (m *DBepository) AddCounter(counter *bizmodels.Counter) *bizmodels.Counter {
	ctx, cancel := context.WithTimeout(context.Background(), m.waitSecRespDB)
	defer cancel()

	rows, err := m.conn.Exec(ctx, "UPDATE counters SET value = value + $1 where name=$2", counter.Value, counter.Name)
	if err != nil {
		fmt.Printf("AddCounter->Update error: %v", err)
	}

	if rows.RowsAffected() == 0 {
		_, err := m.conn.Exec(ctx, "INSERT INTO counters (name, value) VALUES ($1, $2)", counter.Name, counter.Value)
		if err != nil {
			fmt.Printf("AddCounter->INSERT INTO error: %v", err)
		}

		return counter
	}

	temp, err := m.GetCounterMetric(counter.Name)
	if err != nil {
		fmt.Printf("AddCounter->m.GetCounterMetric %v", err)
	}

	return temp
}
