package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/elazarl/goproxy"
	"github.com/wantedly/webmock-proxy/cache"
)

type server struct {
	cache *cache.Cache
	proxy *goproxy.ProxyHttpServer
}

func newRecordServer(cache *cache.Cache) *server {
	s := &server{
		cache: cache,
		proxy: goproxy.NewProxyHttpServer(),
	}
	s.proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	s.proxy.OnRequest().DoFunc(
		func(req *http.Request, pctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			log.Printf("[INFO] req %s %s", pctx.Req.Method, pctx.Req.URL.Host+pctx.Req.URL.Path)

			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				log.Printf("failed to copy request: %v", err)
			}
			req.Body.Close()
			pctx.UserData = body
			req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			return req, nil
		})
	s.proxy.OnResponse().Do(
		goproxy.HandleBytes(
			func(respBody []byte, pctx *goproxy.ProxyCtx) []byte {
				log.Printf("[INFO] resp %s", pctx.Resp.Status)
				reqBody := pctx.UserData.([]byte)
				if err := s.cache.Record(reqBody, pctx.Req, respBody, pctx.Resp); err != nil {
					panic(err)
				}
				return respBody
			}))
	return s
}

func newReplayServer(cache *cache.Cache) *server {
	s := &server{
		cache: cache,
		proxy: goproxy.NewProxyHttpServer(),
	}
	s.proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	s.proxy.OnRequest().DoFunc(
		func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			log.Printf("[INFO] req %s %s", ctx.Req.Method, ctx.Req.URL.Host+ctx.Req.URL.Path)
			resp, err := s.cache.Replay(req)
			if err != nil {
				msg := fmt.Sprintf(`{"error": %q}`, err.Error())
				return req, goproxy.NewResponse(ctx.Req, "application/json", http.StatusInternalServerError, msg)
			}
			return req, resp
		})
	return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.proxy.ServeHTTP(w, r)
}
