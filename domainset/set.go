package domainset

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

func NewDomainSet() *DomainSet {
	return &DomainSet{m: make(map[string]DomainSetRule)}
}

func (d *DomainSet) Insert(tag string, rule string) error {
	_, exist := d.m[tag]
	if !exist {
		d.m[tag] = make(DomainSetRule)
	}
	setRule := d.m[tag]
	return setRule.NewRuleItem(rule)
}

func (d *DomainSet) InsertMap(tag string, m DomainSetRule) {
	_, exist := d.m[tag]
	if !exist {
		d.m[tag] = make(DomainSetRule)
	}
	for k, v := range m {
		d.m[tag][k] = v
	}
}

func (d *DomainSet) Match(domain string, tags ...string) bool {
	if tags == nil {
		return false
	}
	for _, tag := range tags {
		rules, exist := d.m[tag]
		if !exist {
			continue
		}
		for _, rule := range rules {
			if rule.Match(domain) {
				return true
			}
		}
	}
	return false
}

func (d *DomainSet) MarshalBinary() ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	encoder := gob.NewEncoder(buffer)
	m := make(map[string]map[string][]byte)
	for tag, rule := range d.m {
		m[tag] = make(map[string][]byte)
		for _, ruleItem := range rule {
			b, err := ruleItem.MarshalBinary()
			if err != nil {
				return nil, err
			}
			m[tag][ruleItem.Type()] = b
		}
	}
	err := encoder.Encode(m)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (d *DomainSet) UnmarshalBinary(data []byte) error {
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)
	m := make(map[string]map[string][]byte)
	err := decoder.Decode(&m)
	if err != nil {
		return err
	}
	d.m = make(map[string]DomainSetRule)
	for tag, rule := range m {
		d.m[tag] = make(DomainSetRule)
		for _type, ruleBytes := range rule {
			var r DomainSetRuleItem
			switch _type {
			case RuleTypeFull:
				r = NewFull()
			case RuleTypeSuffix:
				r = NewSuffix()
			case RuleTypeKeyword:
				r = NewKeyword()
			case RuleTypeRegex:
				r = NewRegex()
			default:
				return fmt.Errorf("unknown type: %s", _type)
			}
			err = r.UnmarshalBinary(ruleBytes)
			if err != nil {
				return err
			}
			d.m[tag][_type] = r
		}
	}
	return nil
}
