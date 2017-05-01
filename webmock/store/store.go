package store

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/pkg/errors"

	yaml "gopkg.in/yaml.v2"
)

type Store struct {
	w   io.Writer
	mu  sync.RWMutex
	exs Exchanges
}

var ErrNoCacheFound error = errors.New("no cache found")

func New(w io.Writer, exs Exchanges) *Store {
	return &Store{
		w:   w,
		exs: exs,
	}
}

func (s *Store) Record(reqBody []byte, req *http.Request, respBody []byte, resp *http.Response) error {

	ex := &exchange{Request: newRecordRequest(reqBody, req), Response: newRecordResponse(respBody, resp)}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.exs = append(s.exs, ex)

	data, err := yaml.Marshal(s.exs)
	if err != nil {
		return errors.Wrap(err, "failed to marshal exchanges into yaml")
	}
	n, err := s.w.Write(data)
	if err != nil {
		return errors.Wrap(err, "failed to write exchanges")
	}
	if n != len(data) {
		return errors.New("failed to write entire bytes of marshalized exchanges")
	}
	return nil
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
