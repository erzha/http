// Copyright 2014 The erzha Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package server

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"

	"github.com/erzha/kernel"
)

type Sapi struct {
	Kernel *kernel.Sapi

	Res    http.ResponseWriter
	Req    *http.Request
	Status int

	Get  url.Values
	Post url.Values
	Form url.Values
	File ParamFiles

	handler       *actionHandler
	chExitRequest chan bool
}

func (p *Sapi) RequestURI() string {
	if nil == p.Req {
		return ""
	}
	return p.Req.URL.Path
}

func (p *Sapi) Cookie(key string) string {
	c, e := p.Req.Cookie(key)
	if nil != e {
		return ""
	}
	return c.Value
}

func (p *Sapi) Header() http.Header {
	return p.Res.Header()
}

func (p *Sapi) Redirect(url string) {
	p.Header().Add("Location", url)
	p.Res.WriteHeader(302)
}

func (p *Sapi) SetCookie(cookie *http.Cookie) {
	http.SetCookie(p.Res, cookie)
}

func NewSapi(res http.ResponseWriter, req *http.Request) *Sapi {

	ret := &Sapi{}
	ret.Res = res
	ret.Req = req
	ret.Status = 200
	ret.chExitRequest = make(chan bool)

	ret.Get = url.Values{}
	ret.Post = url.Values{}
	ret.File = ParamFiles{}

	req.ParseForm()
	req.ParseMultipartForm(10 * 1024 * 1024)
	ret.Get, _ = url.ParseQuery(req.URL.RawQuery)
	if nil != req.MultipartForm {
		for key, val := range req.MultipartForm.Value {
			for _, v := range val {
				ret.Post.Add(key, v)
			}
		}
		ret.File = ParamFiles(req.MultipartForm.File)

	} else {
		ret.Post = req.PostForm
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
