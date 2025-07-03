// Package examples demonstrates how to add new type converters
// This file shows how to add a custom type converter for a hypothetical "duration" type
package examples

import (
	"fmt"

	"github.com/alex-tw-lam/tfschema/internal/extensions"
	"github.com/alex-tw-lam/tfschema/internal/jsonschema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func init() {
	// Register our custom type converter when this package is imported
	extensions.RegisterLegacyTypeConverter("duration", &DurationTypeConverter{})
}

// DurationTypeConverter handles conversion of duration types to JSON Schema
// This demonstrates how to add support for a new Terraform type without modifying existing code
type DurationTypeConverter struct{}

// Convert converts a duration type expression to a JSON Schema
// This would handle expressions like: duration("5m30s") or duration(var.timeout_value)
func (d *DurationTypeConverter) Convert(expr hcl.Expression) (*jsonschema.Schema, error) {
	// Handle both function calls like duration("5m") and plain duration type
	switch e := expr.(type) {
	case *hclsyntax.ScopeTraversalExpr:
		// Handle simple "duration" type
		if e.Traversal.RootName() == "duration" {
			return d.createDurationSchema()
		}
	case *hclsyntax.FunctionCallExpr:
		// Handle duration("5m30s") function calls
		if e.Name == "duration" {
			return d.createDurationSchema()
		}
	}

	return nil, fmt.Errorf("expression is not a duration type")
}

// createDurationSchema creates a JSON Schema for duration values
func (d *DurationTypeConverter) createDurationSchema() (*jsonschema.Schema, error) {
	// Duration values are typically represented as strings with specific patterns
	// This follows Go's time.Duration format: "1h30m45s", "5m", "30s", etc.
	return &jsonschema.Schema{
		Type:        "string",
		Pattern:     `^([0-9]+(\.[0-9]+)?(ns|us|µs|ms|s|m|h))+$`,
		Description: "A duration string in Go format (e.g., '1h30m', '5s', '100ms')",
	}, nil
}
