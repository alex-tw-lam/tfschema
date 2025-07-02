package main

// CLI entry point: read one .tf file, print its JSON Schema.

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/alex-tw-lam/tfschema/internal/converter"
)

var version = "dev"

func main() {
	versionFlag := flag.Bool("version", false, "Print the version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		return
	}
	if flag.NArg() != 1 {
		fmt.Println("Usage: tfschema <file.tf>")
		os.Exit(1)
	}

	schema, err := converter.ConvertFile(flag.Arg(0))
	if err != nil {
		fmt.Printf("Error converting file: %v\n", err)
		os.Exit(1)
	}

	output, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling to JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(output))
}
