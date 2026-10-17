package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strconv"

	"github.com/assadigit/second-brain-project/internal/helpers"
)

// PyDumps is the PROMOTED canonical JSON encoder — the encoder the
// constants golden tests pioneered (pyCanonical, T-10), now production
// code because write_config_file needs Python json.dumps parity:
// ensure_ascii=False (raw UTF-8, no HTML escaping), indent=2, integral
// floats keep their ".0", and Python float repr semantics (scientific
// notation only from 1e16 up and below 1e-5; doc 05 §3 items 9 and 4).
//
// It encodes:
//   - *OrderedMap  → insertion order (dict semantics)
//   - map[string]any → keys sorted (deterministic; matches the golden-
//     convention dump json.dumps(sort_keys=True) — byte order equals
//     Python's code-point sort for valid UTF-8)
//   - json.Number  → its literal verbatim (arbitrary-precision int parity)
//   - int/uint/float/bool/string/slice/array/pointer/nil/any as Python
//     json.dumps would print them
//
// Like json.dumps, the result carries NO trailing newline.
func PyDumps(v any) string {
	var b bytes.Buffer
	pyEncode(&b, reflect.ValueOf(v), "")
	return b.String()
}

// PyDumpsCompact mirrors Python json.dumps(v) with the DEFAULT
// settings — the WIRE shape GitCurator's LLM clients put on the
// network (T-30's payload goldens pin it byte-for-byte):
// ensure_ascii=True (every non-ASCII rune escapes \uXXXX with lowercase
// hex, astral runes as surrogate pairs, DEL escaped like Python's
// ESCAPE_ASCII table), item separator ", ", key separator ": ", no
// indent. Same type rules as PyDumps (OrderedMap → insertion order,
// map → sorted keys, json.Number verbatim, Python float repr).
func PyDumpsCompact(v any) string {
	var b bytes.Buffer
	pyEncodeCompact(&b, reflect.ValueOf(v))
	return b.String()
}

// PyDumpsUTF8 mirrors Python json.dumps(v, ensure_ascii=False) — the
// WIRE shape directory_sync's publish leg puts on the network
// (T-49): raw UTF-8 strings (no \uXXXX escapes), item separator
// ", ", key separator ": ", no indent. Same type rules as PyDumps
// (OrderedMap → insertion order, map → sorted keys, json.Number
// verbatim, Python float repr).
func PyDumpsUTF8(v any) string {
	var b bytes.Buffer
	pyEncodeUTF8Compact(&b, reflect.ValueOf(v))
	return b.String()
}

