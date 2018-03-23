package mockhttp

import (
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// HandlerRT directly inject an http.Handler to the
// RoundTripper. You may directly test your handler
// in this mocked transport.
func HandlerRT(handler http.Handler) RoundTripperFunc {
	return func(r *http.Request) (resp *http.Response, err error) {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		resp = &http.Response{
			Status:        http.StatusText(w.Code),
			StatusCode:    w.Code,
			Proto:         r.Proto,
			ProtoMajor:    r.ProtoMajor,
			ProtoMinor:    r.ProtoMinor,
			ContentLength: int64(w.Body.Len()),
			Request:       r,
			Header:        w.Header(),
			Body:          ioutil.NopCloser(w.Body),
		}
		return
	}
}

// ServerErrorRT always return a server response of
// the supplied status code with nil error.
func ServerErrorRT(status int) RoundTripperFunc {
	statusText := http.StatusText(status)
	size := int64(len(statusText))

	return func(r *http.Request) (resp *http.Response, err error) {

		// mock header
		header := make(http.Header)
		header.Add("Content-Length", fmt.Sprintf("%d", size))
		header.Add("Content-Type", "text/html")
		header.Add("Date", time.Now().Format(time.RFC1123))

		// mock response
		resp = &http.Response{
			Status:        statusText,
			StatusCode:    status,
			Proto:         r.Proto,
			ProtoMajor:    r.ProtoMajor,
			ProtoMinor:    r.ProtoMinor,
			ContentLength: size,
			Request:       r,
			Header:        header,
			Body:          ioutil.NopCloser(strings.NewReader(statusText)),
		}
		return
	}
}

// TransportErrorRT always return nil server response
// wtih the supplied error
func TransportErrorRT(err error) RoundTripperFunc {
	return func(r *http.Request) (resp *http.Response, err error) {
		return nil, err
	}
}

// StaticResponseRT returns an http.RoundTripper that
// returns the same response body no matter what the
// request is.
func StaticResponseRT(content, contentType string) RoundTripperFunc {
	size := int64(len(content))
	now := time.Now().Format(time.RFC1123)
	return func(r *http.Request) (resp *http.Response, err error) {

		// mock header
		header := make(http.Header)
		header.Add("Content-Length", fmt.Sprintf("%d", size))
		header.Add("Content-Type", contentType)
		header.Add("Date", now)

		// mock response
		resp = &http.Response{
			Status:        http.StatusText(http.StatusOK),
			StatusCode:    http.StatusOK,
			Proto:         r.Proto,
			ProtoMajor:    r.ProtoMajor,
			ProtoMinor:    r.ProtoMinor,
			ContentLength: size,
			Request:       r,
			Header:        header,
			Body:          ioutil.NopCloser(strings.NewReader(content)),
		}
		return
	}
}

// FileSystemRT implements http.RoundTripper by returning
// contents of files in a given folder (as defined as `root`).
func FileSystemRT(root string) RoundTripperFunc {
	return func(r *http.Request) (resp *http.Response, err error) {

		path := filepath.Join(root, r.URL.Path)
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("error openning path: %s",
				err)
		}

		s, err := f.Stat()
		if err != nil {
			return nil, fmt.Errorf("error getting file stat: %s",
				err)
		}

		// detect content type by extension
		contentType := mime.TypeByExtension(filepath.Ext(path))

		// mock header
		header := make(http.Header)
		header.Add("Content-Length", fmt.Sprintf("%d", s.Size()))
		header.Add("Content-Type", contentType)
		header.Add("Date", s.ModTime().Format(time.RFC1123))

		// mock response
		resp = &http.Response{
			Status:        http.StatusText(http.StatusOK),
			StatusCode:    http.StatusOK,
			Proto:         r.Proto,
			ProtoMajor:    r.ProtoMajor,
			ProtoMinor:    r.ProtoMinor,
			ContentLength: s.Size(),
			Request:       r,
			Header:        header,
			Body:          f,
		}
		return
	}
}
