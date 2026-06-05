package generator

// methodTemplates maps a directive name to its Go template string.
// Templates receive a parser.StructInfo as data.
// Available funcs: title (capitalize first char), lower (lowercase first char)
var methodTemplates = map[string]string{
	"getter":             getterTemplate,
	"setter":             setterTemplate,
	"constructor":        constructorTemplate,
	"builder":            builderTemplate,
	"stringer":           stringerTemplate,
	"equals":             equalsTemplate,
	"clone":              cloneTemplate,
	"functional_options": functionalOptionsTemplate,
	"json":               jsonTemplate,
	"validate":           validateTemplate,
}

// getter generates a Get<Field>() method for every field.
const getterTemplate = `
{{- range .Fields}}
func (s *{{$.Name}}) Get{{title .Name}}() {{.Type}} {
	return s.{{.Name}}
}
{{end}}`

// setter generates a fluent Set<Field>(v) *T method for every field.
const setterTemplate = `
{{- range .Fields}}
func (s *{{$.Name}}) Set{{title .Name}}(v {{.Type}}) *{{$.Name}} {
	s.{{.Name}} = v
	return s
}
{{end}}`

// constructor generates a NewT(field1, field2, ...) *T function.
const constructorTemplate = `
func New{{.Name}}({{range $i, $f := .Fields}}{{if $i}}, {{end}}{{lower $f.Name}} {{$f.Type}}{{end}}) *{{.Name}} {
	return &{{.Name}}{
		{{- range .Fields}}
		{{.Name}}: {{lower .Name}},
		{{- end}}
	}
}
`

// builder generates a TBuilder struct with a fluent API and Build() method.
const builderTemplate = `
// {{.Name}}Builder builds a {{.Name}} using a fluent API.
type {{.Name}}Builder struct {
	obj {{.Name}}
}

// New{{.Name}}Builder returns a new, zero-value builder.
func New{{.Name}}Builder() *{{.Name}}Builder {
	return &{{.Name}}Builder{}
}

{{range .Fields -}}
func (b *{{$.Name}}Builder) {{title .Name}}(v {{.Type}}) *{{$.Name}}Builder {
	b.obj.{{.Name}} = v
	return b
}
{{end}}
// Build returns a pointer to the constructed {{.Name}}.
func (b *{{.Name}}Builder) Build() *{{.Name}} {
	c := b.obj
	return &c
}
`

// stringer generates a String() method that prints all fields.
// Uses {{lbrace}} / {{rbrace}} helpers to avoid the {{{ template parsing ambiguity.
const stringerTemplate = `
func (s *{{.Name}}) String() string {
	return fmt.Sprintf("{{.Name}}{{lbrace}}{{range $i, $f := .Fields}}{{if $i}}, {{end}}{{$f.Name}}=%v{{end}}{{rbrace}}",
		{{range $i, $f := .Fields}}{{if $i}}, {{end}}s.{{$f.Name}}{{end}})
}
`

// equals generates an Equal(*T) bool method using reflect.DeepEqual.
// Works correctly for structs containing slices, maps, and nested types.
const equalsTemplate = `
func (s *{{.Name}}) Equal(other *{{.Name}}) bool {
	if other == nil {
		return false
	}
	return reflect.DeepEqual(*s, *other)
}
`

// clone generates a shallow Clone() *T method.
const cloneTemplate = `
// Clone returns a shallow copy of {{.Name}}.
// For deep copies of slice/map fields, copy them manually.
func (s *{{.Name}}) Clone() *{{.Name}} {
	c := *s
	return &c
}
`

// functional_options generates the functional options pattern:
// - A {{.Name}}Option type
// - With{{Field}}(v) option functions for each field
// - New{{.Name}}(opts ...{{.Name}}Option) constructor
// Note: conflicts with "constructor" — use one or the other.
const functionalOptionsTemplate = `
// {{.Name}}Option configures a {{.Name}} instance.
type {{.Name}}Option func(*{{.Name}})

{{range .Fields -}}
// With{{title .Name}} sets the {{.Name}} field.
func With{{title .Name}}(v {{.Type}}) {{$.Name}}Option {
	return func(s *{{$.Name}}) {
		s.{{.Name}} = v
	}
}
{{end}}
// New{{.Name}} creates a {{.Name}} with the given options applied.
func New{{.Name}}(opts ...{{.Name}}Option) *{{.Name}} {
	s := &{{.Name}}{}
	for _, opt := range opts {
		opt(s)
	}
	return s
}
`

// json generates ToJSON() and FromJSON() helper methods.
const jsonTemplate = `
// ToJSON serializes {{.Name}} to JSON bytes.
func (s *{{.Name}}) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}

// {{.Name}}FromJSON deserializes JSON bytes into a {{.Name}}.
func {{.Name}}FromJSON(data []byte) (*{{.Name}}, error) {
	var s {{.Name}}
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}
`

// validate generates a Validate() method that checks fields tagged with validate:"required".
const validateTemplate = `
// Validate checks required fields and returns an error if any are missing.
func (s *{{.Name}}) Validate() error {
	{{- range .Fields}}
	{{- if isRequired .Tag}}
	if {{zeroCheck .Name .Type}} {
		return fmt.Errorf("{{$.Name}}.{{.Name}} is required")
	}
	{{- end}}
	{{- end}}
	return nil
}
`
