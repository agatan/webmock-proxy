package webmock

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/elazarl/goproxy"
	"github.com/wantedly/webmock-proxy/cache"
)

type Server struct {
	config *Config
	proxy  *goproxy.ProxyHttpServer
	body   string
	head   map[string][]string
}

func NewServer(config *Config) (*Server, error) {
	return &Server{
		config: config,
		proxy:  goproxy.NewProxyHttpServer(),
		body:   "",
		head:   make(map[string][]string),
	}, nil
}

func newErrorResponse(req *http.Request, err error) *http.Response {
	msg := fmt.Sprintf(`{"error": %q"}`, err.Error())
	return goproxy.NewResponse(req, "application/json", http.StatusInternalServerError, msg)
}

func (s *Server) connectionCacheHandler() {
	s.proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	s.proxy.OnRequest().DoFunc(
		func(req *http.Request, pctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			log.Printf("[INFO] req %s %s", pctx.Req.Method, pctx.Req.URL.Host+pctx.Req.URL.Path)

			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				log.Printf("failed to copy request: %v", err)
			}
			req.Body.Close()
			pctx.UserData = &Context{RequestBody: body}
			req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			return req, nil
		})
	s.proxy.OnResponse().Do(
		goproxy.HandleBytes(
			func(b []byte, pctx *goproxy.ProxyCtx) []byte {
				log.Printf("[INFO] resp %s", pctx.Resp.Status)
				ctx := pctx.UserData.(*Context)
				c, err := cache.New(".webmock")
				if err != nil {
					panic(err)
				}
				if err := c.Record(ctx.RequestBody, pctx.Req, b, pctx.Resp); err != nil {
					panic(err)
				}
				return b
			}))
}

func (s *Server) mockOnlyHandler() {
	s.proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	s.proxy.OnRequest().DoFunc(
		func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			log.Printf("[INFO] req %s %s", ctx.Req.Method, ctx.Req.URL.Host+ctx.Req.URL.Path)
			c, err := cache.New(".webmock")
			if err != nil {
				panic(err)
			}
			resp, err := c.Replay(req)
			if err != nil {
				return req, newErrorResponse(req, err)
			}
			return req, resp
		})
}

func (s *Server) Start() {
	if s.config.record == true {
		log.Println("[INFO] All HTTP/S request and response is cached.")
		s.connectionCacheHandler()
	} else {
		s.mockOnlyHandler()
	}
	log.Println("[INFO] Running...")
	http.ListenAndServe(":8080", s.proxy)
}
