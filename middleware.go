package mockhttp

import (
	"net/http"
)

// Middleware warps an http.RoundTripper
// and modify the input / output behaviour.
type Middleware interface {
	Wrap(http.RoundTripper) http.RoundTripper
}

// MiddlewareFunc turns a simple function value into
// Middleware implementation.
type MiddlewareFunc func(http.RoundTripper) http.RoundTripper

// Wrap implements Middleware
func (fn MiddlewareFunc) Wrap(inner http.RoundTripper) http.RoundTripper {
	return fn(inner)
}

// ResponseSetStatus sets the response, if presents, status code
// to the given status
func ResponseSetStatus(status int) ResponseModifier {
	return func(resp *http.Response, err error) (*http.Response, error) {
		if resp != nil {
			resp.StatusCode = status
			resp.Status = http.StatusText(status)
		}
		return resp, err
	}
}

// ResponseSetHeader sets the response, if presents, header
// with given key-value pair
func ResponseSetHeader(key, value string) ResponseModifier {
	return func(resp *http.Response, err error) (*http.Response, error) {
		if resp != nil {
			if resp.Header == nil {
				resp.Header = make(http.Header)
			}
			resp.Header.Set(key, value)
		}
		return resp, err
	}
}

// ResponseAddHeader adds the response, if presents, header
// with given key-value pair
func ResponseAddHeader(key, value string) ResponseModifier {
	return func(resp *http.Response, err error) (*http.Response, error) {
		if resp != nil {
			if resp.Header == nil {
				resp.Header = make(http.Header)
			}
			resp.Header.Add(key, value)
		}
		return resp, err
	}
}

// ResponseModifier implements Middleware by modifying http.Response and/or error output
// of inner http.RoundTripper output
type ResponseModifier func(resp *http.Response, err error) (*http.Response, error)

// Wrap implements Middleware
func (modifier ResponseModifier) Wrap(inner http.RoundTripper) http.RoundTripper {
	return RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return modifier(inner.RoundTrip(r))
	})
}

// UseResponseModifier converts a function that fulfills ResponseModifier
// signature into Middleware
func UseResponseModifier(modifier ResponseModifier) Middleware {
	return modifier
}

// Chain wraps the middlware, from outter-most to inner-most, into
// a combined middleware.
func Chain(middlewares ...Middleware) Middleware {
	return MiddlewareFunc(func(roundTripper http.RoundTripper) http.RoundTripper {
		for i := len(middlewares) - 1; i >= 0; i-- {
			roundTripper = middlewares[i].Wrap(roundTripper)
		}
		return roundTripper
	})
}
