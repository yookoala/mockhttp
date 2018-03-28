package mockhttp_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/yookoala/mockhttp"
)

func TestUseResponseStatus(t *testing.T) {
	client := http.Client{
		Transport: mockhttp.
			UseResponseStatus(http.StatusOK).
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

func TestUseResponseSetHeader(t *testing.T) {
	client := http.Client{
		Transport: mockhttp.
			UseResponseSetHeader("Content-Type", "text/html").
			Wrap(mockhttp.StaticResponseRT(`<html>hello world</html>`, "text/plain")),
	}
	resp, _ := client.Get("http://foobar.com")
	if want, have := "text/html", resp.Header.Get("Content-Type"); want != have {
		t.Errorf("expected %#v, got %#v", want, have)
	}
}

func TestUseResponseAddHeader(t *testing.T) {
	client := http.Client{
		Transport: mockhttp.
			UseResponseAddHeader("Content-Type", "text/html").
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
			mockhttp.UseResponseModifier(func(
				resp *http.Response, err error) (*http.Response, error) {
				return resp, err
			}),
			mockhttp.UseResponseSetHeader("Content-Type", "text/html"),
		).Wrap(mockhttp.ServerErrorRT(http.StatusInternalServerError)),
	}
	client.Get("https://api.foobar.com/users/1")
}
