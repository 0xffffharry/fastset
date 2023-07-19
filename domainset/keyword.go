package domainset

import (
	"bytes"
	"fmt"
	"strings"
)

type Keyword struct {
	m map[string]*struct{}
}

func NewKeyword() *Keyword {
	return &Keyword{m: make(map[string]*struct{})}
}

func (k *Keyword) Type() string {
	return RuleTypeKeyword
}

func (k *Keyword) Insert(rule string) error {
	_, exist := k.m[rule]
	if !exist {
		k.m[rule] = (*struct{})(nil)
	}
	return nil
}

func (k *Keyword) Match(domain string) bool {
	for rule := range k.m {
		if strings.Contains(domain, rule) {
			return true
		}
	}
	return false
}

func (k *Keyword) MarshalBinary() ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	for rule := range k.m {
		ruleLen := len(rule)
		if ruleLen > 255 {
			return nil, fmt.Errorf("rule length %d is too long", ruleLen)
		}
		buffer.WriteByte(byte(ruleLen))
		buffer.WriteString(rule)
	}
	return buffer.Bytes(), nil
}

func (k *Keyword) UnmarshalBinary(data []byte) error {
	k.m = make(map[string]*struct{})
	buffer := bytes.NewBuffer(data)
	for buffer.Len() > 0 {
		ruleLen, err := buffer.ReadByte()
		if err != nil {
			return err
		}
		rule := buffer.Next(int(ruleLen))
		k.m[string(rule)] = (*struct{})(nil)
	}
	return nil
}
