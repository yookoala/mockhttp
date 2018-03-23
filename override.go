package mockhttp

import (
	"net/http"
	"sync"
)

var lastDefaultTransport http.RoundTripper
var transportLock *sync.RWMutex

func init() {
	transportLock = &sync.RWMutex{}
}

// UseTransport use the given roundtripper as the
// http.DefaultTransport, and store away
// the current http.DefaultTransport for restoration.
func UseTransport(rt http.RoundTripper) {
	transportLock.Lock() // prevent multiple use of UseTransport
	lastDefaultTransport, http.DefaultTransport = http.DefaultTransport, rt
}

// RestoreTransport restore the http.DefaultTransport
// to the one before last call of UseTransport().
func RestoreTransport() {
	http.DefaultTransport = lastDefaultTransport
	transportLock.Unlock()
}
