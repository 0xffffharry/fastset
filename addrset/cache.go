package addrset

import (
	"net"
	"net/netip"
	"sync"
)

type CacheReader struct {
	*AddrSetReader
	cache sync.Map
}

func NewCacheReader(reader *AddrSetReader) *CacheReader {
	return &CacheReader{
		AddrSetReader: reader,
	}
}

func (r *CacheReader) LookupIP(ip net.IP) ([]string, error) {
	var result []string
	resultAny, exist := r.cache.Load(ip.String())
	if exist {
		result = resultAny.([]string)
		return result, nil
	}
	err := r.Reader.Lookup(ip, &result)
	if err != nil {
		return nil, err
	}
	r.cache.Store(ip.String(), result)
	return result, nil
}

func (r *CacheReader) LookupAddr(ip netip.Addr) ([]string, error) {
	return r.LookupIP(ip.AsSlice())
}
