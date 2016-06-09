package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/elazarl/goproxy"
)

func main() {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true

	// MEMO(munisystem):
	// Request Body を取得できるタイミングが OnRequest 内しか存在しない。
	// チャネル作って OnResponse 内に送る。
	// make でチャネルを作るときに第二引数を与えないと buffer size 0 となり、
	// ロックされてレスポンスが一生帰ってこなくなる。
	c := make(chan string, 1)
	useCache := make(chan bool, 1)

	proxy.OnRequest().DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			body := readRequestBody(r)
			if validateRequest(r, body) {
				resp := newResponse(r)
				ctx.Logf("webmock-proxy use http request cache!!")
				useCache <- true
				return r, resp
			}
			c <- body
			r.Body = ioutil.NopCloser(bytes.NewBufferString(body))
			return r, nil
		})

	proxy.OnResponse().Do(
		goproxy.HandleBytes(
			func(b []byte, ctx *goproxy.ProxyCtx) []byte {
				select {
				case respBody := <-c:
					createCacheFile(respBody, b, ctx)
				case <-useCache:
				}
				return b
			}))
	log.Fatal(http.ListenAndServe(":8080", proxy))
}
