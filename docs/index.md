# tfschema

A Go tool that converts Terraform variable definitions (`variable` blocks) to JSON Schema format, enabling validation of `.tfvars.json` files against their corresponding Terraform variable constraints.

## Features

### Core Type Support

- **Primitive types**: `string`, `number`, `bool`, `any`
- **Collection types**: `list(type)`, `set(type)`, `map(type)`
- **Structural types**: `object({ ... })`, `tuple([...])`
- **Optional types**: `optional(type)` for object properties

### Validation Support

- **String validation**: `minLength`, `maxLength`, `pattern` (regex)
- **Number validation**: `minimum`, `maximum`, enumeration
- **Collection validation**: `minItems`, `maxItems`, `uniqueItems`
- **Object validation**: `minProperties`, strict property enforcement
- **Enum validation**: Predefined value sets for any type
- **Indexed validation**: Direct access to tuple/array elements (e.g., `var.payload[0]`, `var.data[2].field`)
- **Iterative validation**: `alltrue` with `for` expressions for complex collection validation (e.g. `alltrue([for item in var.my_list : item > 0])`)

### Schema Features

- **Flexible Object Schemas**: Object types use `additionalProperties: true` by default for compatibility
- **Type-specific Map Schemas**: Map types allow additional properties with type constraints
- **JSON Schema Draft 7**: Full compliance with modern JSON Schema standards
- **Comprehensive Validation**: Both Terraform and JSON Schema validation support

## Extensible Architecture

`tfschema` follows a **plugin-based, extensible architecture** that enables adding new features without modifying existing code. The "new file" principle allows developers to extend functionality by simply creating new files.

### Key Architecture Features

- **Extension Registry System**: Centralized plugin coordination
- **Type Converter Plugins**: Add support for new Terraform types
- **Validation Rule Plugins**: Implement custom validation logic
- **Pre/Post Processors**: Transform schemas during conversion
- **Legacy Compatibility**: Seamless integration with existing code

### Adding Extensions

Create new files to add functionality:

```go
// File: internal/extensions/examples/my_feature.go
func init() {
    extensions.RegisterLegacyTypeConverter("mytype", &MyTypeConverter{})
    extensions.RegisterLegacyValidationRule(parseMyValidationRule)
}
```

For complete details, see [docs/architecture.md](./architecture.md).

## Installation

To install the `tfschema` command-line tool, you can use `go install`:

```bash
go install github.com/alex-tw-lam/tfschema/cmd/tfschema@latest
```

Or build from source:

```bash
git clone https://github.com/alex-tw-lam/tfschema.git
cd tfschema
go build -o tfschema ./cmd/tfschema
```

## Usage

### Command Line

```bash
# Convert a Terraform file to JSON Schema
tfschema variables.tf > schema.json

# Validate a tfvars.json file against the schema
# First install a JSON Schema validator like ajv-cli:
npm install -g ajv-cli

# Then validate
ajv validate -s schema.json -d terraform.tfvars.json
```

### Programmatic Usage

```go
package main

import (
	"fmt"
	"github.com/alex-tw-lam/tfschema/internal/converter"
)

func main() {
	c := converter.New()
	schema, err := c.ConvertFile("variables.tf")
	if err != nil {
		panic(err)
	}

	// Use the schema for validation or documentation
	fmt.Printf("%s\n", schema)
}
```

### Using the Extension System

```go
package main

import (
    "github.com/alex-tw-lam/tfschema/internal/converter"
    _ "github.com/alex-tw-lam/tfschema/internal/extensions/examples" // Load extensions
)

func main() {
    // Create a new converter instance
    c := converter.New()
    schema, err := c.ConvertFile("variables.tf")
    if err != nil {
        log.Fatal(err)
    }

    // Output the schema as JSON
    output, _ := json.MarshalIndent(schema, "", "  ")
    fmt.Println(string(output))
}
```

For more information on creating extensions, see the [Architecture](./architecture.md) documentation.

## Examples

### Basic Types

```hcl
variable "app_name" {
  type        = string
  description = "The application name"
  default     = "my-app"
}

variable "instance_count" {
  type        = number
  description = "Number of instances"
  default     = 3
}
```

Generates:

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "additionalProperties": true,
  "properties": {
    "app_name": {
      "type": "string",
      "description": "The application name",
      "default": "my-app"
    },
    "instance_count": {
      "type": "number",
      "description": "Number of instances",
      "default": 3
    }
  },
  "required": []
}
```

### Complex Types with Validation

```hcl
variable "server_config" {
  type = object({
    name     = string
    port     = number
    enabled  = bool
    tags     = list(string)
    metadata = map(string)
  })

  validation {
    condition     = length(var.server_config.name) > 0
    error_message = "Server name cannot be empty."
  }

  validation {
    condition     = var.server_config.port >= 1 && var.server_config.port <= 65535
    error_message = "Port must be between 1 and 65535."
  }
}
```

Generates:

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "additionalProperties": true,
  "properties": {
    "server_config": {
      "type": "object",
      "additionalProperties": true,
      "properties": {
        "name": {
          "type": "string",
          "minLength": 1
        },
        "port": {
          "type": "number",
          "minimum": 1,
          "maximum": 65535
        },
        "enabled": {
          "type": "boolean"
        },
        "tags": {
          "type": "array",
          "items": {
            "type": "string"
          }
        },
        "metadata": {
          "type": "object",
          "additionalProperties": {
            "type": "string"
          }
        }
      },
      "required": ["name", "port", "enabled", "tags", "metadata"]
    }
  },
  "required": ["server_config"]
}
```

