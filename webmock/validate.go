package webmock

import (
	"encoding/json"
	"net/http"
	"reflect"
)

func validateRequest(req *http.Request, conn *Connection, body []byte) (bool, error) {
	var header interface{}
	b := []byte(conn.Request.Header)
	if err := json.Unmarshal(b, &header); err != nil {
		return false, err
	}
	if (string(body) == conn.Request.String) &&
		(reflect.DeepEqual(mapToMapInterface(req.Header), header) == true) &&
		(req.Method == conn.Request.Method) {
		return true, nil
	}
	return false, nil
}