func pyEncodeUTF8Compact(b *bytes.Buffer, v reflect.Value) {
	if v.IsValid() {
		if n, ok := v.Interface().(json.Number); ok {
			b.WriteString(string(n))
			return
		}
		if o, ok := v.Interface().(*OrderedMap); ok {
			pyEncodeOrderedUTF8(b, o)
			return
		}
	}
	switch v.Kind() {
	case reflect.Invalid:
		b.WriteString("null")
	case reflect.Bool:
		if v.Bool() {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		b.WriteString(strconv.FormatInt(v.Int(), 10))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		b.WriteString(strconv.FormatUint(v.Uint(), 10))
	case reflect.Float32, reflect.Float64:
		b.WriteString(PyFloatRepr(v.Float()))
	case reflect.String:
		pyEncodeString(b, v.String())
	case reflect.Slice, reflect.Array:
		if v.Kind() == reflect.Slice && v.IsNil() {
			b.WriteString("null") // Python None
			return
		}
		b.WriteByte('[')
		for i := 0; i < v.Len(); i++ {
			if i != 0 {
				b.WriteString(", ")
			}
			pyEncodeUTF8Compact(b, v.Index(i))
		}
		b.WriteByte(']')
	case reflect.Map:
		if v.Len() == 0 {
			b.WriteString("{}")
			return
		}
		keys := make([]string, 0, v.Len())
		for _, k := range v.MapKeys() {
			keys = append(keys, k.String())
		}
		sort.Strings(keys)
		b.WriteByte('{')
		for i, k := range keys {
			if i != 0 {
				b.WriteString(", ")
			}
			pyEncodeString(b, k)
			b.WriteString(": ")
			pyEncodeUTF8Compact(b, v.MapIndex(reflect.ValueOf(k)))
		}
		b.WriteByte('}')
	case reflect.Pointer:
		if v.IsNil() {
			b.WriteString("null")
			return
		}
		pyEncodeUTF8Compact(b, v.Elem())
	case reflect.Interface:
		pyEncodeUTF8Compact(b, v.Elem())
	default:
		panic("pyEncodeUTF8Compact: unsupported kind " + v.Kind().String())
	}
}

func pyEncodeOrderedUTF8(b *bytes.Buffer, o *OrderedMap) {
	if o.Len() == 0 {
		b.WriteString("{}")
		return
	}
	b.WriteByte('{')
	for i, k := range o.Keys() {
		if i != 0 {
			b.WriteString(", ")
		}
		pyEncodeString(b, k)
		b.WriteString(": ")
		v, _ := o.Get(k)
		pyEncodeUTF8Compact(b, reflect.ValueOf(v))
	}
	b.WriteByte('}')
}

func pyEncodeCompact(b *bytes.Buffer, v reflect.Value) {
	if v.IsValid() {
		if n, ok := v.Interface().(json.Number); ok {
			b.WriteString(string(n))
			return
		}
		if o, ok := v.Interface().(*OrderedMap); ok {
			pyEncodeOrderedCompact(b, o)
			return
		}
	}
	switch v.Kind() {
	case reflect.Invalid:
		b.WriteString("null")
	case reflect.Bool:
		if v.Bool() {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		b.WriteString(strconv.FormatInt(v.Int(), 10))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		b.WriteString(strconv.FormatUint(v.Uint(), 10))
	case reflect.Float32, reflect.Float64:
		b.WriteString(PyFloatRepr(v.Float()))
	case reflect.String:
		pyCompactString(b, v.String())
	case reflect.Slice, reflect.Array:
		if v.Kind() == reflect.Slice && v.IsNil() {
			b.WriteString("null") // Python None
			return
		}
		b.WriteByte('[')
		for i := 0; i < v.Len(); i++ {
			if i != 0 {
				b.WriteString(", ")
			}
			pyEncodeCompact(b, v.Index(i))
		}
		b.WriteByte(']')
	case reflect.Map:
		if v.Len() == 0 {
			b.WriteString("{}")
			return
		}
		keys := make([]string, 0, v.Len())
		for _, k := range v.MapKeys() {
			keys = append(keys, k.String())
		}
		sort.Strings(keys)
		b.WriteByte('{')
		for i, k := range keys {
			if i != 0 {
				b.WriteString(", ")
			}
			pyCompactString(b, k)
			b.WriteString(": ")
			pyEncodeCompact(b, v.MapIndex(reflect.ValueOf(k)))
		}
		b.WriteByte('}')
	case reflect.Pointer:
		if v.IsNil() {
			b.WriteString("null")
			return
		}
		pyEncodeCompact(b, v.Elem())
	case reflect.Interface:
		pyEncodeCompact(b, v.Elem())
	default:
		panic("pyEncodeCompact: unsupported kind " + v.Kind().String())
	}
}

func pyEncodeOrderedCompact(b *bytes.Buffer, o *OrderedMap) {
	if o.Len() == 0 {
		b.WriteString("{}")
		return
	}
	b.WriteByte('{')
	for i, k := range o.Keys() {
		if i != 0 {
			b.WriteString(", ")
		}
		pyCompactString(b, k)
		b.WriteString(": ")
		v, _ := o.Get(k)
		pyEncodeCompact(b, reflect.ValueOf(v))
	}
	b.WriteByte('}')
}

// pyCompactString mirrors json.dumps(ensure_ascii=True): printable
// ASCII (space through tilde, minus the short-escaped set) stays raw;
// everything else — controls, DEL, non-ASCII — escapes \uXXXX (lowercase
// hex), astral runes as surrogate pairs.
func pyCompactString(b *bytes.Buffer, s string) {
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString("\\\"")
		case '\\':
			b.WriteString("\\\\")
		case '\n':
			b.WriteString("\\n")
		case '\r':
			b.WriteString("\\r")
		case '\t':
			b.WriteString("\\t")
		case '\b':
			b.WriteString("\\b")
		case '\f':
			b.WriteString("\\f")
		default:
			if r >= 0x20 && r < 0x7f {
				b.WriteRune(r)
			} else if r <= 0xffff {
				fmt.Fprintf(b, "\\u%04x", r)
			} else {
				// astral plane: the surrogate pair Python emits
				r -= 0x10000
				hi := 0xd800 + (r >> 10)
				lo := 0xdc00 + (r & 0x3ff)
				fmt.Fprintf(b, "\\u%04x\\u%04x", hi, lo)
			}
		}
	}
	b.WriteByte('"')
}

// DecodeJSONValue parses ONE JSON document into a Python-faithful value
// tree — the value-level sibling of DecodeJSON (which requires an
// object at the top). Python's json.loads accepts any top-level value
// ("5", "[1,2]", "\"x\"", "null", …) and so does the app's extract_json
// family; the number/shape rules are DecodeJSON's (int64 for integral
// literals, float64 otherwise, *OrderedMap for objects, []any arrays).
func DecodeJSONValue(data []byte) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	v, err := decodeValue(dec)
	if err != nil {
		return nil, err
	}
	if _, err := dec.Token(); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("decodeJSONValue: trailing data after the JSON document")
		}
		return nil, err
	}
	return v, nil
}

