// Copyright 2014 The erzha Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package server

import (
	"errors"
	"net/http"
	"net/url"
	"reflect"
	"runtime/debug"
	"strings"

	"github.com/erzha/kernel"

	"context"
)

var (
	confDefaultAction  string = "index"
	confRewriteEnabled bool
)

type actionHandler struct {
	flag      bool
	hasDoGet  bool
	hasDoPost bool

	//websocket only
	hasDoWebsocket bool
	creater        func() ActionInterface
}

type actionDoGet interface {
	DoGet(ctx context.Context, sapi *Sapi)
}

type actionDoPost interface {
	DoPost(ctx context.Context, sapi *Sapi)
}

type actionDoWebsocket interface {
	DoWebsocket(ctx context.Context, sapi *WebsocketSapi)
}

type ActionInterface interface {
	Execute(ctx context.Context, sapi *Sapi)
	Init(ctx context.Context, sapi *Sapi) error
	InitWebsocket(ctx context.Context, sapi *WebsocketSapi) error
}

type Action struct {
}

func (action *Action) Execute(ctx context.Context, sapi *Sapi) {

}

func (action *Action) Init(ctx context.Context, sapi *Sapi) error {
	return nil
}

func (action *Action) InitWebsocket(ctx context.Context, sapi *WebsocketSapi) error {
	return nil
}

var actionMap map[string]*actionHandler

func Router(url string, name string, creater func() ActionInterface) {
	routerAddRule(url, name)
	handler := &actionHandler{}
	handler.creater = creater
	actionMap[name] = handler
}

func do(ctx context.Context, sapi *kernel.Sapi) {
	httpsapi := sapi.Ext.(*Sapi)

	defer func() {
		r := recover()
		if nil != r {
			sapi.Server.Logger.Warning("server_internal_error ", r)
			sapi.Server.Logger.Warning(string(debug.Stack()))
			http.Error(httpsapi.Res, "server internal error", http.StatusInternalServerError)
		}
	}()

	handler := httpsapi.handler
	obj := handler.creater()
	if false == handler.flag {
		handler.flag = true
		t := reflect.TypeOf(obj)
		_, handler.hasDoGet = t.MethodByName("DoGet")
		_, handler.hasDoPost = t.MethodByName("DoPost")

	}

	if nil != obj.Init(ctx, httpsapi) {
		return
	}

	switch {
	case "GET" == httpsapi.Req.Method && handler.hasDoGet:
		obj.(actionDoGet).DoGet(ctx, httpsapi)
	case "POST" == httpsapi.Req.Method && handler.hasDoPost:
		obj.(actionDoPost).DoPost(ctx, httpsapi)
	default:
		obj.Execute(ctx, httpsapi)
	}
}

func doWebsocket(ctx context.Context, sapi *kernel.Sapi) {
	wsapi := sapi.Ext.(*WebsocketSapi)

	defer func() {
		r := recover()
		if nil != r {
			sapi.Server.Logger.Warning("server_internal_error ", r)
		}
	}()

	handler := wsapi.handler
	obj := handler.creater()
	if false == handler.flag {
		handler.flag = true
		t := reflect.TypeOf(obj)
		_, handler.hasDoWebsocket = t.MethodByName("DoWebsocket")
	}

	if false == handler.hasDoWebsocket {
		sapi.Server.Logger.Warning("DoWebsocket method missed")
		return
	}

	if nil != obj.InitWebsocket(ctx, wsapi) {
		return
	}

	obj.(actionDoWebsocket).DoWebsocket(ctx, wsapi)
}

func parseRequestURI(uri string, httpGetParam url.Values) (*actionHandler, error) {

	//get action name from requesturi
	var actionName string
	var param map[string]string
	var ok bool
	var handler *actionHandler

	uri = strings.TrimLeft(uri, "/\\")

	if uri == "" {
		actionName = confDefaultAction
	} else {
		uri = "/" + uri
		_, ok = httpGetParam["r"]
		if ok {
			actionName = httpGetParam.Get("r")
		} else if confRewriteEnabled {
			actionName, param = urlToAction(uri)
			if "" != actionName && nil != param {
				for key, val := range param {
					httpGetParam.Set(key, val)
				}
			}
		} else {
			actionName = uri
		}
	}

	handler, ok = actionMap[actionName]
	if !ok {
		return nil, errors.New("cannot find action creater named " + actionName + " requesturi:" + uri)
	}
	return handler, nil
}

func init() {
	actionMap = make(map[string]*actionHandler)
	confRewriteEnabled = true
}
