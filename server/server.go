// Copyright 2014 The erzha Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package server

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/erzha/kernel"

	"code.google.com/p/go.net/context"
)

type Handler struct {
	disabled bool

	server *kernel.Server

	//listener
	Ln net.Listener

	maxChildren int64
	currentChildren int64
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

	p.Ln, err = net.Listen("tcp", "127.0.0.1:9999")

	if nil != err {
		return //exit
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

	ctx, cancel := context.WithTimeout(serverCtx, 100)
	defer cancel()

	actionDone := make(chan bool)

	sapiobj := NewSapi(res, req)
	kernelSapi := &kernel.Sapi{}
	kernelSapi.Stdout = res
	kernelSapi.Stderr = res
	kernelSapi.Ext = sapiobj
	sapiobj.Kernel = kernelSapi
	if nil != InitHttpRequest(sapiobj) {
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

func NewHandler() *Handler {
	ret := &Handler{}
	ret.maxChildren = 1024
	return ret

}

func Boot() {
	kernel.Boot(NewHandler())
}
