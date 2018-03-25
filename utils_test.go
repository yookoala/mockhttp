package mockhttp_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/yookoala/mockhttp"
)

func TestHandlerRT(t *testing.T) {

	var resp *http.Response
	var err error
	var content []byte

	// some http.Handler you want to test with
	handler := http.NewServeMux()
	handler.HandleFunc("/item/1", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "content of item 1")
	})
	handler.HandleFunc("/item/2", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "content of item 2")
	})
	handler.HandleFunc("/item/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, http.StatusText(http.StatusForbidden))
	})

	client := &http.Client{
		Transport: mockhttp.HandlerRT(handler),
	}

	tests := []struct {
		status  int
		url     string
		content string
	}{
		{
			status:  http.StatusOK,
			url:     "https://something.com/item/1",
			content: "content of item 1",
		},
		{
			status:  http.StatusOK,
			url:     "https://something.com/item/2",
			content: "content of item 2",
		},
		{
			status:  http.StatusForbidden,
			url:     "https://something.com/item/3",
			content: http.StatusText(http.StatusForbidden),
		},
	}

	for i, test := range tests {
		if resp, err = client.Get(test.url); err != nil {
			t.Errorf("[%d] unexpected error: %s", i, err)
			continue
		}
		if want, have := test.status, resp.StatusCode; want != have {
			t.Errorf("[%d] expected %#v, got %#v", i, want, have)
		}
		if content, err = ioutil.ReadAll(resp.Body); err != nil {
			t.Errorf("[%d] unexpected error: %s", i, err)
			continue
		}
		if want, have := test.content, string(content); want != have {
			t.Errorf("[%d] expected %#v, got %#v", i, want, have)
		}
	}
}

func ExampleHandlerRT() {

	// some http.Handler you want to test with
	handler := http.NewServeMux()
	handler.HandleFunc("/item/1", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "content of item 1")
	})
	handler.HandleFunc("/item/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, http.StatusText(http.StatusForbidden))
	})

	client := &http.Client{
		Transport: mockhttp.HandlerRT(handler),
	}

	resp, _ := client.Get("https://something.com/item/1")
	content, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("result 1: %s\n", content)

	resp, _ = client.Get("https://something.com/item/2")
	content, _ = ioutil.ReadAll(resp.Body)
	fmt.Printf("result 2: %s\n", content)

	// Output:
	// result 1: content of item 1
	// result 2: Forbidden
}

func TestServerErrorRT(t *testing.T) {

	var resp *http.Response
	var err error
	var content []byte

	mock := mockhttp.NewMuxRoundTripper()
	mock.Add("google.com", mockhttp.ServerErrorRT(http.StatusForbidden))
	mock.Add("facebook.com", mockhttp.ServerErrorRT(http.StatusBadGateway))
	mock.Add("yahoo.com", mockhttp.ServerErrorRT(http.StatusBadRequest))
	client := mock.NewClient()

	tests := []struct {
		status  int
		url     string
		content string
	}{
		{
			status:  http.StatusForbidden,
			url:     "https://google.com/mail",
			content: http.StatusText(http.StatusForbidden),
		},
		{
			status:  http.StatusBadGateway,
			url:     "https://facebook.com/user",
			content: http.StatusText(http.StatusBadGateway),
		},
		{
			status:  http.StatusBadRequest,
			url:     "https://yahoo.com/group",
			content: http.StatusText(http.StatusBadRequest),
		},
	}

	for i, test := range tests {
		if resp, err = client.Get(test.url); err != nil {
			t.Errorf("[%d] unexpected error: %s", i, err)
			continue
		}
		if want, have := test.status, resp.StatusCode; want != have {
			t.Errorf("[%d] expected %#v, got %#v", i, want, have)
		}
		if content, err = ioutil.ReadAll(resp.Body); err != nil {
			t.Errorf("[%d] unexpected error: %s", i, err)
			continue
		}
		if want, have := test.content, string(content); want != have {
			t.Errorf("[%d] expected %#v, got %#v", i, want, have)
		}
	}
}

