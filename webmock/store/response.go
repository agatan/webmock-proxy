package store

import (
	"io/ioutil"
	"net/http"
	"strings"
)

type response struct {
	Status     string      `yaml:"status"`
	StatusCode int         `yaml:"status_code"`
	Proto      string      `yaml:"proto"` // e.g. "HTTP/1.1"
	Header     http.Header `yaml:"header,omitempty"`
	Body       string      `yaml:"body,omitempty"`
}

func copyHeader(h http.Header) http.Header {
	dst := make(http.Header)
	for k, vs := range h {
		dvs := make([]string, len(vs))
		for i, v := range vs {
			dvs[i] = v
		}
		dst[k] = dvs
	}
	return dst
}

func newRecordResponse(body []byte, resp *http.Response) *response {
	return &response{
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Proto:      resp.Proto,
		Header:     copyHeader(resp.Header),
		Body:       string(body), // FIXME(agatan): binary body
	}
}

func (r *response) httpResponse() *http.Response {
	return &http.Response{
		Status:     r.Status,
		StatusCode: r.StatusCode,
		Proto:      r.Proto,
		Header:     copyHeader(r.Header),
		Body:       ioutil.NopCloser(strings.NewReader(r.Body)),
	}
}
