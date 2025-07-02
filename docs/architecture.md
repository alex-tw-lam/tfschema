# tfschema Architecture Guide

## Overview

`tfschema` follows a **plugin-based, extensible architecture** that adheres to the **"new file" principle** - new features can be added by creating new files without modifying existing code. This design enables safe, incremental feature development and maintains backward compatibility.

## Core Architecture Components

### 1. Extension Registry System (`internal/extensions/`)

The heart of the extensible architecture is the unified extension registry that coordinates all plugin components.

```go
type ExtensionRegistry struct {
    typeConverters      *types.TypeConverterRegistry
    validationRules     []validation.RuleParser
    attributeAppliers   map[string]AttributeApplier
    postProcessors      []PostProcessor
    preProcessors       []PreProcessor
}
```

#### Key Features:

- **Centralized Registration**: All extension points go through a single registry
- **Automatic Discovery**: Components register themselves via `init()` functions
- **Priority-based Execution**: Processors run in configurable priority order
- **Legacy Bridge**: Seamless compatibility with existing code

### 2. Extension Points

#### A. Type Converters

Convert Terraform types to JSON Schema:

```go
type TypeConverter interface {
    Convert(expr hcl.Expression) (*jsonschema.Schema, error)
}
```

**Adding a new type converter (example):**

```go
// File: internal/extensions/examples/custom_type_converter.go
func init() {
    extensions.RegisterLegacyTypeConverter("duration", &DurationTypeConverter{})
}

type DurationTypeConverter struct{}

func (d *DurationTypeConverter) Convert(expr hcl.Expression) (*jsonschema.Schema, error) {
    return &jsonschema.Schema{
        Type:    "string",
        Pattern: `^([0-9]+(\.[0-9]+)?(ns|us|µs|ms|s|m|h))+$`,
    }, nil
}
```

#### B. Validation Rules

Parse Terraform validation expressions. Rules can be registered with a priority to control the order of execution.

```go
type RuleParser func(expr hcl.Expression, varName string) (Rule, []string, error)
```

**Adding a new validation rule (example):**

```go
// File: internal/extensions/examples/custom_validation_rule.go
func init() {
    // Register with a specific priority. Higher numbers run first.
    extensions.RegisterValidationRule(parseContainsSubstringRule, 10)
}

func parseContainsSubstringRule(expr hcl.Expression, varName string) (validation.Rule, []string, error) {
    // Parse contains(var.field, "substring") expressions
    // Return validation rule that generates JSON Schema pattern
}
```

**Note on `alltrue` expressions**: The architecture supports recursive parsing. For example, the `alltrue` parser (with a high priority) can invoke other registered parsers (like `regex`, `range`, etc.) on the inner expression of a `for` loop.

#### C. Attribute Appliers

Apply HCL attributes to JSON Schema:

```go
type AttributeApplier interface {
    Name() string
    Apply(schema interface{}, attribute interface{}) error
}
```

#### D. Pre/Post Processors

Transform schemas before/after conversion:

```go
type PreProcessor interface {
    Name() string
    Process(input interface{}) (interface{}, error)
    Priority() int
}

type PostProcessor interface {
    Name() string
    Process(schema interface{}) error
    Priority() int
}
```

### 3. Complete Extension Interface

For complex plugins that need multiple components:

```go
type Extension interface {
    Info() ExtensionInfo
    Register(registry *ExtensionRegistry) error
}

type ExtensionInfo struct {
    Name        string
    Version     string
    Author      string
    Description string
    Type        string
}
```

## Architecture Principles

### 1. "New File" Principle

**✅ Correct Way - Add Features by Creating New Files:**

```
internal/extensions/examples/
├── custom_validation_rule.go     # New validation rule
├── custom_type_converter.go      # New type converter
└── advanced_preprocessor.go      # New preprocessing logic
```

**❌ Wrong Way - Modifying Existing Files:**

- Adding cases to existing switch statements
- Modifying core converter logic
- Editing existing registry implementations

### 2. Registration Pattern

All components self-register using Go's `init()` function:

```go
func init() {
    // This runs automatically when the package is imported
    extensions.RegisterLegacyTypeConverter("mytype", &MyTypeConverter{})
    extensions.RegisterLegacyValidationRule(parseMyValidationRule)
}
```

### 3. Interface-Based Design

All extension points are defined as interfaces, enabling:

- **Testability**: Easy mocking and unit testing
- **Flexibility**: Multiple implementations of the same interface
- **Decoupling**: Components don't depend on concrete implementations

### 4. Backward Compatibility

The legacy bridge ensures existing code continues to work:

```go
// Legacy functions still work
RegisterTypeConverter("string", &StringConverter{})
RegisterValidationRule(parseStringRule)

// But internally route through new extension system
func RegisterLegacyTypeConverter(name string, converter TypeConverter) {
    GetGlobalRegistry().RegisterTypeConverter(name, converter)
}
```

## Implementation Examples

### Example 1: `alltrue` Validation Rule

The `alltrue` parser is a good example of a complex, recursive validation rule. It has a high priority so it can process `alltrue([...])` expressions before other parsers.

`internal/validation/alltrue_rule.go`:

```go
package validation

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func init() {
	// Register with a high priority to ensure it runs first
	RegisterRuleParserWithPriority(parseAllTrueRule, 20)
}

func parseAllTrueRule(expr hcl.Expression, varName string) (Rule, []string, error) {
	call, ok := expr.(*hclsyntax.FunctionCallExpr)
	if !ok || call.Name != "alltrue" {
		return nil, nil, nil // Not an alltrue call
	}

	forExpr, ok := call.Args[0].(*hclsyntax.ForExpr)
	if !ok {
		return nil, nil, nil // Not a for expression
	}

	// Extract collection path (e.g., "var.my_list")
	collectionPath, err := pathHandler.ExtractPathFromExpression(forExpr.CollExpr, varName)
	if err != nil {
		return nil, nil, err
	}

	// Recursively parse the inner condition (e.g., "each.value > 0")
	// The `forExpr.KeyVar` is the loop variable (e.g., "each")
	for _, subParser := range GetParsers() {
		innerRule, innerPath, err := subParser(forExpr.CondExpr, forExpr.KeyVar)
		if err != nil {
			return nil, nil, err
		}
		if innerRule != nil {
			// Combine paths: e.g., "my_list" + "*" + "field"
			fullPath := append(append(collectionPath, "*"), innerPath...)
			return innerRule, fullPath, nil
		}
	}

	return nil, nil, nil // No sub-parser matched
}
```

### Example 2: Adding URL Validation

Create `internal/extensions/examples/url_validation.go`:

```go
package examples

import (
    "github.com/atwlam/tfschema/internal/extensions"
    "github.com/atwlam/tfschema/internal/validation"
    // ... other imports
)

func init() {
    extensions.RegisterLegacyValidationRule(parseURLValidationRule)
}

func parseURLValidationRule(expr hcl.Expression, varName string) (validation.Rule, []string, error) {
    // Parse expressions like: contains(var.website_url, "https://")
    // Return rule that adds URL format validation to JSON Schema
}

type URLValidationRule struct {
    RequireHTTPS bool
}

func (r *URLValidationRule) Apply(schema *jsonschema.Schema) error {
    if r.RequireHTTPS {
        schema.Pattern = `^https://.*`
    } else {
        schema.Pattern = `^https?://.*`
    }
    return nil
}
```