func pyEncode(b *bytes.Buffer, v reflect.Value, indent string) {
	// json.Number is a string kind — intercept before the kind switch so
	// arbitrary-precision ints dump their digits verbatim.
	if v.IsValid() {
		if n, ok := v.Interface().(json.Number); ok {
			b.WriteString(string(n))
			return
		}
		// *OrderedMap is a struct kind — intercept for insertion order.
		if o, ok := v.Interface().(*OrderedMap); ok {
			pyEncodeOrdered(b, o, indent)
			return
		}
	}
	switch v.Kind() {
	case reflect.Invalid:
		b.WriteString("null")
	case reflect.Bool:
		if v.Bool() {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		b.WriteString(strconv.FormatInt(v.Int(), 10))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		b.WriteString(strconv.FormatUint(v.Uint(), 10))
	case reflect.Float32, reflect.Float64:
		pyEncodeFloat(b, v.Float())
	case reflect.String:
		pyEncodeString(b, v.String())
	case reflect.Slice, reflect.Array:
		if v.Kind() == reflect.Slice && v.IsNil() {
			b.WriteString("null") // Python None
			return
		}
		if v.Len() == 0 {
			b.WriteString("[]")
			return
		}
		inner := indent + "  "
		b.WriteString("[\n")
		for i := 0; i < v.Len(); i++ {
			b.WriteString(inner)
			pyEncode(b, v.Index(i), inner)
			if i != v.Len()-1 {
				b.WriteByte(',')
			}
			b.WriteByte('\n')
		}
		b.WriteString(indent + "]")
	case reflect.Map:
		if v.Len() == 0 {
			b.WriteString("{}")
			return
		}
		keys := make([]string, 0, v.Len())
		for _, k := range v.MapKeys() {
			keys = append(keys, k.String())
		}
		sort.Strings(keys) // byte order == code-point order for valid UTF-8
		inner := indent + "  "
		b.WriteString("{\n")
		for i, k := range keys {
			b.WriteString(inner)
			pyEncodeString(b, k)
			b.WriteString(": ")
			pyEncode(b, v.MapIndex(reflect.ValueOf(k)), inner)
			if i != len(keys)-1 {
				b.WriteByte(',')
			}
			b.WriteByte('\n')
		}
		b.WriteString(indent + "}")
	case reflect.Pointer:
		if v.IsNil() {
			b.WriteString("null")
			return
		}
		pyEncode(b, v.Elem(), indent)
	case reflect.Interface:
		pyEncode(b, v.Elem(), indent)
	default:
		// OrderedMap is a struct kind — handled below before this switch
		// would reject it. See pyEncodeOrdered.
		panic("pyEncode: unsupported kind " + v.Kind().String())
	}
}

// pyEncodeOrdered writes an *OrderedMap in insertion order.
func pyEncodeOrdered(b *bytes.Buffer, o *OrderedMap, indent string) {
	if o.Len() == 0 {
		b.WriteString("{}")
		return
	}
	inner := indent + "  "
	b.WriteString("{\n")
	keys := o.Keys()
	for i, k := range keys {
		b.WriteString(inner)
		pyEncodeString(b, k)
		b.WriteString(": ")
		v, _ := o.Get(k)
		pyEncode(b, reflect.ValueOf(v), inner)
		if i != len(keys)-1 {
			b.WriteByte(',')
		}
		b.WriteByte('\n')
	}
	b.WriteString(indent + "}")
}

func pyEncodeFloat(b *bytes.Buffer, f float64) {
	b.WriteString(PyFloatRepr(f))
}

// PyFloatRepr delegates to the canonical helpers home (the T-17 cycle
// break: helpers is the leaf; storage→dryrun→helpers would cycle if
// helpers imported storage back — the implementation MOVED, this alias
// keeps the T-12/T-13 call sites stable).
func PyFloatRepr(f float64) string { return helpers.PyFloatRepr(f) }

// pyEncodeString mirrors json.dumps(ensure_ascii=False): raw UTF-8, only
// the JSON-mandated escapes. Go's stock encoder escapes <, >, & in HTML
// mode — hand-rolling the string encoder keeps the bytes Python-identical
// (doc 05 §3 item 9).
func pyEncodeString(b *bytes.Buffer, s string) {
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString("\\\"")
		case '\\':
			b.WriteString("\\\\")
		case '\n':
			b.WriteString("\\n")
		case '\r':
			b.WriteString("\\r")
		case '\t':
			b.WriteString("\\t")
		case '\b':
			b.WriteString("\\b")
		case '\f':
			b.WriteString("\\f")
		default:
			if r < 0x20 {
				fmt.Fprintf(b, "\\u%04x", r)
			} else {
				b.WriteRune(r)
			}
		}
	}
	b.WriteByte('"')
}
