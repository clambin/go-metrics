package server_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/clambin/go-metrics/server"
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
	s := server.New(0)

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
	s := server.NewWithHandlers(0, []server.Handler{{
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

	_, err = httpGet(fmt.Sprintf("http://127.0.0.1:%d/hello", s.Port))
	require.Error(t, err)
	require.Equal(t, "405 Method Not Allowed", err.Error())

	err = s.Shutdown(30 * time.Second)
	require.NoError(t, err)

	wg.Wait()
}

func TestServer_Metrics(t *testing.T) {
	s := server.NewWithHandlers(0, []server.Handler{
		{
			Path: "/hello",
			Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				// don't call WriteHeader to test that middleware metrics defaults to HTTP 200
				_, _ = w.Write([]byte("hello!"))
			}),
			Methods: []string{},
		},
		{
			Path: "/hello2",
			Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte("hello2!"))
			}),
			Methods: []string{http.MethodPost},
		},
	})

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err := s.Run()
		require.True(t, errors.Is(err, http.ErrServerClosed))
		wg.Done()
	}()

	body, err := httpGet(fmt.Sprintf("http://127.0.0.1:%d/hello", s.Port))
	require.NoError(t, err)
	assert.Equal(t, "hello!", body)

	body, err = httpPost(fmt.Sprintf("http://127.0.0.1:%d/hello2", s.Port), "", nil)
	require.NoError(t, err)
	assert.Equal(t, "hello2!", body)

	body, err = httpGet(fmt.Sprintf("http://127.0.0.1:%d/metrics", s.Port))
	require.NoError(t, err)
	assert.Contains(t, body, `
http_duration_seconds_count{method="GET",path="/hello",status_code="200"} `)
	assert.Contains(t, body, `
http_duration_seconds_count{method="POST",path="/hello2",status_code="201"} `)
}

func TestServer_Panics(t *testing.T) {
	s := server.New(0)
	assert.Panics(t, func() { _ = server.New(s.Port) })
}

func TestGetRouter(t *testing.T) {
	listener, err := net.Listen("tcp4", ":0")
	require.NoError(t, err)

	r := server.GetRouter()
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

	target := "http://" + listener.Addr().String()

	var body string
	body, err = httpGet(target + "/hello")
	require.NoError(t, err)
	assert.Equal(t, "hello!", body)

	body, err = httpGet(target + "/metrics")
	require.NoError(t, err)
	assert.Contains(t, body, `
http_duration_seconds_sum{method="GET",path="/hello",status_code="200"} `)

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
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
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
