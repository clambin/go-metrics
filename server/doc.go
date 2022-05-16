/*
Package server provides utilities for running and testing a prometheus scrape server.


Server provides an HTTP server with a Prometheus metrics endpoint configured. Running it is as simple as:

	server := metrics.New(8080)
	server.Run()

This will start an HTTP server on port 8080, with a /metrics endpoint for Prometheus scraping.

For HTTP servers, you may add additional handlers by using NewWithHandlers:

	server := metrics.NewWithHandlers(8080, []metrics.Handler{
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

*/
package server
