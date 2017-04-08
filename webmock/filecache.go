package webmock

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"

	"github.com/pkg/errors"
)

type fileCache struct {
	root string
}

func NewFileCache(root string) Cache {
	return &fileCache{root: root}
}

func (fc *fileCache) write(conn *Connection, req *http.Request) error {
	dst := filepath.Join(fc.root, req.URL.String(), req.Method+".json")
	bs, err := json.Marshal(conn)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dst, bs, 0644)
}

func (fc *fileCache) Save(reqBody []byte, req *http.Request, respBody []byte, resp *http.Response) error {
	reqStruct, err := requestStruct(string(reqBody), req)
	if err != nil {
		return err
	}
	respStruct, err := responseStruct(respBody, resp)
	if err != nil {
		return err
	}
	conn := &Connection{Request: reqStruct, Response: respStruct, RecordedAt: resp.Header.Get("Date")}
	if err := fc.write(conn, req); err != nil {
		return err
	}
	log.Printf("[INFO] Create HTTP/S connection cache.")
	return nil
}

func (fc *fileCache) Find(req *http.Request) (*http.Response, error) {
	reqBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read request body")
	}
	req.Body.Close()
	var (
		url  = req.URL.Host + req.URL.Path
		file = req.Method + ".json"
		dst  = filepath.Join(fc.root, url, file)
	)
	b, err := ioutil.ReadFile(dst)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read local cache file")
	}
	conn := new(Connection)
	if err := json.Unmarshal(b, conn); err != nil {
		return nil, errors.Wrap(err, "failed to read serialized local cache")
	}
	is, err := validateRequest(req, conn, reqBody)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find cache")
	} else if is == false {
		return nil, fmt.Errorf("cache not found for %s %s", req.Method, req.URL.String())
	}
	resp, err := createHttpResponse(req, conn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create http response")
	}
	return resp, nil
}
