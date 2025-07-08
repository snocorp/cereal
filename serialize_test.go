package cereal

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestSerialize(t *testing.T) {
	b, err := Serialize(map[string]any{}, "1")
	if err != nil {
		t.Error(err)
	}

	if string(b) != "1{}" {
		t.Error("expected '1{}' but got", string(b))
	}
}

func TestSerialize_TooLongVersion(t *testing.T) {
	_, err := Serialize(map[string]any{}, "😀")
	if err == nil {
		t.Error("expected an error")
	}
	msg := err.Error()
	if msg != "version must be exactly one byte" {
		t.Error("expected error to be 'version must be exactly one byte' but got", msg)
	}
}

func TestSerialize_InvalidVersion(t *testing.T) {
	_, err := Serialize(map[string]any{}, "0")
	if err == nil {
		t.Error("expected an error")
	}
	msg := err.Error()
	if msg != "invalid version 0" {
		t.Error("expected error to be 'invalid version 0' but got", msg)
	}
}

func TestSerializeV1_Unsupported(t *testing.T) {
	channel := make(chan string)
	err := serializeV1(map[string]any{"x": channel}, &bytes.Buffer{})
	if err == nil {
		t.Error("expected an error")
	}
	msg := err.Error()
	if !strings.HasPrefix(msg, "<root>.x: unsupported value type chan") {
		t.Error("expected error to be '<root>.x: unsupported value type chan' but got", msg)
	}
}

func TestSerializeV1_Pointer(t *testing.T) {
	var x *string
	err := serializeV1(map[string]any{"x": x}, &bytes.Buffer{})
	if err == nil {
		t.Error("expected an error")
	}
	msg := err.Error()
	if msg != "<root>.x: unsupported value type ptr for <nil>" {
		t.Error("expected error to be '<root>.x: unsupported value type ptr for <nil>' but got", msg)
	}
}

func TestSerializeV1_Nil(t *testing.T) {
	err := serializeV1(map[string]any{"x": nil}, &bytes.Buffer{})
	if err == nil {
		t.Error("expected an error")
	}
	msg := err.Error()
	if msg != "<root>.x: unsupported value type invalid for <invalid reflect.Value>" {
		t.Error("expected error to be '<root>.x: unsupported value type invalid for <invalid reflect.Value>' but got", msg)
	}
}

func TestSerializeV1_EmptyMap(t *testing.T) {
	buf := bytes.Buffer{}
	err := serializeV1(map[string]any{}, &buf)
	if err != nil {
		t.Error(err)
	}

	if buf.String() != "{}" {
		t.Error("expected '{}' but got", buf.String())
	}
}

func TestSerializeV1_MultiValue(t *testing.T) {
	buf := bytes.Buffer{}
	err := serializeV1(map[string]any{"a": 1, "b": true, "c": "hello"}, &buf)
	if err != nil {
		t.Error(err)
	}

	// map keys are not consistently ordered so we need to parse the result
	m, err := parseV1(&buf)
	if err != nil {
		t.Error(err)
	}

	if m["a"] != 1 {
		t.Error("expected x.a to be '1' but got", m["a"])
	}
	if m["b"] != true {
		t.Error("expected x.b to be 'true' but got", m["b"])
	}
	if m["c"] != "hello" {
		t.Error("expected x.c to be 'hello' but got", m["c"])
	}
}

func TestSerializeV1_SingleBoolTrue(t *testing.T) {
	buf := bytes.Buffer{}
	err := serializeV1(map[string]any{"b": true}, &buf)
	if err != nil {
		t.Error(err)
	}

	if buf.String() != "{b:b1}" {
		t.Error("expected '{b:b1}' but got", buf.String())
	}
}

func TestSerializeV1_SingleBoolFalse(t *testing.T) {
	buf := bytes.Buffer{}
	err := serializeV1(map[string]any{"b": false}, &buf)
	if err != nil {
		t.Error(err)
	}

	if buf.String() != "{b:b0}" {
		t.Error("expected '{b:b0}' but got", buf.String())
	}
}

func TestSerializeV1_SingleInt(t *testing.T) {
	buf := bytes.Buffer{}
	err := serializeV1(map[string]any{"n": 25}, &buf)
	if err != nil {
		t.Error(err)
	}

	if buf.String() != "{n:i25}" {
		t.Error("expected '{n:i25}' but got", buf.String())
	}
}

func TestSerializeV1_SingleFloat32(t *testing.T) {
	buf := bytes.Buffer{}
	err := serializeV1(map[string]any{"x": float32(1.234)}, &buf)
	if err != nil {
		t.Error(err)
	}

	if buf.String() != "{x:f1.234}" {
		t.Error("expected '{x:f1.234}' but got", buf.String())
	}
}

