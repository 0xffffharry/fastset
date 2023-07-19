package main

import (
	"github.com/yaotthaha/fastset/build/addrset"
	"github.com/yaotthaha/fastset/build/domainset"
)

func main() {
	addrset.BuildAddrSet()
	domainset.BuildDomainSet()
}
