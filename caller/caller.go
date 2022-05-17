package caller

import (
	"net/http"
	"sync"
)

// Caller interface of a generic API caller
type Caller interface {
	Do(req *http.Request) (resp *http.Response, err error)
}

// BaseClient performs the actual HTTP request
type BaseClient struct {
	HTTPClient *http.Client
	lock       sync.Mutex
}

var _ Caller = &BaseClient{}

// Do performs the actual HTTP request
func (b *BaseClient) Do(req *http.Request) (resp *http.Response, err error) {
	b.lock.Lock()
	if b.HTTPClient == nil {
		b.HTTPClient = http.DefaultClient
	}
	b.lock.Unlock()
	return b.HTTPClient.Do(req)
}
