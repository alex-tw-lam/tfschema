package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/atwlam/tfschema/internal/converter"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: tfschema <file.tf>")
		os.Exit(1)
	}

	filepath := os.Args[1]
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
