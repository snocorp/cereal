package cereal

import (
	"bytes"
	"testing"
)

func TestParse_EmptyInput(t *testing.T) {
	_, err := Parse(bytes.NewBuffer([]byte{}))
	if err == nil {
		t.Error("expected an error")
	}
	msg := err.Error()
	if msg != "expected a version in the first byte" {
		t.Error("expected error to be 'expected a version in the first byte' but got", msg)
	}
}

func TestParse_BadVersion(t *testing.T) {
	_, err := Parse(bytes.NewBuffer([]byte{0}))
	if err == nil {
		t.Error("expected an error")
	}
	msg := err.Error()
	if msg != "unexpected version '0'" {
		t.Error("expected error to be \"unexpected version '0'\" but got", msg)
	}
}

func TestParse(t *testing.T) {
	result, err := Parse(bytes.NewBuffer([]byte{'1', '{', 'b', ':', 'b', '1', '}'}))
	if err != nil {
		t.Error(err)
	}

	if len(result) != 1 {
		t.Error("expected single entry")
	}
	if result["b"] != true {
		t.Error("expected 'b' to be true")
	}
}

func TestParseBytesV1_EmptyInput(t *testing.T) {
	_, err := parseV1(bytes.NewBuffer([]byte{}))
	if err == nil {
		t.Error("expected an error")
	}
	msg := err.Error()
	if msg != "<root>: unexpected end of input" {
		t.Error("expected error to be '<root>: unexpected end of input' but got", msg)
	}
}

func TestParseBytesV1_BadInput(t *testing.T) {
	_, err := parseV1(bytes.NewBuffer([]byte{'X'}))
	if err == nil {
		t.Error("expected an error")
	}
	msg := err.Error()
	if msg != "<root>: expected '{'" {
		t.Error("expected error to be \"<root>: expected '{'\" but got", msg)
	}
}

func TestParseBytesV1_EmptyMap(t *testing.T) {
	result, err := parseV1(bytes.NewBuffer([]byte{'{', '}'}))
	if err != nil {
		t.Error(err)
	}

	if len(result) != 0 {
		t.Error("expected an empty map")
	}
}

func TestParseMapV1_BadInput(t *testing.T) {
	_, err := parseMapV1(bytes.NewBuffer([]byte{}), []string{"<root>"})
	if err == nil {
		t.Error("expected an error")
	}
	msg := err.Error()
	if msg != "<root>: unexpected end of input" {
		t.Error("expected error to be '<root>: unexpected end of input' but got", msg)
	}
}

func TestParseMapV1_EscapedKey(t *testing.T) {
	result, err := parseMapV1(bytes.NewBuffer([]byte{'\\', '}', ':', 'b', '1', '}'}), []string{})
	if err != nil {
		t.Error(err)
	}

	if len(result) != 1 {
		t.Error("expected single entry")
	}
	if result["}"] != true {
		t.Error("expected '}' to be true")
	}
}

func TestParseMapV1_EscapedValue(t *testing.T) {
	result, err := parseMapV1(bytes.NewBuffer([]byte{'a', ':', '"', '\\', ',', '}'}), []string{})
	if err != nil {
		t.Error(err)
	}

	if len(result) != 1 {
		t.Error("expected single entry")
	}
	if result["a"] != "," {
		t.Error("expected 'a' to be ','")
	}
}

func TestParseMapV1_EscapedKeyEscape(t *testing.T) {
	result, err := parseMapV1(bytes.NewBuffer([]byte{'\\', '\\', ':', 'b', '1', '}'}), []string{})
	if err != nil {
		t.Error(err)
	}

	if len(result) != 1 {
		t.Error("expected single entry")
	}
	if result["\\"] != true {
		t.Error("expected '\\' to be true")
	}
}

func TestParseMapV1_SingleBool(t *testing.T) {
	result, err := parseMapV1(bytes.NewBuffer([]byte{'b', ':', 'b', '1', '}'}), []string{})
	if err != nil {
		t.Error(err)
	}

	if len(result) != 1 {
		t.Error("expected single entry")
	}
	if result["b"] != true {
		t.Error("expected 'b' to be true")
	}
}

func TestParseMapV1_SingleFloat32(t *testing.T) {
	result, err := parseMapV1(bytes.NewBuffer([]byte{'f', ':', 'f', '1', '}'}), []string{})
	if err != nil {
		t.Error(err)
	}

	if len(result) != 1 {
		t.Error("expected single entry")
	}
	if result["f"] != float32(1) {
		t.Error("expected 'b' to be 1")
	}
}

func TestParseMapV1_String(t *testing.T) {
	result, err := parseMapV1(bytes.NewBuffer([]byte("b:\"😀}")), []string{})
	if err != nil {
		t.Error(err)
	}

	if len(result) != 1 {
		t.Error("expected single entry")
	}
	if result["b"] != "😀" {
		t.Error("expected 'b' to be '😀'")
	}
}

