package mockhttp

import "net/http"

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

// UseResponseStatus sets the response, if presents, status code
// to the given status
func UseResponseStatus(status int) Middleware {
	return MiddlewareFunc(func(inner http.RoundTripper) http.RoundTripper {
		return RoundTripperFunc(func(r *http.Request) (resp *http.Response, err error) {
			if resp, err = inner.RoundTrip(r); err != nil {
				return
			}
			resp.StatusCode = status
			resp.Status = http.StatusText(status)
			return
		})
	})
}

// UseResponseSetHeader sets the response, if presents, header
// with given key-value pair
func UseResponseSetHeader(key, value string) Middleware {
	return MiddlewareFunc(func(inner http.RoundTripper) http.RoundTripper {
		return RoundTripperFunc(func(r *http.Request) (resp *http.Response, err error) {
			if resp, err = inner.RoundTrip(r); err != nil {
				return
			}
			resp.Header.Set(key, value)
			return
		})
	})
}

// UseResponseAddHeader adds the response, if presents, header
// with given key-value pair
func UseResponseAddHeader(key, value string) Middleware {
	return MiddlewareFunc(func(inner http.RoundTripper) http.RoundTripper {
		return RoundTripperFunc(func(r *http.Request) (resp *http.Response, err error) {
			if resp, err = inner.RoundTrip(r); err != nil {
				return
			}
			resp.Header.Add(key, value)
			return
		})
	})
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
