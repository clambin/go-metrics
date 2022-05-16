package server

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net"
	"net/http"
	"strconv"
	"time"
)

// Server runs an HTTP Server for a Prometheus scrape (i.e. /metric) endpoint. It includes a "http_duration_seconds"
// Counter metric that measures the time of each HTTP server request.
type Server struct {
	// the Port that the HTTP server listens on
	Port     int
	listener net.Listener
	server   http.Server
}

// New creates a new Server, which will listen on the specified TCP port. If Port is zero, Server will listen on
// a randomly chosen free port.  The selected can be found in Server's Port field.
func New(port int) (server *Server) {
	return NewWithHandlers(port, []Handler{})
}

// Handler contains an endpoint to be registered in the Server's HTTP server, using NewWithHandlers.
type Handler struct {
	// Path of the endpoint (e.g. "/health"). Must include the leading /
	Path string
	// Handler that implements the endpoint
	Handler http.Handler
	// Methods that the handler should support. If empty, http.MethodGet is the default
	Methods []string
}

// NewWithHandlers creates a new Server with additional handlers. If Port is zero, Server will listen on
// a randomly chosen free port.  The selected can be found in Server's Port field.
func NewWithHandlers(port int, handlers []Handler) *Server {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic("unable to create prometheus metrics server")
	}

	if port == 0 {
		port = listener.Addr().(*net.TCPAddr).Port
	}

	r := GetRouter()
	for _, handler := range handlers {
		methods := handler.Methods
		if handler.Methods == nil || len(handler.Methods) == 0 {
			methods = []string{http.MethodGet}
		}
		r.Path(handler.Path).Handler(handler.Handler).Methods(methods...)
	}

	return &Server{
		Port:     port,
		listener: listener,
		server:   http.Server{Handler: r},
	}
}

// GetRouter returns an HTTP router with a prometheus metrics endpoint. Use this if you do not want to use Server.Run(),
// but instead prefer to incorporate the metrics endpoint into your application's existing HTTP server:
//
//		r := metrics.GetRouter()
//		r.Path("/some-endpoint").Handler(someHandler)
//		server := http.Server{
//			Addr: ":8080",
//
func GetRouter() (router *mux.Router) {
	router = mux.NewRouter()
	router.Use(prometheusMiddleware)
	router.Path("/metrics").Handler(promhttp.Handler())
	return
}

// Run starts the HTTP Server. This calls server's http.Server's Serve method and returns that method's return value.
func (server *Server) Run() (err error) {
	return server.server.Serve(server.listener)
}

// Shutdown performs a graceful shutdown of the HTTP Server.
func (server *Server) Shutdown(timeout time.Duration) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return server.server.Shutdown(ctx)
}

// Prometheus metrics
var (
	httpDuration = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name: "http_duration_seconds",
		Help: "Duration of HTTP requests",
	}, []string{"path", "method", "status_code"})
)

// prometheusMiddleware measures the time it takes to perform a /metric call.
func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()
		lrw := newLoggingResponseWriter(w)
		start := time.Now()
		next.ServeHTTP(lrw, r)
		httpDuration.WithLabelValues(path, r.Method, strconv.Itoa(lrw.statusCode)).Observe(time.Since(start).Seconds())
	})
}

// loggingResponseWriter records the HTTP status code of a ResponseWriter, so we can use it to log response times for
// individual status codes.
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// newLoggingResponseWriter creates a new loggingResponseWriter.
func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK, // if the handler doesn't call WriteHeader(), default to HTTP 200
	}
}

// WriteHeader implements the http.ResponseWriter interface.
func (w *loggingResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// Write implements the http.ResponseWriter interface.
func (w *loggingResponseWriter) Write(body []byte) (int, error) {
	return w.ResponseWriter.Write(body)
}
