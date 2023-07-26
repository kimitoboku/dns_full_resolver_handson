package main

import (
	"flag"
	"github.com/miekg/dns"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

var (
	port = flag.Int("port", 15353, "Run DNS port")
	host = flag.String("host", "0.0.0.0", "Run DNS host")
)

var resolver *FullResolver

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

	rr := resolver.IterativeSearch(q.Name, q.Qtype)

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

func main() {
	resolver = NewFullReoslver()
	startDNSServer()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit
}
