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

func (r *request) match(body []byte, req *http.Request) bool {
	if r.Path != req.URL.Path || r.Method != req.Method {
		return false
	}
	for key, vs := range r.Header {
		reqvs, ok := req.Header[key]
		if !ok {
			return false
		}
		for _, v := range vs {
			ok := false
			for _, reqv := range reqvs {
				if v.match(reqv) {
					ok = true
					break
				}
			}
			if !ok {
				return false
			}
		}
	}
	return r.Body == string(body)
}

func (h *headerValue) match(t string) bool {
	if h == nil {
		return true
	}
	if h.compiledRegexp != nil {
		return h.compiledRegexp.MatchString(t)
	}
	if h.Text != "" {
		return h.Text == t
	}
	return true
}