func TestSerializeV1_SingleFloat64(t *testing.T) {
	buf := bytes.Buffer{}
	err := serializeV1(map[string]any{"x": float64(1.234)}, &buf)
	if err != nil {
		t.Error(err)
	}

	if buf.String() != "{x:d1.234}" {
		t.Error("expected '{x:d1.234}' but got", buf.String())
	}
}

func TestSerializeV1_SingleString(t *testing.T) {
	buf := bytes.Buffer{}
	err := serializeV1(map[string]any{"x": "abcdef😀"}, &buf)
	if err != nil {
		t.Error(err)
	}

	if buf.String() != "{x:\"abcdef😀}" {
		t.Error("expected '{x:\"abcdef😀}' but got", buf.String())
	}
}

func TestSerializeV1_SingleArray(t *testing.T) {
	buf := bytes.Buffer{}
	err := serializeV1(map[string]any{"x": []int{1, 2, 3}}, &buf)
	if err != nil {
		t.Error(err)
	}

	if buf.String() != "{x:[i1,i2,i3]}" {
		t.Error("expected '{x:[i1,i2,i3]}' but got", buf.String())
	}
}

func TestSerializeV1_EmptyArray(t *testing.T) {
	buf := bytes.Buffer{}
	err := serializeV1(map[string]any{"x": []int{}}, &buf)
	if err != nil {
		t.Error(err)
	}

	if buf.String() != "{x:[]}" {
		t.Error("expected '{x:[]}' but got", buf.String())
	}
}

func TestSerializeV1_SingleMap(t *testing.T) {
	buf := bytes.Buffer{}
	err := serializeV1(map[string]any{"x": map[string]int{"a": 1, "b": 2, "c": 3}}, &buf)
	if err != nil {
		t.Error(err)
	}

	// map keys are not consistently ordered so we need to parse the result
	m, err := parseV1(&buf)
	if err != nil {
		t.Error(err)
	}

	m = m["x"].(map[string]any)
	if m["a"] != 1 {
		t.Error("expected x.a to be '1' but got", m["a"])
	}
	if m["b"] != 2 {
		t.Error("expected x.b to be '2' but got", m["b"])
	}
	if m["c"] != 3 {
		t.Error("expected x.c to be '3' but got", m["c"])
	}
}

func TestSerializeV1_EmptySubMap(t *testing.T) {
	buf := bytes.Buffer{}
	err := serializeV1(map[string]any{"x": map[string]int{}}, &buf)
	if err != nil {
		t.Error(err)
	}

	// map keys are not consistently ordered so we need to parse the result
	m, err := parseV1(&buf)
	if err != nil {
		t.Error(err)
	}

	m = m["x"].(map[string]any)
	if len(m) != 0 {
		t.Error("expected x to be empty but got", m)
	}
}

func TestSerializeV1_InvalidMapKey(t *testing.T) {
	buf := bytes.Buffer{}
	err := serializeV1(map[string]any{"x": map[int]int{5: 5}}, &buf)
	if err == nil {
		t.Error("expected an error")
	}
	msg := err.Error()
	if msg != "<root>.x: map key type must be string, not int" {
		t.Error("expected error to be '<root>.x: map key type must be string, not int' but got", msg)
	}
}

func TestSerializeV1_InterfaceTypes(t *testing.T) {
	var v map[string]any
	json.NewDecoder(strings.NewReader(`{"x":{"d":1.0,"i":2,"s":"a","b":true}}`)).Decode(&v)
	buf := bytes.Buffer{}
	err := serializeV1(v, &buf)
	if err != nil {
		t.Error(err)
	}

	// map keys are not consistently ordered so we need to parse the result
	m, err := parseV1(&buf)
	if err != nil {
		t.Error(err)
	}

	m = m["x"].(map[string]any)
	if m["d"] != 1.0 {
		t.Error("expected d to be '1.0' but got", m["a"])
	}
	if m["i"] != 2.0 {
		t.Error("expected i to be '2' but got", m["i"])
	}
	if m["s"] != "a" {
		t.Error("expected s to be 'a' but got", m["s"])
	}
	if m["b"] != true {
		t.Error("expected b to be 'true' but got", m["b"])
	}
}

func TestSerializeV1_InvalidInterfaceTypes(t *testing.T) {
	var v map[string]any
	json.NewDecoder(strings.NewReader(`{"x":{"nil":null}}`)).Decode(&v)
	buf := bytes.Buffer{}
	err := serializeV1(v, &buf)
	if err == nil {
		t.Error("expected an error")
	}
	msg := err.Error()
	if msg != "<root>.x.nil: unsupported value type interface for <nil>" {
		t.Error("expected error to be '<root>.x.nil: unsupported value type interface for <nil>' but got", msg)
	}
}
