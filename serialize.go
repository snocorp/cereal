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
	switch kind {
	case reflect.Bool:
		writeBool(buf, value.Bool())
	case reflect.Int:
		writeInt(buf, value.Int())
	case reflect.Float32:
		writeFloat(buf, value.Float())
	case reflect.Float64:
		writeDouble(buf, value.Float())
	case reflect.String:
		writeString(buf, value.String())
	case reflect.Slice:
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
	case reflect.Map:
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
				err := writeValue(mapValue, buf, append(path, mapKey.String()))
				if err != nil {
					return err
				}

				if i < value.Len()-1 {
					buf.Write([]byte{','})
				} else {
					buf.Write([]byte{'}'})
				}
			}
		}
	// handle decoded JSON
	case reflect.Interface:
		v := value.Interface()
		if i, ok := v.(int64); ok {
			writeInt(buf, i)
		} else if f, ok := v.(float64); ok {
			writeDouble(buf, f)
		} else if str, ok := v.(string); ok {
			writeString(buf, str)
		} else if b, ok := v.(bool); ok {
			writeBool(buf, b)
		} else {
			return fmt.Errorf("%v: unsupported value type %v for %v", strings.Join(path, "."), kind, value)
		}
	default:
		return fmt.Errorf("%v: unsupported value type %v for %v", strings.Join(path, "."), kind, value)
	}
	return nil
}

func writeBool(buf io.Writer, value bool) {
	buf.Write([]byte{'b'})
	if value {
		buf.Write([]byte{'1'})
	} else {
		buf.Write([]byte{'0'})
	}
}

func writeString(buf io.Writer, value string) {
	buf.Write([]byte{'"'})
	buf.Write([]byte(value))
}

func writeInt(buf io.Writer, value int64) {
	buf.Write([]byte{'i'})
	buf.Write([]byte(strconv.FormatInt(value, 10)))
}

func writeFloat(buf io.Writer, value float64) {
	buf.Write([]byte{'f'})
	buf.Write([]byte(strconv.FormatFloat(value, 'g', -1, 32)))
}

func writeDouble(buf io.Writer, value float64) {
	buf.Write([]byte{'d'})
	buf.Write([]byte(strconv.FormatFloat(value, 'g', -1, 64)))
}
