package mockhttp_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/yookoala/mockhttp"
)

func TestResponseSetStatus(t *testing.T) {
	client := http.Client{
		Transport: mockhttp.
			UseResponseModifier(mockhttp.ResponseSetStatus(http.StatusOK)).
			Wrap(mockhttp.ServerErrorRT(http.StatusBadGateway)),
	}
	resp, _ := client.Get("http://foobar.com")
	content, _ := ioutil.ReadAll(resp.Body)

	if want, have := http.StatusOK, resp.StatusCode; want != have {
		t.Errorf("expected %#v, got %#v", want, have)
	}
	if want, have := http.StatusText(http.StatusBadGateway), string(content); want != have {
		t.Errorf("expected %#v, got %#v", want, have)
	}
}

func TestResponseSetHeader(t *testing.T) {
	client := http.Client{
		Transport: mockhttp.
			UseResponseModifier(mockhttp.ResponseSetHeader("Content-Type", "text/html")).
			Wrap(mockhttp.StaticResponseRT(`<html>hello world</html>`, "text/plain")),
	}
	resp, _ := client.Get("http://foobar.com")
	if want, have := "text/html", resp.Header.Get("Content-Type"); want != have {
		t.Errorf("expected %#v, got %#v", want, have)
	}
}

func TestResponseAddHeader(t *testing.T) {
	client := http.Client{
		Transport: mockhttp.
			UseResponseModifier(mockhttp.ResponseAddHeader("Content-Type", "text/html")).
			Wrap(mockhttp.StaticResponseRT(`<html>hello world</html>`, "text/plain")),
	}
	resp, _ := client.Get("http://foobar.com")
	if want, have := "text/plain", resp.Header["Content-Type"][0]; want != have {
		t.Errorf("expected %#v, got %#v", want, have)
	}
	if want, have := "text/html", resp.Header["Content-Type"][1]; want != have {
		t.Errorf("expected %#v, got %#v", want, have)
	}
}

func TestUseResponseModifier(t *testing.T) {
	client := http.Client{
		Transport: mockhttp.
			UseResponseModifier(func(resp *http.Response, err error) (*http.Response, error) {
				newContent := "<html>hello world</html>"
				resp.StatusCode = http.StatusOK
				resp.ContentLength = int64(len(newContent))
				resp.Header.Set("Content-Length", fmt.Sprintf("%d", resp.ContentLength))
				resp.Header.Set("Content-Type", "text/html")
				resp.Body = ioutil.NopCloser(strings.NewReader(newContent))
				return resp, err
			}).
			Wrap(mockhttp.ServerErrorRT(http.StatusInternalServerError)),
	}
	resp, _ := client.Get("http://foobar.com")
	content, _ := ioutil.ReadAll(resp.Body)

	if want, have := http.StatusOK, resp.StatusCode; want != have {
		t.Errorf("expected %#v, got %#v", want, have)
	}
	if want, have := int64(24), resp.ContentLength; want != have {
		t.Errorf("expected %#v, got %#v", want, have)
	}
	if want, have := "text/html", resp.Header.Get("Content-Type"); want != have {
		t.Errorf("expected %#v, got %#v", want, have)
	}
	if want, have := "<html>hello world</html>", string(content); want != have {
		t.Errorf("expected %#v, got %#v", want, have)
	}
}

func TestChain(t *testing.T) {
	client := http.Client{
		Transport: mockhttp.Chain(
			mockhttp.UseResponseModifier(
				mockhttp.ResponseSetHeader("Content-Type", "text/html"),
			),
			mockhttp.UseResponseModifier(
				mockhttp.ResponseSetHeader("Content-Type", "application/json"),
			),
			mockhttp.UseResponseModifier(
				mockhttp.ResponseSetStatus(http.StatusOK),
			),
			mockhttp.UseResponseModifier(func(
				resp *http.Response, err error) (*http.Response, error) {
				resp = &http.Response{
					Body: ioutil.NopCloser(strings.NewReader("hello world")),
				}
				err = nil
				return resp, err
			}),
		).Wrap(mockhttp.ServerErrorRT(http.StatusInternalServerError)),
	}

	resp, err := client.Get("https://api.foobar.com/users/1")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if want, have := http.StatusOK, resp.StatusCode; want != have {
		t.Errorf("expected %#v, got %#v", want, have)
	}
	if want, have := http.StatusText(http.StatusOK), resp.Status; want != have {
		t.Errorf("expected %#v, got %#v", want, have)
	}
	if want, have := "hello world", string(content); want != have {
		t.Errorf("expected %#v, got %#v", want, have)
	}
}
