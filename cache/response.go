package cache

import "net/http"

type response struct {
	Status     string              `yaml:"status"`
	StatusCode int                 `yaml:"status_code"`
	Proto      string              `yaml:"proto"` // e.g. "HTTP/1.1"
	Header     map[string][]string `yaml:"header,omitempty"`
	Body       string              `yaml:"body,omitempty"`
}

func newRecordResponse(body []byte, resp *http.Response) *response {
	r := &response{
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Proto:      resp.Proto,
		Header:     make(map[string][]string),
		Body:       string(body), // FIXME(agatan): binary body
	}

	for k, vs := range resp.Header {
		rs := make([]string, len(vs))
		for i, v := range vs {
			rs[i] = v
		}
		r.Header[k] = rs
	}

	return r
}
