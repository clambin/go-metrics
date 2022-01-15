// Package metrics provides utilities for running and testing a prometheus scrape server.
//
// Server provides an HTTP server with a Prometheus metrics endpoint configured. Running it is as simple as:
//
//		func main() {
//			...
//			initialize your application
//			...
//			server := metrics.NewServer(8080)
//			server.Run()
//		}
//
// This will start an HTTP server on port 8080, with a /metrics endpoint for Prometheus scraping.  To add
// a metric endpoint to an existing http.Server, use the GetRouter() function instead:
//
// 		r := metrics.GetRouter()
//		r.Path("/hello").Handler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
//			_, _ = w.Write([]byte("hello!"))
//		}))
//		s := &http.Server{Handler: r}
// 		listener, _  := net.Listen("tcp", ":8080")
// 		err := s.Serve(listener)
//
// APIClientMetrics provides a standard way of measuring client metrics of API calls. Currently, latency &
// errors are supported:
//
// 		cfg = metrics.APIClientMetrics{
//			Latency: promauto.NewSummaryVec(prometheus.SummaryOpts{
//				Name: "latency_metrics",
//				Help: "Latency metric",
//			}, []string{}),
//			Errors: promauto.NewCounterVec(prometheus.CounterOpts{
//				Name: "error_metric",
//				Help: "Error metric",
//			}, []string{}),
//		}
//		timer := cfg.MakeLatencyTimer()
// 		err := apiCall()
// 		timer.ObserveDuration() // observe latency
//		cfg.ReportErrors(err)   // report any errors
//
// MetricName, MetricLabel and MetricValue provide an easy way to validate metrics during unit testing:
//
//		func TestMetricTools(t *testing.T) {
//			...
//			ch := make(chan prometheus.Metric)
//			go yourCollector.Collect(ch)
//
//			m := <-ch
//			assert.Equal(t, "foo_bar_gauge", metrics.MetricName(m))
//			assert.Equal(t, 1.0, metrics.MetricValue(m).GetGauge().GetValue())
//			assert.Equal(t, "valueA", metrics.MetricLabel(m, "labelA"))
//		}
//
package metrics
