package gox

import (
	"testing"
)

func TestParseJSX(t *testing.T) {
	tests := []struct {
		name    string
		jsx     string
		wantErr bool
	}{
		{
			name:    "valid JSX",
			jsx:     "`:begin-gox: <button>Click me</button>` :end-gox:",
			wantErr: false,
		},
		{
			name:    "valid JSX with props",
			jsx:     "`:begin-gox: <button onClick={props.OnClick}>{props.Text}</button>` :end-gox:",
			wantErr: false,
		},
		{
			name:    "invalid JSX",
			jsx:     "`:begin-gox: <button>Click me</button>` :end-gox:",
			wantErr: true,
		},
		{
			name:    "empty JSX",
			jsx:     "`:begin-gox: ` :end-gox:",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := ParseJSX(tt.jsx)
			if err != nil && !tt.wantErr {
				t.Errorf("ParseJSX() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && expr == nil {
				t.Error("ParseJSX() returned nil expression for valid JSX")
			}
		})
	}
}

func TestParseFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{
			name:     "valid gox file",
			filename: "testdata/valid.gox",
			wantErr:  false,
		},
		{
			name:     "invalid gox file",
			filename: "testdata/invalid.gox",
			wantErr:  true,
		},
		{
			name:     "non-existent file",
			filename: "testdata/doesnotexist.gox",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := ParseFile(tt.filename)
			if err != nil && !tt.wantErr {
				t.Errorf("ParseFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && file == nil {
				t.Error("ParseFile() returned nil file for valid input")
			}
		})
	}
}
