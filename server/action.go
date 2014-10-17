package server

import (
	"strings"
	"errors"

	"github.com/erzha/kernel"

	"code.google.com/p/go.net/context"
)


var (
	confDefaultAction string
	confRewriteEnabled bool
)


type ActionInterface interface {
	Execute(ctx context.Context, sapi *kernel.Sapi)
}

type Action struct {
}

func (action *Action) Execute(ctx context.Context, sapi *kernel.Sapi) {

}


type ActionSpec struct {
	Action
}

func (action *ActionSpec) Execute(ctx context.Context, sapi *kernel.Sapi) {

}

func (action *ActionSpec) DoGet(ctx context.Context, sapi *kernel.Sapi) {

}

func (action *ActionSpec) DoPost(ctx context.Context, sapi *kernel.Sapi) {

}

var actionMap map[string]ActionInterface

func Router(url string, name string, obj ActionInterface) {
	routerAddRule(url, name)
	actionMap[name] = obj
}

func do(ctx context.Context, sapi *kernel.Sapi) {
}

func InitHttpRequest(sapi *kernel.Sapi) error  {

	//get action name from requesturi
	var httpsapi *Sapi
	var actionName, uri string
	var param map[string]string
	var ok bool

	httpsapi = sapi.Ext.(*Sapi)

	uri = httpsapi.RequestURI()
	uri = strings.TrimLeft(uri, "/\\")

	if uri == "" {
		actionName = confDefaultAction
	} else if confRewriteEnabled {
		actionName, param = urlToAction(uri)
		if "" != actionName && nil != param {
			for key, val := range param {
				httpsapi.GET.Set(key, val)
			}
		}
	}

	if "" == actionName {
		actionName = uri
	}

	httpsapi.actionObj, ok = actionMap[actionName]
	if !ok {
		return errors.New("")
	}
	return nil
}

func init() {
	actionMap = make(map[string]ActionInterface)
}
