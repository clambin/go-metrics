package metrics_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/clambin/go-metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	s := metrics.NewServer(0)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err := s.Run()
		require.True(t, errors.Is(err, http.ErrServerClosed))
		wg.Done()
	}()

	body, err := httpGet(fmt.Sprintf("http://127.0.0.1:%d/metrics", s.Port))
	require.NoError(t, err)
	assert.Contains(t, body, `
# HELP`)

	err = s.Shutdown(30 * time.Second)
	require.NoError(t, err)

	wg.Wait()
}

func TestNewServerWithHandlers(t *testing.T) {
	s := metrics.NewServerWithHandlers(0, []metrics.Handler{{
		Path: "/hello",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req.Method == "POST" {
				_, _ = w.Write([]byte("hello!"))
				return
			}
			http.Error(w, "only POST is allowed", http.StatusBadRequest)
		}),
		Methods: []string{http.MethodPost},
	}})

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err := s.Run()
		require.True(t, errors.Is(err, http.ErrServerClosed))
		wg.Done()
	}()

	body, err := httpPost(fmt.Sprintf("http://127.0.0.1:%d/hello", s.Port), "", nil)
	require.NoError(t, err)
	assert.Equal(t, "hello!", body)

	body, err = httpGet(fmt.Sprintf("http://127.0.0.1:%d/hello", s.Port))
	require.Error(t, err)
	require.Equal(t, "405 Method Not Allowed", err.Error())

	err = s.Shutdown(30 * time.Second)
	require.NoError(t, err)

	wg.Wait()
}

func TestServer_Metrics(t *testing.T) {
	s := metrics.NewServerWithHandlers(0, []metrics.Handler{{
		Path: "/hello",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req.Method == "POST" {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("hello!"))
				return
			}
			http.Error(w, "only POST is allowed", http.StatusBadRequest)
		}),
		Methods: []string{http.MethodPost},
	}})

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err := s.Run()
		require.True(t, errors.Is(err, http.ErrServerClosed))
		wg.Done()
	}()

	body, err := httpPost(fmt.Sprintf("http://127.0.0.1:%d/hello", s.Port), "", nil)
	require.NoError(t, err)
	assert.Equal(t, "hello!", body)

	body, err = httpGet(fmt.Sprintf("http://127.0.0.1:%d/metrics", s.Port))
	require.NoError(t, err)
	assert.Contains(t, body, `
http_duration_seconds_count{method="POST",path="/hello",status_code="200"} `)
}

func TestServer_Panics(t *testing.T) {
	s := metrics.NewServer(0)
	assert.Panics(t, func() { _ = metrics.NewServer(s.Port) })
}

func TestGetRouter(t *testing.T) {
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	r := metrics.GetRouter()
	r.Path("/hello").Handler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("hello!"))
	}))

	s := &http.Server{Handler: r}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err2 := s.Serve(listener)
		require.True(t, errors.Is(err2, http.ErrServerClosed))
		wg.Done()
	}()

	var body string
	body, err = httpGet("http://" + listener.Addr().String() + "/hello")
	require.NoError(t, err)
	assert.Equal(t, "hello!", body)

	body, err = httpGet("http://" + listener.Addr().String() + "/metrics")
	require.NoError(t, err)
	assert.Contains(t, body, `
http_duration_seconds_sum{method="GET",path="/metrics",status_code="200"} `)

	err = s.Shutdown(context.Background())
	require.NoError(t, err)
	wg.Wait()
}

func httpGet(url string) (response string, err error) {
	var resp *http.Response
	if resp, err = http.Get(url); err == nil {
		return httpParseRequest(resp)
	}
	return
}

func httpPost(url string, contentType string, requestBody io.Reader) (response string, err error) {
	var resp *http.Response
	if resp, err = http.Post(url, contentType, requestBody); err == nil {
		return httpParseRequest(resp)
	}
	return
}

func httpParseRequest(resp *http.Response) (response string, err error) {
	if resp.StatusCode == http.StatusOK {
		var body []byte
		if body, err = io.ReadAll(resp.Body); err == nil {
			response = string(body)
		}
	} else {
		err = errors.New(resp.Status)
	}
	_ = resp.Body.Close()
	return
}