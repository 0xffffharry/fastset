package domainset

import (
	"bytes"
	"net/http"
	"os"
	"strings"

	"github.com/yaotthaha/fastset/domainset"
	"github.com/yaotthaha/go-log"

	"github.com/v2fly/v2ray-core/v5/app/router/routercommon"
	"google.golang.org/protobuf/proto"
)

const filePath = "/workdir/fastset/domain.set"

var codes = []string{
	"cn",
	"gfw",
	"greatfire",
	"microsoft",
	"microsoft-dev",
	"microsoft-pki",
	"github",
	"apple",
	"apple-cn",
	"google",
	"google-cn",
	"private",
	"netflix",
	"disney",
	"amazon",
	"facebook",
	"twitter",
	"youtube",
	"telegram",
	"spotify",
	"category-scholar-!cn",
	"category-scholar-cn",
}

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

const geoSiteUrl = "https://github.com/Loyalsoldier/v2ray-rules-dat/blob/release/geosite.dat?raw=true"

func BuildDomainSet() {
	resp, err := http.Get(geoSiteUrl)
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
	set := domainset.NewDomainSet()
	var geoSiteList routercommon.GeoSiteList
	err = proto.Unmarshal(buffer.Bytes(), &geoSiteList)
	if err != nil {
		logger.Fatal(err)
		return
	}
	for _, geoSiteEntry := range geoSiteList.Entry {
		tag := strings.ToLower(geoSiteEntry.CountryCode)
		if !Contains(codes, tag) {
			continue
		}
		d := make(domainset.DomainSetRule)
		attributes := make(map[string][]*routercommon.Domain)
		for _, domain := range geoSiteEntry.Domain {
			if len(domain.Attribute) > 0 {
				for _, attribute := range domain.Attribute {
					attributes[attribute.Key] = append(attributes[attribute.Key], domain)
				}
			}
			var err error
			switch domain.Type {
			case routercommon.Domain_Plain:
				err = d.NewRuleItemWithType(domainset.RuleTypeKeyword, domain.Value)
			case routercommon.Domain_Regex:
				err = d.NewRuleItemWithType(domainset.RuleTypeRegex, domain.Value)
			case routercommon.Domain_RootDomain:
				err = d.NewRuleItemWithType(domainset.RuleTypeSuffix, domain.Value)
			case routercommon.Domain_Full:
				err = d.NewRuleItemWithType(domainset.RuleTypeFull, domain.Value)
			}
			if err != nil {
				logger.Fatal(err)
				return
			}
		}
		set.InsertMap(tag, d)
		for attribute, attributeEntries := range attributes {
			dAttribute := make(domainset.DomainSetRule)
			for _, domain := range attributeEntries {
				var err error
				switch domain.Type {
				case routercommon.Domain_Plain:
					err = dAttribute.NewRuleItemWithType(domainset.RuleTypeKeyword, domain.Value)
				case routercommon.Domain_Regex:
					err = dAttribute.NewRuleItemWithType(domainset.RuleTypeRegex, domain.Value)
				case routercommon.Domain_RootDomain:
					err = dAttribute.NewRuleItemWithType(domainset.RuleTypeSuffix, domain.Value)
				case routercommon.Domain_Full:
					err = dAttribute.NewRuleItemWithType(domainset.RuleTypeFull, domain.Value)
				}
				if err != nil {
					logger.Fatal(err)
					return
				}
			}
			set.InsertMap(tag+"@"+attribute, dAttribute)
		}
	}
	//
	logger.Info("build domain set...")
	b, err := set.MarshalBinary()
	if err != nil {
		logger.Fatal(err)
		return
	}
	logger.Info("build domain set success")
	err = os.WriteFile(filePath, b, 0o644)
	if err != nil {
		logger.Fatal(err)
		return
	}
	logger.Info("write addr set file success")
}
