package cache

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	yaml "gopkg.in/yaml.v2"
)

type Cache struct {
	basedir string
	hosts   map[string][]*exchange
}

func New(basedir string) *Cache {
	return &Cache{
		basedir: basedir,
		hosts:   make(map[string][]*exchange),
	}
}

func (c *Cache) Record(req *http.Request) error {
	r := new(request)
	r.Method = req.Method
	r.Path = req.URL.Path

	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return errors.Wrap(err, "failed to read request body")
	}
	r.Body = string(body)
	req.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	r.Header = make(map[string][]*headerValue)
	for k, vs := range req.Header {
		r.Header[k] = make([]*headerValue, len(vs))
		for i, v := range vs {
			r.Header[k][i] = &headerValue{Text: v}
		}
	}

	ex := &exchange{Request: r}

	if old, ok := c.hosts[req.Host]; ok {
		c.hosts[req.Host] = append(old, ex)
	} else {
		c.hosts[req.Host] = []*exchange{ex}
	}

	savedir := filepath.Join(c.basedir, req.Host)
	if err := os.MkdirAll(savedir, 0777); err != nil {
		log.Println(savedir)
		return errors.Wrap(err, "failed to make directory to save")
	}

	savepath := filepath.Join(savedir, "exchanges.yaml")
	data, err := yaml.Marshal(c.hosts[req.Host])
	if err != nil {
		return errors.Wrap(err, "failed to marshal exchanges into yaml")
	}
	return ioutil.WriteFile(savepath, data, 0644)
}
