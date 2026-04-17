package cereal

import (
	"bytes"
	"testing"
)

func TestUnmarshal_EmptyInput(t *testing.T) {
	var m *map[string]any = &map[string]any{}
	err := Unmarshal([]byte{}, &m)
	if err == nil {
		t.Error("expected an error")
	}
	msg := err.Error()
	if msg != "expected a version in the first byte" {
		t.Error("expected error to be 'expected a version in the first byte' but got", msg)
	}
}

func TestUnmarshal_BadVersion(t *testing.T) {
	var m *map[string]any = &map[string]any{}
	err := Unmarshal([]byte{0}, m)
	if err == nil {
		t.Error("expected an error")
	}
	msg := err.Error()
	if msg != "unexpected version '0'" {
		t.Error("expected error to be \"unexpected version '0'\" but got", msg)
	}
}

func TestUnmarshalV1_Map(t *testing.T) {
	m := map[string]any{}
	err := unmarshalV1(bytes.NewBuffer([]byte{'{', 'b', ':', 'b', '1', '}'}), &m)
	if err != nil {
		t.Error(err)
	}

	if len(m) != 1 {
		t.Error("expected single entry")
	}
	if m["b"] != true {
		t.Error("expected 'b' to be true")
	}
}

func TestUnmarshal_EmptyMap(t *testing.T) {
	var emptyMap *map[string]any = &map[string]any{}
	err := unmarshalV1(bytes.NewBuffer([]byte{'{', '}'}), emptyMap)
	if err != nil {
		t.Error(err)
	}

	if len(*emptyMap) != 0 {
		t.Error("expected an empty map")
	}
}

func TestUnmarshalV1_InvalidType(t *testing.T) {
	type Struct struct {
		B bool
	}
	s := Struct{}
	err := unmarshalV1(bytes.NewBuffer([]byte{'{', 'B', ':', 'X', '1', '}'}), &s)
	if err == nil {
		t.Error("expected an error")
	}

	msg := err.Error()
	if msg != "<root>: invalid type marker 'X'" {
		t.Error("expected error to be \"<root>: invalid type marker 'X'\" but got", msg)
	}
}

func TestUnmarshalV1_InvalidValue(t *testing.T) {
	type Struct struct {
		B bool
	}
	s := Struct{}
	err := unmarshalV1(bytes.NewBuffer([]byte{'{', 'B', ':', 'b', '2', '}'}), &s)
	if err == nil {
		t.Error("expected an error")
	}

	msg := err.Error()
	if msg != "<root>.B: invalid bool '2'" {
		t.Error("expected error to be \"<root>.B: invalid bool '2'\" but got", msg)
	}
}

func TestUnmarshalV1_InvalidField(t *testing.T) {
	type Struct struct {
		B bool
	}
	s := Struct{}
	err := unmarshalV1(bytes.NewBuffer([]byte{'{', 'X', ':', 'b', '1', '}'}), &s)
	if err == nil {
		t.Error("expected an error")
	}

	msg := err.Error()
	if msg != "<root>: unexpected field name 'X'" {
		t.Error("expected error to be \"<root>: unexpected field name 'X'\" but got", msg)
	}
}

func TestUnmarshalV1_Struct(t *testing.T) {
	type Struct struct {
		B bool
		I int
	}
	s := Struct{}
	err := unmarshalV1(bytes.NewBuffer([]byte{'{', 'B', ':', 'b', '1', ',', 'I', ':', 'i', '2', '}'}), &s)
	if err != nil {
		t.Error(err)
	}

	if s.B != true {
		t.Error("expected 'B' to be true")
	}
	if s.I != 2 {
		t.Error("expected 'I' to be 2")
	}
}

func TestUnmarshalV1_NestedStruct(t *testing.T) {
	type NestedStruct struct {
		B bool
	}
	type OuterStruct struct {
		A NestedStruct
	}
	s := OuterStruct{}
	err := unmarshalV1(bytes.NewBuffer([]byte{'{', 'A', ':', '{', 'B', ':', 'b', '1', '}', '}'}), &s)
	if err != nil {
		t.Error(err)
	}

	if s.A.B != true {
		t.Error("expected 'B' to be true")
	}
}

func TestUnmarshalV1_NestedMap(t *testing.T) {
	type Struct struct {
		A map[string]any
	}
	s := Struct{}
	err := unmarshalV1(bytes.NewBuffer([]byte{'{', 'A', ':', '{', 'B', ':', 'b', '1', '}', '}'}), &s)
	if err != nil {
		t.Error(err)
	}

	if s.A["B"] != true {
		t.Error("expected 'B' to be true")
	}
}

func TestUnmarshalV1_NestedMapError(t *testing.T) {
	type Struct struct {
		A map[string]any
	}
	s := Struct{}
	err := unmarshalV1(bytes.NewBuffer([]byte{'{', 'A', ':', '{', 'B', ':', 'X', '1', '}', '}'}), &s)
	if err == nil {
		t.Error("expected an error")
	}

	msg := err.Error()
	if msg != "<root>.A: invalid type marker 'X'" {
		t.Error("expected error to be \"<root>.A: invalid type marker 'X'\" but got", msg)
	}
}

