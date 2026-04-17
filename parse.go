package cereal

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

// State represents the current state of the parser.
type State int

const (
	ReadingKey State = iota
	ReadingType
	ReadingValue
)

// ValueType represents the type of value being parsed.
type ValueType int

const (
	Bool ValueType = iota
	Int
	Float32
	Float64
	String
	Map
	Array
)

// Parse reads from the provided io.Reader and returns a map representation of the data.
func Parse(reader io.Reader) (map[string]any, error) {
	result := map[string]any{}
	b := make([]byte, 1)
	n, err := reader.Read(b)
	if err != nil && err != io.EOF {
		return result, err
	}
	if n == 0 {
		return result, errors.New("expected a version in the first byte")
	}

	versionByte := b[0]
	if versionByte == '1' {
		return parseV1(reader)
	}

	return result, fmt.Errorf("unexpected version '%v'", versionByte)
}

func parseV1(reader io.Reader) (map[string]any, error) {
	result := map[string]any{}
	b := make([]byte, 1)
	n, err := reader.Read(b)
	if err != nil && err != io.EOF {
		return result, err
	}
	if n == 0 {
		return result, errors.New("<root>: unexpected end of input")
	}

	if b[0] != '{' {
		return result, errors.New("<root>: expected '{'")
	}

	return parseMapV1(reader, []string{"<root>"})
}

func parseMapV1(reader io.Reader, path []string) (map[string]any, error) {
	result := map[string]any{}
	b := make([]byte, 1)

	state := ReadingKey
	key := strings.Builder{}
	var valueType ValueType
	value := strings.Builder{}
	escaped := false
	for {
		n, err := reader.Read(b)
		if err != nil && err != io.EOF {
			return result, err
		}
		if n == 0 {
			return result, fmt.Errorf("%v: unexpected end of input", strings.Join(path, "."))
		}

		if state == ReadingKey {
			if escaped {
				escaped = false
				key.Write(b)
			} else if b[0] == '}' && key.Len() == 0 {
				return result, nil
			} else if b[0] == ',' && key.Len() == 0 {
				continue
			} else if b[0] == ':' {
				state = ReadingType
			} else if b[0] == '\\' {
				escaped = true
			} else {
				key.Write(b)
			}
		} else if state == ReadingType {
			valueType, err = parseValueType(b[0], path)
			if err != nil {
				return result, err
			}

			switch valueType {
			case Map:
				result[key.String()], err = parseMapV1(reader, append(path, key.String()))
				if err != nil {
					return result, err
				}

				// set the state for the next k/v pair
				state = ReadingKey
				key = strings.Builder{}
				value = strings.Builder{}
			case Array:
				result[key.String()], err = parseArrayV1(reader, append(path, key.String()))
				if err != nil {
					return result, err
				}

				// set the state for the next k/v pair
				state = ReadingKey
				key = strings.Builder{}
				value = strings.Builder{}
			default:
				state = ReadingValue
			}
		} else if state == ReadingValue {
			if escaped {
				escaped = false
				value.Write(b)
			} else if b[0] == ',' {
				result[key.String()], err = parseValue(value.String(), valueType, append(path, key.String()))
				if err != nil {
					return result, err
				}

				// set the state for the next k/v pair
				state = ReadingKey
				key = strings.Builder{}
				value = strings.Builder{}
			} else if b[0] == '}' {
				result[key.String()], err = parseValue(value.String(), valueType, append(path, key.String()))
				return result, err
			} else if b[0] == '\\' {
				escaped = true
			} else {
				value.Write(b)
			}
		} else {
			return result, fmt.Errorf("%v: invalid state", strings.Join(path, "."))
		}
	}
}

