package store

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/pkg/errors"

	yaml "gopkg.in/yaml.v2"
)

type Store struct {
	basedir string
	mu      sync.RWMutex
	exs     []*exchange
}

var ErrNoCacheFound error = errors.New("no cache found")

func New(basedir string) (*Store, error) {
	s := &Store{
		basedir: basedir,
	}
	yamlpath := filepath.Join(basedir, "default.yaml")
	f, err := os.Open(yamlpath)
	if err == nil {
		defer f.Close()
		s.exs, err = loadExchanges(f)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to load yaml %s", yamlpath)
		}
	} else if !os.IsNotExist(err) {
		return nil, errors.Wrapf(err, "failed to open %s", yamlpath)
	}
	return s, nil
}

func (s *Store) Record(reqBody []byte, req *http.Request, respBody []byte, resp *http.Response) error {

	ex := &exchange{Request: newRecordRequest(reqBody, req), Response: newRecordResponse(respBody, resp)}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.exs = append(s.exs, ex)

	if err := os.MkdirAll(s.basedir, 0777); err != nil {
		return errors.Wrap(err, "failed to make directory to save")
	}

	savepath := filepath.Join(s.basedir, "default.yaml")
	data, err := yaml.Marshal(s.exs)
	if err != nil {
		return errors.Wrap(err, "failed to marshal exchanges into yaml")
	}
	return ioutil.WriteFile(savepath, data, 0644)
}

func (s *Store) Replay(req *http.Request) (*http.Response, error) {
	body, err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	req.Body = ioutil.NopCloser(bytes.NewReader(body))
	if err != nil {
		return nil, errors.Wrap(err, "failed to read request body")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, ex := range s.exs {
		if ex.Request.match(body, req) {
			return ex.Response.httpResponse(), nil
		}
	}
	return nil, ErrNoCacheFound
}
