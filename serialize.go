package cereal

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

// Serialize takes a map and a version string, and returns the serialized byte representation.
func Serialize(value map[string]any, version string) ([]byte, error) {
	versionBytes := []byte(version)
	if len(versionBytes) != 1 {
		return []byte{}, errors.New("version must be exactly one byte")
	}
	buf := bytes.NewBuffer(versionBytes)

	if versionBytes[0] == '1' {
		err := serializeV1(value, buf)
		if err != nil {
			return []byte{}, err
		}
	} else {
		return buf.Bytes(), fmt.Errorf("invalid version %v", version)
	}

	return buf.Bytes(), nil
}

func serializeV1(value map[string]any, buf io.Writer) error {
	buf.Write([]byte{'{'})

	i := 0
	for k, v := range value {
		buf.Write([]byte(k))
		buf.Write([]byte{':'})

		err := writeValue(reflect.ValueOf(v), buf, []string{"<root>", k})
		if err != nil {
			return err
		}

		if i < len(value)-1 {
			buf.Write([]byte{','})
		}

		i++
	}

	buf.Write([]byte{'}'})
	return nil
}

func writeValue(value reflect.Value, buf io.Writer, path []string) error {
	kind := value.Kind()
	if kind == reflect.Bool {
		buf.Write([]byte{'b'})
		if value.Bool() {
			buf.Write([]byte{'1'})
		} else {
			buf.Write([]byte{'0'})
		}
	} else if kind == reflect.Int {
		buf.Write([]byte{'i'})
		buf.Write([]byte(strconv.FormatInt(value.Int(), 10)))
	} else if kind == reflect.Float32 {
		buf.Write([]byte{'f'})
		buf.Write([]byte(strconv.FormatFloat(value.Float(), 'g', -1, 32)))
	} else if kind == reflect.Float64 {
		buf.Write([]byte{'d'})
		buf.Write([]byte(strconv.FormatFloat(value.Float(), 'g', -1, 64)))
	} else if kind == reflect.String {
		buf.Write([]byte{'"'})
		buf.Write([]byte(value.String()))
	} else if kind == reflect.Slice {
		buf.Write([]byte{'['})
		if value.Len() == 0 {
			buf.Write([]byte{']'})
		} else {
			for i := 0; i < value.Len(); i++ {
				elemValue := value.Index(i)

				writeValue(elemValue, buf, append(path, strconv.Itoa(i)))
				if i < value.Len()-1 {
					buf.Write([]byte{','})
				} else {
					buf.Write([]byte{']'})
				}
			}
		}
	} else if kind == reflect.Map {
		buf.Write([]byte{'{'})
		mapKeys := value.MapKeys()
		if len(mapKeys) == 0 {
			buf.Write([]byte{'}'})
		} else {
			for i := 0; i < len(mapKeys); i++ {
				mapKey := mapKeys[i]
				if mapKey.Kind() != reflect.String {
					return fmt.Errorf("%v: map key type must be string, not %v", strings.Join(path, "."), mapKey.Kind())
				}
				buf.Write([]byte(mapKey.String()))
				buf.Write([]byte{':'})

				mapValue := value.MapIndex(mapKey)
				writeValue(mapValue, buf, append(path, mapKey.String()))

				if i < value.Len()-1 {
					buf.Write([]byte{','})
				} else {
					buf.Write([]byte{'}'})
				}
			}
		}
	} else {
		return fmt.Errorf("%v: unsupported value type %v", strings.Join(path, "."), kind)
	}
	return nil
}
