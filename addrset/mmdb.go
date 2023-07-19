package addrset

import (
	"bytes"
	"fmt"
	"net"
	"net/netip"
	"sort"

	"github.com/maxmind/mmdbwriter"
	"github.com/maxmind/mmdbwriter/inserter"
	"github.com/maxmind/mmdbwriter/mmdbtype"
	"github.com/oschwald/maxminddb-golang"
)

const DatabaseType = "addrSet"

type AddrSetTree struct {
	*mmdbwriter.Tree
}

func NewAddrSetTree() (*AddrSetTree, error) {
	tree, err := mmdbwriter.New(mmdbwriter.Options{
		DatabaseType:            DatabaseType,
		IncludeReservedNetworks: true,
		IPVersion:               6,
		Inserter:                mergeValue,
	})
	if err != nil {
		return nil, err
	}
	return &AddrSetTree{tree}, nil
}

func mergeValue(newValue mmdbtype.DataType) inserter.Func {
	return func(oldValue mmdbtype.DataType) (mmdbtype.DataType, error) {
		if oldValue == nil {
			return newValue, nil
		}
		if newValue == nil {
			return oldValue, nil
		}
		_newValue, ok := newValue.(mmdbtype.Slice)
		if !ok {
			return nil, fmt.Errorf("invalid type: %T", newValue)
		}
		_oldValue, ok := oldValue.(mmdbtype.Slice)
		if !ok {
			return nil, fmt.Errorf("invalid type: %T", oldValue)
		}
		m := make(map[string]int)
		for _, v := range _oldValue {
			str, ok := v.(mmdbtype.String)
			if !ok {
				return nil, fmt.Errorf("invalid type: %T", v)
			}
			m[string(str)]++
		}
		for _, v := range _newValue {
			str, ok := v.(mmdbtype.String)
			if !ok {
				return nil, fmt.Errorf("invalid type: %T", v)
			}
			m[string(str)]++
		}
		newStrSlice := make([]string, 0)
		for k := range m {
			newStrSlice = append(newStrSlice, k)
		}
		sort.Slice(newStrSlice, func(i, j int) bool {
			return newStrSlice[i] < newStrSlice[j]
		})
		newSlice := make(mmdbtype.Slice, len(newStrSlice))
		for i, v := range newStrSlice {
			newSlice[i] = mmdbtype.String(v)
		}
		return newSlice, nil
	}
}

func (a *AddrSetTree) InsertMap(m map[string][]netip.Prefix) error {
	if m == nil {
		return fmt.Errorf("map is nil")
	}
	for t, p := range m {
		if p == nil {
			continue
		}
		for _, prefix := range p {
			_, ipNet, err := net.ParseCIDR(prefix.Masked().String())
			if err != nil {
				return err
			}
			err = a.Insert(ipNet, mmdbtype.Slice{mmdbtype.String(t)})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type AddrSetReader struct {
	*maxminddb.Reader
}

func NewAddrSetReaderFromContent(content []byte) (*AddrSetReader, error) {
	reader, err := maxminddb.FromBytes(content)
	if err != nil {
		return nil, err
	}
	if reader.Metadata.DatabaseType != DatabaseType {
		reader.Close()
		return nil, fmt.Errorf("incorrect database type, expected %s, got %s", DatabaseType, reader.Metadata.DatabaseType)
	}
	return &AddrSetReader{reader}, nil
}

func NewAddrSetReaderFromFile(path string) (*AddrSetReader, error) {
	reader, err := maxminddb.Open(path)
	if err != nil {
		return nil, err
	}
	if reader.Metadata.DatabaseType != DatabaseType {
		reader.Close()
		return nil, fmt.Errorf("incorrect database type, expected %s, got %s", DatabaseType, reader.Metadata.DatabaseType)
	}
	return &AddrSetReader{reader}, nil
}

func NewAddrSetReaderFromTree(tree *AddrSetTree) (*AddrSetReader, error) {
	buffer := bytes.NewBuffer(nil)
	defer buffer.Reset()
	_, err := tree.WriteTo(buffer)
	if err != nil {
		return nil, err
	}
	reader, err := maxminddb.FromBytes(buffer.Bytes())
	if err != nil {
		return nil, err
	}
	return &AddrSetReader{reader}, nil
}

func (r *AddrSetReader) LookupIP(ip net.IP) ([]string, error) {
	var result []string
	err := r.Lookup(ip, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *AddrSetReader) LookupAddr(ip netip.Addr) ([]string, error) {
	return r.LookupIP(ip.AsSlice())
}
