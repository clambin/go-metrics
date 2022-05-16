package tools_test

import (
	"github.com/clambin/go-metrics/tools"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"testing"
)

type Collector struct {
	gauge   *prometheus.Desc
	counter *prometheus.Desc
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(c.gauge, prometheus.GaugeValue, 1.0, "valueA", "valueB")
	ch <- prometheus.MustNewConstMetric(c.counter, prometheus.CounterValue, 2.0, "valueA", "valueB")
}

func TestMetricTools(t *testing.T) {
	c := Collector{
		gauge: prometheus.NewDesc(
			prometheus.BuildFQName("foo", "bar", "gauge"),
			"Dummy metric",
			[]string{"labelA", "labelB"},
			prometheus.Labels{"labelC": "valueC"},
		),
		counter: prometheus.NewDesc(
			prometheus.BuildFQName("foo", "bar", "counter"),
			"Dummy metric",
			[]string{"labelA", "labelB"},
			prometheus.Labels{"labelC": "valueC"},
		),
	}

	ch := make(chan prometheus.Metric)
	go c.Collect(ch)

	m := <-ch
	assert.Equal(t, "foo_bar_gauge", tools.MetricName(m))
	assert.Equal(t, 1.0, tools.MetricValue(m).GetGauge().GetValue())
	assert.Equal(t, "valueA", tools.MetricLabel(m, "labelA"))
	assert.Equal(t, "valueB", tools.MetricLabel(m, "labelB"))
	assert.Equal(t, "valueC", tools.MetricLabel(m, "labelC"))

	m = <-ch
	assert.Equal(t, "foo_bar_counter", tools.MetricName(m))
	assert.Equal(t, 2.0, tools.MetricValue(m).GetCounter().GetValue())
	assert.Equal(t, "valueA", tools.MetricLabel(m, "labelA"))
	assert.Equal(t, "valueB", tools.MetricLabel(m, "labelB"))
	assert.Equal(t, "valueC", tools.MetricLabel(m, "labelC"))
}
