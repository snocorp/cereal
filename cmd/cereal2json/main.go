package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/snocorp/cereal"
)

// Parse a cereal string and output a JSON string
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
		fmt.Fprintln(os.Stderr, "input is empty")
		os.Exit(1)
	}

	obj, err := cereal.Parse(file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	encoder := json.NewEncoder(os.Stdout)
	err = encoder.Encode(obj)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage:\n - cereal2json <filename>\n - cat file.cereal | cereal2json -")
}
