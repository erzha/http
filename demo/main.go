package main

import (
	"fmt"
	"github.com/erzha/http/server"

	"code.google.com/p/go.net/context"
)

type IndexAction struct {
	server.Action
}

func (action *IndexAction) Execute(ctx context.Context, sapi *server.Sapi) {
	sapi.Kernel.Print("good")
	n, e := sapi.Res.Write([]byte("good sapi"))
	fmt.Println(n, e)
}

func main() {
	server.Boot()
}

func init() {
	server.Router("/blog", "blog", &IndexAction{})
}