func TestUnmarshalV1_NestedStructArray(t *testing.T) {
	type InnerStruct struct {
		B bool
	}
	type Struct struct {
		A []InnerStruct
	}
	s := Struct{}
	err := unmarshalV1(bytes.NewBuffer([]byte{'{', 'A', ':', '[', '{', 'B', ':', 'b', '1', '}', ']', '}'}), &s)
	if err != nil {
		t.Error(err)
	}

	if s.A[0].B != true {
		t.Errorf("expected 'A[0].B' to be true: %v", s)
	}
}

func TestUnmarshalV1_NestedStructArrayInvalid(t *testing.T) {
	type InnerStruct struct {
		B bool
	}
	type Struct struct {
		A []InnerStruct
	}
	s := Struct{}
	err := unmarshalV1(bytes.NewBuffer([]byte{'{', 'A', ':', '[', '{', 'B', ':', 'i', '1', '}', ']', '}'}), &s)
	if err == nil {
		t.Error("expected an error")
	}

	if err.Error() != "<root>.A.0: type int cannot be assigned to field B with type bool" {
		t.Errorf("Unexpected error message: %v", err.Error())
	}
}

func TestUnmarshalV1_NestedArray(t *testing.T) {
	type Struct struct {
		B []bool
		I []int
		F []float32
		D []float64
		S []string
		M []map[string]any
		A [][]int
	}
	s := Struct{}
	err := unmarshalV1(bytes.NewBuffer([]byte{'{',
		'B', ':', '[', 'b', '1', ',', ',', 'b', '0', ']',
		'I', ':', '[', 'i', '2', ']',
		'F', ':', '[', 'f', '3', ']',
		'D', ':', '[', 'd', '4', ']',
		'S', ':', '[', '"', 'a', ']',
		'M', ':', '[', '{', 'm', ':', 'i', '6', '}', ']',
		'A', ':', '[', '[', 'i', '7', ']', ']',
		'}'}), &s)
	if err != nil {
		t.Error(err)
	}

	if s.B[0] != true {
		t.Errorf("expected 'B[0]' to be true, but got %v", s.B[0])
	}
	if s.B[1] != false {
		t.Errorf("expected 'B[0]' to be true, but got %v", s.B[0])
	}
	if s.I[0] != 2 {
		t.Error("expected 'I[0]' to be 2")
	}
	if s.F[0] != 3.0 {
		t.Error("expected 'F[0]' to be 3.0")
	}
	if s.D[0] != 4.0 {
		t.Error("expected 'D[0]' to be 4.0")
	}
	if s.S[0] != "a" {
		t.Error("expected 'D[0]' to be \"a\"")
	}
	if s.M[0]["m"] != 6 {
		t.Error("expected 'M[0][\"m\"]' to be 6")
	}
}

func TestUnmarshalV1_NestedArrayMapInvalid(t *testing.T) {
	type Struct struct {
		M []map[string]any
	}
	s := Struct{}
	err := unmarshalV1(bytes.NewBuffer([]byte{'{',
		'M', ':', '[', '{', ':', ':', 'i', '6', '}', ']',
		'}'}), &s)
	if err == nil {
		t.Error("expected an error")
	}

	if err.Error() != "<root>.M.0: invalid type marker ':'" {
		t.Errorf("Unexpected error message: %v", err.Error())
	}
}

func TestUnmarshalV1_Pointer(t *testing.T) {
	type Struct struct {
		P *int
	}
	s := Struct{}
	err := unmarshalV1(bytes.NewBuffer([]byte{'{',
		'P', ':', 'i', '1',
		'}'}), &s)
	if err == nil {
		t.Error("expected an error")
	}

	if err.Error() != "<root>: type int cannot be assigned to field P with type *int" {
		t.Errorf("Unexpected error message: %v", err.Error())
	}
}

func TestUnmarshalV1_NestedPointer(t *testing.T) {
	type InnerStruct struct {
		P *int
	}
	type Struct struct {
		S InnerStruct
	}
	s := Struct{}
	err := unmarshalV1(bytes.NewBuffer([]byte{'{',
		'S', ':', '{', 'P', ':', 'i', '1', '}',
		'}'}), &s)
	if err == nil {
		t.Error("expected an error")
	}

	if err.Error() != "<root>.S: type int cannot be assigned to field P with type *int" {
		t.Errorf("Unexpected error message: %v", err.Error())
	}
}

func TestUnmarshalV1_NestedArrayMalformed(t *testing.T) {
	type Struct struct {
		B []bool
	}
	s := Struct{}
	err := unmarshalV1(bytes.NewBuffer([]byte{'{',
		'B', ':', '['}), &s)
	if err == nil {
		t.Error("expected an error")
	}

	if err.Error() != "<root>.B: unexpected end of input" {
		t.Errorf("Unexpected error message: %v", err.Error())
	}
}

