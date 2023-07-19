package domainset

import (
	"fmt"
	"strings"
	"sync"
)

type DomainSet struct {
	m     map[string]DomainSetRule
	cache sync.Map
}

type DomainSetRule map[string]DomainSetRuleItem

type DomainSetRuleItem interface {
	Type() string
	Insert(rule string) error
	Match(domain string) bool
	MarshalBinary() ([]byte, error)
	UnmarshalBinary(data []byte) error
}

const (
	RuleTypeFull    = "full"
	RuleTypeSuffix  = "suffix"
	RuleTypeKeyword = "keyword"
	RuleTypeRegex   = "regex"
)

func (d *DomainSetRule) NewRuleItemWithType(_type string, rule string) error {
	_, exist := (*d)[_type]
	if !exist {
		switch _type {
		case RuleTypeFull:
			(*d)[_type] = NewFull()
		case RuleTypeSuffix:
			(*d)[_type] = NewSuffix()
		case RuleTypeKeyword:
			(*d)[_type] = NewKeyword()
		case RuleTypeRegex:
			(*d)[_type] = NewRegex()
		default:
			return fmt.Errorf("unknown type: %s", _type)
		}
	}
	return (*d)[_type].Insert(rule)
}

func (d *DomainSetRule) NewRuleItem(rule string) error {
	switch {
	case strings.HasPrefix(rule, RuleTypeFull+":"):
		rule = strings.TrimPrefix(rule, RuleTypeFull+":")
		return d.NewRuleItemWithType(RuleTypeFull, rule)
	case strings.HasPrefix(rule, RuleTypeSuffix+":"):
		rule = strings.TrimPrefix(rule, RuleTypeSuffix+":")
		return d.NewRuleItemWithType(RuleTypeSuffix, rule)
	case strings.HasPrefix(rule, RuleTypeKeyword+":"):
		rule = strings.TrimPrefix(rule, RuleTypeKeyword+":")
		return d.NewRuleItemWithType(RuleTypeKeyword, rule)
	case strings.HasPrefix(rule, RuleTypeRegex+":"):
		rule = strings.TrimPrefix(rule, RuleTypeRegex+":")
		return d.NewRuleItemWithType(RuleTypeRegex, rule)
	default:
		return d.NewRuleItemWithType(RuleTypeFull, rule)
	}
}

func (d *DomainSetRule) Merge(others ...*DomainSetRule) error {
	if others == nil {
		return nil
	}
	var err error
	for _, other := range others {
		for _type, items := range *other {
			if _, exist := (*d)[_type]; !exist {
				switch _type {
				case RuleTypeFull:
					(*d)[_type] = NewFull()
				case RuleTypeSuffix:
					(*d)[_type] = NewSuffix()
				case RuleTypeKeyword:
					(*d)[_type] = NewKeyword()
				case RuleTypeRegex:
					(*d)[_type] = NewRegex()
				}
			}
			switch _items := items.(type) {
			case *Full:
				for item := range _items.m {
					err = (*d)[_type].Insert(item)
					if err != nil {
						return err
					}
				}
			case *Suffix:
				for item := range _items.m {
					err = (*d)[_type].Insert(item)
					if err != nil {
						return err
					}
				}
			case *Keyword:
				for item := range _items.m {
					err = (*d)[_type].Insert(item)
					if err != nil {
						return err
					}
				}
			case *Regex:
				for _, item := range _items.arr {
					err = (*d)[_type].Insert(item.String())
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
