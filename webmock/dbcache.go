package webmock

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

type dbCache struct {
	db *gorm.DB
}

func NewDBCache(db *gorm.DB) Cache {
	return &dbCache{db: db}
}

func (dc *dbCache) Save(reqBody []byte, req *http.Request, respBody []byte, resp *http.Response) error {
	url := req.URL.Host + req.URL.Path

	reqStruct, err := requestStruct(string(reqBody), req)
	if err != nil {
		return err
	}
	respStruct, err := responseStruct(respBody, resp)
	if err != nil {
		return err
	}
	conn := Connection{Request: reqStruct, Response: respStruct, RecordedAt: resp.Header.Get("Date")}
	conns := []Connection{conn}
	endpoint := &Endpoint{
		URL:         url,
		Connections: conns,
		Update:      time.Now(),
	}
	ce := readEndpoint(url, dc.db)
	if len(ce.Connections) != 0 {
		for _, v := range ce.Connections {
			deleteConnection(&v, dc.db)
			if v.Request.Method == req.Method {
				continue
			}
			conns = append(conns, v)
		}
		endpoint.Connections = conns
		updateEndpoint(ce, endpoint, dc.db)

		log.Printf("[INFO] Update HTTP/S connection cache.")
		return nil
	}
	if err := insertEndpoint(endpoint, dc.db); err != nil {
		return err
	}
	log.Printf("[INFO] Create HTTP/S connection cache.")
	return nil
}

func (dc *dbCache) Find(req *http.Request) (*http.Response, error) {
	reqBody, err := ioReader(req.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read request body")
	}
	endpoint := findEndpoint(req.Method, req.URL.Host+req.URL.Path, dc.db)
	if len(endpoint.Connections) == 0 {
		return nil, fmt.Errorf("cache not found for %s %s", req.Method, req.URL.String())
	}
	conn := &endpoint.Connections[0]
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
