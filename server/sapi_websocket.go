// Copyright 2014 The erzha Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package server

import (
	"net/http"
	"net/url"

	"github.com/erzha/kernel"

	"golang.org/x/net/websocket"
)


type WebsocketSapi struct {

	Kernel *kernel.Sapi

	Conn *websocket.Conn
	Req *http.Request
	Status int

	Get	url.Values

	handler *actionHandler
}

func (p *WebsocketSapi) RequestURI() string {
	if nil == p.Req {
		return ""
	}
	return p.Req.URL.Path
}

func (p *WebsocketSapi) Cookie(key string)  string {
	c, e := p.Req.Cookie(key)
	if nil!=e {
		return ""
	}
	return c.Value;
}

func NewWebsocketSapi(req *http.Request) *WebsocketSapi {

	ret := &WebsocketSapi{}
	ret.Req = req
	ret.Status = 200

	ret.Get = url.Values{}
	return ret
}
