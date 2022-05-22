package client

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestCacheTable_ShouldCache(t *testing.T) {
	table := CacheTable{Table: []CacheTableEntry{
		{Endpoint: `/foo`},
		{Endpoint: `/foo/[\d+]`, IsRegExp: true},
		{Endpoint: `/bar`, Methods: []string{http.MethodGet}},
	}}

	type testcase struct {
		input  *http.Request
		expiry time.Duration
		match  bool
	}
	for _, tc := range []testcase{
		{input: &http.Request{URL: &url.URL{Path: "/foo"}}, match: true},
		{input: &http.Request{URL: &url.URL{Path: "/foo/123"}}, match: true},
		{input: &http.Request{URL: &url.URL{Path: "/foo/bar"}}, match: false},
		{input: &http.Request{URL: &url.URL{Path: "/bar"}, Method: http.MethodGet}, match: true},
		{input: &http.Request{URL: &url.URL{Path: "/bar"}, Method: http.MethodPost}, match: false},
		{input: &http.Request{URL: &url.URL{Path: "/foobar"}}, match: false},
	} {
		found, expiry := table.shouldCache(tc.input)
		assert.Equal(t, tc.match, found, tc.input)
		assert.Equal(t, tc.expiry, expiry, tc.input)

	}

	assert.True(t, table.compiled)
	for _, entry := range table.Table {
		if entry.IsRegExp {
			assert.NotNil(t, entry.compiledRegExp)
		} else {
			assert.Nil(t, entry.compiledRegExp)
		}
	}
}

func TestCacheTable_CacheEverything(t *testing.T) {
	table := CacheTable{}

	found, _ := table.shouldCache(&http.Request{URL: &url.URL{Path: "/"}})
	assert.True(t, found)
}

func TestCacheTable_Invalid_Input(t *testing.T) {
	table := CacheTable{Table: []CacheTableEntry{
		{Endpoint: `/foo/[\d+`, IsRegExp: true},
	}}

	assert.Panics(t, func() { table.shouldCache(&http.Request{URL: &url.URL{Path: "/foo"}}) })
}