func TestParseMapV1_InvalidBool(t *testing.T) {
	_, err := parseMapV1(bytes.NewBuffer([]byte{'b', ':', 'b', '2', '}'}), []string{"<root>"})
	if err == nil {
		t.Error("expected an error")
	}
	msg := err.Error()
	if msg != "<root>.b: invalid bool '2'" {
		t.Error("expected error to be \"<root>.b: invalid bool '2'\" but got", msg)
	}
}

func TestParseMapV1_InvalidInt(t *testing.T) {
	_, err := parseMapV1(bytes.NewBuffer([]byte{'b', ':', 'i', 'a', ',', 'c', ':', 'i', '0', '}'}), []string{"<root>"})
	if err == nil {
		t.Error("expected an error")
	}
	msg := err.Error()
	if msg != "<root>.b: invalid int 'a'" {
		t.Error("expected error to be \"<root>.b: invalid int 'a'\" but got", msg)
	}
}

func TestParseMapV1_InvalidFloat64(t *testing.T) {
	_, err := parseMapV1(bytes.NewBuffer([]byte{'c', ':', 'd', '1', '.', 'a', '}'}), []string{"<root>"})
	if err == nil {
		t.Error("expected an error")
	}
	msg := err.Error()
	if msg != "<root>.c: invalid float64 '1.a'" {
		t.Error("expected error to be \"<root>.c: invalid float64 '1.a'\" but got", msg)
	}
}

func TestParseMapV1_Multi(t *testing.T) {
	result, err := parseMapV1(bytes.NewBuffer([]byte{'b', ':', 'i', '1', ',', 'c', ':', 'd', '3', '.', '0', '}'}), []string{})
	if err != nil {
		t.Error(err)
	}

	if len(result) != 2 {
		t.Error("expected two entries but got", len(result))
	}
	if result["b"] != 1 {
		t.Error("expected 'b' to be '1' but got", result["b"])
	}
	if result["c"] != 3.0 {
		t.Error("expected 'c' to be '3.0' but got", result["c"])
	}
}

func TestParseMapV1_InvalidType(t *testing.T) {
	_, err := parseMapV1(bytes.NewBuffer([]byte{'b', ':', 'X', '2', '}'}), []string{"<root>"})
	if err == nil {
		t.Error("expected an error")
	}
	msg := err.Error()
	if msg != "<root>: invalid type marker 'X'" {
		t.Error("expected error to be \"<root>: invalid type marker 'X'\" but got", msg)
	}
}

func TestParseMapV1_SingleMap(t *testing.T) {
	result, err := parseMapV1(bytes.NewBuffer([]byte{'m', ':', '{', 'b', ':', 'b', '1', '}', '}'}), []string{})
	if err != nil {
		t.Error(err)
	}

	if len(result) != 1 {
		t.Error("expected single entry")
	}
	if result["m"].(map[string]any)["b"] != true {
		t.Error("expected 'b' to be true")
	}
}

func TestParseMapV1_MultiMap(t *testing.T) {
	result, err := parseMapV1(bytes.NewBuffer([]byte{'m', ':', '{', 'b', ':', 'b', '1', '}', ',', 'n', ':', '{', 'b', ':', 'b', '0', '}', '}'}), []string{})
	if err != nil {
		t.Error(err)
	}

	if len(result) != 2 {
		t.Error("expected two entries but got", len(result))
	}
	if result["m"].(map[string]any)["b"] != true {
		t.Error("expected 'm.b' to be true")
	}
	if result["n"].(map[string]any)["b"] != false {
		t.Error("expected 'n.b' to be false")
	}
}

func TestParseMapV1_BadMap(t *testing.T) {
	_, err := parseMapV1(bytes.NewBuffer([]byte{'m', ':', '{', 'b', ':'}), []string{"<root>"})
	if err == nil {
		t.Error("expected an error")
	}
	msg := err.Error()
	if msg != "<root>.m: unexpected end of input" {
		t.Error("expected error to be \"<root>.m: unexpected end of input\" but got", msg)
	}
}

func TestParseMapV1_SingleArray(t *testing.T) {
	result, err := parseMapV1(bytes.NewBuffer([]byte{'m', ':', '[', 'b', '0', ',', 'b', '1', ']', '}'}), []string{})
	if err != nil {
		t.Error(err)
	}

	if len(result) != 1 {
		t.Error("expected single entry")
	}
	if result["m"].([]any)[0] != false {
		t.Error("expected 'm[0]' to be false")
	}
	if result["m"].([]any)[1] != true {
		t.Error("expected 'm[1]' to be true")
	}
}

func TestParseMapV1_BadArray(t *testing.T) {
	_, err := parseMapV1(bytes.NewBuffer([]byte{'m', ':', '[', 'b', '0'}), []string{"<root>"})
	if err == nil {
		t.Error("expected an error")
	}
	msg := err.Error()
	if msg != "<root>.m: unexpected end of input" {
		t.Error("expected error to be \"<root>.m: unexpected end of input\" but got", msg)
	}
}

