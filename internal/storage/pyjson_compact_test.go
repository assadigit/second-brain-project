// pyjson_compact_test.go — T-30 storage additions: PyDumpsCompact (the
// Python json.dumps(v) DEFAULT wire shape: ensure_ascii=True, ", "/": "
// separators, no indent) and DecodeJSONValue (value-level json.loads).
//
// Byte-parity pins derived from the Python reference
// (python3 -c 'import json; print(json.dumps(...))') — the wire bodies
// GitCurator's LLM clients send are pinned byte-for-byte by the
// internal/llmclient payload goldens; these unit tests pin the encoder
// mechanics themselves.
package storage

import (
	"encoding/json"
	"testing"
)

func TestPyDumpsCompactBasics(t *testing.T) {
	om := NewOrderedMap()
	om.Set("model", "m")
	msgs := []any{}
	m1 := NewOrderedMap()
	m1.Set("role", "user")
	m1.Set("content", "hi")
	msgs = append(msgs, m1)
	om.Set("messages", msgs)
	om.Set("temperature", 0.7)
	want := `{"model": "m", "messages": [{"role": "user", "content": "hi"}], "temperature": 0.7}`
	if got := PyDumpsCompact(om); got != want {
		t.Fatalf("PyDumpsCompact = %s, want %s", got, want)
	}
}

func TestPyDumpsCompactEnsureASCII(t *testing.T) {
	om := NewOrderedMap()
	om.Set("wörld", "ä✓")
	// Python: json.dumps({"w\u00f6rld": "ä✓"}) →
	// '{"w\\u00f6rld": "\\u00e4\\u2713"}'
	want := `{"w\u00f6rld": "\u00e4\u2713"}`
	if got := PyDumpsCompact(om); got != want {
		t.Fatalf("ensure_ascii: got %s, want %s", got, want)
	}
	// Astral runes → surrogate pairs (Python's emit shape).
	om2 := NewOrderedMap()
	om2.Set("emoji", "🦙")
	want2 := `{"emoji": "\ud83e\udd99"}`
	if got := PyDumpsCompact(om2); got != want2 {
		t.Fatalf("surrogates: got %s, want %s", got, want2)
	}
	// DEL (0x7f) escapes like Python's ESCAPE_ASCII table; 0x20-0x7e raw.
	om3 := NewOrderedMap()
	om3.Set("del", "a\x7fb")
	want3 := `{"del": "a\u007fb"}`
	if got := PyDumpsCompact(om3); got != want3 {
		t.Fatalf("DEL: got %s, want %s", got, want3)
	}
	// Control chars without short escapes → \u00xx lowercase.
	om4 := NewOrderedMap()
	om4.Set("c", "\x01")
	want4 := `{"c": "\u0001"}`
	if got := PyDumpsCompact(om4); got != want4 {
		t.Fatalf("ctrl: got %s, want %s", got, want4)
	}
}

func TestPyDumpsCompactNumbersAndShapes(t *testing.T) {
	cases := []struct {
		v    any
		want string
	}{
		{int64(5), "5"},
		{json.Number("123456789012345678901234567890"), "123456789012345678901234567890"},
		{0.7, "0.7"},
		{1.0, "1.0"},
		{0.2, "0.2"},
		{true, "true"},
		{nil, "null"},
		{[]any{}, "[]"},
		{[]any{int64(1), int64(2)}, "[1, 2]"},
		{map[string]any{"b": int64(1), "a": int64(2)}, `{"a": 2, "b": 1}`},
	}
	for _, c := range cases {
		if got := PyDumpsCompact(c.v); got != c.want {
			t.Errorf("PyDumpsCompact(%v) = %s, want %s", c.v, got, c.want)
		}
	}
	// nil slice → null (Python None); the wire never carries it, but the
	// PyDumps law must hold.
	var nilSlice []any
	if got := PyDumpsCompact(nilSlice); got != "null" {
		t.Errorf("nil slice: got %s, want null", got)
	}
}

func TestDecodeJSONValueTopLevelShapes(t *testing.T) {
	cases := []struct {
		in   string
		want any
	}{
		{"5", int64(5)},
		{"5.5", 5.5},
		{"[1, 2, 3]", []any{int64(1), int64(2), int64(3)}},
		{`"x"`, "x"},
		{"null", nil},
		{"true", true},
	}
	for _, c := range cases {
		got, err := DecodeJSONValue([]byte(c.in))
		if err != nil {
			t.Fatalf("DecodeJSONValue(%s): %v", c.in, err)
		}
		switch g := got.(type) {
		case int64:
			if g != c.want.(int64) {
				t.Errorf("%s → %v, want %v", c.in, got, c.want)
			}
		case float64:
			if g != c.want.(float64) {
				t.Errorf("%s → %v, want %v", c.in, got, c.want)
			}
		case string:
			if g != c.want.(string) {
				t.Errorf("%s → %v, want %v", c.in, got, c.want)
			}
		case bool:
			if g != c.want.(bool) {
				t.Errorf("%s → %v, want %v", c.in, got, c.want)
			}
		case nil:
			if c.want != nil {
				t.Errorf("%s → nil, want %v", c.in, c.want)
			}
		case []any:
			w := c.want.([]any)
			if len(g) != len(w) {
				t.Fatalf("%s → %v, want %v", c.in, got, c.want)
			}
			for i := range g {
				if g[i] != w[i] {
					t.Errorf("%s[%d] → %v, want %v", c.in, i, g[i], w[i])
				}
			}
		default:
			t.Errorf("%s → unexpected type %T", c.in, got)
		}
	}
	// Objects → *OrderedMap with insertion order.
	v, err := DecodeJSONValue([]byte(`{"b": 1, "a": 2}`))
	if err != nil {
		t.Fatal(err)
	}
	om, ok := v.(*OrderedMap)
	if !ok {
		t.Fatalf("object decode → %T", v)
	}
	if keys := om.Keys(); keys[0] != "b" || keys[1] != "a" {
		t.Errorf("insertion order lost: %v", keys)
	}
	// Trailing data rejected (Python json.loads raises).
	if _, err := DecodeJSONValue([]byte(`{} {}`)); err == nil {
		t.Error("trailing data accepted")
	}
	// Malformed rejected.
	if _, err := DecodeJSONValue([]byte(`{"choices": [{"mes`)); err == nil {
		t.Error("malformed accepted")
	}
}
