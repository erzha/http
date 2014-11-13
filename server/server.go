// Copyright 2014 The erzha Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package server

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/erzha/kernel"

	"golang.org/x/net/context"
	"golang.org/x/net/websocket"
)

var handlerObj *Handler
type Handler struct {
	disabled bool

	server *kernel.Server

	//listener
	Ln net.Listener

	maxChildren int64
	currentChildren int64

	staticPrefix string
	staticDir string

	confTimeout time.Duration

	ws websocket.Server
}

func (p *Handler) shutdown() {
	p.disabled = true
	for p.currentChildren > 0 {
		/*
		p.pServer.Logger.Infof(
			"wait for currentChildren stop, remains %d. use [ kill -9 %d ] if you want to kill it at once.",
			p.currentChildren,
			os.Getpid(),
		)
		*/
		time.Sleep(1*time.Second)
	}
	p.Ln.Close()
}

var serverCtx context.Context
var serverCtxCancel context.CancelFunc

func (p *Handler) Serve(ctx context.Context, pServer *kernel.Server) {

	serverCtx, serverCtxCancel = context.WithCancel(ctx)
	p.server = pServer

	var err error
	
	listenNet := pServer.Conf.String("erzha.http.net", "tcp")
	listenAddr := pServer.Conf.String("erzha.http.laddr", ":8989")
	p.Ln, err = net.Listen(listenNet, listenAddr)

	if nil != err {
		pServer.Logger.Fatalf("erzha_http_server_listen_error %s", err.Error())
		return //exit
	}

	p.staticPrefix = pServer.Conf.String("erzha.http.static_prefix", "/static/")
	p.staticDir = pServer.Conf.String("erzha.http.static_dir", "static/")

	p.confTimeout, err = time.ParseDuration(pServer.Conf.String("erzha.http.timeout", "6s"))
	if nil != err {
		pServer.Logger.Fatal("erzha_http_server_timeout_conf_error ", err)
		return
	}

	go func() {
		server := &http.Server{}
		server.Handler = p
		server.Serve(p.Ln)
	}()
	<-serverCtx.Done()
	p.shutdown()
}

func (p *Handler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if p.disabled {
		http.Error(res, "the server is shutting down", 503)
		return
	}

	if p.currentChildren >= p.maxChildren {
		errorMsg := fmt.Sprintf("currentChildren has reached %d, please raise the wgf.sapi.maxChildren", p.currentChildren)
		http.Error(res, errorMsg, 503)
		return
	}

	//manage currentChildren
	defer func(){ p.currentChildren-- }()
	p.currentChildren++

	isWebsocket := req.Header.Get("Upgrade") == "websocket"
	if false == isWebsocket {
		p.serveHttpRequest(res, req)
	} else {
		p.ws.ServeHTTP(res, req)
	}
}

func (p *Handler) serveHttpRequest(res http.ResponseWriter, req *http.Request) {
	var err error

	ctx, cancel := context.WithTimeout(serverCtx, p.confTimeout)
	defer cancel()

	actionDone := make(chan bool)


	//whether is a static resourse
	if req.RequestURI == "/favicon.ico" || strings.HasPrefix(req.RequestURI, p.staticPrefix) {
		http.ServeFile(res, req, p.staticDir + req.RequestURI)
		return
	}

	sapiobj := NewSapi(res, req)
	kernelSapi := kernel.NewSapi()
	kernelSapi.Stdout = res
	kernelSapi.Stderr = res
	kernelSapi.Ext = sapiobj
	sapiobj.Kernel = kernelSapi

	sapiobj.handler, err = parseRequestURI(sapiobj.RequestURI(), sapiobj.Get)
	if err != nil {
		kernelSapi.Server.Logger.Warning(err.Error())
		return
	}

	go func() {
		kernel.FireAction(ctx, kernelSapi, do)
		close(actionDone)
	}()

	select {
	case <-actionDone:
	case <-ctx.Done():
	case <-res.(http.CloseNotifier).CloseNotify(): //client disconnected
	}
}

func (p *Handler) serveWebsocket(conn *websocket.Conn) {
	var err error

	ctx, cancel := context.WithTimeout(serverCtx, p.confTimeout)
	defer cancel()

	sapiobj := NewWebsocketSapi(conn.Request())
	kernelSapi := kernel.NewSapi()
	kernelSapi.Ext = sapiobj
	sapiobj.Kernel = kernelSapi
	sapiobj.Conn = conn

	sapiobj.handler, err = parseRequestURI(sapiobj.RequestURI(), sapiobj.Get)
	if err != nil {
		kernelSapi.Server.Logger.Warning(err.Error())
		return
	}

	actionDone := make(chan bool)
	go func() {
		kernel.FireAction(ctx, kernelSapi, doWebsocket)
		close(actionDone)
	}()

	select {
	case <-actionDone:
	case <-ctx.Done():
	}
}

func NewHandler() *Handler {
	handlerObj := &Handler{}
	handlerObj.maxChildren = 1024
	handlerObj.ws = websocket.Server{}
	handlerObj.ws.Handler = handlerObj.serveWebsocket
	return handlerObj

}

func Boot() {
	kernel.Boot(NewHandler())
}
