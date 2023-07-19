package addrset

import (
	"bytes"
	"fmt"
	"net/netip"
	"strings"
)

func ParseContent(content []byte) (*map[string][]netip.Prefix, error) {
	lines := bytes.Split(content, []byte("\n"))
	m := new(map[string][]netip.Prefix)
	*m = make(map[string][]netip.Prefix)
	var tag string
	var hasRule bool
	var rules []netip.Prefix
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if string(line) == "" {
			continue
		}
		b := bytes.SplitN(line, []byte(" "), 2)[0]
		if strings.HasPrefix(string(b), "#") {
			continue
		}
		if strings.HasPrefix(string(b), "[") {
			if !strings.HasSuffix(string(b), "]") {
				return nil, fmt.Errorf("invalid format")
			}
			if hasRule {
				(*m)[tag] = rules
				rules = nil
				hasRule = false
			}
			tag = string(b[1 : len(b)-1])
			continue
		}
		ip, err := netip.ParseAddr(string(b))
		if err == nil {
			hasRule = true
			bits := 32
			if ip.Is6() {
				bits = 128
			}
			rules = append(rules, netip.PrefixFrom(ip, bits))
		}
		prefix, err := netip.ParsePrefix(string(b))
		if err == nil {
			hasRule = true
			rules = append(rules, prefix)
		}
	}
	if hasRule {
		(*m)[tag] = rules
	}
	return m, nil
}

func Merge(ms ...*map[string][]netip.Prefix) *map[string][]netip.Prefix {
	m := new(map[string][]netip.Prefix)
	*m = make(map[string][]netip.Prefix)
	for _, mm := range ms {
		for tag, p := range *mm {
			if _, exist := (*m)[tag]; !exist {
				(*m)[tag] = make([]netip.Prefix, 0)
			}
			(*m)[tag] = append((*m)[tag], p...)
		}
	}
	return m
}
