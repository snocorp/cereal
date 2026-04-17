package main

import (
	"fmt"
	"strings"

	"github.com/snocorp/cereal"
)

func main() {
	serialized := "1{key:\"value,num:i42,flag:b1}"

	data, err := cereal.Parse(strings.NewReader(serialized))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Parsed:", data)
	// Parsed: map[flag:true key:value num:42]
}
