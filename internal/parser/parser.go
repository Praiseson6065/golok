package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

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
				for _, name := range field.Names {
					info.Fields = append(info.Fields, FieldInfo{
						Name: name.Name,
						Type: exprToString(field.Type),
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
