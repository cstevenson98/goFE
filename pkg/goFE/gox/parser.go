package gox

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

// ComponentWrapper represents a wrapper around a goFE Component
type ComponentWrapper struct {
	Component interface{}
}

// LoadGoxFromFile loads the contents of a .gox file as a string
func LoadGoxFromFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// ParseFile parses a .gox file and returns the AST
func ParseFile(filename string) (*ast.File, error) {
	content, err := LoadGoxFromFile(filename)
	if err != nil {
		return nil, err
	}

	// Validate GoX markers in the content
	if err := validateGoxString(content); err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	return parser.ParseFile(fset, filename, content, parser.ParseComments)
}

// ParseJSX parses JSX-like syntax in a string
func ParseJSX(jsx string) (*ast.CallExpr, error) {
	// Remove the $` prefix and ` suffix if present
	jsx = strings.TrimPrefix(jsx, "$`")
	jsx = strings.TrimSuffix(jsx, "`")

	// Create a temporary Go file with the JSX wrapped in a function
	tempGo := fmt.Sprintf("package temp\nfunc temp() {\n\treturn template.New(\"%s\")\n}", jsx)

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", tempGo, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	// Extract the function body
	if len(file.Decls) == 0 {
		return nil, fmt.Errorf("no declarations found")
	}

	funcDecl, ok := file.Decls[0].(*ast.FuncDecl)
	if !ok {
		return nil, fmt.Errorf("expected function declaration")
	}

	if len(funcDecl.Body.List) == 0 {
		return nil, fmt.Errorf("empty function body")
	}

	// Get the first statement
	stmt := funcDecl.Body.List[0]

	// Convert to return statement
	retStmt, ok := stmt.(*ast.ReturnStmt)
	if !ok {
		return nil, fmt.Errorf("expected return statement")
	}

	if len(retStmt.Results) == 0 {
		return nil, fmt.Errorf("empty return statement")
	}

	// Convert to call expression
	callExpr, ok := retStmt.Results[0].(*ast.CallExpr)
	if !ok {
		return nil, fmt.Errorf("expected call expression")
	}

	return callExpr, nil
}
