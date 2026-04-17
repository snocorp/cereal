package cereal

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
)

func Unmarshal(data []byte, v any) error {
	reader := bytes.NewBuffer(data)

	b := make([]byte, 1)
	n, err := reader.Read(b)
	if err != nil && err != io.EOF {
		return err
	}
	if n == 0 {
		return errors.New("expected a version in the first byte")
	}

	versionByte := b[0]
	if versionByte == '1' {
		return unmarshalV1(reader, v)
	}

	return fmt.Errorf("unexpected version '%v'", versionByte)
}

func unmarshalV1(reader io.Reader, v any) error {
	b := make([]byte, 1)
	n, err := reader.Read(b)
	if err != nil && err != io.EOF {
		return err
	}
	if n == 0 {
		return errors.New("<root>: unexpected end of input")
	}

	if b[0] != '{' {
		return errors.New("<root>: expected '{'")
	}

	value := reflect.ValueOf(v)
	if value.Kind() != reflect.Pointer || value.IsNil() {
		return fmt.Errorf("Cannot unmarshal to non-pointer variable")
	}

	elem := value.Elem()
	k := elem.Kind()
	switch k {
	case reflect.Map:
		m, err := parseMapV1(reader, []string{"<root>"})
		if err != nil {
			return err
		}

		elem.Set(reflect.ValueOf(m))
	case reflect.Struct:
		err := parseStructV1(reader, elem, []string{"<root>"})
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported root type %v", k)
	}

	return nil
}
