# tfschema

A Go tool that converts Terraform variable definitions (`variable` blocks) to
[JSON Schema (draft-07)](https://json-schema.org/draft-07/json-schema-core), so
`.tfvars.json` files can be validated against their Terraform variable
constraints.

```hcl
variable "server_config" {
  type = object({
    name = string
    port = number
  })

  validation {
    condition     = var.server_config.port >= 1 && var.server_config.port <= 65535
    error_message = "Port must be between 1 and 65535."
  }
}
```

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "server_config": {
      "type": "object",
      "properties": {
        "name": { "type": "string" },
        "port": { "type": "number", "minimum": 1, "maximum": 65535 }
      },
      "required": ["name", "port"],
      "additionalProperties": true
    }
  },
  "required": ["server_config"],
  "additionalProperties": true
}
```

## Install

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

Typical workflow:

1. Point `tfschema` at a `.tf` file and save the schema:

   ```bash
   tfschema variables.tf > schema.json
   ```

2. Install a JSON Schema validator once. Any draft-07 validator works; `ajv-cli` shown here:

   ```bash
   npm install -g ajv-cli
   ```

3. Validate a tfvars file against the schema:

   ```bash
   ajv validate -s schema.json -d terraform.tfvars.json
   ```

`tfschema -version` prints the build version.

## What is supported

### Types

- `string`, `number`, `bool`, `any`
- `list(type)`, `set(type)`, `map(type)`
- `object({ field = type, ... })` with `optional(type)` attributes
- `tuple([type1, type2, ...])`

### Variable attributes

- `description` → `description`
- `default` → `default` (also keeps a variable out of `required`)
- `sensitive = true` → `sensitive: true`
- `nullable = true` → an `anyOf` choice between `null` and the type

### `validation` blocks

- String length: `length(var.field) > N` (and `>=`, `<`, `<=`, `==`)
- Regex: `can(regex("pattern", var.field))` → `pattern`
- Number range: `var.field >= N && var.field <= M`
- Enum: `contains(["a", "b"], var.field)` or `var.field == "a" || var.field == "b"` → `enum`
- Per-element: `alltrue([for item in var.list : ...])` applies the inner rule to every element
- Indexed access: `var.tuple[0]`, `var.data[2].field`

A condition no parser understands is skipped with a warning on stderr
(pointing at the file and line); the JSON on stdout stays clean.

## How it works

The conversion pipeline, in order:

1. Parse the `.tf` file into an HCL body (`converter.go`)
2. Convert each variable's type expression (`typemap.go`)
3. Copy `description` / `default` / `sensitive` / `nullable` onto the schema (`attributes.go`)
4. Apply each `validation` block to the matching spot in the schema (`validation.go`)
5. Print the assembled schema as JSON

The code base is small on purpose: each file has one job. Reading order:

| File | Job |
|---|---|
| `cmd/tfschema/main.go` | CLI: read file, print JSON schema |
| `internal/jsonschema/schema.go` | The JSON Schema data structure |
| `internal/converter/converter.go` | Parses HCL, walks `variable` blocks, assembles the root schema |
| `internal/converter/typemap.go` | Type expressions → `cty.Type` (official `typeexpr` library) → JSON Schema |
| `internal/converter/attributes.go` | `description`/`default`/`sensitive`/`nullable` → schema fields |
| `internal/converter/validation.go` | Applies each parsed rule to the right spot in the schema |
| `internal/validation/validation.go` | Extracts `validation` blocks, tries the rule parsers in order |
| `internal/validation/path_expression.go` | Reads `var.foo.bar[0]` paths out of condition expressions |
| `internal/validation/expression.go` | Shared helpers for reading HCL expressions (literals, variable refs, parens) |
| `internal/validation/length_rule.go` | `length(var.x) > n` rules |
| `internal/validation/range_rule.go` | number comparison rules (`var.x >= 1 && var.x <= 10`) |
| `internal/validation/enum_rule.go` | enum rules (`contains([...], var.x)`, `==`/`\|\|` chains) |
| `internal/validation/regex_rule.go` | `can(regex(...))` rules |
| `internal/validation/alltrue_rule.go` | `alltrue([for ...])` rules |

The rule parsers are tried in the order listed in `validation.go` until one
recognises a condition. Adding a new rule kind means writing one parser
function and adding it to that list.

## Testing

```bash
go test ./...
```

The `tests/` directory holds 28 end-to-end scenarios (a `test.tf` plus an
expected `test.schema.json` each); the suite converts each one and compares
the output. See [docs/testing.md](./docs/testing.md) for the scenario
catalogue and [docs/releasing.md](./docs/releasing.md) for the release
process.

## License

Apache License 2.0. See [LICENSE](./LICENSE).
