package main

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
		os.Exit(0)
	}

	if len(flag.Args()) != 1 {
		fmt.Println("Usage: tfschema <file.tf>")
		os.Exit(1)
	}

	filepath := flag.Arg(0)
	c := converter.New()
	schema, err := c.ConvertFile(filepath)
	if err != nil {
		fmt.Printf("Error converting file: %v\n", err)
		os.Exit(1)
	}

	jsonOutput, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling to JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonOutput))
}
