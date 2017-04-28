package cache

import "regexp"

type exchange struct {
	Request *request `yaml:"request"`
}

type request struct {
	Method string                    `yaml:"method"`
	Path   string                    `yaml:"path"`
	Header map[string][]*headerValue `yaml:"header,omitempty"`
	Body   string                    `yaml:"body,omitempty"`
}

type headerValue struct {
	Text           string `yaml:"text,omitempty"`
	Regexp         string `yaml:"regexp,omitempty"`
	compiledRegexp *regexp.Regexp
}
