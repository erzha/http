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


type ActionSpec struct {
	Action
}

func (action *ActionSpec) Execute(ctx context.Context, sapi *Sapi) {

}

func (action *ActionSpec) DoGet(ctx context.Context, sapi *Sapi) {

}

func (action *ActionSpec) DoPost(ctx context.Context, sapi *Sapi) {

}

var actionMap map[string]func()ActionInterface

func Router(url string, name string, creater func()ActionInterface) {
	routerAddRule(url, name)
	actionMap[name] = creater
}

func do(ctx context.Context, sapi *kernel.Sapi) {
	httpsapi := sapi.Ext.(*Sapi)
	httpsapi.actionObj.Execute(ctx, httpsapi)
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

	var creater func()ActionInterface
	creater, ok = actionMap[actionName]
	if !ok {
		return errors.New("")
	}
	httpsapi.actionObj = creater()
	return nil
}

func init() {
	actionMap = make(map[string]func()ActionInterface)
}

