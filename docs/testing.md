# tf-var-go Test Suite

This directory contains comprehensive end-to-end tests for the `tf-var-go` converter, organized using category partition analysis methodology.

## Test Organization

The tests are organized into 4 main categories based on complexity and feature focus:

### 1. Basic Features (`1-basic-features/`)

Core Terraform type support without validation:

- **01-string-none-basic**: Basic string variable
- **02-number-none-basic**: Basic number variable
- **03-bool-none-basic**: Basic boolean variable
- **04-list-none-basic**: Basic list variable
- **05-object-none-basic**: Basic object variable
- **06-map-none-basic**: Basic map variable
- **07-bool-enum-basic**: Boolean with enum validation
- **08-string-enum-basic**: String with enum validation

### 2. Simple Validation (`2-simple-validation/`)

Single validation rules applied to basic types:

- **09-string-length-basic**: String with length constraints
- **10-string-regex-basic**: String with regex pattern
- **11-number-range-basic**: Number with range constraints
- **12-number-enum-basic**: Number with enum validation
- **13-list-length-basic**: List with length constraints
- **14-object-length-basic**: Object with property count constraints

### 3. Advanced Features (`3-advanced-features/`)

Complex type combinations and nested validation:

- **15-list-enum-advanced**: List of objects with enum validation
- **16-object-enum-advanced**: Object with nested enum validation
- **17-map-length-advanced**: Map with length constraints
- **18-map-enum-advanced**: Map with enum validation

### 4. Complex Validation (`4-complex-validation/`)

Highly nested structures with multiple validation rules:

- **19-highly-nested-complex**: Multi-level object with various validation types, including `alltrue` for collection items.
- **20-set-length-basic**: Set with length constraints and uniqueness
- **21-tuple-nested-validation-complex**: Tuple with indexed validation

## Test Structure

Each test case contains three files:

- `test.tf` - Terraform variable definition
- `test.schema.json` - Expected JSON Schema output
- `test.tfvar.json` - Sample JSON values for validation

## Category Partition Analysis

### Type Categories

1. **Primitive Types**: string, number, bool
2. **Collection Types**: list, set, map
3. **Structural Types**: object, tuple
4. **Special Types**: optional (for object properties)

### Validation Categories

1. **Length Constraints**: minLength, maxLength, minItems, maxItems
2. **Range Constraints**: minimum, maximum (for numbers)
3. **Pattern Constraints**: regex patterns (for strings)
4. **Enum Constraints**: predefined value lists
5. **Property Constraints**: minProperties, maxProperties (for objects)

### Complexity Categories

1. **None**: No validation rules
2. **Basic**: Single validation rule
3. **Advanced**: Multiple validation rules or nested structures
4. **Complex**: Highly nested with multiple validation types

## Testing Methodology

### Coverage Matrix

The test suite uses a category partition approach to ensure comprehensive coverage:

| Type   | None | Basic                   | Advanced         | Complex |
| ------ | ---- | ----------------------- | ---------------- | ------- |
| string | ✓    | ✓ (length, regex, enum) | ✓                | ✓       |
| number | ✓    | ✓ (range, enum)         | -                | ✓       |
| bool   | ✓    | ✓ (enum)                | -                | ✓       |
| list   | ✓    | ✓ (length)              | ✓ (enum)         | ✓       |
| object | ✓    | ✓ (length)              | ✓ (enum)         | ✓       |
| map    | ✓    | -                       | ✓ (length, enum) | ✓       |
| set    | -    | ✓ (length)              | -                | ✓       |
| tuple  | -    | -                       | -                | ✓       |

### Design Principles

1. **Systematic Coverage**: Each combination of type and validation complexity is tested
2. **Isolation**: Each test focuses on a specific feature or combination
3. **Realistic Examples**: Test cases use realistic variable definitions
4. **Incremental Complexity**: Tests progress from simple to complex scenarios

### Key Test Categories

#### Constraint Types

- **Length**: String length, array/object size limits
- **Range**: Numeric minimum/maximum values
- **Pattern**: Regular expression validation
- **Enum**: Predefined value sets
- **Type**: Strict type checking with additionalProperties

#### Nesting Levels

- **Flat**: Single-level structures
- **Nested**: 2-3 levels of nesting
- **Complex**: 4+ levels with mixed types

#### Validation Scope

- **Root**: Validation on the root variable
- **Property**: Validation on object properties
- **Indexed**: Validation on specific array/tuple elements

## Running Tests

### End-to-End Tests

```bash
# Run all E2E tests
go test ./tests -v

# Run specific test category
go test ./tests -v -run "1-basic-features"
```

### Validation Script

```bash
# Validate all test cases with both Terraform and JSON Schema
./scripts/validate_tests.sh
```

### Individual Test Case

```bash
# Generate schema for a specific test
./tfschema tests/1-basic-features/01-string-none-basic/test.tf

# Validate against expected output
diff <(./tfschema tests/1-basic-features/01-string-none-basic/test.tf) tests/1-basic-features/01-string-none-basic/test.schema.json
```

## Schema Features Tested

### JSON Schema Draft 7 Features

- **Basic Types**: string, number, boolean, object, array
- **Validation Keywords**: minLength, maxLength, pattern, minimum, maximum, enum
- **Object Features**: properties, required, additionalProperties
- **Array Features**: items, minItems, maxItems, uniqueItems
- **Tuple Features**: prefixItems, additionalItems (custom extension)

### Terraform-Specific Extensions

- **Sensitive Values**: Custom `sensitive` field
- **Optional Properties**: Support for `optional()` type modifier
- **Type Inference**: Schema generation from default values
- **Path-Based Validation**: Support for nested property validation

## Test File Format

### Terraform Files (`test.tf`)

```hcl
variable "example_var" {
  type        = string
  description = "An example variable"
  default     = "default_value"

  validation {
    condition     = length(var.example_var) > 5
    error_message = "Must be longer than 5 characters"
  }
}
```

### Expected Schema (`test.schema.json`)

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "additionalProperties": false,
  "properties": {
    "example_var": {
      "type": "string",
      "description": "An example variable",
      "default": "default_value",
      "minLength": 6
    }
  }
}
```

### Test Values (`test.tfvar.json`)

```json
{
  "example_var": "valid_value"
}
```

## Maintenance

### Adding New Tests

1. Create new directory following naming convention: `XX-feature-description/`
2. Add three required files: `test.tf`, `test.schema.json`, `test.tfvar.json`
3. Update this README if introducing new categories
4. Run tests to ensure compatibility

### Updating Expected Outputs

```bash
# Regenerate all expected schemas
find tests -name "test.tf" -exec sh -c 'dir=$(dirname "{}"); ./tfschema "{}" > "$dir/test.schema.json"' \;
```

This systematic approach ensures comprehensive coverage of the converter's functionality while maintaining clear organization and documentation of test cases.
