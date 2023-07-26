package main

import (
	"fmt"
	"github.com/miekg/dns"
	"log"
	"time"
)

type CacheData struct {
	Expire time.Time
	Data   dns.Msg
}

type CacheKey struct {
	qname string
	qtype uint16
}

type DataCache struct {
	cache map[CacheKey]CacheData
}

func (c *DataCache) Get(qname string, qtype uint16) (dns.Msg, error) {
	d, ok := c.cache[CacheKey{
		qname: qname,
		qtype: qtype,
	}]
	if !ok {
		log.Printf("data is not found in cache: %s", qname)
		return dns.Msg{}, fmt.Errorf("data is not found")
	}
	now := time.Now()
	sub := d.Expire.Sub(now)
	ttl := int(sub.Seconds())
	log.Println(ttl)
	if ttl < 1 {
		log.Printf("data is expired in cache: %s", qname)
		return dns.Msg{}, fmt.Errorf("data is expired")
	}

	log.Printf("Fetch DNS Cache %s", qname)
	for i := range d.Data.Answer {
		d.Data.Answer[i].Header().Ttl = uint32(ttl)
	}
	for i := range d.Data.Ns {
		d.Data.Ns[i].Header().Ttl = uint32(ttl)
	}
	for i := range d.Data.Extra {
		d.Data.Extra[i].Header().Ttl = uint32(ttl)
	}
	return d.Data, nil
}

func (c *DataCache) Add(qname string, qtype uint16, data dns.Msg) {
	var ttl uint32
	if data.MsgHdr.Rcode == dns.RcodeSuccess {
		log.Printf("Success %s %d, minial: %d", qname, qtype, ttl)
		ttl = data.Answer[0].Header().Ttl
	} else if data.MsgHdr.Rcode == dns.RcodeNameError {
		headTtl := data.Ns[0].Header().Ttl
		soa := data.Ns[0].(*dns.SOA)
		minimal := soa.Minttl
		if headTtl < minimal {
			ttl = headTtl
		} else {
			ttl = minimal
		}
		log.Printf("NXRrset %s %d, minial: %d", qname, qtype, ttl)
	} else {
		log.Printf("Can not add new Cache %s, Rcode: %d", qname, data.MsgHdr.Rcode)
		return
	}
	log.Printf("Add new Cache %s", qname)

	expiredTime := time.Now().Add(time.Second * time.Duration(ttl))
	cacheData := CacheData{
		Data:   data,
		Expire: expiredTime,
	}
	c.cache[CacheKey{
		qname: qname,
		qtype: qtype,
	}] = cacheData
}

func NewDataCache() *DataCache {
	cache := &DataCache{
		cache: map[CacheKey]CacheData{},
	}
	return cache
}