func TestParseArrayV1_EmptyInput(t *testing.T) {
	_, err := parseArrayV1(bytes.NewBuffer([]byte{}), []string{"<root>"})
	if err == nil {
		t.Error("expected an error")
	}
	msg := err.Error()
	if msg != "<root>: unexpected end of input" {
		t.Error("expected error to be \"<root>: unexpected end of input\" but got", msg)
	}
}

func TestParseArrayV1_InvalidValueType(t *testing.T) {
	_, err := parseArrayV1(bytes.NewBuffer([]byte{'X', 'a', ']'}), []string{"<root>"})
	if err == nil {
		t.Error("expected an error")
	}
	msg := err.Error()
	if msg != "<root>.0: invalid type marker 'X'" {
		t.Error("expected error to be \"<root>.0: invalid type marker 'X'\" but got", msg)
	}
}

func TestParseArrayV1_BadMap(t *testing.T) {
	_, err := parseArrayV1(bytes.NewBuffer([]byte{'{', 'a', '}', ']'}), []string{"<root>"})
	if err == nil {
		t.Error("expected an error")
	}
	msg := err.Error()
	if msg != "<root>.0: unexpected end of input" {
		t.Error("expected error to be \"<root>.0: unexpected end of input\" but got", msg)
	}
}

func TestParseArrayV1_SingleMap(t *testing.T) {
	result, err := parseArrayV1(bytes.NewBuffer([]byte{'{', 'a', ':', 'b', '0', '}', ']'}), []string{})
	if err != nil {
		t.Error(err)
	}

	if len(result) != 1 {
		t.Error("Expected an array of length 1 but got", len(result))
	}
	if result[0].(map[string]any)["a"] != false {
		t.Error("Expected an value of false but got", result[0].(map[string]any)["a"])
	}
}

func TestParseArrayV1_MultiMap(t *testing.T) {
	result, err := parseArrayV1(bytes.NewBuffer([]byte{'{', 'a', ':', 'b', '0', '}', ',', '{', 'b', ':', 'b', '1', '}', ']'}), []string{})
	if err != nil {
		t.Error(err)
	}

	if len(result) != 2 {
		t.Error("Expected an array of length 2 but got", len(result))
	}
	if result[0].(map[string]any)["a"] != false {
		t.Error("Expected a value of false but got", result[0].(map[string]any)["a"])
	}
	if result[1].(map[string]any)["b"] != true {
		t.Error("Expected a value of true but got", result[0].(map[string]any)["a"])
	}
}

func TestParseArrayV1_BadArray(t *testing.T) {
	_, err := parseArrayV1(bytes.NewBuffer([]byte{'['}), []string{"<root>"})
	if err == nil {
		t.Error("expected an error")
	}
	msg := err.Error()
	if msg != "<root>.0: unexpected end of input" {
		t.Error("expected error to be \"<root>.0: unexpected end of input\" but got", msg)
	}
}

func TestParseArrayV1_EmptyArray(t *testing.T) {
	result, err := parseArrayV1(bytes.NewBuffer([]byte{']'}), []string{})
	if err != nil {
		t.Error(err)
	}

	if len(result) != 0 {
		t.Error("Expected an array of length 0 but got", len(result))
	}
}

func TestParseArrayV1_SingleArray(t *testing.T) {
	result, err := parseArrayV1(bytes.NewBuffer([]byte{'[', 'b', '0', ']', ']'}), []string{})
	if err != nil {
		t.Error(err)
	}

	if len(result) != 1 {
		t.Error("Expected an array of length 1 but got", len(result))
	}
	if result[0].([]any)[0] != false {
		t.Error("Expected an value of false but got", result[0].([]any)[0])
	}
}

func TestParseArrayV1_Escaped(t *testing.T) {
	result, err := parseArrayV1(bytes.NewBuffer([]byte{'[', '"', '\\', ']', ']', ']'}), []string{})
	if err != nil {
		t.Error(err)
	}

	if len(result) != 1 {
		t.Error("Expected an array of length 1 but got", len(result))
	}
	if result[0].([]any)[0] != "]" {
		t.Error("Expected an value of ']' but got", result[0].([]any)[0])
	}
}

func TestParseArrayV1_BadValue(t *testing.T) {
	_, err := parseArrayV1(bytes.NewBuffer([]byte{'f', 'r', ']'}), []string{"<root>"})
	if err == nil {
		t.Error("expected an error")
	}
	msg := err.Error()
	if msg != "<root>.0: invalid float32 'r'" {
		t.Error("expected error to be \"<root>.0: invalid float32 'r'\" but got", msg)
	}
}

func TestParseArrayV1_BadValueMulti(t *testing.T) {
	_, err := parseArrayV1(bytes.NewBuffer([]byte{'f', 'r', ',', 'f', '4', ']'}), []string{"<root>"})
	if err == nil {
		t.Error("expected an error")
	}
	msg := err.Error()
	if msg != "<root>.0: invalid float32 'r'" {
		t.Error("expected error to be \"<root>.0: invalid float32 'r'\" but got", msg)
	}
}
