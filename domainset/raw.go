package domainset

import (
	"bytes"
	"fmt"
	"strings"
)

func ParseContent(content []byte) (*map[string]DomainSetRule, error) {
	lines := bytes.Split(content, []byte("\n"))
	m := new(map[string]DomainSetRule)
	*m = make(map[string]DomainSetRule)
	var tag string
	var hasRule bool
	rules := new(DomainSetRule)
	*rules = make(DomainSetRule)
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
				if _, exist := (*m)[tag]; exist {
					rule := (*m)[tag]
					err := rule.Merge(rules)
					if err != nil {
						return nil, err
					}
				} else {
					(*m)[tag] = *rules
				}
				rules = new(DomainSetRule)
				*rules = make(DomainSetRule)
				hasRule = false
			}
			tag = string(b[1 : len(b)-1])
			continue
		}
		err := rules.NewRuleItem(string(b))
		if err == nil {
			hasRule = true
		}
	}
	if hasRule {
		if _, exist := (*m)[tag]; exist {
			rule := (*m)[tag]
			err := rule.Merge(rules)
			if err != nil {
				return nil, err
			}
		} else {
			(*m)[tag] = *rules
		}
	}
	return m, nil
}

func Merge(ms ...*map[string]DomainSetRule) *map[string]DomainSetRule {
	if ms == nil {
		return nil
	}
	m := new(map[string]DomainSetRule)
	*m = make(map[string]DomainSetRule)
	for _, mm := range ms {
		for tag, rule := range *mm {
			if _, exist := (*m)[tag]; exist {
				r := (*m)[tag]
				err := r.Merge(&rule)
				if err != nil {
					continue
				}
				(*m)[tag] = r
			} else {
				(*m)[tag] = rule
			}
		}
	}
	return m
}
