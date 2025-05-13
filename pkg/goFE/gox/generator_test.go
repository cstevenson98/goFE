package gox

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestGenerateComponent(t *testing.T) {
	tests := []struct {
		name    string
		info    ComponentInfo
		want    string
		wantErr bool
	}{
		{
			name: "simple component",
			info: ComponentInfo{
				Name:  "Button",
				Props: []string{"Text", "OnClick"},
			},
			want:    "type Button struct",
			wantErr: false,
		},
		{
			name: "component with children",
			info: ComponentInfo{
				Name:     "Container",
				Props:    []string{"Title"},
				Children: []ast.Expr{},
			},
			want:    "type Container struct",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateComponent(tt.info)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateComponent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !strings.Contains(got, tt.want) {
				t.Errorf("GenerateComponent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateGoCode(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		want    string
		wantErr bool
	}{
		{
			name:    "simple component file",
			file:    "testdata/valid.gox",
			want:    "func Button(props *ButtonProps) string",
			wantErr: false,
		},
		{
			name:    "invalid gox file",
			file:    "testdata/invalid.gox",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Load test data
			source, err := LoadGoxFromFile(tt.file)
			if err != nil {
				t.Fatalf("Failed to load test file %s: %v", tt.file, err)
			}

			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, tt.file, source, parser.ParseComments)
			if err != nil && !tt.wantErr {
				t.Fatalf("Failed to parse test source: %v", err)
			}

			got, err := GenerateGoCode(file)
			if err != nil && !tt.wantErr {
				t.Errorf("GenerateGoCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !strings.Contains(got, tt.want) {
				t.Errorf("GenerateGoCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExprToString(t *testing.T) {
	tests := []struct {
		name string
		expr string
		want string
	}{
		{
			name: "simple identifier",
			expr: "Button",
			want: "Button",
		},
		{
			name: "selector expression",
			expr: "goFE.Component",
			want: "goFE.Component",
		},
		{
			name: "pointer type",
			expr: "*ButtonProps",
			want: "*ButtonProps",
		},
		{
			name: "array type",
			expr: "[]Component",
			want: "[]Component",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			expr, err := parser.ParseExprFrom(fset, "", tt.expr, 0)
			if err != nil {
				t.Fatalf("Failed to parse test expression: %v", err)
			}

			got := exprToString(expr)
			if got != tt.want {
				t.Errorf("exprToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStmtToString(t *testing.T) {
	tests := []struct {
		name    string
		stmt    string
		want    string
		wantErr bool
	}{
		{
			name:    "simple return",
			stmt:    "return `:begin-gox:<button>Click me</button>:end-gox:`",
			want:    "return `<button>Click me</button>`",
			wantErr: false,
		},
		{
			name:    "return with multiple values",
			stmt:    "return a, b",
			want:    "return a, b",
			wantErr: false,
		},
		{
			name:    "invalid gox string",
			stmt:    "return `:begin-gox:<button>Click me</button>`",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary Go file with the statement
			tempGo := fmt.Sprintf("package temp\nfunc temp() {\n\t%s\n}", tt.stmt)

			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "", tempGo, parser.ParseComments)
			if err != nil {
				t.Fatalf("Failed to parse test statement: %v", err)
			}

			// Get the function body's first statement
			funcDecl := file.Decls[0].(*ast.FuncDecl)
			stmt := funcDecl.Body.List[0]

			got, err := stmtToString(stmt)
			if (err != nil) != tt.wantErr {
				t.Errorf("stmtToString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("stmtToString() = %v, want %v", got, tt.want)
			}
		})
	}
}