func TestUnmarshalV1_NestedArrayInvalidValue(t *testing.T) {
	type Struct struct {
		B []bool
	}
	s := Struct{}
	err := unmarshalV1(bytes.NewBuffer([]byte{'{',
		'B', ':', '[', 'b', '2', ']',
		'}'}), &s)
	if err == nil {
		t.Error("expected an error")
	}

	if err.Error() != "<root>.B.0: invalid bool '2'" {
		t.Errorf("Unexpected error message: %v", err.Error())
	}
}

func TestUnmarshalV1_NestedArrayWrongType(t *testing.T) {
	type Struct struct {
		B []bool
	}
	s := Struct{}
	err := unmarshalV1(bytes.NewBuffer([]byte{'{',
		'B', ':', '[', 'i', '1', ']',
		'}'}), &s)
	if err == nil {
		t.Error("expected an error")
	}

	if err.Error() != "<root>: cannot assign slice of type []int to slice field 'B' of type []bool" {
		t.Errorf("Unexpected error message: %v", err.Error())
	}
}

func TestUnmarshalV1_NestedArrayWrongTypeMap(t *testing.T) {
	type Struct struct {
		B []bool
	}
	s := Struct{}
	err := unmarshalV1(bytes.NewBuffer([]byte{'{',
		'B', ':', '[', '{', 'B', ':', 'b', '1', ']',
		'}'}), &s)
	if err == nil {
		t.Error("expected an error")
	}

	if err.Error() != "<root>.B: a struct or map cannot be inserted into slice of type []bool" {
		t.Errorf("Unexpected error message: %v", err.Error())
	}
}

func TestUnmarshalV1_NestedArrayInvalidType(t *testing.T) {
	type Struct struct {
		B []bool
	}
	s := Struct{}
	err := unmarshalV1(bytes.NewBuffer([]byte{'{',
		'B', ':', '[', 'X', '1', ']',
		'}'}), &s)
	if err == nil {
		t.Error("expected an error")
	}

	if err.Error() != "<root>.B.0: invalid type marker 'X'" {
		t.Errorf("Unexpected error message: %v", err.Error())
	}
}

func TestUnmarshalV1_NestedArrayMixedTypes(t *testing.T) {
	type Struct struct {
		B []bool
	}
	s := Struct{}
	err := unmarshalV1(bytes.NewBuffer([]byte{'{',
		'B', ':', '[', 'b', '1', ',', 'i', '1', ']',
		'}'}), &s)
	if err == nil {
		t.Error("expected an error")
	}

	if err.Error() != "<root>.B: arrays in structs must contain only elements of the same type" {
		t.Errorf("Unexpected error message: %v", err.Error())
	}
}

func TestUnmarshalV1_StructInvalidBasicType(t *testing.T) {
	type Struct struct {
		B bool
	}
	s := Struct{}
	err := unmarshalV1(bytes.NewBuffer([]byte{'{', 'B', ':', 'i', '1', '}'}), &s)
	if err == nil {
		t.Error("expected an error")
	}

	if err.Error() != "<root>: type int cannot be assigned to field B with type bool" {
		t.Errorf("Unexpected error message: %v", err.Error())
	}
}

func TestUnmarshalV1_StructInvalidMapType(t *testing.T) {
	type Struct struct {
		B bool
	}
	s := Struct{}
	err := unmarshalV1(bytes.NewBuffer([]byte{'{', 'B', ':', '{', '}', '}'}), &s)
	if err == nil {
		t.Error("expected an error")
	}

	if err.Error() != "<root>: a struct or map cannot be assigned to field B with type bool" {
		t.Errorf("Unexpected error message: %v", err.Error())
	}
}

func TestUnmarshalV1_EmptyInput(t *testing.T) {
	s := Struct{}
	err := unmarshalV1(bytes.NewBuffer([]byte{}), &s)
	if err == nil {
		t.Error("expected an error")
	}

	if err.Error() != "<root>: unexpected end of input" {
		t.Errorf("Unexpected error message: %v", err.Error())
	}
}

func TestUnmarshalV1_NonPointer(t *testing.T) {
	var s Struct
	err := unmarshalV1(bytes.NewBuffer([]byte{'{', '}'}), s)
	if err == nil {
		t.Error("expected an error")
	}

	if err.Error() != "Cannot unmarshal to non-pointer variable" {
		t.Errorf("Unexpected error message: %v", err.Error())
	}
}

func TestUnmarshalV1_Nil(t *testing.T) {
	var s *Struct = nil
	err := unmarshalV1(bytes.NewBuffer([]byte{'{', '}'}), s)
	if err == nil {
		t.Error("expected an error")
	}

	if err.Error() != "Cannot unmarshal to non-pointer variable" {
		t.Errorf("Unexpected error message: %v", err.Error())
	}
}

func TestUnmarshalV1_StructEndOfInput(t *testing.T) {
	var s *Struct = &Struct{}
	err := unmarshalV1(bytes.NewBuffer([]byte{'{'}), s)
	if err == nil {
		t.Error("expected an error")
	}

	if err.Error() != "<root>: unexpected end of input" {
		t.Errorf("Unexpected error message: %v", err.Error())
	}
}
