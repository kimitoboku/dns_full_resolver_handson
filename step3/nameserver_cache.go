package main

import (
	"fmt"
	"log"
	"strings"
	"time"
)

type CacheNameserver struct {
	Expire      time.Time
	Nameservers []string
	Ipaddress   []string
}

type NsCacheKey struct {
	qname string
}

type NsCache struct {
	cache map[NsCacheKey]CacheNameserver
}

func (c *NsCache) Get(qname string) ([]string, []string, error) {
	dnsLevels := strings.Split(qname, ".")
	for i := 0; i < len(dnsLevels)-1; i++ {
		keyDomain := strings.Join(dnsLevels[i:len(dnsLevels)], ".")
		d, ok := c.cache[NsCacheKey{
			qname: keyDomain,
		}]
		if !ok {
			continue
		}
		now := time.Now()
		sub := d.Expire.Sub(now)
		ttl := int(sub.Seconds())
		log.Println(ttl)
		if ttl < 1 {
			log.Printf("data is expired in cache: %s", qname)
			continue
		}

		log.Printf("Fetch Ns DNS Cache %s", qname)

		return d.Nameservers, d.Ipaddress, nil
	}
	return nil, nil, fmt.Errorf("Data is not found")
}

func (c *NsCache) Add(keyDomain string, nameservers []string, ipadress []string, ttl uint32) {
	log.Printf("Add NS Cache: %s -> %v, %v", keyDomain, nameservers, ipadress)
	expiredTime := time.Now().Add(time.Second * time.Duration(ttl))
	cacheData := CacheNameserver{
		Nameservers: nameservers,
		Ipaddress:   ipadress,
		Expire:      expiredTime,
	}
	c.cache[NsCacheKey{
		qname: keyDomain,
	}] = cacheData
}

func NewNsCache() *NsCache {
	cache := &NsCache{
		cache: map[NsCacheKey]CacheNameserver{},
	}
	return cache
}
