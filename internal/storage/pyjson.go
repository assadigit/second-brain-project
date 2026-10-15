package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
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
	b.WriteString(pyFloatRepr(f))
}

// PyFloatRepr exposes pyFloatRepr (Python float repr — which equals
// Python's str(float) in Python 3) for the parity helpers that need
// str(value) semantics on non-string slot values (internal/prompts'
// pyStr; the T-13 golden pins 5 → "5" vs 5.0 → "5.0").
func PyFloatRepr(f float64) string { return pyFloatRepr(f) }

// pyFloatRepr reproduces Python's float repr (CPython float_repr_style
// short): the shortest digit string that round-trips, in positional
// notation while -3 <= decpt <= 16, otherwise scientific with a signed,
// at-least-two-digit exponent. Integral positional values keep ".0".
//
// Verified against the generated corpus golden (storage_corpus.golden.json
// float_repr): 1e15 → "1000000000000000.0", 1e16 → "1e+16",
// 1e-05 → "1e-05", 0.0001 → "0.0001", 0.5 → "0.5", 2.0 → "2.0".
// Go's own 'g' format would print 1e15 as "1e+15" — divergent.
func pyFloatRepr(f float64) string {
	// Python json.dumps spells non-finite floats Infinity/-Infinity/NaN
	// (never in real configs; kept exact for completeness).
	switch {
	case math.IsNaN(f):
		return "NaN"
	case math.IsInf(f, 1):
		return "Infinity"
	case math.IsInf(f, -1):
		return "-Infinity"
	}
	neg := false
	if f < 0 || (f == 0 && math.Signbit(f)) {
		neg = true
		f = -f
	}
	// Shortest round-trip digits + decimal exponent via 'e' format.
	sci := strconv.FormatFloat(f, 'e', -1, 64)
	eIdx := strings.IndexByte(sci, 'e')
	mant := sci[:eIdx]
	exp10, _ := strconv.Atoi(sci[eIdx+1:]) // value = d.ddd × 10^exp10
	digits := strings.Replace(mant, ".", "", 1)
	// decpt: position of the decimal point relative to the digit string
	// (value = 0.digits × 10^decpt).
	decpt := exp10 + 1

	var out string
	if decpt < -3 || decpt > 16 {
		// scientific: d[.ddd]e±XX (exponent = decpt-1, ≥ 2 digits)
		m := digits[:1]
		if len(digits) > 1 {
			m += "." + digits[1:]
		}
		sign := "+"
		e := decpt - 1
		if e < 0 {
			sign, e = "-", -e
		}
		es := strconv.Itoa(e)
		if len(es) < 2 {
			es = "0" + es
		}
		out = m + "e" + sign + es
	} else if decpt <= 0 {
		out = "0." + strings.Repeat("0", -decpt) + digits
	} else if decpt >= len(digits) {
		out = digits + strings.Repeat("0", decpt-len(digits)) + ".0"
	} else {
		out = digits[:decpt] + "." + digits[decpt:]
	}
	if neg {
		return "-" + out
	}
	return out
}

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
