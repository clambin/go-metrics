/*
Package tools provides utilities for writing tests for code that produces Prometheus metrics.


MetricName, MetricLabel and MetricValue provide an easy way to validate metrics during unit testing:

	func TestMetrics(t *testing.T) {
		...
		ch := make(chan prometheus.Metric)
		go yourMetric.Collect(ch)

		m := <-ch
		assert.Equal(t, "foo_bar_gauge", metrics.MetricName(m))
		assert.Equal(t, 1.0, metrics.MetricValue(m).GetGauge().GetValue())
		assert.Equal(t, "valueA", metrics.MetricLabel(m, "labelA"))
	}

*/
package tools
