package codegeneration

import "testing"

func TestProcessElementName(t *testing.T) {
	tests := []struct {
		name          string
		fullName      string
		dependencies  []string
		wantPseudo    string
		wantFormatted string
	}{
		{
			name:          "Basic snake_case",
			fullName:      "dependency.module.example_name(abc)",
			dependencies:  []string{"dependency", "module"},
			wantPseudo:    "example_name",
			wantFormatted: "example name",
		},
		{
			name:          "Basic CamelCase",
			fullName:      "dependency.module.ExampleName(abc)",
			dependencies:  []string{"dependency", "module"},
			wantPseudo:    "ExampleName",
			wantFormatted: "Example Name",
		},
		{
			name:          "No dependencies",
			fullName:      "example_name(abc)",
			dependencies:  []string{},
			wantPseudo:    "example_name",
			wantFormatted: "example name",
		},
		{
			name:          "No parentheses",
			fullName:      "dependency.module.ExampleName",
			dependencies:  []string{"dependency", "module"},
			wantPseudo:    "ExampleName",
			wantFormatted: "Example Name",
		},
		{
			name:          "Complex dependencies",
			fullName:      "dependency.module.submodule.SomeValue(abc)",
			dependencies:  []string{"dependency", "module", "submodule"},
			wantPseudo:    "SomeValue",
			wantFormatted: "Some Value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPseudo, gotFormatted, _ := ProcessElementName(tt.fullName, tt.dependencies)
			if gotPseudo != tt.wantPseudo {
				t.Errorf("ProcessElementName() pseudocode = %v, want %v", gotPseudo, tt.wantPseudo)
			}
			if gotFormatted != tt.wantFormatted {
				t.Errorf("ProcessElementName() formatted = %v, want %v", gotFormatted, tt.wantFormatted)
			}
		})
	}
}
