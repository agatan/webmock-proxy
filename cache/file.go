package cache

import (
	"io"
	"io/ioutil"
	"regexp"

	"github.com/pkg/errors"

	yaml "gopkg.in/yaml.v2"
)

type exchange struct {
	Request *request `yaml:"request"`
}

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

func load(r io.Reader) ([]*exchange, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read yaml")
	}
	var es []*exchange
	if err := yaml.Unmarshal(data, &es); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal exchanges")
	}

	// compile regexps
	for _, e := range es {
		for k, vs := range e.Request.Header {
			for _, v := range vs {
				if v.Regexp != "" {
					v.compiledRegexp, err = regexp.Compile(v.Regexp)
					if err != nil {
						return nil, errors.Wrapf(err, "failed to compile regexp %q for %q in the request to %s %q", v.Regexp, k, e.Request.Method, e.Request.Path)
					}
				}
			}
		}
	}
	return es, nil
}
