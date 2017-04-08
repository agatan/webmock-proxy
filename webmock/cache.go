package webmock

import (
	"net/http"
)

type Cache interface {
	Save(reqBody []byte, req *http.Request, respBody []byte, resp *http.Response) error
	Find(req *http.Request) (*http.Response, error)
}
