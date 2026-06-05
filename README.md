# golok

A Go code generation tool inspired by Lombok. Annotate your structs with a single comment and get standard methods generated into a `_golok.go` file.

Zero runtime dependencies. Pure `go generate`.

---

## Install

```bash
go install github.com/praiseson6065/golok/cmd/golok@latest
```

---

## Usage

Add a directive comment above your struct, then a `go:generate` line to wire the tool:

```go
//go:generate golok -file=$GOFILE

// +golok:constructor,getter,setter,builder,stringer,equals,clone
type User struct {
    Name  string
    Email string
    Age   int
}
```

Run it:
```bash
go generate ./...
```

This creates `user_golok.go` in the same package.

### Directory mode

Process all Go files in a directory at once:
```bash
golok -dir=./models
```

---

## Supported Directives

| Directive            | What it generates                                          |
|----------------------|------------------------------------------------------------|
| `all`                | Shorthand for all standard directives below                |
| `constructor`        | `NewT(field1, field2, ...) *T`                             |
| `getter`             | `GetField() T` for every field                             |
| `setter`             | `SetField(v T) *T` (fluent) for every field                |
| `builder`            | `TBuilder` with fluent API + `Build() *T`                  |
| `stringer`           | `String() string` using `fmt.Sprintf`                      |
| `equals`             | `Equal(other *T) bool` via `reflect.DeepEqual`             |
| `clone`              | `Clone() *T` shallow copy                                  |
| `functional_options` | `WithField(v)` option funcs + `NewT(opts...)` constructor  |
| `json`               | `ToJSON() ([]byte, error)` + `TFromJSON([]byte) (*T, error)` |
| `validate`           | `Validate() error` — checks fields tagged `validate:"required"` |
| `interface`          | `TInterface` with method signatures for all generated methods |
| `mapper=Target`      | `ToTarget() *Target` + `TFromTarget(*Target) *T`           |

Stack as many as you need, comma-separated:
```go
// +golok:constructor,stringer,equals
```

### Per-field control

Skip individual fields from code generation:
```go
// +golok:getter,setter
type User struct {
    Name     string
    Email    string
    password string // +golok:skip
}
```

### Validation with struct tags

```go
// +golok:validate
type User struct {
    Name  string `validate:"required"`
    Email string `validate:"required"`
    Age   int
}
```

### Functional options pattern

```go
// +golok:functional_options,stringer
type Server struct {
    Host string
    Port int
    TLS  bool
}
```
Generates `WithHost()`, `WithPort()`, `WithTLS()` option functions and `NewServer(opts ...)`.

> **Note:** `functional_options` conflicts with `constructor` — use one or the other.

### Struct mapper

Convert between structs with shared fields:
```go
// +golok:getter,mapper=UserDTO
type User struct {
    Name  string
    Email string
    Age   int
}

type UserDTO struct {
    Name  string
    Email string
}
```
Generates `ToUserDTO()` and `UserFromUserDTO()` — only copies fields that match by name and type.

### Interface generation

Auto-generate an interface from the struct's generated methods:
```go
// +golok:getter,stringer,interface
type User struct {
    Name  string
    Email string
}
```
Generates:
```go
type UserInterface interface {
    GetName() string
    GetEmail() string
    String() string
}
```

---

## Architecture

```
golok/
├── cmd/golok/main.go              # CLI: parses flags, orchestrates
├── internal/
│   ├── parser/parser.go           # go/ast: reads structs + directives
│   └── generator/
│       ├── generator.go           # Executes templates, calls gofmt
│       └── templates.go           # One template constant per method
└── example/
    ├── user.go                    # Input: annotated structs
    └── user_golok.go              # Output: generated methods
```

### How it works

1. **Parse** — `go/ast` walks the file. Any `GenDecl` with a `// +golok:...` comment gets collected with its fields, types, and struct tags.
2. **Generate** — For each struct × method combo, a `text/template` is executed with `StructInfo` as data. Mapper uses cross-struct resolution via `allStructs` index.
3. **Format** — `go/format.Source()` runs gofmt on the output so whitespace is always clean.
4. **Write** — `<input>_golok.go` is written in the same package (so unexported fields are accessible too).

---

## Extending with a new method

1. Add a template constant to `internal/generator/templates.go`:
```go
const myTemplate = `
func (s *{{.Name}}) MyMethod() {
    {{range .Fields}}
    _ = s.{{.Name}}
    {{end}}
}
`
```

2. Register it in `methodTemplates`:
```go
var methodTemplates = map[string]string{
    // ...existing...
    "mymethod": myTemplate,
}
```

3. If you need a new import, add a case in `neededImports()`.

That's it. No codegen framework needed.

---

## Known limitations

- **Shallow clone** — `Clone()` does a value copy (`c := *s`). Slice/map fields share the underlying array. For deep copies, do it manually.
- **`==` works for scalar fields** — `Equal()` uses `reflect.DeepEqual` which handles slices and maps correctly.
- **Mapper requires same file** — Both source and target structs must be in the same `.go` file.
- **Embedded fields** — Skipped. Only named fields are processed.

---

## Running tests

```bash
go test ./...
```
