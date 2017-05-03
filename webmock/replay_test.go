package webmock

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestReplayGetRoot(t *testing.T) {
	px, err := NewProxy(ReplayMode, BaseDir("../_test"), Namespace("replay/getroot"))
	if err != nil {
		t.Fatal(err)
	}
	defer px.Close()

	testcases := []struct {
		host       string
		statusCode int
		body       string
	}{
		{"http://200.example.com", 200, "ok"},
		{"http://502.example.com", 502, ""},
	}

	for _, tc := range testcases {
		t.Run(fmt.Sprintf("getting %q", tc.host), func(st *testing.T) {
			resp, err := px.Client.Get(tc.host)
			if err != nil {
				st.Fatal(err)
			}
			if resp.StatusCode != tc.statusCode {
				st.Fatalf("expected status code is %d, but got %d", tc.statusCode, resp.StatusCode)
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				st.Fatalf("failed to read body: %v", err)
			}
			if string(body) != tc.body {
				st.Fatalf(`expected response is %q, but got %q`, tc.body, string(body))
			}
		})
	}
}

func TestReplayNotFound(t *testing.T) {
	px, err := NewProxy(ReplayMode, BaseDir("../_test"))
	if err != nil {
		t.Fatal(err)
	}
	defer px.Close()

	resp, err := px.Client.Get("http://no.such.host.server")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusTeapot {
		t.Fatalf("expected status is 418 I'm a teapot, but got %d: %s", resp.StatusCode, resp.Status)
	}
}
