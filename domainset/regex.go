package domainset

import (
	"bytes"
	"fmt"

	"github.com/dlclark/regexp2"
)

type Regex struct {
	arr []*regexp2.Regexp
}

func NewRegex() *Regex {
	return &Regex{}
}

func (r *Regex) Type() string {
	return RuleTypeRegex
}

func (r *Regex) Insert(rule string) error {
	re, err := regexp2.Compile(rule, regexp2.RE2)
	if err != nil {
		return err
	}
	r.arr = append(r.arr, re)
	return nil
}

func (r *Regex) Match(domain string) bool {
	var (
		match bool
		err   error
	)
	for _, rule := range r.arr {
		match, err = rule.MatchString(domain)
		if err == nil && match {
			return true
		}
	}
	return false
}

func (r *Regex) MarshalBinary() ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	var ruleStr string
	for _, rule := range r.arr {
		ruleStr = rule.String()
		ruleLen := len(ruleStr)
		if ruleLen > 255 {
			return nil, fmt.Errorf("rule length %d is too long", ruleLen)
		}
		buffer.WriteByte(byte(ruleLen))
		buffer.WriteString(ruleStr)
	}
	return buffer.Bytes(), nil
}

func (r *Regex) UnmarshalBinary(data []byte) error {
	buffer := bytes.NewBuffer(data)
	for buffer.Len() > 0 {
		ruleLen, err := buffer.ReadByte()
		if err != nil {
			return err
		}
		rule := buffer.Next(int(ruleLen))
		re, err := regexp2.Compile(string(rule), regexp2.RE2)
		if err != nil {
			return fmt.Errorf("compile regex %s failed: %s", string(rule), err)
		}
		r.arr = append(r.arr, re)
	}
	return nil
}
