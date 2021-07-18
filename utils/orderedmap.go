package utils

import (
	"fmt"
	"sort"
)

// OrderedMap implements ordered map data structure.
type OrderedMap struct {
	keys []interface{}
	m map[interface{}]interface{}
}

// NewOrderedMap function returns a pointer to the initialized OrderedMap.
func NewOrderedMap() *OrderedMap {
	return &OrderedMap{
		keys: make([]interface{}, 0),
		m: make(map[interface{}]interface{}),
	}
}

// Keys method returns internal ordered keys.
func (p *OrderedMap) Keys() []interface{} {
	return p.keys
}

// RawMap method returns internal unordered map.
func (p *OrderedMap) RawMap() map[interface{}]interface{} {
	return p.m
}

// Exist method checks whether key exists.
func (p *OrderedMap) Exist(k interface{}) bool {
	_, e := p.m[k]
	return e
}

// Val method returns the value if key exists, and returns nil otherwise.
func (p *OrderedMap) Val(k interface{}) interface{} {
	v, e := p.m[k]
	if e {
		return v
	} else {
		return nil
	}
}

// Add method adds (k,v) pair to the ordered map, and if key(=k) already exist, v will overwrite current value.
func (p *OrderedMap) Add(k interface{}, v interface{}) {
	_, e := p.m[k]
	if e {
		p.m[k] = v
	} else {
		p.keys = append(p.keys, k)
		p.m[k] = v
	}
}

// Remove method removes (k,v) pair from the ordered map.
func (p *OrderedMap) Remove(k interface{}) {
	_, e := p.m[k]
	if e {
		for i, k2 := range p.keys {
			if k2 == k {
				p.keys = append(p.keys[:i], p.keys[i+1:]...)
				break
			}
		}
		delete(p.m, k)
	} else {
		return
	}
}

// Len method returns size of the ordered map.
func (p *OrderedMap) Len() int {
	return len(p.keys)
}

// Sort method sorts internal keys.
func (p *OrderedMap) Sort() []string {
	keys := make([]string, 0)
	for _, k := range p.keys {
		keys = append(keys, k.(string))
	}

	sort.Strings(keys)
	return keys
}

// String method implements the Stringer interface
func (p *OrderedMap) String() string {
	var s string = "OrderedMap{"
	for i, k := range p.keys {
		if i < len(p.keys)-1 {
			s += fmt.Sprintf("%v:%v, ", k, p.m[k])
		} else {
			s += fmt.Sprintf("%v:%v", k, p.m[k])
		}
	}
	s += "}"

	return s
}