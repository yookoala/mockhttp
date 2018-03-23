package mockhttp

import (
	"fmt"
	"net/http"
)

// RoundTripperFunc is a simplified way to implement http.RoundTripper
type RoundTripperFunc func(*http.Request) (*http.Response, error)

// RoundTrip implements http.RoundTripper
func (rt RoundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return rt(r)
}

// MuxRoundTripper mux http.RoundTripper by the request's URL.Host field
type MuxRoundTripper map[string]http.RoundTripper

// Add an http.RoundTripper to the mux with reference to the host.
// Please note that a fallback http.RoundTripper can be set with
// host = "*"
func (mux MuxRoundTripper) Add(host string, rt http.RoundTripper) {
	mux[host] = rt
}

// AddFunc add an RoundTripperFunc to the mux with reference to the host
// Please note that a fallback http.RoundTripper can be set with
// host = "*"
func (mux MuxRoundTripper) AddFunc(host string, fn RoundTripperFunc) {
	mux[host] = fn
}

// Get the http.RoundTripper for the given host
func (mux MuxRoundTripper) Get(host string) (http.RoundTripper, error) {
	if rt, found := mux[host]; found {
		return rt, nil // RoundTripper with match host
	}
	if rt, found := mux["*"]; found {
		return rt, nil // fallback RoundTripper
	}
	return nil, fmt.Errorf("no http.RoundTripper found for host %s",
		host)
}

// RoundTrip implements http.RoundTripper
func (mux MuxRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	rt, err := mux.Get(r.URL.Host)
	if err != nil {
		return nil, err
	}
	return rt.RoundTrip(r)
}

// NewClient returns a new http.Client with the mux as transport
func (mux MuxRoundTripper) NewClient() *http.Client {
	return &http.Client{
		Transport: mux,
	}
}
