package codegeneration

import (
	"reflect"
	"testing"
)

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

// TestCreateReturnList tests the CreateReturnList function.
func TestCreateReturnList(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
		hasError bool
	}{
		// Nested types classes
		{"list[list[class.subclass]]", []string{"class.subclass"}, false},
		{"tuple[tuple[class.subclass, int], float]", []string{"class.subclass"}, false},

		// Complex types (fully qualified class names)
		{"list[ansys.library.core.class.subclass.Parameter]",
			[]string{"ansys.library.core.class.subclass.Parameter"}, false},
		{"tuple[ansys.library.core.class.subclass.Parameter, int]",
			[]string{"ansys.library.core.class.subclass.Parameter"}, false},
		{"dict[str, ansys.library.core.class]",
			[]string{"ansys.library.core.class"}, false},

		// Pipe-separated union types
		{"class.subclass | otherclass.subclass",
			[]string{"class.subclass", "otherclass.subclass"}, false},
		{"int | str", nil, false},
		{"tuple[ansys.library.core.class | otherclass, int]",
			[]string{"ansys.library.core.class", "otherclass"}, false},
		{"list[class.subclass | otherclass.subclass]",
			[]string{"class.subclass", "otherclass.subclass"}, false},

		// Edge cases
		{"", nil, false},                              // Empty input (should return nil, no error)
		{"[]", nil, false},                            // Empty brackets (should return nil, no error)
		{"tuple[  ]", nil, false},                     // Empty tuple (should return nil, no error)
		{"randomtype", []string{"randomtype"}, false}, // Single random type

		// Cases with a single return type
		{"ansys.library.core.class.subclass.Parameter",
			[]string{"ansys.library.core.class.subclass.Parameter"}, false},
	}

	for _, test := range tests {
		result, err := CreateReturnList(test.input)

		if test.hasError {
			if err == nil {
				t.Errorf("Expected error for input %q but got none", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for input %q: %v", test.input, err)
			}
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("For input %q, expected %v but got %v", test.input, test.expected, result)
			}
		}
	}
}