### `alltrue` Validation

`tfschema` can parse `alltrue` expressions with `for` loops to apply validation rules to each element in a list or map.

```hcl
variable "user_profiles" {
  type = list(object({
    username = string
    level    = number
  }))

  validation {
    condition = alltrue([
      for profile in var.user_profiles : length(profile.username) > 3
    ])
    error_message = "All usernames must be longer than 3 characters."
  }

  validation {
    condition = alltrue([
      for profile in var.user_profiles : profile.level >= 1 && profile.level <= 5
    ])
    error_message = "All user levels must be between 1 and 5."
  }
}
```

This generates a schema that applies the validation rules to each item in the `user_profiles` array:

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "user_profiles": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "username": {
            "type": "string",
            "minLength": 4
          },
          "level": {
            "type": "number",
            "minimum": 1,
            "maximum": 5
          }
        },
        "required": ["username", "level"],
        "additionalProperties": true
      }
    }
  },
  "required": ["user_profiles"],
  "additionalProperties": true
}
```

### Tuple Types with Indexed Validation

```hcl
variable "mixed_payload" {
  type = tuple([string, number, object({
    name     = string
    enabled  = bool
    retries  = number
  })])

  validation {
    condition     = length(var.mixed_payload[0]) == 36
    error_message = "First element must be exactly 36 characters (UUID format)."
  }

  validation {
    condition     = var.mixed_payload[1] >= 1 && var.mixed_payload[1] <= 8
    error_message = "Second element must be between 1 and 8."
  }

  validation {
    condition     = var.mixed_payload[2].retries >= 0 && var.mixed_payload[2].retries <= 3
    error_message = "Retries must be between 0 and 3."
  }
}
```

Generates:

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "additionalProperties": true,
  "properties": {
    "mixed_payload": {
      "type": "array",
      "items": [
        {
          "type": "string",
          "minLength": 36,
          "maxLength": 36
        },
        {
          "type": "number",
          "minimum": 1,
          "maximum": 8
        },
        {
          "type": "object",
          "additionalProperties": true,
          "properties": {
            "name": { "type": "string" },
            "enabled": { "type": "boolean" },
            "retries": {
              "type": "number",
              "minimum": 0,
              "maximum": 3
            }
          },
          "required": ["name", "enabled", "retries"]
        }
      ],
      "minItems": 3,
      "maxItems": 3
    }
  },
  "required": ["mixed_payload"]
}
```

## Supported Terraform Features

### Types

- ✅ `string`, `number`, `bool`, `any`
- ✅ `list(type)`, `set(type)`, `map(type)`
- ✅ `object({ field = type, ... })`
- ✅ `tuple([type1, type2, ...])`
- ✅ `optional(type)` in object definitions

### Validations

- ✅ String length: `length(var.field) > N`
- ✅ String regex: `can(regex("pattern", var.field))`
- ✅ Number range: `var.field >= N && var.field <= M`
- ✅ Enum validation: `contains(["a", "b", "c"], var.field)`
- ✅ Collection length: `length(var.list) > N`
- ✅ Indexed access: `var.tuple[0]`, `var.list[1].field`
- ✅ Complex expressions with logical operators

### Attributes

- ✅ `description` → JSON Schema `description`
- ✅ `default` → JSON Schema `default`
- ✅ `sensitive` → Metadata annotation (planned)

## Testing

The project includes comprehensive end-to-end tests covering 28 different scenarios:

```bash
# Run all tests
go test -v ./...

# Run only end-to-end tests
go test -v ./tests

# Run validation script (requires terraform and ajv-cli)
./scripts/validate_tests.sh
```

### Test Categories

1. **Basic Features** (9 tests): Core type support including `any` type
2. **Simple Validation** (6 tests): Basic validation rules
3. **Advanced Features** (4 tests): Complex type combinations
4. **Complex Validation** (4 tests): Highly nested scenarios with tuple support
5. **Edge Cases** (1 test): Special validation scenarios
6. **Terraschema Compatibility** (4 tests): Legacy compatibility testing

For detailed testing documentation, see [docs/testing.md](./testing.md).

## Architecture & Development

- **Architecture Guide**: [docs/architecture.md](./architecture.md) - Complete guide to the extensible architecture
- **Extension Examples**: [internal/extensions/examples/](./internal/extensions/examples/) - Sample implementations
- **Contributing**: Follow the "new file" principle - add features by creating new files
- **Testing Strategy**: Comprehensive unit, integration, and end-to-end testing

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

1. Fork the repository
2. Create your feature branch following the "new file" principle
3. Add comprehensive tests for new functionality
4. Ensure all existing tests pass: `go test -v ./...`
5. Update documentation as needed
6. Submit a pull request

For architectural guidance, see [docs/architecture.md](./architecture.md).
