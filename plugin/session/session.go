package session

import (
	"encoding/gob"
	"math/rand"
	"net/http"
	"strconv"
	"bytes"
	"time"

	"github.com/erzha/kernel"
	"github.com/erzha/http/server"

	"golang.org/x/net/context"
)

func uuid() string {
	now := time.Now()
	unixtimestamp := now.Unix()
	rand.Seed(unixtimestamp)

	pre := strconv.FormatInt(unixtimestamp, 36)
	suf := strconv.FormatInt(rand.Int63(), 36)
	return pre + suf
}

var sessionHandler Handler

type Session struct {
	httpsapi *server.Sapi
	id         string
	hasStarted bool
	h Handler
}

func (s *Session) Id() string {
	if "" == s.id {
		s.id = uuid()
	}
	return s.id
}

func (s *Session) Start() {
	if s.hasStarted {
		return
	}

	s.hasStarted = true
	id := s.httpsapi.Cookie("SID")

	if id == "" {
		id = uuid()
		newcookie := http.Cookie{Name: "SID", Value: id}
		s.httpsapi.SetCookie(&newcookie)
	}
	s.id = id
}

func (s *Session) Get(key string, dst interface{}) bool {
	valInStore, ok := sessionHandler.Get(s.id, key)
	if false == ok || nil != gob.NewDecoder(bytes.NewReader(valInStore)).Decode(dst) {
		return false
	}
	return true

}

func (s *Session) Set(key string, value interface{}) bool {
	buf := new(bytes.Buffer)
	if nil != gob.NewEncoder(buf).Encode(value) {
		return false
	}
	return sessionHandler.Set(s.id, key, buf.Bytes())
}

func (s *Session) Del(key string) bool {
	return sessionHandler.Del(s.id, key)
}

func (s *Session) Destory() bool {
	return sessionHandler.Destory(s.id)
}

func sessionCreater() (interface{}, error) {
	return &Session{hasStarted: false}, nil
}

func serverInit(ctx context.Context, s *kernel.Server) error {
	if nil == sessionHandler {
		sessionHandler = newDefaultHandler()
	}
	sessionHandler.SetExpireTime(1200)
	sessionHandler.Start()
	return nil
}

func serverShutdown(ctx context.Context, s *kernel.Server) error {
	sessionHandler.Stop()
	return nil
}

func requestInit(ctx context.Context, sapi *kernel.Sapi, obj interface{}) error {
	session := obj.(*Session)
	session.httpsapi = sapi.Ext.(*server.Sapi)
	return nil
}

func newInfo() kernel.PluginInfo {
	info := kernel.PluginInfo{}
	info.Creater = sessionCreater
	info.RequestInit = requestInit
	info.ServerInit = serverInit
	info.ServerShutdown = serverShutdown
	return info
}

func RegisterPlugin() {
	kernel.RegisterPlugin("session", newInfo())
}
