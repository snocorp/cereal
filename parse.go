package cereal

import (
	"errors"
	"fmt"
	"io"
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

			if valueType == Map {
				result[key.String()], err = parseMapV1(reader, append(path, key.String()))
				if err != nil {
					return result, err
				}

				// set the state for the next k/v pair
				state = ReadingKey
				key = strings.Builder{}
				value = strings.Builder{}
			} else if valueType == Array {
				result[key.String()], err = parseArrayV1(reader, append(path, key.String()))
				if err != nil {
					return result, err
				}

				// set the state for the next k/v pair
				state = ReadingKey
				key = strings.Builder{}
				value = strings.Builder{}
			} else {
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
				continue
			}

			valueType, err = parseValueType(b[0], append(path, strconv.Itoa(index)))
			if err != nil {
				return result, err
			}

			if valueType == Map {
				r, err := parseMapV1(reader, append(path, strconv.Itoa(index)))
				if err != nil {
					return result, err
				}

				result = append(result, r)

				// set the state to parse another element
				state = ReadingType
				value = strings.Builder{}
			} else if valueType == Array {
				r, err := parseArrayV1(reader, append(path, strconv.Itoa(index)))
				if err != nil {
					return result, err
				}

				result = append(result, r)

				// set the state to parse another element
				state = ReadingType
				value = strings.Builder{}
			} else {
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
	if b == 'b' {
		valueType = Bool
	} else if b == 'i' {
		valueType = Int
	} else if b == 'd' {
		valueType = Float64
	} else if b == 'f' {
		valueType = Float32
	} else if b == '"' {
		valueType = String
	} else if b == '{' {
		valueType = Map
	} else if b == '[' {
		valueType = Array
	} else {
		err = fmt.Errorf("%v: invalid type marker '%v'", strings.Join(path, "."), string(b))
	}

	return
}

func parseValue(s string, valueType ValueType, path []string) (any, error) {
	if valueType == Bool {
		if s == "0" {
			return false, nil
		} else if s == "1" {
			return true, nil
		} else {
			return nil, fmt.Errorf("%v: invalid bool '%v'", strings.Join(path, "."), s)
		}
	} else if valueType == Int {
		v, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return 0, fmt.Errorf("%v: invalid int '%v'", strings.Join(path, "."), s)
		}
		return int(v), nil
	} else if valueType == Float32 {
		v, err := strconv.ParseFloat(s, 32)
		if err != nil {
			return 0, fmt.Errorf("%v: invalid float32 '%v'", strings.Join(path, "."), s)
		}
		return float32(v), nil
	} else if valueType == Float64 {
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, fmt.Errorf("%v: invalid float64 '%v'", strings.Join(path, "."), s)
		}
		return v, nil
	} else if valueType == String {
		return s, nil
	}

	return nil, fmt.Errorf("invalid type '%v' for '%v'", valueType, s)
}
