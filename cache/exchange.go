package cache

import (
	"io"
	"io/ioutil"
	"regexp"

	"github.com/pkg/errors"

	yaml "gopkg.in/yaml.v2"
)

type exchange struct {
	Request  *request  `yaml:"request"`
	Response *response `yaml:"response"`
}

func loadExchanges(r io.Reader) ([]*exchange, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read yaml")
	}
	var es []*exchange
	if err := yaml.Unmarshal(data, &es); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal exchanges")
	}

	// compile regexps in requests
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
