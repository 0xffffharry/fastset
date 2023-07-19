package domainset

import (
	"bytes"
	"fmt"
)

type Full struct {
	m map[string]*struct{}
}

func NewFull() *Full {
	return &Full{m: make(map[string]*struct{})}
}

func (f *Full) Type() string {
	return RuleTypeFull
}

func (f *Full) Insert(rule string) error {
	_, exist := f.m[rule]
	if !exist {
		f.m[rule] = (*struct{})(nil)
	}
	return nil
}

func (f *Full) Match(domain string) bool {
	_, exist := f.m[domain]
	return exist
}

func (f *Full) MarshalBinary() ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	for rule := range f.m {
		ruleLen := len(rule)
		if ruleLen > 255 {
			return nil, fmt.Errorf("rule length %d is too long", ruleLen)
		}
		buffer.WriteByte(byte(ruleLen))
		buffer.WriteString(rule)
	}
	return buffer.Bytes(), nil
}

func (f *Full) UnmarshalBinary(data []byte) error {
	f.m = make(map[string]*struct{})
	buffer := bytes.NewBuffer(data)
	for buffer.Len() > 0 {
		ruleLen, err := buffer.ReadByte()
		if err != nil {
			return err
		}
		rule := buffer.Next(int(ruleLen))
		f.m[string(rule)] = (*struct{})(nil)
	}
	return nil
}
