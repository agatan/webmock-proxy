package store

import (
	"bytes"
	"io/ioutil"
	"log"
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
	hosts   map[string][]*exchange
}

var ErrNoCacheFound error = errors.New("no cache found")

func New(basedir string) (*Store, error) {
	s := &Store{
		basedir: basedir,
		hosts:   make(map[string][]*exchange),
	}
	hostdirs, err := ioutil.ReadDir(basedir)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, err
	}
	for _, hostdir := range hostdirs {
		fullpath := filepath.Join(basedir, hostdir.Name(), "exchanges.yaml")
		f, err := os.Open(fullpath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, errors.Wrapf(err, "failed to open %s", fullpath)
		}
		defer f.Close()
		es, err := loadExchanges(f)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to load yaml for %s", hostdir.Name())
		}
		s.hosts[hostdir.Name()] = es
	}
	return s, nil
}

func (s *Store) Record(reqBody []byte, req *http.Request, respBody []byte, resp *http.Response) error {

	ex := &exchange{Request: newRecordRequest(reqBody, req), Response: newRecordResponse(respBody, resp)}

	s.mu.Lock()
	defer s.mu.Unlock()
	if old, ok := s.hosts[req.Host]; ok {
		s.hosts[req.Host] = append(old, ex)
	} else {
		s.hosts[req.Host] = []*exchange{ex}
	}

	savedir := filepath.Join(s.basedir, req.Host)
	if err := os.MkdirAll(savedir, 0777); err != nil {
		log.Println(savedir)
		return errors.Wrap(err, "failed to make directory to save")
	}

	savepath := filepath.Join(savedir, "exchanges.yaml")
	data, err := yaml.Marshal(s.hosts[req.Host])
	if err != nil {
		return errors.Wrap(err, "failed to marshal exchanges into yaml")
	}
	return ioutil.WriteFile(savepath, data, 0644)
}

func (s *Store) Replay(req *http.Request) (*http.Response, error) {
	s.mu.RLock()
	exchanges, ok := s.hosts[req.Host]
	s.mu.RUnlock()
	if !ok {
		return nil, ErrNoCacheFound
	}
	body, err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	req.Body = ioutil.NopCloser(bytes.NewReader(body))
	if err != nil {
		return nil, errors.Wrap(err, "failed to read request body")
	}
	for _, ex := range exchanges {
		if ex.Request.match(body, req) {
			return ex.Response.httpResponse(), nil
		}
	}
	return nil, ErrNoCacheFound
}
