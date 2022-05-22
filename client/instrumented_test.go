package client_test

import (
	"encoding/json"
	"fmt"
	"github.com/clambin/go-metrics/client"
	"github.com/clambin/go-metrics/tools"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_Do(t *testing.T) {
	metrics := client.NewMetrics("foo", "bar")
	s := httptest.NewServer(http.HandlerFunc(handler))
	c := &client.InstrumentedClient{
		Options:     client.Options{PrometheusMetrics: metrics},
		Application: "foo",
	}

	response, err := doCall(c, s.URL+"/foo")
	require.NoError(t, err)
	assert.Equal(t, "bar", response.Name)
	assert.Equal(t, 42, response.Age)

	response, err = doCall(c, s.URL+"/bar")
	require.Error(t, err)

	s.Close()
	response, err = doCall(c, s.URL+"/foo")
	require.Error(t, err)

	ch := make(chan prometheus.Metric)
	go metrics.Latency.Collect(ch)

	expectedLatencyCounters := map[string]uint64{
		"/foo": 2,
		"/bar": 1,
	}

	for range expectedLatencyCounters {
		desc := <-ch
		endpoint := *tools.MetricValue(desc).GetLabel()[1].Value
		expected, ok := expectedLatencyCounters[endpoint]
		require.True(t, ok)
		assert.Equal(t, expected, tools.MetricValue(desc).GetSummary().GetSampleCount(), endpoint)
	}

	ch = make(chan prometheus.Metric)
	go metrics.Errors.Collect(ch)
	expectedErrorCounters := map[string]float64{
		"/foo": 1,
		"/bar": 0,
	}

	for range expectedErrorCounters {
		desc := <-ch
		endpoint := *tools.MetricValue(desc).GetLabel()[1].Value
		expected, ok := expectedErrorCounters[endpoint]
		require.True(t, ok)
		assert.Equal(t, expected, tools.MetricValue(desc).GetCounter().GetValue(), endpoint)

	}

}

type testStruct struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func doCall(c client.Caller, url string) (response testStruct, err error) {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	var resp *http.Response
	if resp, err = c.Do(req); err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		return response, fmt.Errorf("call failed: %s", resp.Status)
	}
	defer func() { _ = resp.Body.Close() }()

	err = json.NewDecoder(resp.Body).Decode(&response)
	return
}

func handler(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/foo" {
		http.Error(w, "invalid endpoint", http.StatusNotFound)
		return
	}

	_ = json.NewEncoder(w).Encode(testStruct{
		Name: "bar",
		Age:  42,
	})
}
