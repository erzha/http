// Copyright 2014 The erzha Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package server

import (
	"strings"
	"errors"
	"reflect"
	"net/http"

	"github.com/erzha/kernel"

	"golang.org/x/net/context"
)


var (
	confDefaultAction string = "index"
	confRewriteEnabled bool
)

type actionHandler struct {
	flag bool
	hasDoGet bool
	hasDoPost bool
	creater func()ActionInterface
}

type actionDoGet interface {
	DoGet(ctx context.Context, sapi *Sapi)
}

type actionDoPost interface {
	DoPost(ctx context.Context, sapi *Sapi)
}

type ActionInterface interface {
	Execute(ctx context.Context, sapi *Sapi)
	Init(ctx context.Context, sapi *Sapi) error
}

type Action struct {
}

func (action *Action) Execute(ctx context.Context, sapi *Sapi) {

}

func (action *Action) Init(ctx context.Context, sapi *Sapi) error {
	return nil
}

var actionMap map[string]*actionHandler

func Router(url string, name string, creater func()ActionInterface) {
	routerAddRule(url, name)
	handler := &actionHandler{}
	handler.creater = creater
	actionMap[name] = handler
}

func do(ctx context.Context, sapi *kernel.Sapi) {
	httpsapi := sapi.Ext.(*Sapi)

	defer func() {
		r := recover()
		if nil!=r {
			sapi.Server.Logger.Warning("server_internal_error ", r)
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

	switch  {
	case "GET" == httpsapi.Req.Method && handler.hasDoGet:
		obj.(actionDoGet).DoGet(ctx, httpsapi)
	case "POST" == httpsapi.Req.Method && handler.hasDoPost:
		obj.(actionDoPost).DoPost(ctx, httpsapi)
	default :
		obj.Execute(ctx, httpsapi)
	}
}

func InitHttpRequest(httpsapi *Sapi) error  {

	//get action name from requesturi
	var actionName, uri string
	var param map[string]string
	var ok bool

	uri = httpsapi.RequestURI()
	uri = strings.TrimLeft(uri, "/\\")
	
	if uri == "" {
		actionName = confDefaultAction
	} else {
		uri = "/" + uri
		_, ok = httpsapi.Get["r"]
		if ok {
			actionName = httpsapi.Get.Get("r")
		} else if confRewriteEnabled {
			actionName, param = urlToAction(uri)
			if "" != actionName && nil != param {
				for key, val := range param {
					httpsapi.Get.Set(key, val)
				}
			}
		} else {
			actionName = uri
		}
	}

	httpsapi.handler, ok = actionMap[actionName]
	if !ok {
		return errors.New("cannot find action creater named " + actionName)
	}
	return nil
}

func init() {
	actionMap = make(map[string]*actionHandler)
	confRewriteEnabled = true
}

