package functiondefinitions

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"regexp"
	"strings"

	"github.com/ansys/allie-flowkit/pkg/internalstates"
	"github.com/ansys/allie-sharedtypes/pkg/allieflowkitgrpc"
)

// ExtractFunctionDefinitionsFromPackage parses the given file for public functions and populates internalstates.AvailableFunctions.
// The function definitions are stored in the internalstates.AvailableFunctions map.
// The key is the function name and the value is a FunctionDefinition struct.
// The FunctionDefinition struct contains the function's name, description, package, inputs, and outputs.
// The inputs and outputs are stored as FunctionInput and FunctionOutput structs, respectively.
// The FunctionInput and FunctionOutput structs contain the name, type, and GoType of the input/output.
// The GoType is the Go type of the input/output, while the Type is a simplified type string (e.g., "string", "number", "boolean", "json").
//
// The function returns an error if the file cannot be parsed.
//
// Parameters:
//   - packagePath: the path to the package file to parse.
//
// Returns:
//   - error: an error if the file cannot be parsed.
func ExtractFunctionDefinitionsFromPackage(content string, category string) error {
	fset := token.NewFileSet() // positions are relative to fset

	// Parse the file given by filePath
	node, err := parser.ParseFile(fset, "", content, parser.ParseComments)
	if err != nil {
		return err
	}

	enumerable_definitions := make(map[string][]string)

	// Iterate over all declarations in the file to extract type and const definitions
	for _, decl := range node.Decls {

		// get general declarations
		if gd, isGd := decl.(*ast.GenDecl); isGd {

			// differentiating between type and const declarations
			switch gd.Tok {

			case token.TYPE:
				// type declaration - only consider string type
				spec := gd.Specs[0]
				type_spec, is_type_spec := spec.(*ast.TypeSpec)
				if is_type_spec {
					ident, isIdent := type_spec.Type.(*ast.Ident)
					if isIdent {
						// check whether name is string -> exclude all other types including structs
						if ident.Name == "string" {
							// add to enumerable_definition
							enumerable_definitions[type_spec.Name.Name] = []string{}
						}
					}
				}

			case token.CONST:
				// const declaration (enumerable entries)

				// extract consts
				for _, spec := range gd.Specs {

					// should be value spec
					value_spec, is_value_spec := spec.(*ast.ValueSpec)
					if is_value_spec {
						// check whether enumerable is defined
						_, is_enumerable_defined := enumerable_definitions[value_spec.Type.(*ast.Ident).Name]
						if is_enumerable_defined {
							// add to enumerable_definition
							enumerable_definitions[value_spec.Type.(*ast.Ident).Name] = append(enumerable_definitions[value_spec.Type.(*ast.Ident).Name], value_spec.Names[0].Name)
						}
					}

				}

			}

		}
	}

	// Iterate over all declarations in the file
	for _, decl := range node.Decls {
		// Filter function declarations
		if fn, isFn := decl.(*ast.FuncDecl); isFn {
			// Check if the function is exported
			if fn.Name.IsExported() {
				funcDef := &allieflowkitgrpc.FunctionDefinition{
					Name:        fn.Name.Name,
					Description: fn.Doc.Text(),
					Category:    category,
					Input:       []*allieflowkitgrpc.FunctionInputDefinition{},
					Output:      []*allieflowkitgrpc.FunctionOutputDefinition{},
				}

				// Handle inputs (parameters)
				if fn.Type.Params != nil {
					for _, param := range fn.Type.Params.List {
						if len(param.Names) == 0 {
							funcDef.Input = append(funcDef.Input, &allieflowkitgrpc.FunctionInputDefinition{
								Name:   typeExprToString(param.Type),
								Type:   typeExprToSimpleType(param.Type),
								GoType: typeExprToString(param.Type),
							})
						} else {
							for _, paramName := range param.Names {
								// skip if endpoint
								if paramName.Name == "llmHandlerEndpoint" || paramName.Name == "knowledgeDbEndpoint" {
									continue
								}

								// check if enumerable
								goType := typeExprToString(param.Type)
								options := []string{}
								enumerable, is_enumerable := enumerable_definitions[goType]
								if is_enumerable {
									options = enumerable
									goType = "string"
								}

								funcDef.Input = append(funcDef.Input, &allieflowkitgrpc.FunctionInputDefinition{
									Name:    paramName.Name,
									Type:    typeExprToSimpleType(param.Type),
									GoType:  goType,
									Options: options,
								})
							}
						}
					}
				}

				// Handle outputs (results)
				if fn.Type.Results != nil {
					for _, result := range fn.Type.Results.List {
						if len(result.Names) == 0 {
							goType := typeExprToString(result.Type)
							funcDef.Output = append(funcDef.Output, &allieflowkitgrpc.FunctionOutputDefinition{
								Name:   goType,
								Type:   typeExprToSimpleType(result.Type),
								GoType: goType,
							})
						} else {
							for _, resultName := range result.Names {
								goType := typeExprToString(result.Type)
								funcDef.Output = append(funcDef.Output, &allieflowkitgrpc.FunctionOutputDefinition{
									Name:   resultName.Name,
									Type:   typeExprToSimpleType(result.Type),
									GoType: goType,
								})
							}
						}
					}
				}

				// Store the function definition
				internalstates.AvailableFunctions[funcDef.Name] = funcDef
			}
		}
	}
	return nil
}

