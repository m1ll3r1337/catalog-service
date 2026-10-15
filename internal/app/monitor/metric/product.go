package mmetric

import (
	"fmt"
	"time"

	"github.com/m1ll3r1337/catalog-service/internal/app/monitor"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	labelResult = "result"
	labelReason = "reason"
)

type createResult string

const (
	createResultSuccess createResult = "success"
	createResultError   createResult = "error"
)

type reason string

const reasonNone reason = "none"

type Product struct {
	createTotal           *prometheus.CounterVec   // labels: result, reason
	createDurationSeconds *prometheus.HistogramVec // labels: result, reason
	createdTotal          prometheus.Counter
	createdAmountTotal    prometheus.Counter
}

var _ monitor.ProductMetrics = (*Product)(nil)

func NewPrometheusProduct(registerer prometheus.Registerer) (*Product, error) {
	if registerer == nil {
		registerer = prometheus.DefaultRegisterer
	}

	createTotalName := "create_total"
	createTotalCounterVec, err := counterVec(registerer, createTotalName, "", labelResult, labelReason)
	if err != nil {
		return nil, fmt.Errorf("register %s: %w", createTotalName, err)
	}

	createDurationSecondsName := "create_duration_seconds"
	createDurationSecondsHistogramVec, err := histogramVec(registerer, createDurationSecondsName, "", prometheus.DefBuckets, labelResult, labelReason)
	if err != nil {
		return nil, fmt.Errorf("register %s: %w", createDurationSecondsName, err)
	}

	createdTotalName := "created_total"
	createdTotalCounter, err := counter(registerer, createdTotalName, "")
	if err != nil {
		return nil, fmt.Errorf("register %s: %w", createdTotalName, err)
	}

	createdAmountTotalName := "created_amount_cents_total"
	createdAmountTotalCounter, err := counter(registerer, createdAmountTotalName, "")
	if err != nil {
		return nil, fmt.Errorf("register %s: %w", createdAmountTotalName, err)
	}

	return &Product{
		createTotal:           createTotalCounterVec,
		createDurationSeconds: createDurationSecondsHistogramVec,
		createdTotal:          createdTotalCounter,
		createdAmountTotal:    createdAmountTotalCounter,
	}, nil
}

type productCreateMetric struct {
	parent  *Product
	started time.Time
}

func (m *Product) Create() monitor.ProductCreateMetric {
	return &productCreateMetric{
		parent:  m,
		started: time.Now(),
	}
}

func (m *productCreateMetric) Success(price int64) {
	duration := time.Since(m.started).Seconds()
	m.parent.createTotal.WithLabelValues(string(createResultSuccess), string(reasonNone)).Inc()
	m.parent.createDurationSeconds.WithLabelValues(string(createResultSuccess), string(reasonNone)).Observe(duration)
	m.parent.createdTotal.Inc()
	if price > 0 {
		m.parent.createdAmountTotal.Add(float64(price))
	}
}

func (m *productCreateMetric) Failed(err error) {
	duration := time.Since(m.started).Seconds()
	r := createReason(err)
	m.parent.createTotal.WithLabelValues(string(createResultError), r).Inc()
	m.parent.createDurationSeconds.WithLabelValues(string(createResultError), r).Observe(duration)
}