func parseStructV1(reader io.Reader, rv reflect.Value, path []string) error {
	b := make([]byte, 1)

	state := ReadingKey
	key := strings.Builder{}
	var fv reflect.Value
	var valueType ValueType
	value := strings.Builder{}
	escaped := false
	for {
		n, err := reader.Read(b)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			return fmt.Errorf("%v: unexpected end of input", strings.Join(path, "."))
		}

		if state == ReadingKey {
			if escaped {
				escaped = false
				key.Write(b)
			} else if b[0] == '}' && key.Len() == 0 {
				return nil
			} else if b[0] == ',' && key.Len() == 0 {
				continue
			} else if b[0] == ':' {
				state = ReadingType
			} else if b[0] == '\\' {
				escaped = true
			} else {
				key.Write(b)
			}
		} else if state == ReadingType {
			fv = rv.FieldByName(key.String())
			if !fv.IsValid() {
				return fmt.Errorf("%v: unexpected field name '%v'", strings.Join(path, "."), key.String())
			}

			valueType, err = parseValueType(b[0], path)
			if err != nil {
				return err
			}

			switch valueType {
			case Map:
				k := fv.Kind()
				switch k {
				case reflect.Struct:
					err := parseStructV1(reader, fv, append(path, key.String()))
					if err != nil {
						return err
					}
				case reflect.Map:
					result, err := parseMapV1(reader, append(path, key.String()))
					if err != nil {
						return err
					}
					fv.Set(reflect.ValueOf(result))
				default:
					return fmt.Errorf(
						"%v: a struct or map cannot be assigned to field %v with type %v",
						strings.Join(path, "."),
						key.String(),
						fv.Type(),
					)
				}

				// set the state for the next k/v pair
				state = ReadingKey
				key = strings.Builder{}
				value = strings.Builder{}
			case Array:
				sliceValue, err := parseTypedArrayV1(reader, fv, append(path, key.String()))
				if err != nil {
					return err
				}

				if sliceValue.Type() != fv.Type() {
					return fmt.Errorf("%v: cannot assign slice of type %v to slice field '%v' of type %v", strings.Join(path, "."), sliceValue.Type(), key.String(), fv.Type())
				}

				fv.Set(sliceValue)

				// set the state for the next k/v pair
				state = ReadingKey
				key = strings.Builder{}
				value = strings.Builder{}
			default:
				state = ReadingValue
			}
		} else if state == ReadingValue {
			if escaped {
				escaped = false
				value.Write(b)
			} else if b[0] == ',' || b[0] == '}' {
				result, err := parseValue(value.String(), valueType, append(path, key.String()))
				if err != nil {
					return err
				}

				resultValue := reflect.ValueOf(result)
				if fv.Kind() == resultValue.Kind() {
					fv.Set(resultValue)
				} else {
					return fmt.Errorf(
						"%v: type %v cannot be assigned to field %v with type %v",
						strings.Join(path, "."),
						resultValue.Type(),
						key.String(),
						fv.Type(),
					)
				}

				if b[0] == ',' {
					// set the state for the next k/v pair
					state = ReadingKey
					key = strings.Builder{}
					value = strings.Builder{}
				} else {
					return nil
				}
			} else if b[0] == '\\' {
				escaped = true
			} else {
				value.Write(b)
			}
		} else {
			return fmt.Errorf("%v: invalid state", strings.Join(path, "."))
		}
	}
}

func parseTypedArrayV1(reader io.Reader, arrayValue reflect.Value, path []string) (reflect.Value, error) {
	var sliceValue reflect.Value
	b := make([]byte, 1)

	var valueType ValueType = -1
	var origValueType ValueType = -1
	state := ReadingType
	value := strings.Builder{}
	escaped := false

	for {
		var index string
		if !sliceValue.IsValid() {
			index = "0"
		} else {
			index = strconv.Itoa(sliceValue.Len())
		}

		n, err := reader.Read(b)
		if err != nil && err != io.EOF {
			return sliceValue, err
		}
		if n == 0 {
			return sliceValue, fmt.Errorf("%v: unexpected end of input", strings.Join(path, "."))
		}

		if state == ReadingType {
			if b[0] == ']' {
				return sliceValue, nil
			} else if b[0] == ',' && sliceValue.Len() > 0 {
				// looking for a type, but found an unexpected comma, just try again
				continue
			}

			valueType, err = parseValueType(b[0], append(path, index))
			if err != nil {
				return sliceValue, err
			}

			if origValueType < 0 {
				origValueType = valueType
			} else if origValueType != valueType {
				return sliceValue, fmt.Errorf("%v: arrays in structs must contain only elements of the same type", strings.Join(path, "."))
			}

			switch valueType {
			case Map:
				var elemValue reflect.Value
				k := arrayValue.Type().Elem().Kind()
				switch k {
				case reflect.Struct:
					structPtrValue := reflect.New(arrayValue.Type().Elem())
					err := parseStructV1(reader, structPtrValue.Elem(), append(path, index))
					if err != nil {
						return sliceValue, err
					}

					elemValue = structPtrValue.Elem()
				case reflect.Map:
					result, err := parseMapV1(reader, append(path, index))
					if err != nil {
						return sliceValue, err
					}
					elemValue = reflect.ValueOf(result)
				default:
					return sliceValue, fmt.Errorf(
						"%v: a struct or map cannot be inserted into slice of type %v",
						strings.Join(path, "."),
						arrayValue.Type(),
					)
				}

				if !sliceValue.IsValid() {
					sliceValue = reflect.MakeSlice(arrayValue.Type(), 0, 1)
				}
				sliceValue = reflect.Append(sliceValue, elemValue)

				// set the state to parse another element
				state = ReadingType
				value = strings.Builder{}
			case Array:
				subValue := reflect.MakeSlice(arrayValue.Type().Elem(), 0, 0)
				subSliceValue, err := parseTypedArrayV1(reader, subValue, append(path, index))
				if err != nil {
					return sliceValue, err
				}

				if !sliceValue.IsValid() {
					sliceValue = reflect.MakeSlice(reflect.SliceOf(subSliceValue.Type()), 0, 1)
				}
				sliceValue = reflect.Append(sliceValue, subSliceValue)

				// set the state to parse another element
				state = ReadingType
				value = strings.Builder{}
			default:
				state = ReadingValue
			}
		} else if state == ReadingValue {
			if escaped {
				escaped = false
				value.Write(b)
			} else if b[0] == ',' || b[0] == ']' {
				nextByte := b[0]
				r, err := parseValue(value.String(), valueType, append(path, index))
				if err != nil {
					return sliceValue, err
				}

				if !sliceValue.IsValid() {
					sliceValue = reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(r)), 0, 1)
				}
				sliceValue = reflect.Append(sliceValue, reflect.ValueOf(r))

				if nextByte == ',' {
					// set the state to parse another element
					state = ReadingType
					value = strings.Builder{}
				} else {
					return sliceValue, nil
				}
			} else if b[0] == '\\' {
				escaped = true
			} else {
				value.Write(b)
			}
		} else {
			return sliceValue, errors.New("invalid state")
		}
	}
}

