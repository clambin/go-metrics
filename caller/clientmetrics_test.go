package caller_test

import (
	"errors"
	"github.com/clambin/go-metrics/caller"
	"github.com/clambin/go-metrics/tools"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestAPIClientMetrics_MakeLatencyTimer(t *testing.T) {
	cfg := caller.ClientMetrics{}

	// MakeLatencyTimer returns nil if no Latency metric is set
	timer := cfg.MakeLatencyTimer()
	assert.Nil(t, timer)

	cfg = caller.ClientMetrics{
		Latency: promauto.NewSummaryVec(prometheus.SummaryOpts{
			Name: "latency_metrics",
			Help: "Latency metric",
		}, []string{}),
	}

	// collect metrics
	timer = cfg.MakeLatencyTimer()
	require.NotNil(t, timer)
	time.Sleep(10 * time.Millisecond)
	timer.ObserveDuration()

	// one measurement should be collected
	ch := make(chan prometheus.Metric)
	go cfg.Latency.Collect(ch)
	m := <-ch
	assert.Equal(t, uint64(1), tools.MetricValue(m).GetSummary().GetSampleCount())
	assert.NotZero(t, tools.MetricValue(m).GetSummary().GetSampleSum())
}

func TestAPIClientMetrics_ReportErrors(t *testing.T) {
	cfg := caller.ClientMetrics{}

	// ReportErrors doesn't crash when no Errors metric is set
	cfg.ReportErrors(nil)

	cfg = caller.ClientMetrics{
		Errors: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "error_metric",
			Help: "Error metric",
		}, []string{}),
	}

	// collect metrics
	cfg.ReportErrors(nil)

	// do a measurement
	ch := make(chan prometheus.Metric)
	go cfg.Errors.Collect(ch)
	m := <-ch
	assert.Equal(t, 0.0, tools.MetricValue(m).GetCounter().GetValue())

	// record an error
	cfg.ReportErrors(errors.New("some error"))

	// counter should now be 1
	ch = make(chan prometheus.Metric)
	go cfg.Errors.Collect(ch)
	m = <-ch
	assert.Equal(t, 1.0, tools.MetricValue(m).GetCounter().GetValue())
}

func TestAPIClientMetrics_Nil(t *testing.T) {
	cfg := caller.ClientMetrics{}

	timer := cfg.MakeLatencyTimer("foo")
	assert.Nil(t, timer)
	cfg.ReportErrors(nil, "foo")
}