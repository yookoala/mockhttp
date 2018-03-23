package mockhttp_test

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

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