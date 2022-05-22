package client_test

import (
	"errors"
	"github.com/clambin/go-metrics/client"
	"github.com/clambin/go-metrics/tools"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"time"
)

func TestClientMetrics_MakeLatencyTimer(t *testing.T) {
	cfg := client.Metrics{}

	// MakeLatencyTimer returns nil if no Latency metric is set
	timer := cfg.MakeLatencyTimer()
	assert.Nil(t, timer)

	cfg = client.NewMetrics("foo", "")

	// collect metrics
	timer = cfg.MakeLatencyTimer("foo", "/bar", http.MethodGet)
	require.NotNil(t, timer)
	time.Sleep(10 * time.Millisecond)
	timer.ObserveDuration()

	// one measurement should be collected
	ch := make(chan prometheus.Metric)
	go cfg.Latency.Collect(ch)
	m := <-ch
	assert.Contains(t, m.Desc().String(), `{fqName: "foo_api_latency", `)
	assert.Equal(t, uint64(1), tools.MetricValue(m).GetSummary().GetSampleCount())
	assert.NotZero(t, tools.MetricValue(m).GetSummary().GetSampleSum())
}

func TestClientMetrics_ReportErrors(t *testing.T) {
	cfg := client.Metrics{}

	// ReportErrors doesn't crash when no Errors metric is set
	cfg.ReportErrors(nil)

	cfg = client.NewMetrics("bar", "")

	// collect metrics
	cfg.ReportErrors(nil, "foo", "/bar", http.MethodGet)

	// do a measurement
	ch := make(chan prometheus.Metric)
	go cfg.Errors.Collect(ch)
	m := <-ch
	assert.Equal(t, 0.0, tools.MetricValue(m).GetCounter().GetValue())

	// record an error
	cfg.ReportErrors(errors.New("some error"), "foo", "/bar", http.MethodGet)

	// counter should now be 1
	ch = make(chan prometheus.Metric)
	go cfg.Errors.Collect(ch)
	m = <-ch
	assert.Equal(t, 1.0, tools.MetricValue(m).GetCounter().GetValue())
}

func TestClientMetrics_Nil(t *testing.T) {
	cfg := client.Metrics{}

	timer := cfg.MakeLatencyTimer("snafu")
	assert.Nil(t, timer)
	cfg.ReportErrors(nil, "foo")
}
