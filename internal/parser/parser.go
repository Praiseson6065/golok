package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// allMethods is the full set of directives, used to expand "all".
// Note: functional_options is excluded because it conflicts with constructor.
var allMethods = []string{
	"constructor", "getter", "setter", "builder",
	"stringer", "equals", "clone", "json", "validate",
}

// StructInfo holds all data needed to generate methods for one struct.
type StructInfo struct {
	Name    string
	Package string
	Fields  []FieldInfo
	Methods []string // e.g. ["getter", "setter", "builder"]
}

// FieldInfo represents a single struct field.
type FieldInfo struct {
	Name string
	Type string
	Tag  string // raw struct tag, e.g. json:"name" validate:"required"
}

// ParseFile parses a .go file and returns all structs annotated with // +golok:...
func ParseFile(filename string) ([]StructInfo, string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, "", err
	}

	pkg := f.Name.Name
	var structs []StructInfo

	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		methods := parseDirective(genDecl.Doc)
		if len(methods) == 0 {
			continue // no golok directive, skip
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			info := StructInfo{
				Name:    typeSpec.Name.Name,
				Package: pkg,
				Methods: methods,
			}

			for _, field := range structType.Fields.List {
				if len(field.Names) == 0 {
					continue // skip embedded fields
				}
				if hasSkipDirective(field) {
					continue // field marked with +golok:skip
				}
				// Extract struct tag if present (strip backticks)
				var tag string
				if field.Tag != nil {
					tag = strings.Trim(field.Tag.Value, "`")
				}
				for _, name := range field.Names {
					info.Fields = append(info.Fields, FieldInfo{
						Name: name.Name,
						Type: exprToString(field.Type),
						Tag:  tag,
					})
				}
			}

			structs = append(structs, info)
		}
	}

	return structs, pkg, nil
}

// parseDirective extracts method names from a comment like:
//
//	// +golok:getter,setter,builder
func parseDirective(doc *ast.CommentGroup) []string {
	if doc == nil {
		return nil
	}
	const prefix = "// +golok:"
	for _, c := range doc.List {
		text := strings.TrimSpace(c.Text)
		if !strings.HasPrefix(text, prefix) {
			continue
		}
		raw := strings.TrimPrefix(text, prefix)
		var methods []string
		for _, p := range strings.Split(raw, ",") {
			if p = strings.TrimSpace(p); p != "" {
				methods = append(methods, p)
			}
		}
		// Expand "all" shorthand to every known directive
		for _, m := range methods {
			if m == "all" {
				return allMethods
			}
		}
		return methods
	}
	return nil
}

// exprToString converts an AST type expression to its string representation.
func exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + exprToString(t.X)
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + exprToString(t.Elt)
		}
		return "[...]" + exprToString(t.Elt)
	case *ast.MapType:
		return "map[" + exprToString(t.Key) + "]" + exprToString(t.Value)
	case *ast.InterfaceType:
		return "any"
	case *ast.ChanType:
		return "chan " + exprToString(t.Value)
	default:
		return "any"
	}
}

// hasSkipDirective checks whether a struct field has a // +golok:skip annotation
// in either its doc comment (above) or inline comment (trailing).
func hasSkipDirective(field *ast.Field) bool {
	for _, group := range []*ast.CommentGroup{field.Doc, field.Comment} {
		if group == nil {
			continue
		}
		for _, c := range group.List {
			if strings.Contains(c.Text, "+golok:skip") {
				return true
			}
		}
	}
	return false
}
