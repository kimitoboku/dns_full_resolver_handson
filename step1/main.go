package main

import (
	"flag"
	"github.com/miekg/dns"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

var (
	port = flag.Int("port", 15353, "Run DNS port")
	host = flag.String("host", "0.0.0.0", "Run DNS host")
)

func serveDNS(server *dns.Server) {
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func startDNSServer() {
	dns.HandleFunc(".", handlerQuery)
	udpServer := &dns.Server{
		Addr: *host + ":" + strconv.Itoa(*port),
		Net:  "udp",
	}
	defer udpServer.Shutdown()
	go serveDNS(udpServer)

	tcpServer := &dns.Server{
		Addr: *host + ":" + strconv.Itoa(*port),
		Net:  "tcp",
	}
	defer tcpServer.Shutdown()
	go serveDNS(tcpServer)

}

func handlerQuery(w dns.ResponseWriter, r *dns.Msg) {
	q := r.Question[0]
	log.Printf("QName: %s, Qtype: %d", q.Name, q.Qtype)

	rr := IterativeSearch(q.Name, q.Qtype)

	// set Query ID
	rr.Id = r.Id

	// is recursion query response
	rr.RecursionDesired = true
	rr.RecursionAvailable = true
	// is not handle DNSSEC
	rr.AuthenticatedData = false
	// is DNS Cache response
	rr.Authoritative = false
	// is Query response
	rr.Response = true

	err := w.WriteMsg(rr)
	if err != nil {
		log.Println(err)
	}
}

func queryIterative(nameserver string, qname string, qtype uint16) *dns.Msg {
	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(qname), qtype)

	r, _, err := c.Exchange(m, net.JoinHostPort(nameserver, "53"))
	if err != nil {
		log.Println(err.Error())
		r.MsgHdr.Rcode = dns.RcodeRefused
	}

	return r
}

func getNameServer(msg *dns.Msg) ([]string, bool) {
	var nameservers []string
	for _, ns := range msg.Ns {
		nsRr := ns.(*dns.NS)
		nameservers = append(nameservers, nsRr.Ns)
	}

	if len(msg.Extra) == 0 {
		return nameservers, false
	}
	var ipaddress []string
	for _, extra := range msg.Extra {
		if extra.Header().Rrtype == dns.TypeA {
			aRr := extra.(*dns.A)
			ipAddr := aRr.A.String()
			ipaddress = append(ipaddress, ipAddr)
		}
	}
	if len(ipaddress) == 0 {
		return nameservers, false
	}

	return ipaddress, true
}

func IterativeSearch(qname string, qtype uint16) *dns.Msg {
	rootNs := "202.12.27.33"

	ns := rootNs

	for {
		resp := queryIterative(ns, qname, qtype)
		if resp.MsgHdr.Rcode == dns.RcodeSuccess {
			if resp.MsgHdr.Authoritative {
				return resp
			}
			nameservers, hasIpAddr := getNameServer(resp)
			if hasIpAddr {
				ns = nameservers[0]
			} else {
				nsArr := IterativeSearch(nameservers[0], dns.TypeA)
				ns = nsArr.Answer[0].(*dns.A).A.String()
			}
		} else {
			return resp
		}
	}
	return nil
}

func main() {
	startDNSServer()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit
}
