package store

import (
	"net/http"
	"regexp"

	"github.com/pkg/errors"
)

type request struct {
	Host   string
	Method string
	Path   string
	Header map[string][]*headerValue `yaml:"header,omitempty"`
	Body   string                    `yaml:"body,omitempty"`
}

type headerValue struct {
	Text           string `yaml:"text,omitempty"`
	Regexp         string `yaml:"regexp,omitempty"`
	compiledRegexp *regexp.Regexp
}

func newRecordRequest(body []byte, req *http.Request) *request {
	r := &request{
		Host:   req.Host,
		Method: req.Method,
		Path:   req.URL.Path,
		Body:   string(body),
	}

	r.Header = make(map[string][]*headerValue)
	for k, vs := range req.Header {
		r.Header[k] = make([]*headerValue, len(vs))
		for i, v := range vs {
			r.Header[k][i] = &headerValue{Text: v}
		}
	}

	return r
}

func (r *request) compile() (err error) {
	for k, vs := range r.Header {
		for _, v := range vs {
			if v.Regexp != "" {
				v.compiledRegexp, err = regexp.Compile(v.Regexp)
				if err != nil {
					return errors.Wrapf(err, "failed to compile regexp %q for %s", v.Regexp, k)
				}
			}
		}
	}
	return
}

func (r *request) match(body []byte, req *http.Request) bool {
	if r.Host != req.Host || r.Path != req.URL.Path || r.Method != req.Method {
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
