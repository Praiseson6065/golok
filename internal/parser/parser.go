package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// allMethods is the full set of directives, used to expand "all".
// Note: functional_options is excluded because it conflicts with constructor.
// Note: interface and mapper are opt-in only.
var allMethods = []string{
	"constructor", "getter", "setter", "builder",
	"stringer", "equals", "clone", "json", "validate",
}

// StructInfo holds all data needed to generate methods for one struct.
type StructInfo struct {
	Name          string
	Package       string
	Fields        []FieldInfo
	Methods       []string // e.g. ["getter", "setter", "builder"]
	MapperTargets []string // e.g. ["UserDTO"] from mapper=UserDTO
}

// FieldInfo represents a single struct field.
type FieldInfo struct {
	Name string
	Type string
	Tag  string // raw struct tag, e.g. json:"name" validate:"required"
}

// ParseFile parses a .go file and returns:
//   - annotated: structs with // +golok:... directives (for code generation)
//   - allStructs: every struct in the file by name (for mapper target resolution)
//   - pkg: the package name
func ParseFile(filename string) (annotated []StructInfo, allStructs map[string]StructInfo, pkg string, err error) {
	fset := token.NewFileSet()
	f, parseErr := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if parseErr != nil {
		return nil, nil, "", parseErr
	}

	pkg = f.Name.Name
	allStructs = make(map[string]StructInfo)

	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		methods, mapperTargets := parseDirective(genDecl.Doc)

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
				Name:          typeSpec.Name.Name,
				Package:       pkg,
				Methods:       methods,
				MapperTargets: mapperTargets,
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

			// Always add to allStructs for mapper lookups
			allStructs[info.Name] = info

			// Only add to annotated if it has directives
			if len(methods) > 0 || len(mapperTargets) > 0 {
				annotated = append(annotated, info)
			}
		}
	}

	return annotated, allStructs, pkg, nil
}

// parseDirective extracts method names and mapper targets from a comment like:
//
//	// +golok:getter,setter,mapper=UserDTO
func parseDirective(doc *ast.CommentGroup) (methods []string, mapperTargets []string) {
	if doc == nil {
		return nil, nil
	}
	const prefix = "// +golok:"
	for _, c := range doc.List {
		text := strings.TrimSpace(c.Text)
		if !strings.HasPrefix(text, prefix) {
			continue
		}
		raw := strings.TrimPrefix(text, prefix)
		for _, p := range strings.Split(raw, ",") {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			// Handle mapper=Target directives
			if strings.HasPrefix(p, "mapper=") {
				target := strings.TrimPrefix(p, "mapper=")
				if target != "" {
					mapperTargets = append(mapperTargets, target)
				}
				continue
			}
			methods = append(methods, p)
		}
		// Expand "all" shorthand to every known directive
		for _, m := range methods {
			if m == "all" {
				return allMethods, mapperTargets
			}
		}
		return methods, mapperTargets
	}
	return nil, nil
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