func ExampleServerErrorRT() {

	client := &http.Client{
		Transport: mockhttp.ServerErrorRT(http.StatusForbidden),
	}

	resp, _ := client.Get("https://www.google.com")
	content, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("result 1: %s\n", content)

	resp, _ = client.Get("https://www.facebook.com")
	content, _ = ioutil.ReadAll(resp.Body)
	fmt.Printf("result 2: %s\n", content)

	// Output:
	// result 1: Forbidden
	// result 2: Forbidden
}

func TestTransportErrorRT(t *testing.T) {

	var resp *http.Response
	var err error
	var content []byte

	mock := mockhttp.NewMuxRoundTripper()
	mock.Add("google.com", mockhttp.TransportErrorRT(fmt.Errorf("domain not found")))
	mock.Add("facebook.com", mockhttp.TransportErrorRT(fmt.Errorf("domain expired")))
	mock.Add("yahoo.com", mockhttp.TransportErrorRT(fmt.Errorf("network not found")))
	client := mock.NewClient()

	tests := []struct {
		url string
		msg string
	}{
		{
			url: "https://google.com",
			msg: "Get https://google.com: domain not found",
		},
		{
			url: "https://facebook.com",
			msg: "Get https://facebook.com: domain expired",
		},
		{
			url: "https://yahoo.com",
			msg: "Get https://yahoo.com: network not found",
		},
	}

	for i, test := range tests {
		if resp, err = client.Get(test.url); err == nil {
			t.Errorf("[%d] expected error, got nil", i)
			continue
		} else if resp != nil {
			if resp.Body != nil {
				content, err = ioutil.ReadAll(resp.Body)
				t.Errorf("[%d] expected nil, got %s", i, content)
			} else {
				t.Errorf("[%d] expected nil, got response with nil Body", i)
			}
		}
		if want, have := test.msg, err.Error(); want != have {
			t.Errorf("[%d] expected %#v, got %#v", i, want, have)
		}
	}
}

func ExampleTransportErrorRT() {
	client := &http.Client{
		Transport: mockhttp.TransportErrorRT(fmt.Errorf("domain not found")),
	}

	_, err := client.Get("https://www.google.com")
	fmt.Printf("result 1: %s\n", err.Error())

	_, err = client.Post("https://www.facebook.com", "text/plain", nil)
	fmt.Printf("result 2: %s\n", err.Error())

	// Output:
	// result 1: Get https://www.google.com: domain not found
	// result 2: Post https://www.facebook.com: domain not found
}

func TestStaticResponseRT(t *testing.T) {

	var resp *http.Response
	var err error
	var content []byte

	client := &http.Client{
		Transport: mockhttp.StaticResponseRT("hello world", "text/plain"),
	}

	tests := []struct {
		status  int
		url     string
		content string
	}{
		{
			status:  http.StatusOK,
			url:     "https://something.com/item/1",
			content: "hello world",
		},
		{
			status:  http.StatusOK,
			url:     "https://google.com/",
			content: "hello world",
		},
		{
			status:  http.StatusOK,
			url:     "https://facebook.com/",
			content: "hello world",
		},
	}

	for i, test := range tests {
		if resp, err = client.Get(test.url); err != nil {
			t.Errorf("[%d] unexpected error: %s", i, err)
			continue
		}
		if want, have := test.status, resp.StatusCode; want != have {
			t.Errorf("[%d] expected %#v, got %#v", i, want, have)
		}
		if content, err = ioutil.ReadAll(resp.Body); err != nil {
			t.Errorf("[%d] unexpected error: %s", i, err)
			continue
		}
		if want, have := test.content, string(content); want != have {
			t.Errorf("[%d] expected %#v, got %#v", i, want, have)
		}
	}
}

