# Cereal

Cereal is a Go library for serializing and deserializing structured data.

Goals:

- Improve on JSON's problems
- Type safe: What you serialize is what you will get back
- Compact: Save bytes when storing your data
- Human-readable: Uses utf8 characters instead of a raw binary format

## Installation

To use Cereal, add it to your project:

```sh
go get github.com/snocorp/cereal
```

## Usage

### Serialize

The `Serialize` function converts a `map[string]any` into a serialized byte slice. It requires a version string to specify the serialization format. Currently the only supported version is "1".

#### Function Signature

```go
func Serialize(value map[string]any, version string) ([]byte, error)
```

#### Example: Serialize a Simple Map

```go
package main

import (
	"fmt"
	"github.com/snocorp/cereal"
)

func main() {
	data := map[string]any{
		"key": "value",
		"num": 42,
		"flag": true,
	}

	serialized, err := cereal.Serialize(data, "1")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Serialized:", string(serialized))
  // Serialized: 1{key:"value,num:i42,flag:b1}
}
```

### Parse

The `Parse` function reads serialized data from an `io.Reader` and converts it back into a `map[string]any`. It automatically detects the version from the first byte of the input.

#### Function Signature

```go
func Parse(reader io.Reader) (map[string]any, error)
```

#### Example: Parse Serialized Data

```go
package main

import (
	"fmt"
	"github.com/snocorp/cereal"
  "strings"
)

func main() {
	serialized := "1{key:\"value,num:i42,flag:b1}"

	data, err := cereal.Parse(strings.NewReader(serialized))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Parsed:", data)
}
```

## Supported Data Types

Cereal supports the following data types for serialization and parsing:

- **Boolean**: Serialized as `b1` (true) or `b0` (false).
- **Integer**: Serialized as `i<number>`.
- **Float32**: Serialized as `f<number>`.
- **Float64**: Serialized as `d<number>`.
- **String**: Serialized as `"string`.
- **Array**: Serialized as `[value1,value2,...]`.
- **Map**: Serialized as `{key1:value1,key2:value2,...}`.

## Error Handling

Both `Serialize` and `Parse` return detailed error messages when they encounter invalid input or unsupported data types. For example:

- `Serialize`: `"version must be exactly one byte"`, `"<root>.key: unsupported value type chan"`.
- `Parse`: `"<root>: unexpected end of input"`, `"<root>.key: invalid type marker 'X'"`.

## Testing

The library includes comprehensive unit tests to ensure correctness. To run the tests:

```sh
go test ./...
```

## License

Cereal is licensed under the MIT License. See the `LICENSE` file for details.
