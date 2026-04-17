package main

import (
	"fmt"

	"github.com/snocorp/cereal"
)

type Example struct {
	Key  string
	Num  int
	Flag bool
}

func main() {
	serialized := []byte("1{Key:\"value,Num:i42,Flag:b1}")

	example := Example{}
	err := cereal.Unmarshal(serialized, &example)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Parsed:", example)
	// Parsed: {value 42 true}
}