// typeExprToSimpleType translates an ast.Expr (which represents a type in Go's AST) into a simple type string,
// treating user-defined types and any complex structures as "json".
// The simple types are "string", "number", "boolean", and "json".
// This function is used to simplify the type representation for function inputs and outputs.
//
// The function returns a simple type string.
//
// Parameters:
//   - expr: the ast.Expr representing the type.
//
// Returns:
//   - string: the simple type string.
func typeExprToSimpleType(expr ast.Expr) string {
	switch typ := expr.(type) {
	case *ast.Ident: // Ident covers basic types and user-defined types.
		name := typ.Name
		// Check for basic types
		if name == "string" {
			return "string"
		} else if strings.HasPrefix(name, "int") || strings.HasPrefix(name, "uint") || name == "float32" || name == "float64" {
			return "number"
		} else if name == "bool" {
			return "boolean"
		} else {
			// Assume any other Ident not matching the basic types is a user-defined type (struct, etc.), which we categorize as "json".
			return "json"
		}
	case *ast.ArrayType, *ast.SliceExpr:
		return "json" // Arrays and slices are treated as JSON.
	case *ast.StructType:
		return "json" // Explicitly defined structs in the AST are treated as JSON.
	case *ast.MapType:
		return "json" // Maps are key-value pairs, treated as JSON.
	case *ast.InterfaceType:
		return "json" // Interface{} can hold any type, treated as JSON.
	case *ast.StarExpr: // Pointer type
		// Recursively call goTypeToSimpleType on the base type of the pointer.
		return typeExprToSimpleType(typ.X)
	}

	return "json"
}

// typeExprToString converts an ast.Expr that represents a type into a string representation.
// This function aims to produce more readable type strings for complex types.
//
// The function returns a string representation of the type.
//
// Parameters:
//   - expr: the ast.Expr representing the type.
//
// Returns:
//   - string: the string representation of the type.
func typeExprToString(expr ast.Expr) string {
	// Use a buffer to write the type's string representation
	var buf bytes.Buffer
	// Print the type expression into the buffer
	err := format.Node(&buf, token.NewFileSet(), expr)
	if err != nil {
		// In case of an error (which should be rare), fallback to a simple representation.
		return "unknown"
	}
	// Get the raw string representation of the type
	typeStr := buf.String()

	// Remove package prefixes using regular expressions
	// This regex matches package paths followed by a dot (e.g., "sharedtypes.")
	packagePattern := regexp.MustCompile(`\b[a-zA-Z_]\w*\.`)
	cleanedTypeStr := packagePattern.ReplaceAllString(typeStr, "")

	return cleanedTypeStr
}