func parseArrayV1(reader io.Reader, path []string) ([]any, error) {
	result := []any{}
	b := make([]byte, 1)

	var valueType ValueType
	state := ReadingType
	value := strings.Builder{}
	escaped := false

	for {
		index := len(result)
		n, err := reader.Read(b)
		if err != nil && err != io.EOF {
			return result, err
		}
		if n == 0 {
			return result, fmt.Errorf("%v: unexpected end of input", strings.Join(path, "."))
		}

		if state == ReadingType {
			if b[0] == ']' {
				return result, nil
			} else if b[0] == ',' && len(result) > 0 {
				// looking for a type, but found an unexpected comma, just try again
				continue
			}

			valueType, err = parseValueType(b[0], append(path, strconv.Itoa(index)))
			if err != nil {
				return result, err
			}

			switch valueType {
			case Map:
				r, err := parseMapV1(reader, append(path, strconv.Itoa(index)))
				if err != nil {
					return result, err
				}

				result = append(result, r)

				// set the state to parse another element
				state = ReadingType
				value = strings.Builder{}
			case Array:
				r, err := parseArrayV1(reader, append(path, strconv.Itoa(index)))
				if err != nil {
					return result, err
				}

				result = append(result, r)

				// set the state to parse another element
				state = ReadingType
				value = strings.Builder{}
			default:
				state = ReadingValue
			}
		} else if state == ReadingValue {
			if escaped {
				escaped = false
				value.Write(b)
			} else if b[0] == ',' {
				r, err := parseValue(value.String(), valueType, append(path, strconv.Itoa(index)))
				if err != nil {
					return result, err
				}

				result = append(result, r)

				// set the state to parse another element
				state = ReadingType
				value = strings.Builder{}
			} else if b[0] == ']' {
				r, err := parseValue(value.String(), valueType, append(path, strconv.Itoa(index)))
				if err != nil {
					return result, err
				}

				result = append(result, r)

				return result, nil
			} else if b[0] == '\\' {
				escaped = true
			} else {
				value.Write(b)
			}
		} else {
			return result, errors.New("invalid state")
		}
	}
}

func parseValueType(b byte, path []string) (valueType ValueType, err error) {
	switch b {
	case 'b':
		valueType = Bool
	case 'i':
		valueType = Int
	case 'd':
		valueType = Float64
	case 'f':
		valueType = Float32
	case '"':
		valueType = String
	case '{':
		valueType = Map
	case '[':
		valueType = Array
	default:
		err = fmt.Errorf("%v: invalid type marker '%v'", strings.Join(path, "."), string(b))
	}

	return
}

func parseValue(s string, valueType ValueType, path []string) (any, error) {
	switch valueType {
	case Bool:
		switch s {
		case "0":
			return false, nil
		case "1":
			return true, nil
		default:
			return nil, fmt.Errorf("%v: invalid bool '%v'", strings.Join(path, "."), s)
		}
	case Int:
		v, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return 0, fmt.Errorf("%v: invalid int '%v'", strings.Join(path, "."), s)
		}
		return int(v), nil
	case Float32:
		v, err := strconv.ParseFloat(s, 32)
		if err != nil {
			return 0, fmt.Errorf("%v: invalid float32 '%v'", strings.Join(path, "."), s)
		}
		return float32(v), nil
	case Float64:
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, fmt.Errorf("%v: invalid float64 '%v'", strings.Join(path, "."), s)
		}
		return v, nil
	case String:
		return s, nil
	}

	return nil, fmt.Errorf("invalid type '%v' for '%v'", valueType, s)
}
