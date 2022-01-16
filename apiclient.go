package metrics

import "github.com/prometheus/client_golang/prometheus"

// APIClientMetrics contains Prometheus metrics to capture during API calls.
type APIClientMetrics struct {
	Latency *prometheus.SummaryVec // measures latency of an API call
	Errors  *prometheus.CounterVec // measures any errors returned by an API call
}

// ReportErrors measures any API client call failures:
//
//	err := callAPI(server, endpoint)
//	pm.ReportErrors(err)
func (pm *APIClientMetrics) ReportErrors(err error, labelValues ...string) {
	if pm == nil || pm.Errors == nil {
		return
	}

	var value float64
	if err != nil {
		value = 1.0
	}
	pm.Errors.WithLabelValues(labelValues...).Add(value)
}

// MakeLatencyTimer creates a prometheus.Timer to measure the duration (latency) of an API client call
// If no Latency metric was created, timer will be nil:
//
//	timer := pm.MakeLatencyTimer(server, endpoint)
//	callAPI(server, endpoint)
//	if timer != nil {
//		timer.ObserveDuration()
//	}
func (pm *APIClientMetrics) MakeLatencyTimer(labelValues ...string) (timer *prometheus.Timer) {
	if pm != nil && pm.Latency != nil {
		timer = prometheus.NewTimer(pm.Latency.WithLabelValues(labelValues...))
	}
	return
}
