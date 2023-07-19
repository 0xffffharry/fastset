package addrset

import (
	"bytes"
	"net"
	"net/http"
	"net/netip"
	"os"
	"strings"

	"github.com/yaotthaha/fastset/addrset"
	"github.com/yaotthaha/go-log"

	"github.com/oschwald/geoip2-golang"
	"github.com/oschwald/maxminddb-golang"
)

const filePath = "/workdir/fastset/addr.set"

const GeoCountryUrl = "https://github.com/Loyalsoldier/geoip/blob/release/Country.mmdb?raw=true"

var codes = []string{
	"cn",
	"hk",
	"mo",
	"tw",
	"sg",
	"jp",
	"kr",
	"my",
	"us",
	"ca",
	"au",
	"de",
	"fr",
	//
	"private",
	"cloudflare",
	"cloudfront",
	"facebook",
	"fastly",
	"netflix",
	"google",
	"telegram",
	"twitter",
}

type extAddFunc func() (map[string][]netip.Prefix, error)

var extAddrFuncSlice = []extAddFunc{}

var logger log.Logger

func init() {
	simpleLogger := log.NewSimpleLogger()
	simpleLogger.SetOutput(os.Stdout)
	simpleLogger.SetLevel(log.LevelDebug)
	simpleLogger.SetFormatFunc(log.DefaultFormatFunc)
	simpleLogger.SetErrOutput(os.Stderr)
	simpleLogger.SetErrDone(true)
	logger = simpleLogger
}

func BuildAddrSet() {
	resp, err := http.Get(GeoCountryUrl)
	if err != nil {
		logger.Fatal(err)
		return
	}
	logger.Infof("request success, code: %d", resp.StatusCode)
	buffer := bytes.NewBuffer(nil)
	_, err = buffer.ReadFrom(resp.Body)
	if err != nil {
		logger.Fatal(err)
		return
	}
	_ = resp.Body.Close()
	logger.Infof("read body success, size: %d", buffer.Len())
	//
	mmdbReader, err := maxminddb.FromBytes(buffer.Bytes())
	if err != nil {
		logger.Fatal(err)
		return
	}
	networks := mmdbReader.Networks(maxminddb.SkipAliasedNetworks)
	m := make(map[string][]netip.Prefix)
	var country geoip2.Enterprise
	var ipNet *net.IPNet
	for networks.Next() {
		ipNet, err = networks.Network(&country)
		if err != nil {
			logger.Fatalf("read network error: %s", err)
			return
		}
		var code string
		if country.Country.IsoCode != "" {
			code = strings.ToLower(country.Country.IsoCode)
		} else if country.RegisteredCountry.IsoCode != "" {
			code = strings.ToLower(country.RegisteredCountry.IsoCode)
		} else if country.RepresentedCountry.IsoCode != "" {
			code = strings.ToLower(country.RepresentedCountry.IsoCode)
		} else if country.Continent.Code != "" {
			code = strings.ToLower(country.Continent.Code)
		} else {
			continue
		}
		if !Contains(codes, code) {
			continue
		}
		prefix, err := netip.ParsePrefix(ipNet.String())
		if err != nil {
			logger.Fatalf("parse prefix error: %s", err)
			return
		}
		m[code] = append(m[code], prefix)
	}
	_ = mmdbReader.Close()
	logger.Info("read mmdb success")
	if extAddrFuncSlice != nil {
		for _, f := range extAddrFuncSlice {
			_m, err := f()
			if err != nil {
				logger.Fatal(err)
				return
			}
			if _m == nil {
				continue
			}
			for tag, p := range _m {
				if _, exist := m[tag]; !exist {
					m[tag] = make([]netip.Prefix, 0)
				}
				m[tag] = append(m[tag], p...)
			}
		}
	}
	//
	logger.Info("build addr set...")
	tree, err := addrset.NewAddrSetTree()
	if err != nil {
		logger.Fatal(err)
		return
	}
	err = tree.InsertMap(m)
	if err != nil {
		logger.Fatal(err)
		return
	}
	treeBuffer := bytes.NewBuffer(nil)
	_, err = tree.WriteTo(treeBuffer)
	if err != nil {
		logger.Fatal(err)
		return
	}
	logger.Info("build addr set success")
	err = os.WriteFile(filePath, treeBuffer.Bytes(), 0o644)
	if err != nil {
		logger.Fatal(err)
		return
	}
	logger.Info("write addr set file success")
}
