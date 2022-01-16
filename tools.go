package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	pcg "github.com/prometheus/client_model/go"
	"regexp"
)

// MetricName returns the name of the provided Prometheus metric.
func MetricName(metric prometheus.Metric) (name string) {
	desc := metric.Desc().String()
	r := regexp.MustCompile(`fqName: "([a-z,_]+)"`)
	match := r.FindStringSubmatch(desc)
	if len(match) >= 2 {
		name = match[1]
	}
	return
}

// MetricValue converts a Prometheus metric to a prometheus/client_model/go metric, so the caller can read its value:
//
//		ch := make(chan prometheus.Metric)
//		go c.Collect(ch)
//
//		m := <-ch
//		metric := metrics.MetricValue(m)
//		assert.Equal(t, 1.0, metric.GetGauge().GetValue())
//
// Panics if metric is not a valid Prometheus metric.
func MetricValue(metric prometheus.Metric) *pcg.Metric {
	m := &pcg.Metric{}
	if metric.Write(m) != nil {
		panic("failed to parse metric")
	}
	return m
}

// MetricLabel returns the value of a metric's label. Panics if metric is not a valid Prometheus metric.
func MetricLabel(metric prometheus.Metric, labelName string) (value string) {
	m := MetricValue(metric)
	for _, label := range m.GetLabel() {
		if label.GetName() == labelName {
			value = label.GetValue()
			break
		}
	}
	return value
}
