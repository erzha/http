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

	Get	url.Values
	Post	url.Values
	Form	url.Values
	File	ParamFiles

	actionObj ActionInterface
}

func (p *Sapi) RequestURI() string {
	if nil == p.Req {
		return ""
	}
	return p.Req.URL.Path
}

func (p *Sapi) Cookie(key string)  string {
	c, e := p.Req.Cookie(key)
	if nil!=e {
		return ""
	}
	return c.Value;
}

func (p *Sapi) Header() http.Header {
	return p.Res.Header
}

func (p *Sapi) SetCookie(cookie *http.Cookie) {
	http.SetCookie(p.Res, cookie)
}

func NewSapi(res http.ResponseWriter, req *http.Request) *Sapi {

	ret := &Sapi{}
	ret.Res = res
	ret.Req = req
	ret.Status = 200

	p.Get = url.Values{}
	p.Post = url.Values{}
	p.File = ParamFiles{}

	ret.Get, _ = url.ParseQuery(req.URL.RawQuery)
	if nil != req.MultipartForm {
		for key, val := range req.MultipartForm.Value {
			for _, v := range val {
				p.Post.Add(key, v)
			}
		}
		p.File = ParamFiles(sapi.Req.MultipartForm.File)

	} else {
		p.Post = req.PostForm
	}
	ret.Form = req.Form

	return ret
}

type ParamFiles map[string][]*multipart.FileHeader

//获取请求中名字为key的第一个文件
func (pf ParamFiles) Get(key string) (f multipart.File, name string, err error) {
	fhList, ok := pf[key]
	if !ok || len(fhList) <= 0 {
		err = errors.New("none file named " + key)
		return
	}

	name = fhList[0].Filename
	f, err = fhList[0].Open()
	return
}

//将上传的文件move到指定位置
func (pf ParamFiles) Move(key string, path string) error {
	var f multipart.File
	var err error
	var pFile *os.File

	f, _, err = pf.Get(key)
	if nil != err {
		return err
	}

	pFile, err = os.Create(path)
	if nil != err {
		return err
	}

	_, err = io.Copy(pFile, f)
	if nil != err {
		return err
	}

	return nil

}
