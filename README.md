This repository is a hands-on implementation for DNS full resolver implements.

## step0 
step0 is an implementation of a DNS server that does nothing.
The DNS server response DNS Query with query response and recursive response flags.

## step1
step1 is simple full resolver implementation.
The DNS server can handle A records.
The DNS server execute iterative search.

## step2
step2 is implementation of data cache.
The DNS server cache user query data response.

## step3
step3 is implementation of nameserver cache
The DNS server cache NS record until iterative search.