package store

import (
	"io"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"

	yaml "gopkg.in/yaml.v2"
)

type Exchanges []*exchange

type exchange struct {
	Request  *request  `yaml:"request"`
	Response *response `yaml:"response"`
}

func LoadExchanges(r io.Reader) (Exchanges, error) {
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
		if err := e.Request.compile(); err != nil {
			return nil, errors.Wrapf(err, "failed to load request %s %s/%s", e.Request.Method, e.Request.Host, e.Request.Path)
		}
	}
	return es, nil
}

func (exs Exchanges) Lookup(body []byte, req *http.Request) *http.Response {
	for _, ex := range exs {
		if ex.Request.match(body, req) {
			return ex.Response.httpResponse()
		}
	}
	return nil
}
