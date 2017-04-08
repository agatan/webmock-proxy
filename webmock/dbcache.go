package webmock

import (
	"log"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
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
