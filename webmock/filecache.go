package webmock

import (
	"errors"
	"log"
	"net/http"
	"path/filepath"
)

type fileCache struct {
	root string
}

func NewFileCache(root string) Cache {
	return &fileCache{root: root}
}

func (fc *fileCache) Save(reqBody []byte, req *http.Request, respBody []byte, resp *http.Response) error {
	var (
		url  = req.URL.Host + req.URL.Path
		file = req.Method + ".json"
		dst  = filepath.Join(fc.root, url, file)
	)

	reqStruct, err := requestStruct(string(reqBody), req)
	if err != nil {
		return err
	}
	respStruct, err := responseStruct(respBody, resp)
	if err != nil {
		return err
	}
	conn := Connection{Request: reqStruct, Response: respStruct, RecordedAt: resp.Header.Get("Date")}
	byteArr, err := structToJSON(conn)
	if err != nil {
		return err
	}
	if err := writeFile(string(byteArr), dst); err != nil {
		return err
	}
	log.Printf("[INFO] Create HTTP/S connection cache.")
	return nil
}

func (fc *fileCache) Find(req *http.Request) (*http.Response, error) {
	return nil, errors.New("not implemented yet")
}
