# tfschema Test Suite

`go test ./...` runs unit tests per package plus one data-driven end-to-end
suite. Each e2e scenario is a directory under `tests/` containing a `test.tf`
and the expected `test.schema.json`; the suite converts the input and compares
the generated schema against the expectation, so behaviour changes show up as
plain JSON diffs.

To regenerate every expectation after an intentional behaviour change:

```bash
./scripts/regenerate_test_schemas.sh
```

## Scenario catalogue (28 scenarios)

### 1. Basic features (`1-basic-features/`)

Core type support without validation:

- **01-string-none-basic**: Basic string variable
- **02-number-none-basic**: Basic number variable
- **03-bool-none-basic**: Basic boolean variable
- **04-list-none-basic**: Basic list variable
- **05-object-none-basic**: Basic object variable
- **06-map-none-basic**: Basic map variable
- **07-bool-enum-basic**: Boolean with enum validation
- **08-string-enum-basic**: String with enum validation
- **23-any-type-basic**: Variable with `any` type

### 2. Simple validation (`2-simple-validation/`)

One validation rule on a basic type:

- **09-string-length-basic**: String with length constraints
- **10-string-regex-basic**: String with regex pattern
- **11-number-range-basic**: Number with range constraints
- **12-number-enum-basic**: Number with enum validation
- **13-list-length-basic**: List with length constraints
- **14-object-length-basic**: Object with property count constraints

### 3. Advanced features (`3-advanced-features/`)

Complex type combinations:

- **15-list-enum-advanced**: List of objects with enum validation using `alltrue`
- **16-object-enum-advanced**: Object with nested enum validation
- **17-map-length-advanced**: Map with length constraints
- **18-map-enum-advanced**: Map with enum validation using `alltrue`

### 4. Complex validation (`4-complex-validation/`)

Deeply nested structures:

- **19-highly-nested-complex**: Multi-level object with various validation types
- **20-set-length-basic**: Set with length constraints and uniqueness
- **21-tuple-nested-validation-complex**: Tuple with indexed validation
- **22-ultra-complex-nesting**: Tuples, sets and deep nesting combined

### 5. Edge cases (`5-edge-cases/`)

- **24-regex-with-or-condition**: Regex validation with an OR condition (empty string allowed)

### 6. Terraschema compatibility (`6-terraschema-compat/`)

Outputs must stay compatible with the original
[terraschema](https://github.com/larryclampton/terraschema) project:

- **25-terraschema-simple**: Basic terraschema compatibility
- **26-terraschema-simple-types**: All basic types including nullable
- **27-terraschema-complex-types**: Complex nested types with optional fields
- **28-terraschema-custom-validation**: Various validation patterns

## Test design

The scenarios are a subset of the covering array generated from a
category-partition model of the inputs (type × element × depth × validation
kind); see `tests/model.pict` and `tests/test-spec.csv`. New scenarios should
extend the model rather than adding isolated cases.

## Optional: cross-validation with terraform and the jsonschema CLI

`./scripts/validate_tests.sh` additionally runs `terraform plan` on each
scenario and checks the tfvars files against the generated schemas with the
`jsonschema` CLI; it requires both tools on `PATH` and is not part of
`go test`.
