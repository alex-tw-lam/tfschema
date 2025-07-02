# tfschema Test Suite

This directory contains comprehensive end-to-end tests for the `tfschema` converter, organized using category partition analysis methodology.

## Test Organization

The tests are organized into 6 main categories based on complexity and feature focus:

### 1. Basic Features (`1-basic-features/`) - 9 tests

Core Terraform type support without validation:

- **01-string-none-basic**: Basic string variable
- **02-number-none-basic**: Basic number variable
- **03-bool-none-basic**: Basic boolean variable
- **04-list-none-basic**: Basic list variable
- **05-object-none-basic**: Basic object variable
- **06-map-none-basic**: Basic map variable
- **07-bool-enum-basic**: Boolean with enum validation
- **08-string-enum-basic**: String with enum validation
- **23-any-type-basic**: Variable with `any` type

### 2. Simple Validation (`2-simple-validation/`) - 6 tests

Single validation rules applied to basic types:

- **09-string-length-basic**: String with length constraints
- **10-string-regex-basic**: String with regex pattern
- **11-number-range-basic**: Number with range constraints
- **12-number-enum-basic**: Number with enum validation
- **13-list-length-basic**: List with length constraints
- **14-object-length-basic**: Object with property count constraints

### 3. Advanced Features (`3-advanced-features/`) - 4 tests

Complex type combinations and nested validation:

- **15-list-enum-advanced**: List of objects with enum validation using `alltrue`
- **16-object-enum-advanced**: Object with nested enum validation
- **17-map-length-advanced**: Map with length constraints
- **18-map-enum-advanced**: Map with enum validation using `alltrue`

### 4. Complex Validation (`4-complex-validation/`) - 4 tests

Highly nested structures with multiple validation rules:

- **19-highly-nested-complex**: Multi-level object with various validation types, including `alltrue` for collection items
- **20-set-length-basic**: Set with length constraints and uniqueness
- **21-tuple-nested-validation-complex**: Tuple with indexed validation
- **22-ultra-complex-nesting**: Ultra-complex nested structure with tuples, sets, and deep nesting

### 5. Edge Cases (`5-edge-cases/`) - 1 test

Special validation scenarios and edge cases:

- **24-regex-with-or-condition**: Regex validation with OR condition (empty string allowed)

### 6. Terraschema Compatibility (`6-terraschema-compat/`) - 4 tests

Legacy compatibility testing with terraschema format:

- **25-terraschema-simple**: Basic terraschema compatibility
- **26-terraschema-simple-types**: All basic types including nullable
- **27-terraschema-complex-types**: Complex nested types with optional fields
- **28-terraschema-custom-validation**: Various validation patterns

## Test Structure

Each test case contains three files:

- `test.tf` - Terraform variable definition
- `test.schema.json` - Expected JSON Schema output
- `test.tfvar.json` - Sample JSON values for validation

## Category Partition Analysis

### Type Categories

1. **Primitive Types**: string, number, bool, any
2. **Collection Types**: list, set, map
3. **Structural Types**: object, tuple
4. **Special Types**: optional (for object properties), nullable

### Validation Categories

1. **Length Constraints**: minLength, maxLength, minItems, maxItems
2. **Range Constraints**: minimum, maximum (for numbers)
3. **Pattern Constraints**: regex patterns (for strings)
4. **Enum Constraints**: predefined value lists
5. **Property Constraints**: minProperties, maxProperties (for objects)

### Complexity Categories

1. **Basic**: Core type support with minimal validation
2. **Simple**: Single validation rule on basic types
3. **Advanced**: Multiple validation rules or nested structures with `alltrue`
4. **Complex**: Highly nested with multiple validation types
5. **Edge Cases**: Special scenarios and conditional validation
6. **Terraschema**: Legacy format compatibility testing

## Testing Methodology

### Coverage Matrix

The test suite uses a category partition approach to ensure comprehensive coverage across 6 categories:

| Type   | Basic | Simple | Advanced | Complex | Edge Cases | Terraschema |
| ------ | ----- | ------ | -------- | ------- | ---------- | ----------- |
| string | ✓     | ✓      | ✓        | ✓       | ✓          | ✓           |
| number | ✓     | ✓      | -        | ✓       | -          | ✓           |
| bool   | ✓     | -      | -        | ✓       | -          | ✓           |
| list   | ✓     | ✓      | ✓        | ✓       | -          | ✓           |
| object | ✓     | ✓      | ✓        | ✓       | -          | ✓           |
| map    | ✓     | -      | ✓        | ✓       | -          | ✓           |
| set    | -     | -      | -        | ✓       | -          | ✓           |
| tuple  | -     | -      | -        | ✓       | -          | ✓           |
| any    | ✓     | -      | -        | -       | -          | -           |

### Design Principles

1. **Systematic Coverage**: Each combination of type and validation complexity is tested
2. **Isolation**: Each test focuses on a specific feature or combination
3. **Realistic Examples**: Test cases use realistic variable definitions
4. **Incremental Complexity**: Tests progress from simple to complex scenarios

### Key Test Categories

#### Total Test Count: 28 tests across 6 categories

1. **Basic Features** (9 tests): Core type support including `any` type
2. **Simple Validation** (6 tests): Single validation rules on basic types
3. **Advanced Features** (4 tests): Complex type combinations with `alltrue` validation
4. **Complex Validation** (4 tests): Highly nested scenarios with tuple support
5. **Edge Cases** (1 test): Special validation scenarios and regex with OR conditions
6. **Terraschema Compatibility** (4 tests): Legacy compatibility with terraschema format

#### Constraint Types

- **Length**: String length, array/object size limits (minLength, maxLength, minItems, maxItems)
- **Range**: Numeric minimum/maximum values (minimum, maximum, exclusiveMinimum, exclusiveMaximum)
- **Pattern**: Regular expression validation with `can(regex())`
- **Enum**: Predefined value sets using `contains()` function
- **Collection Validation**: `alltrue` with `for` expressions for array/map validation

#### Nesting Levels

- **Flat**: Single-level structures (basic features)
- **Nested**: 2-3 levels of nesting (advanced features)
- **Complex**: 4+ levels with mixed types (complex validation)
- **Ultra-Complex**: Deep nesting with tuples, sets, and multiple validation rules

#### Validation Scope

- **Root**: Validation on the root variable
- **Property**: Validation on object properties and nested fields
- **Indexed**: Validation on specific array/tuple elements (e.g., `var.tuple[0]`)
- **Iterative**: Validation on all collection elements using `alltrue` and `for`

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
- **Validation Keywords**: minLength, maxLength, pattern, minimum, maximum, enum, exclusiveMinimum, exclusiveMaximum
- **Object Features**: properties, required, additionalProperties, minProperties
- **Array Features**: items, minItems, maxItems, uniqueItems
- **Tuple Features**: items array with positional schemas, minItems, maxItems for exact length

### Terraform-Specific Extensions

- **Optional Properties**: Support for `optional()` type modifier in object definitions
- **Nullable Types**: Support for `nullable = true` with `anyOf` schemas
- **Type Inference**: Schema generation from default values and type definitions
- **Path-Based Validation**: Support for nested property validation and indexed access
- **Complex Validation**: `alltrue` with `for` expressions for collection validation

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
  "additionalProperties": true,
  "properties": {
    "example_var": {
      "type": "string",
      "description": "An example variable",
      "default": "default_value",
      "minLength": 6
    }
  },
  "required": []
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
