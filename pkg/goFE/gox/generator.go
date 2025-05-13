package gox

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"strings"
)

// ComponentInfo holds all the necessary information to generate a Go component
// from a JSX-like declaration. This includes the component name, its props,
// children components, and any HTML attributes.
type ComponentInfo struct {
	Name       string            // The name of the component (e.g., "Button")
	Props      []string          // List of prop names that the component accepts
	Children   []ast.Expr        // Child components or expressions within the component
	Attributes map[string]string // HTML attributes and their values
}

// GenerateComponent generates Go code for a component based on the provided ComponentInfo.
// It creates:
// 1. A struct definition for the component
// 2. A constructor function
// 3. Implementation of the goFE.Component interface
// 4. A render method that generates the HTML
func GenerateComponent(info ComponentInfo) (string, error) {
	var sb strings.Builder

	// Generate the component struct with an ID field and all props
	sb.WriteString(fmt.Sprintf("type %s struct {\n", info.Name))
	sb.WriteString("\tid uuid.UUID\n") // Required for goFE.Component interface
	for _, prop := range info.Props {
		sb.WriteString(fmt.Sprintf("\t%s interface{}\n", prop))
	}
	sb.WriteString("}\n\n")

	// Generate the constructor function that takes props and returns a new component instance
	sb.WriteString(fmt.Sprintf("func New%s(props *%sProps) *%s {\n", info.Name, info.Name, info.Name))
	sb.WriteString(fmt.Sprintf("\treturn &%s{\n", info.Name))
	sb.WriteString("\t\tid: uuid.New(),\n") // Generate a new UUID for the component
	for _, prop := range info.Props {
		sb.WriteString(fmt.Sprintf("\t\t%s: props.%s,\n", prop, prop))
	}
	sb.WriteString("\t}\n")
	sb.WriteString("}\n\n")

	// Generate the GetID method required by goFE.Component interface
	sb.WriteString(fmt.Sprintf("func (c *%s) GetID() uuid.UUID {\n", info.Name))
	sb.WriteString("\treturn c.id\n")
	sb.WriteString("}\n\n")

	// Generate the GetChildren method required by goFE.Component interface
	sb.WriteString(fmt.Sprintf("func (c *%s) GetChildren() []goFE.Component {\n", info.Name))
	sb.WriteString("\treturn nil\n") // TODO: Implement child component handling
	sb.WriteString("}\n\n")

	// Generate the InitEventListeners method required by goFE.Component interface
	sb.WriteString(fmt.Sprintf("func (c *%s) InitEventListeners() {\n", info.Name))
	sb.WriteString("\t// TODO: Initialize event listeners\n")
	sb.WriteString("}\n\n")

	// Generate the Render method required by goFE.Component interface
	sb.WriteString(fmt.Sprintf("func (c *%s) Render() string {\n", info.Name))
	sb.WriteString("\treturn c.render()\n")
	sb.WriteString("}\n\n")

	return sb.String(), nil
}

// GenerateGoCode generates Go code from a parsed .gox file.
// It processes the AST and generates equivalent Go code that implements
// the goFE.Component interface and handles the JSX-like syntax.
func GenerateGoCode(file *ast.File) (string, error) {
	var sb strings.Builder

	// Add package declaration from the original file
	sb.WriteString("package " + file.Name.Name + "\n\n")

	// Add required imports for goFE framework
	sb.WriteString("import (\n")
	sb.WriteString("\t\"github.com/cstevenson98/goFE/pkg/goFE\"\n")
	sb.WriteString("\t\"github.com/google/uuid\"\n")
	sb.WriteString("\t\"html/template\"\n")
	sb.WriteString(")\n\n")

	// Process each declaration in the file
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			// Process function declarations
			if d.Recv == nil {
				// This is a top-level function (component definition)
				code, err := generateFunction(d)
				if err != nil {
					return "", err
				}
				sb.WriteString(code)
			}
		case *ast.GenDecl:
			// Process type declarations
			if d.Tok == token.TYPE {
				for _, spec := range d.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						code, err := generateType(typeSpec)
						if err != nil {
							return "", err
						}
						sb.WriteString(code)
					}
				}
			}
		}
	}

	// Format the generated code using go/format
	src := []byte(sb.String())
	formatted, err := format.Source(src)
	if err != nil {
		return "", err
	}

	return string(formatted), nil
}

