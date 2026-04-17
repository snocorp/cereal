package main

import (
	"fmt"

	"github.com/snocorp/cereal"
)

func main() {
	data := map[string]any{
		"key":  "value",
		"num":  42,
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