func ExampleStaticResponseRT() {
	client := &http.Client{
		Transport: mockhttp.StaticResponseRT("hello world", "text/plain"),
	}

	resp, _ := client.Get("http://whatever.com")
	content, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("%s\n", content)

	// Output: hello world
}

func TestFileSystemRT_success(t *testing.T) {

	var resp *http.Response
	var err error
	var content []byte

	client := &http.Client{
		Transport: mockhttp.FileSystemRT("./testdata"),
	}

	tests := []struct {
		status      int
		url         string
		contentType string
		content     string
	}{
		{
			status:      http.StatusOK,
			url:         "https://something.com/test.json",
			contentType: "application/json",
			content: `{
    "foo": "bar",
    "hello": [
        "world 1",
        "world 2"
    ]
}`,
		},
		{
			status:      http.StatusOK,
			url:         "https://google.com/test.txt",
			contentType: "text/plain; charset=utf-8",
			content:     "hello world",
		},
		{
			status:      http.StatusOK,
			url:         "https://facebook.com/persons/1.json",
			contentType: "application/json",
			content: `{
    "id": 1,
    "name": "Elon Musk",
    "cool": true
}`,
		},
	}

	for i, test := range tests {
		if resp, err = client.Get(test.url); err != nil {
			t.Errorf("[%d] unexpected error: %s", i, err)
			continue
		}
		if want, have := test.status, resp.StatusCode; want != have {
			t.Errorf("[%d] expected %#v, got %#v", i, want, have)
		}

		if want, have := test.contentType, resp.Header.Get("Content-Type"); want != have {
			t.Errorf("[%d] expected %#v, got %#v", i, want, have)
		}

		if content, err = ioutil.ReadAll(resp.Body); err != nil {
			t.Errorf("[%d] unexpected error: %s", i, err)
		} else if want, have := test.content, string(content); want != have {
			t.Errorf("[%d] expected %#v, got %#v", i, want, have)
		}
	}
}

func TestFileSystemRT_error(t *testing.T) {

	var resp *http.Response
	var err error
	var content []byte

	client := &http.Client{
		Transport: mockhttp.FileSystemRT("./testdata"),
	}

	tests := []struct {
		status      int
		url         string
		contentType string
		content     string
	}{
		{
			status:      http.StatusNotFound,
			url:         "https://something.com/persons/2.json",
			contentType: "text/plain",
			content:     http.StatusText(http.StatusNotFound),
		},
		{
			status:      http.StatusForbidden,
			url:         "https://google.com/persons/",
			contentType: "text/plain",
			content:     http.StatusText(http.StatusForbidden),
		},
	}

	for i, test := range tests {
		if resp, err = client.Get(test.url); err != nil {
			t.Errorf("[%d] unexpected error: %s", i, err)
			continue
		}
		if want, have := test.status, resp.StatusCode; want != have {
			t.Errorf("[%d] expected %#v, got %#v", i, want, have)
		}
		if want, have := test.contentType, resp.Header.Get("Content-Type"); want != have {
			t.Errorf("[%d] expected %#v, got %#v", i, want, have)
		}
		if content, err = ioutil.ReadAll(resp.Body); err != nil {
			t.Errorf("[%d] unexpected error: %s", i, err)
		} else if want, have := test.content, string(content); want != have {
			t.Errorf("[%d] expected %#v, got %#v", i, want, have)
		}
	}
}

func ExampleFileSystemRT() {
	client := &http.Client{
		Transport: mockhttp.FileSystemRT("./testdata"),
	}

	resp, _ := client.Get("https://www.google.com/persons/2.json")
	content, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("result 1: %s\n", content)

	resp, _ = client.Get("https://www.facebook.com/persons/")
	content, _ = ioutil.ReadAll(resp.Body)
	fmt.Printf("result 2: %s\n", content)

	// Output:
	// result 1: Not Found
	// result 2: Forbidden
}