// generateType generates Go code for a type declaration
func generateType(typeSpec *ast.TypeSpec) (string, error) {
	var sb strings.Builder

	// Generate type declaration
	sb.WriteString("type " + typeSpec.Name.Name + " ")
	sb.WriteString(exprToString(typeSpec.Type))
	sb.WriteString("\n\n")

	return sb.String(), nil
}

// validateGoxString checks if a string contains both :begin-gox: and :end-gox: markers
func validateGoxString(str string) error {
	start := strings.Index(str, ":begin-gox:")
	end := strings.Index(str, ":end-gox:")

	if start == -1 {
		return fmt.Errorf("missing :begin-gox: marker")
	}
	if end == -1 {
		return fmt.Errorf("missing :end-gox: marker")
	}
	if end <= start {
		return fmt.Errorf(":end-gox: marker must come after :begin-gox: marker")
	}
	return nil
}

// stmtToString converts an ast.Stmt to a string representation.
// Currently handles:
// - Return statements
// - Multiple return values
// Will be extended to handle more statement types as needed.
func stmtToString(stmt ast.Stmt) (string, error) {
	switch s := stmt.(type) {
	case *ast.ReturnStmt:
		var sb strings.Builder
		sb.WriteString("return ")
		for i, result := range s.Results {
			if i > 0 {
				sb.WriteString(", ")
			}
			// If the result is a string literal, extract the gox content
			if bl, ok := result.(*ast.BasicLit); ok && bl.Kind == token.STRING {
				str := bl.Value
				// Remove quotes
				if len(str) > 1 && (str[0] == '`' && str[len(str)-1] == '`' || str[0] == '"' && str[len(str)-1] == '"') {
					str = str[1 : len(str)-1]
				}
				if err := validateGoxString(str); err != nil {
					return "", err
				}
				start := strings.Index(str, ":begin-gox:")
				end := strings.Index(str, ":end-gox:")
				goxContent := str[start+len(":begin-gox:") : end]
				sb.WriteString("`" + goxContent + "`")
			} else {
				sb.WriteString(exprToString(result))
			}
		}
		return sb.String(), nil
	default:
		return "", nil
	}
}

// generateFunction generates Go code for a function declaration.
// It handles the conversion of JSX-like syntax to Go code and ensures
// proper implementation of the goFE.Component interface.
func generateFunction(funcDecl *ast.FuncDecl) (string, error) {
	var sb strings.Builder

	// Generate function signature
	sb.WriteString("func " + funcDecl.Name.Name)
	sb.WriteString("(")
	for i, param := range funcDecl.Type.Params.List {
		if i > 0 {
			sb.WriteString(", ")
		}
		for j, name := range param.Names {
			if j > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(name.Name)
		}
		sb.WriteString(" ")
		sb.WriteString(exprToString(param.Type))
	}
	sb.WriteString(") ")

	// Generate return type
	if funcDecl.Type.Results != nil {
		sb.WriteString(exprToString(funcDecl.Type.Results.List[0].Type))
	}
	sb.WriteString(" {\n")

	// Generate function body
	if funcDecl.Body != nil {
		for _, stmt := range funcDecl.Body.List {
			stmtStr, err := stmtToString(stmt)
			if err != nil {
				return "", err
			}
			sb.WriteString("\t" + stmtStr + "\n")
		}
	}

	sb.WriteString("}\n\n")
	return sb.String(), nil
}

// exprToString converts an ast.Expr to a string representation.
// It handles various types of expressions:
// - Identifiers (e.g., "Button")
// - Selector expressions (e.g., "goFE.Component")
// - Pointer types (e.g., "*ButtonProps")
// - Array types (e.g., "[]Component")
// - Call expressions (e.g., "template.New(...)")
func exprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return exprToString(e.X) + "." + e.Sel.Name
	case *ast.StarExpr:
		return "*" + exprToString(e.X)
	case *ast.ArrayType:
		return "[]" + exprToString(e.Elt)
	case *ast.CallExpr:
		var sb strings.Builder
		sb.WriteString(exprToString(e.Fun))
		sb.WriteString("(")
		for i, arg := range e.Args {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(exprToString(arg))
		}
		sb.WriteString(")")
		return sb.String()
	case *ast.BasicLit:
		return e.Value
	case *ast.StructType:
		var sb strings.Builder
		sb.WriteString("struct {\n")
		for _, field := range e.Fields.List {
			for _, name := range field.Names {
				sb.WriteString("\t" + name.Name + " ")
				sb.WriteString(exprToString(field.Type))
				sb.WriteString("\n")
			}
		}
		sb.WriteString("}")
		return sb.String()
	default:
		return ""
	}
}
