package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// OrderedMap is the config tree representation: Python's dict is
// insertion-ordered, and write_config_file's output preserves that order —
// a plain Go map cannot (doc 05 §3 item 1). The config load/merge/save
// paths (T-12/T-16) round-trip through this type so the bytes stay
// stable; typed accessors ride on top elsewhere.
type OrderedMap struct {
	keys []string
	m    map[string]any
}

// NewOrderedMap returns an empty ordered map.
func NewOrderedMap() *OrderedMap {
	return &OrderedMap{m: map[string]any{}}
}

// Set inserts or updates a key. Updating an existing key KEEPS its
// position (Python dict[k] = v semantics); new keys are appended.
func (o *OrderedMap) Set(key string, value any) {
	if _, ok := o.m[key]; !ok {
		o.keys = append(o.keys, key)
	}
	o.m[key] = value
}

// Get returns the value for key.
func (o *OrderedMap) Get(key string) (any, bool) {
	v, ok := o.m[key]
	return v, ok
}

// Delete removes a key (Python dict.pop/del semantics: the key and its
// position both leave; a missing key is a no-op). T-40's hand-delivery
// queue heal (the one-spelling re-key) needs the pop.
func (o *OrderedMap) Delete(key string) bool {
	if _, ok := o.m[key]; !ok {
		return false
	}
	delete(o.m, key)
	for i, k := range o.keys {
		if k == key {
			o.keys = append(o.keys[:i], o.keys[i+1:]...)
			break
		}
	}
	return true
}

// Keys returns the keys in insertion order.
func (o *OrderedMap) Keys() []string {
	out := make([]string, len(o.keys))
	copy(out, o.keys)
	return out
}

// Len returns the number of keys.
func (o *OrderedMap) Len() int { return len(o.keys) }

// Clone returns a SHALLOW copy, mirroring Python dict(existing) — the
// merge law's nested-dict handling makes its own nested copies.
func (o *OrderedMap) Clone() *OrderedMap {
	out := NewOrderedMap()
	for _, k := range o.keys {
		out.Set(k, o.m[k])
	}
	return out
}

// DecodeJSON parses a JSON document into a Python-faithful tree:
// *OrderedMap for objects (insertion order preserved), []any for arrays,
// and NUMBERS the way Python's json.load produces them:
//
//   - integral literals without '.', 'e', 'E'  → int64 (Python int);
//     if the literal overflows int64 it is kept verbatim as json.Number
//     (Python's ints are arbitrary precision and re-dump identically);
//   - everything else → float64, re-encoded by PyDumps with Python
//     float repr semantics.
//
// Note: Python's json accepts NaN/Infinity literals; Go's decoder does
// not. Config files never carry them (recorded here for honesty, not
// fixed — see doc 09 §5).
func DecodeJSON(data []byte) (*OrderedMap, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	v, err := decodeValue(dec)
	if err != nil {
		return nil, err
	}
	if _, err := dec.Token(); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("decodeJSON: trailing data after the JSON document")
		}
		return nil, err
	}
	om, ok := v.(*OrderedMap)
	if !ok {
		if v == nil {
			return NewOrderedMap(), nil
		}
		return nil, &json.UnmarshalTypeError{Value: "non-object JSON document"}
	}
	return om, nil
}

func decodeValue(dec *json.Decoder) (any, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '{':
			om := NewOrderedMap()
			for dec.More() {
				keyTok, err := dec.Token()
				if err != nil {
					return nil, err
				}
				key, ok := keyTok.(string)
				if !ok {
					return nil, &json.UnmarshalTypeError{Value: "non-string object key"}
				}
				v, err := decodeValue(dec)
				if err != nil {
					return nil, err
				}
				om.Set(key, v)
			}
			if _, err := dec.Token(); err != nil { // consume '}'
				return nil, err
			}
			return om, nil
		case '[':
			var arr []any
			for dec.More() {
				v, err := decodeValue(dec)
				if err != nil {
					return nil, err
				}
				arr = append(arr, v)
			}
			if _, err := dec.Token(); err != nil { // consume ']'
				return nil, err
			}
			if arr == nil {
				arr = []any{}
			}
			return arr, nil
		}
		return nil, &json.UnmarshalTypeError{Value: "unexpected delimiter " + string(rune(t))}
	case json.Number:
		return decodeNumber(t), nil
	case string:
		return t, nil
	case bool:
		return t, nil
	case nil:
		return nil, nil
	default:
		return nil, &json.UnmarshalTypeError{Value: "unexpected token"}
	}
}

// decodeNumber normalizes a JSON number literal to Python json.load
// semantics (see DecodeJSON).
func decodeNumber(n json.Number) any {
	lit := string(n)
	if !strings.ContainsAny(lit, ".eE") {
		if i, err := strconv.ParseInt(lit, 10, 64); err == nil {
			return i
		}
		// arbitrary-precision int: keep the literal verbatim; PyDumps
		// re-dumps json.Number as its digits (Python int parity).
		return n
	}
	f, err := n.Float64()
	if err != nil {
		return n // unreachable for valid JSON numbers; kept honest
	}
	return f
}
