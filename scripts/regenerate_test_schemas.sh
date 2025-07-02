#!/bin/bash

# Script to regenerate all test schema files using the updated converter
set -e

echo "Building tfschema binary..."
go build -o tfschema ./cmd/tfschema

echo "Regenerating test schema files..."

# Find all test directories containing .tf files
find tests -name "test.tf" | while read -r tf_file; do
    test_dir=$(dirname "$tf_file")
    schema_file="$test_dir/test.schema.json"

    echo "Processing: $tf_file -> $schema_file"

    # Generate the schema using our converter
    ./tfschema "$tf_file" >"$schema_file"

    # Verify the generated file is valid JSON
    if ! jq empty "$schema_file" 2>/dev/null; then
        echo "ERROR: Generated invalid JSON for $schema_file"
        exit 1
    fi
done

echo "Cleaning up binary..."
rm tfschema

echo "All test schema files regenerated successfully!"
echo "Run 'go test -v ./tests' to verify all tests pass."
