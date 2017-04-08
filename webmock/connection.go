package webmock

import (
	"time"
)

type Endpoint struct {
	ID          uint `gorm:"primary_key;AUTO_INCREMENT" json:"-"`
	URL         string
	Connections []Connection
	Update      time.Time
}

type Connection struct {
	ID         uint     `gorm:"primary_key;AUTO_INCREMENT" json:"-"`
	EndpointID uint     `json:"-"`
	Request    Request  `json:"request"`
	Response   Response `json:"response"`
	RecordedAt string   `json:"recorded_at"`
}

type Request struct {
	ID           uint   `gorm:"primary_key;AUTO_INCREMENT" json:"-"`
	ConnectionID uint   `json:"-"`
	Header       string `json:"header"`
	String       string `json:"string"`
	Method       string `json:"method"`
	URL          string `json:"url"`
}

type Response struct {
	ID           uint   `gorm:"primary_key;AUTO_INCREMENT" json:"-"`
	ConnectionID uint   `json:"-"`
	Status       string `json:"status"`
	Header       string `json:"header"`
	String       string `json:"string"`
}
