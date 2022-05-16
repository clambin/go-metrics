package caller

import (
	"net/http"
)

// Caller interface of a generic API caller
type Caller interface {
	Do(req *http.Request) (resp *http.Response, err error)
}

// BaseClient performs the actual HTTP request
type BaseClient struct {
	HTTPClient *http.Client
}

var _ Caller = &BaseClient{}

// Do performs the actual HTTP request
func (b BaseClient) Do(req *http.Request) (resp *http.Response, err error) {
	return b.HTTPClient.Do(req)
}
