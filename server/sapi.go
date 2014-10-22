package server

import (
	"net/http"
	"net/url"

	"github.com/erzha/kernel"
)


type Sapi struct {

	Kernel *kernel.Sapi

	Res http.ResponseWriter
	Req *http.Request
	Status int

	GET	url.Values
	POST	url.Values
	FORM	url.Values

	actionObj ActionInterface
}

func (p *Sapi) RequestURI() string {
	if nil == p.Req {
		return ""
	}
	return p.Req.URL.Path
}

func (p *Sapi) Cookie(key string) string {
	return p.Req.Cookie(key)
}

func (p *Sapi) SetCookie(cookie *http.Cookie) {
	http.SetCookie(p.Res, cookie)
}

func NewSapi(res http.ResponseWriter, req *http.Request) *Sapi {

	ret := &Sapi{}
	ret.Res = res
	ret.Req = req
	ret.Status = 200

	return ret
}
