package main

import (
	"github.com/miekg/dns"
	"log"
	"net"
)

type FullResolver struct {
	dataCache *DataCache
	nsCache   *NsCache
}

func NewFullReoslver() *FullResolver {
	return &FullResolver{
		dataCache: NewDataCache(),
		nsCache:   NewNsCache(),
	}
}

func (f *FullResolver) queryIterative(nameserver string, qname string, qtype uint16) *dns.Msg {
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

func (f *FullResolver) getNameServer(msg *dns.Msg) ([]string, bool) {
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

func (f *FullResolver) AddNsCache(msg *dns.Msg) {
	var nameservers []string
	for _, ns := range msg.Ns {
		nsRr := ns.(*dns.NS)
		nameservers = append(nameservers, nsRr.Ns)
	}
	var ipaddress []string
	for _, extra := range msg.Extra {
		if extra.Header().Rrtype == dns.TypeA {
			aRr := extra.(*dns.A)
			ipAddr := aRr.A.String()
			ipaddress = append(ipaddress, ipAddr)
		}
	}
	keyDomain := msg.Ns[0].Header().Name
	ttl := msg.Ns[0].Header().Ttl
	log.Printf("Add NS Cache %s %d", keyDomain, ttl)
	f.nsCache.Add(keyDomain, nameservers, ipaddress, ttl)
}

func (f *FullResolver) IterativeSearch(qname string, qtype uint16) *dns.Msg {
	resp, err := f.dataCache.Get(qname, qtype)
	if err == nil {
		return &resp
	}

	rootNs := "202.12.27.33"
	var ns string
	nameservers, ipaddress, err := f.nsCache.Get(qname)
	if err == nil {
		if len(ipaddress) == 0 {
			nsArr := f.IterativeSearch(nameservers[0], dns.TypeA)
			ns = nsArr.Answer[0].(*dns.A).A.String()
		} else {
			ns = ipaddress[0]
		}
	} else {
		ns = rootNs
	}

	for {
		resp := f.queryIterative(ns, qname, qtype)
		if resp.MsgHdr.Rcode == dns.RcodeSuccess {
			if resp.MsgHdr.Authoritative {
				f.dataCache.Add(qname, qtype, *resp)
				return resp
			}
			f.AddNsCache(resp)
			nameservers, hasIpAddr := f.getNameServer(resp)
			if hasIpAddr {
				ns = nameservers[0]
			} else {
				nsArr := f.IterativeSearch(nameservers[0], dns.TypeA)
				ns = nsArr.Answer[0].(*dns.A).A.String()
			}
		} else {
			f.dataCache.Add(qname, qtype, *resp)
			return resp
		}
	}
	return nil
}
