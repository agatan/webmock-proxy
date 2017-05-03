package webmock

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"

	"github.com/elazarl/goproxy"
	"github.com/pkg/errors"
	"github.com/wantedly/webmock-proxy/webmock/store"
)

type Proxy struct {
	IsRecordMode bool
	Addr         string
	BaseDir      string
	URL          *url.URL
	Client       *http.Client
	Server       *http.Server

	listener net.Listener
	wg       sync.WaitGroup
	mu       sync.Mutex
	closed   bool

	store *store.Store
	proxy *goproxy.ProxyHttpServer
}

func newLocalListener(addr string) (net.Listener, error) {
	if addr != "" {
		l, err := net.Listen("tcp", addr)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to listen on %v", addr)
		}
		return l, nil
	}
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if l, err = net.Listen("tcp6", "[::1]:0"); err != nil {
			return nil, errors.Wrap(err, "failed to listen on a port")
		}
	}
	return l, nil
}

func NewProxy(options ...Option) (*Proxy, error) {
	p, err := NewUnstartedProxy(options...)
	if err != nil {
		return nil, err
	}
	p.Start()
	return p, nil
}

func NewUnstartedProxy(options ...Option) (*Proxy, error) {
	gopx := goproxy.NewProxyHttpServer()
	p := &Proxy{
		IsRecordMode: os.Getenv("WEBMOCK_PROXY_RECORD") == "1",
		BaseDir:      ".webmock",
		Server:       &http.Server{Handler: gopx},
		proxy:        gopx,
	}
	for _, op := range options {
		op(p)
	}
	var err error
	p.store, err = store.New(p.BaseDir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize exchange store")
	}
	if p.IsRecordMode {
		p.registerRecordHandlers()
	} else {
		p.registerReplayHandlers()
	}
	p.listener, err = newLocalListener(p.Addr)
	return p, err
}

func (p *Proxy) registerRecordHandlers() {
	p.proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	p.proxy.OnRequest().DoFunc(
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
	p.proxy.OnResponse().Do(
		goproxy.HandleBytes(
			func(respBody []byte, pctx *goproxy.ProxyCtx) []byte {
				log.Printf("[INFO] resp %s", pctx.Resp.Status)
				reqBody := pctx.UserData.([]byte)
				p.store.Record(reqBody, pctx.Req, respBody, pctx.Resp)
				return respBody
			}))
}

func (p *Proxy) registerReplayHandlers() {
	p.proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	p.proxy.OnRequest().DoFunc(
		func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			log.Printf("[INFO] req %s %s", ctx.Req.Method, ctx.Req.URL.Host+ctx.Req.URL.Path)
			resp, err := p.store.Replay(req)
			if err != nil {
				msg := fmt.Sprintf(`{"error": %q}`, err.Error())
				return req, goproxy.NewResponse(ctx.Req, "application/json", http.StatusInternalServerError, msg)
			}
			return req, resp
		})
}

func (p *Proxy) Start() {
	u, err := url.Parse("http://" + p.listener.Addr().String())
	if err != nil {
		panic(errors.Wrap(err, "failed to parse listening address"))
	}
	p.URL = u
	p.Client = &http.Client{
		Transport: &http.Transport{
			Proxy:           http.ProxyURL(u),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	p.goServe()
}

func (p *Proxy) goServe() {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		_ = p.Server.Serve(p.listener)
	}()
}

func (p *Proxy) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.closed {
		p.closed = true
		if err := p.listener.Close(); err != nil {
			return err
		}
		p.Server.SetKeepAlivesEnabled(false)
		if p.IsRecordMode {
			return p.store.Flush()
		}
	}
	return nil
}
