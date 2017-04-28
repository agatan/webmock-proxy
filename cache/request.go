package cache

import (
	"net/http"
	"regexp"
)

type request struct {
	Method string                    `yaml:"method"`
	Path   string                    `yaml:"path"`
	Header map[string][]*headerValue `yaml:"header,omitempty"`
	Body   string                    `yaml:"body,omitempty"`
}

type headerValue struct {
	Text           string `yaml:"text,omitempty"`
	Regexp         string `yaml:"regexp,omitempty"`
	compiledRegexp *regexp.Regexp
}

func newRecordRequest(body []byte, req *http.Request) *request {
	r := new(request)
	r.Method = req.Method
	r.Path = req.URL.Path

	r.Body = string(body)

	r.Header = make(map[string][]*headerValue)
	for k, vs := range req.Header {
		r.Header[k] = make([]*headerValue, len(vs))
		for i, v := range vs {
			r.Header[k][i] = &headerValue{Text: v}
		}
	}

	return r
}
