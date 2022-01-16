/*
Package metrics provides utilities for running and testing a prometheus scrape server.


Server provides an HTTP server with a Prometheus metrics endpoint configured. Running it is as simple as:

	server := metrics.NewServer(8080)
	server.Run()

This will start an HTTP server on port 8080, with a /metrics endpoint for Prometheus scraping.

For HTTP servers, you may add additional handlers by using NewServerWithHandlers:

	server := metrics.NewServerWithHandlers(8080, []metrics.Handler{
		{
			Path: "/hello",
			Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.Method == "POST" {
					_, _ = w.Write([]byte("hello!"))
					return
				}
				http.Error(w, "only POST is allowed", http.StatusBadRequest)
			}),
			Methods: []string{http.MethodPost},
		},
	})
	server.Run()

If you need to build your own HTTP server, you can use GetRouter() instead:

	r := metrics.GetRouter()
	r.Path("/hello").Handler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("hello!"))
	}))

	server := &http.Server{
		Handler: r,
		Addr:    ":8080",
	}
	server.ListenAndServe()

APIClientMetrics provides a standard way of measuring client metrics when performing API calls.
Currently, latency and errors are supported:

	cfg = metrics.APIClientMetrics{
		Latency: promauto.NewSummaryVec(prometheus.SummaryOpts{
			Name: "latency_metrics",
			Help: "Latency metric",
		}, []string{}),
		Errors: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "error_metric",
			Help: "Error metric",
		}, []string{}),
	}
	timer := cfg.MakeLatencyTimer()
	err := apiCall()
	if timer != nil {
		timer.ObserveDuration() // observe latency
	}
	cfg.ReportErrors(err)       // report any errors


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
package metrics
