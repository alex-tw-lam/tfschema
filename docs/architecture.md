# tfschema Architecture

`tfschema` converts Terraform `variable` blocks into a JSON Schema (Draft 7) so
that `.tfvars.json` files can be validated against the same constraints Terraform
enforces. This document describes how the conversion actually flows.

## Package layout

```
cmd/tfschema/         CLI entry point
internal/converter/   Orchestrates HCL → JSON Schema conversion
internal/converter/types/   Type converters (primitive, list, object, map, set, optional, tuple)
internal/validation/  Parses Terraform `validation` blocks into JSON Schema constraints
internal/jsonschema/  The JSON Schema Draft-7 struct and JSON tagging
```

## Conversion pipeline

`cmd/tfschema/main.go` reads a `.tf` file and calls
`converter.Converter.ConvertFile`. The converter then does the following for each
`variable` block:

1. **Parse HCL.** The file is parsed with `hashicorp/hcl/v2` (`hclparse`).
2. **Convert the type.** `Converter.ConvertType` dispatches the `type`
   expression to a converter registered in a small `TypeConverterRegistry`
   (`internal/converter/types`). Each converter knows one Terraform type
   constructor — `string`/`number`/`bool`/`any`, `list(...)`, `set(...)`,
   `map(...)`, `object({...})`, `tuple([...])`, and `optional(...)` — and
   returns the corresponding JSON Schema fragment.
3. **Apply attributes.** An `AttributeProcessor` applies the variable's
   `description`, `default`, `sensitive`, and `nullable` attributes to the
   schema via individual appliers (`description_applier.go`,
   `default_applier.go`, `sensitive_applier.go`, `nullable_applier.go`).
4. **Apply validation rules.** `ValidationProcessor` reads each `validation {}`
   block and asks the validation registry for a matching rule
   (`internal/validation`). Each rule carries a JSON-Schema path so it can be
   attached at the right location — including nested fields and indexed tuple
   positions.
5. **Infer the type when missing.** If a variable has no explicit `type`,
   `type_inference.go` derives a schema from its `default` value.

The root schema is always an `object` whose `properties` map holds one entry per
variable; `required` lists the variables that have no `default`.

## Validation rules

Validation rules are self-registering parsers (`internal/validation`):

| Parser | Matches | Produces |
| ------ | ------- | -------- |
| `length_rule` | `length(var.x) > N` etc. | `minLength`/`maxLength`, `minItems`/`maxItems` |
| `range_rule` | `var.x >= N && var.x <= M`, equality, OR-enums | `minimum`/`maximum`, `enum` |
| `regex_rule` | `can(regex(...))` | `pattern` |
| `enum_rule` | `contains([...], var.x)` | `enum` |
| `alltrue_rule` | `alltrue([for x in var.coll : ...])` | re-dispatches the loop body to the other parsers |

`alltrue` is a **composite** parser: it has the highest priority (so it runs
first) and, for the body of its `for` expression, re-runs every *other*
registered parser via `GetParsersExcept("alltrue")`. The resulting rule is
applied to every element of the collection (path segment `"*"`). This is the only
recursion in the system, and excluding itself by name — rather than by function
identity — keeps it simple and dependency-free.
