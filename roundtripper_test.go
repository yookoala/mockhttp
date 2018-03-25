package mockhttp_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/yookoala/mockhttp"
)

type mockRoundTripper int

func (mrt mockRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, nil
}

func TestMuxRoundTripper_Add(t *testing.T) {
	mux := mockhttp.MuxRoundTripper{}
	fn1 := mockRoundTripper(1)
	fn2 := mockRoundTripper(2)
	fn3 := mockRoundTripper(3)
	mux.Add("www.google.com", fn1)
	mux.Add("www.facebook.com", fn2)
	mux.Add("*", fn3)

	if have, err := mux.Get("www.google.com"); err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if have != fn1 {
		t.Errorf("expected fn1(%#v), got %#v", fn1, have)
	}
	if have, err := mux.Get("www.facebook.com"); err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if have != fn2 {
		t.Errorf("expected fn2(%#v), got %#v", fn2, have)
	}
	if have, err := mux.Get("www.foobar.com"); err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if have != fn3 {
		t.Errorf("expected fn3(%#v), got %#v", fn3, have)
	}
}

func TestMuxRoundTripper_AddFunc(t *testing.T) {
	mux := mockhttp.MuxRoundTripper{}
	mux.AddFunc("www.google.com", func(r *http.Request) (resp *http.Response, err error) {
		resp = &http.Response{
			Body: ioutil.NopCloser(strings.NewReader("This is a mocked google page")),
		}
		return
	})
	mux.AddFunc("www.facebook.com", func(r *http.Request) (resp *http.Response, err error) {
		resp = &http.Response{
			Body: ioutil.NopCloser(strings.NewReader("This is a mocked facebook page")),
		}
		return
	})

	tests := []struct {
		url string
		res string
	}{
		{
			url: "http://www.google.com/helloworld.html",
			res: "This is a mocked google page",
		},
		{
			url: "https://www.facebook.com/helloworld.html",
			res: "This is a mocked facebook page",
		}}

	for _, test := range tests {
		r, err := http.NewRequest("GET", test.url, nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		resp, err := mux.RoundTrip(r)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		c, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if want, have := test.res, string(c); want != have {
			t.Errorf("expected: %#v, got %#v", want, have)
		}
	}
}

func TestMuxRoundTripper_Get_error(t *testing.T) {
	mux := mockhttp.MuxRoundTripper{}
	mux.AddFunc("www.google.com", func(r *http.Request) (resp *http.Response, err error) {
		resp = &http.Response{
			Body: ioutil.NopCloser(strings.NewReader("This is a mocked google page")),
		}
		return
	})
	mux.AddFunc("www.facebook.com", func(r *http.Request) (resp *http.Response, err error) {
		resp = &http.Response{
			Body: ioutil.NopCloser(strings.NewReader("This is a mocked facebook page")),
		}
		return
	})

	rt, err := mux.Get("www.foobar.com")
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if want, have := "no http.RoundTripper found for host www.foobar.com", err.Error(); want != have {
		t.Errorf("expected %#v, got %#v", want, have)
	}
	if rt != nil {
		t.Errorf("expected nil, got %#v", rt)
	}

	req, err := http.NewRequest("GET", "http://www.foobar.com/helloworld", nil)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	resp, err := mux.RoundTrip(req)
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if want, have := "no http.RoundTripper found for host www.foobar.com", err.Error(); want != have {
		t.Errorf("expected %#v, got %#v", want, have)
	}
	if resp != nil {
		t.Errorf("expected nil, got %#v", rt)
	}

}

func TestMuxRoundTripper_NewClient(t *testing.T) {
	mux := mockhttp.MuxRoundTripper{}
	mux.AddFunc("www.google.com", func(r *http.Request) (resp *http.Response, err error) {
		resp = &http.Response{
			Body: ioutil.NopCloser(strings.NewReader("This is a mocked google page")),
		}
		return
	})

	r, err := mux.NewClient().Get("https://www.google.com")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	if r == nil {
		t.Errorf("expected response, got nil")
		return
	}

	c, err := ioutil.ReadAll(r.Body)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	if want, have := "This is a mocked google page", string(c); want != have {
		t.Errorf("expected: %#v, got %#v", want, have)
	}
}

func mustTimePtr(t time.Time, err error) *time.Time {
	if err != nil {
		panic(err)
	}
	return &t
}

func ExampleMuxRoundTripper() {

	// a mock transport layer for testing your API calls / resource download
	mock := mockhttp.NewMuxRoundTripper()
	mock.Add("api.service1.com", mockhttp.StaticResponseRT(`{"status": "OK", "cool": true}`, "application/json"))
	mock.Add("api.service2.com", mockhttp.StaticResponseRT(`{"status": "OK", "cool": true}`, "application/json"))
	mock.Add("api.service3.com", mockhttp.ServerErrorRT(http.StatusInternalServerError))
	mock.Add("mycdn-service.com", mockhttp.FileSystemRT("./testdata"))
	mock.Add("*", mockhttp.TransportErrorRT(fmt.Errorf("no network"))) // as fallback

	// simply produces http.Client with the MuxRoundTripper as transport
	client := mock.NewClient()

	var resp *http.Response
	var body []byte
	var err error

	// ordinary http.NewRequest routine
	req, _ := http.NewRequest("GET", "https://api.service1.com/some/endpoint", nil)
	resp, _ = client.Do(req)
	body, _ = ioutil.ReadAll(resp.Body)
	fmt.Printf("result 1 - %s\n", body)

	// directly use http.Client methods
	resp, _ = client.Get("https://api.service1.com/some/endpoint")
	body, _ = ioutil.ReadAll(resp.Body)
	fmt.Printf("result 2 - %s\n", body)

	// POST with http.Client method
	resp, _ = client.Post("https://api.service2.com/some/endpoint", "application/json", strings.NewReader(`{"submit": true}`))
	body, _ = ioutil.ReadAll(resp.Body)
	fmt.Printf("result 3 - %s\n", body)

	// POST to an endpoint with mock server error
	resp, _ = client.Post("https://api.service3.com/some/endpoint", "application/json", strings.NewReader(`{"submit": true}`))
	body, _ = ioutil.ReadAll(resp.Body)
	fmt.Printf("result 4 - %s\n", body)

	// GET a text file, in local file system, with content "hello world"
	resp, _ = client.Get("https://mycdn-service.com/test.txt")
	body, _ = ioutil.ReadAll(resp.Body)
	fmt.Printf("result 5 - %s\n", body)

	// GET an external service with mock transport,
	// which will be handled by "*" host handler.
	_, err = client.Get("https://api.google.com/some/endpoint")
	fmt.Printf("result 6 - %s\n", err.Error())

	// Output:
	// result 1 - {"status": "OK", "cool": true}
	// result 2 - {"status": "OK", "cool": true}
	// result 3 - {"status": "OK", "cool": true}
	// result 4 - Internal Server Error
	// result 5 - hello world
	// result 6 - Get https://api.google.com/some/endpoint: no network
}
