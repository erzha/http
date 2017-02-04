//测试用的

package main

import (
	"encoding/json"
	"errors"
	"net"
	"time"

	"context"
	"github.com/erzha/elog"
	"github.com/erzha/http/server"
	"github.com/erzha/kernel"

	"github.com/miekg/dns"
)

//main.go
func main() {
	kernel.RegisterPlugin("appInit", RegisterPlugin_APP())
	server.Router("index", "index", func() server.ActionInterface { return &IndexAction{} })
	server.Boot()
}

//app/app.go
var appSysLogger *elog.Logger

func RegisterPlugin_APP() kernel.PluginInfo {
	info := kernel.PluginInfo{}
	info.ServerInit = Plugin_App_ServerInit
	return info
}

func Plugin_App_ServerInit(ctx context.Context, pServer *kernel.Server) error {
	appSysLogger = pServer.Logger
	return nil
}

//app/model/dnsutil.go
type DNSClient struct {
	ns             string
	recursionCount int
}

func newDnsClient(ns string) *DNSClient {
	client := &DNSClient{}
	client.ns = ns
	return client
}

func (client *DNSClient) LookupENDS0_A_Record(requestDomain string, clientip string) ([]string, error) {

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
	nsServer := client.ns
	msg := new(dns.Msg)
	msg.Id = dns.Id()
	msg.RecursionDesired = true
	msg.Question = make([]dns.Question, 1)
	msg.Question[0] = dns.Question{requestDomain, dns.TypeA, dns.ClassINET}
	msg.Extra = append(msg.Extra, dnsOption)

	in, err := dns.Exchange(msg, nsServer)
	if nil != err {
		return nil, err
	}

	//make a record json
	jData := make(map[string][]string, 16)
	for index := range in.Answer {
		record := in.Answer[index]
		switch realHeader := record.(type) {
		case *dns.A:
			jData["A"] = append(jData["A"], realHeader.A.String())
		case *dns.CNAME:
			jData["CNAME"] = append(jData["Target"], realHeader.Target)
		}
	}

	appSysLogger.Info("query_domain NS:", client.ns, " domain:", requestDomain, " A:", jData["A"], " CNAME:", jData["CNAME"])

	if len(jData["A"]) == 0 && len(jData["CNAME"]) > 0 {
		//if  len(jData["CNAME"]) > 0 {
		for index := range jData["CNAME"] {
			r, e := client.LookupENDS0_A_Record(jData["CNAME"][index], clientip)
			if nil != e && len(r) > 0 {
				jData["A"] = append(jData["A"], r...)
			}
		}
	}

	return jData["A"], nil
}

//app/action/IndexAction.go
type IndexAction struct {
	server.Action
	sapi  *server.Sapi
	start time.Time
	err   error
}

func (action *IndexAction) Init(ctx context.Context, isapi *server.Sapi) error {
	action.sapi = isapi
	action.start = time.Now()
	return nil
}

func (action *IndexAction) Response(data interface{}) {
	r := map[string]interface{}{}
	r["time_cost_ms"] = time.Now().Sub(action.start).Nanoseconds() / 1000 / 1000
	if nil != action.err {
		r["error"] = action.err
	}
	r["data"] = data

	j, _ := json.Marshal(r)
	action.sapi.Res.Write(j)
}

func (action *IndexAction) Execute(ctx context.Context, sapi *server.Sapi) {
	//Parse Request
	returnData := map[string]string{}

	requestDomain := sapi.Form.Get("domain")
	nsServer := sapi.Form.Get("ns")
	clientip := sapi.Form.Get("clientip")

	if "" == requestDomain {
		action.err = errors.New("empty domain")
		action.Response(returnData)
		return
	}

	if "" == nsServer {
		nsServer = "8.8.8.8:53" //Google DNS Server
	}

	if "" == clientip {
		clientip, _, _ = net.SplitHostPort(sapi.Req.RemoteAddr)
	}

	jData := make(map[string]interface{}, 16)
	dnsclient := newDnsClient(nsServer)
	dnsResult, err := dnsclient.LookupENDS0_A_Record(requestDomain, clientip)
	if nil != err {
		action.err = err
	}

	if dnsResult != nil {
		jData["ip"] = dnsResult
	}
	action.Response(jData)
}
