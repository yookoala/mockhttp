package mockhttp_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/yookoala/mockhttp"
)

func TestUseTransport_basic(t *testing.T) {
	mock := mockhttp.NewMuxRoundTripper()
	defaultTransport := http.DefaultTransport

	mockhttp.UseTransport(&mock)
	if want, have := &mock, http.DefaultTransport; want != have {
		t.Errorf("expected %#v, got %#v", want, have)
	}

	mockhttp.RestoreTransport()
	if want, have := defaultTransport, http.DefaultTransport; want != have {
		t.Errorf("expected %#v, got %#v", want, have)
	}
}

func TestUseTransport_locking(t *testing.T) {
	errors := make(chan error)
	wg := &sync.WaitGroup{}
	testFn := func(i int, errors chan error) {
		mock := mockhttp.NewMuxRoundTripper()
		signature := fmt.Sprintf("hello: %d", i)
		mock.Add("*", mockhttp.StaticResponseRT(
			signature,
			"text/plain",
		))

		mockhttp.UseTransport(&mock)
		t.Logf("UseTransport %d", i)
		<-time.After(5 * time.Millisecond) // wait for 5 milliseconds
		resp, _ := mock.NewClient().Get("https://www.google.com")

		content, _ := ioutil.ReadAll(resp.Body)
		if want, have := signature, string(content); want != have {
			errors <- fmt.Errorf("expected %#v, got %#v", want, have)
		}
		resp.Body.Close()
		t.Logf("RestoreTransport %d", i)
		mockhttp.RestoreTransport()
		wg.Done()
	}

	t.Logf("expect to see UseTransport [number] follow immediately with RestoreTransport of the same number")
	for i := 1; i <= 50; i++ {
		wg.Add(1)
		go testFn(i, errors)
	}
	go func() {
		wg.Wait()
		close(errors)
		t.Logf("all testFn finished")
	}()

	for err := range errors {
		if err != nil {
			t.Errorf("collision: %s", err.Error())
		}
	}
}

func ExampleUseTransport_simple() {
	mock := mockhttp.StaticResponseRT("hello world", "text/plain")
	mockhttp.UseTransport(&mock)

	resp, _ := http.Get("https://www.google.com")
	content, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("%s", content)

	mockhttp.RestoreTransport()

	// Output: hello world
}
