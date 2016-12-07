//测试用的

package main

import (
	"net"
	"encoding/json"

	"github.com/erzha/http/server"
	"golang.org/x/net/context"

	"github.com/miekg/dns"
)


//main.go
func main() {
	initApp()
}


//app/app.go
func initApp() {
	server.Router("index", "index", func()server.ActionInterface{return &IndexAction{} })
	server.Boot()
}


//app/action/IndexAction.go
type IndexAction struct {
	server.Action
	sapi *server.Sapi
}

func (action *IndexAction) Init(ctx context.Context, isapi *server.Sapi) error {
	action.sapi = isapi
	return nil
}

func (action *IndexAction) Response(data interface{}) {
	j, _ := json.Marshal(data)
	action.sapi.Res.Write(j)
}


func (action *IndexAction) Execute(ctx context.Context, sapi *server.Sapi) { 

	//Parse Request
	returnData := map[string]string{}

	requestDomain := sapi.Form.Get("domain")
	nsServer := sapi.Form.Get("ns")
	clientip := sapi.Form.Get("clientip")
	
	if "" == requestDomain {
		returnData["err"] = "empty domain"
		action.Response(returnData)
		return
	}

	if "" == nsServer {
		nsServer = "8.8.8.8:53" //Google DNS Server
	}

	if "" == clientip {
		clientip, _, _ = net.SplitHostPort(sapi.Req.RemoteAddr)
	}


	//DNS协议的EDNS特性(draft-ietf-dnsop-edns-client-subnet-00)
	edns := new(dns.EDNS0_SUBNET)
	edns.Code = dns.EDNS0SUBNET
	edns.Family = 1
	edns.SourceNetmask = 32
	edns.Address = net.ParseIP(clientip).To4()
	
	dnsOption := new(dns.OPT)
	dnsOption.Hdr.Name = "."
	dnsOption.Hdr.Rrtype = dns.TypeOPT
	dnsOption.Option = append(dnsOption.Option, edns)


	//Do DNS Query
	msg := new(dns.Msg)
	msg.Id = dns.Id()
	msg.RecursionDesired = true
	msg.Question = make([]dns.Question, 1)
	msg.Question[0] = dns.Question{requestDomain, dns.TypeA, dns.ClassINET}
	msg.Extra = append(msg.Extra, dnsOption)
	
	in, err := dns.Exchange(msg, nsServer)
	if (nil != err) {
		returnData["err"] = err.Error()
		action.Response(returnData)
		return
	}

	//make a record json
	jData := make(map[string][]string, 16)
	for index := range in.Answer {
		record := in.Answer[index]
		switch realHeader := record.(type) {
			case *dns.A: jData["A"] = append(jData["A"], realHeader.A.String())
			case *dns.CNAME: jData["CNAME"] = append(jData["Target"], realHeader.Target)
		}
	}
	action.Response(jData)
}