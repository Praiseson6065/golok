package generator

// methodTemplates maps a directive name to its Go template string.
// Templates receive a parser.StructInfo as data.
// Available funcs: title (capitalize first char), lower (lowercase first char)
var methodTemplates = map[string]string{
	"getter":      getterTemplate,
	"setter":      setterTemplate,
	"constructor": constructorTemplate,
	"builder":     builderTemplate,
	"stringer":    stringerTemplate,
	"equals":      equalsTemplate,
	"clone":       cloneTemplate,
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
