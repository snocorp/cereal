package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/snocorp/cereal"
)

// Parse a JSON string and output a cereal string
func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		printUsage()
		os.Exit(1)
	}

	var file *os.File
	var err error
	if args[0] == "-" {
		file = os.Stdin
	} else {
		file, err = os.Open(args[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	stat, err := file.Stat()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if stat.Size() == 0 {
		if len(args) == 0 {
			printUsage()
		} else {
			fmt.Fprintln(os.Stderr, "input is empty")
		}

		os.Exit(1)
	}

	var obj map[string]any

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&obj)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	b, err := cereal.Serialize(obj, "1")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	_, err = os.Stdout.Write(b)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage: json2cereal <filename>")
}
