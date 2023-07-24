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

	// is recursion query response
	r.RecursionDesired = true
	r.RecursionAvailable = true
	// is not handle DNSSEC
	r.AuthenticatedData = false
	// is DNS Cache response
	r.Authoritative = false
	// is Query response
	r.Response = true

	err := w.WriteMsg(r)
	if err != nil {
		log.Println(err)
	}

}

func main() {
	startDNSServer()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit
}
