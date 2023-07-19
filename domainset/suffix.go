package domainset

import (
	"bytes"
	"fmt"
	"strings"
)

type Suffix struct {
	m map[string]*struct{}
}

func NewSuffix() *Suffix {
	return &Suffix{m: make(map[string]*struct{})}
}

func (s *Suffix) Type() string {
	return RuleTypeSuffix
}

func (s *Suffix) Insert(rule string) error {
	if !strings.HasPrefix(rule, ".") {
		rule = "." + rule
	}
	_, exist := s.m[rule]
	if !exist {
		s.m[rule] = (*struct{})(nil)
	}
	return nil
}

func (s *Suffix) Match(domain string) bool {
	if !strings.HasPrefix(domain, ".") {
		domain = "." + domain
	}
	for rule := range s.m {
		if strings.HasSuffix(domain, rule) {
			return true
		}
	}
	return false
}

func (s *Suffix) MarshalBinary() ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	for rule := range s.m {
		ruleLen := len(rule)
		if ruleLen > 255 {
			return nil, fmt.Errorf("rule length %d is too long", ruleLen)
		}
		buffer.WriteByte(byte(ruleLen))
		buffer.WriteString(rule)
	}
	return buffer.Bytes(), nil
}

func (s *Suffix) UnmarshalBinary(data []byte) error {
	s.m = make(map[string]*struct{})
	buffer := bytes.NewBuffer(data)
	for buffer.Len() > 0 {
		ruleLen, err := buffer.ReadByte()
		if err != nil {
			return err
		}
		rule := buffer.Next(int(ruleLen))
		s.m[string(rule)] = (*struct{})(nil)
	}
	return nil
}
