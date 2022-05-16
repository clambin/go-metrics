/*
Package caller provides a standard way of writing API clients. It's meant to be a drop-in replacement for an HTTPClient.
Currently, it supports a way of generating Prometheus metrics when performing API calls, and a means of caching API responses
for one or more endpoints.

InstrumentedClient generates Prometheus metrics when performing API calls. Currently, latency and errors are supported:

	cfg = caller.ClientMetrics{
		Latency: promauto.NewSummaryVec(prometheus.SummaryOpts{
			Name: "request_duration_seconds",
			Help: "Duration of API requests.",
		}, []string{"application", "request"}),
		Errors:  promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "request_errors",
			Help: "Duration of API requests.",
		}, []string{"application", "request"}),

	c := caller.InstrumentedClient{
		BaseClient: caller.BaseClient{HTTPClient: http.DefaultClient},
		Options: cfg,
		Application: "foo",
	}

	req, _ := http.NewRequest(http.MethodGet, url, nil)
	resp, err = c.Do(req)

This will generate Prometheus metrics for every request sent by the InstrumentedClient. The application label will be
as set by the InstrumentedClient object. The request will be set to the Path of the request.


Cacher caches responses to HTTP requests:

	c := caller.NewCacher(
		http.DefaultClient, "foo", caller.Options{},
		[]caller.CacheTableEntry{
			{Endpoint: "/foo"},
		},
		50*time.Millisecond, 0,
	)

This creates a Cacher that will cache the response of called to any request with Path '/foo', for up to 50 msec.

Note: NewCacher will create a Caller that also generates Prometheus metrics by chaining the request to an InstrumentedClient.
To avoid this, create the Cacher object directly:

	c := &Cacher{
		Caller: &BaseClient{HTTPClient: httpClient},
		Table: CacheTable{Table: cacheEntries},
		Cache: cache.New[string, []byte](cacheExpiry, cacheCleanup),

*/
package caller
