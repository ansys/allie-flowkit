package codegeneration

import (
	"fmt"
	"regexp"
	"strings"
)

// RemoveEmptyLines removes empty lines from a string
//
// Parameters:
//   - input: the string to remove empty lines from
//
// Returns:
//   - the string with empty lines removed
func RemoveEmptyLines(input string) string {
	// Split the string into lines
	lines := strings.Split(input, "\n")
	// Filter out empty lines
	var nonEmptyLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" { // Keep only non-empty lines
			nonEmptyLines = append(nonEmptyLines, trimmed)
		}
	}
	// Join the non-empty lines back together
	return strings.Join(nonEmptyLines, "\n")
}

// SplitByCapitalLetters splits a string by capital letters
//
// Parameters:
//   - s: the string to split
//
// Returns:
//   - the string with spaces inserted before capital letters
func SplitByCapitalLetters(s string) string {
	re := regexp.MustCompile(`[A-Z]+[a-z]*|[A-Z]+`)
	words := re.FindAllString(s, -1)
	return strings.Join(words, " ")
}

// CreateReturnListMechanical creates a list of return elements from a string
//
// Parameters:
//   - returnString: the string to create the list from
//
// Returns:
//   - the list of return elements
//   - an error if the string is empty
func CreateReturnListMechanical(returnString string) (returnElementList []string, err error) {
	// Split the string by comma
	returnList := strings.Split(returnString, " ")
	// Trim the spaces from the strings
	for _, returnElement := range returnList {
		// Check if the string contains 'Ansys.'
		if strings.Contains(returnElement, "Ansys.") {
			// Start the string when the 'Ansys.' appears and remove the last ']' if it exists
			returnElement = returnElement[strings.Index(returnElement, "Ansys."):]
			returnElement = strings.TrimSuffix(returnElement, "]")
			// Add the element to the list
			returnElementList = append(returnElementList, returnElement)
		}
	}
	return returnElementList, nil
}

// ProcessElementName processes an element name
//
// Parameters:
//   - fullName: the full name of the element
//   - dependencies: the dependencies of the element
//
// Returns:
//   - the pseudocode name of the element
//   - the formatted name of the element
//   - an error if the name is empty
func ProcessElementName(fullName string, dependencies []string) (namePseudocode string, nameFormatted string, err error) {
	// If the name is empty, return empty error.
	if fullName == "" {
		return "", "", fmt.Errorf("empty name")
	}

	// Remove the prefix from the name.
	if len(dependencies) > 0 {
		prefixNamePseudocode := strings.Join(dependencies, ".") + "."
		namePseudocode = fullName[len(prefixNamePseudocode):]
	} else {
		namePseudocode = fullName
	}
	namePseudocode = strings.Split(namePseudocode, "(")[0]

	// If string contains underscores (snake_case), replace them with spaces.
	if strings.Contains(namePseudocode, "_") {
		nameFormatted = strings.ReplaceAll(namePseudocode, "_", " ")
	} else {
		// Add space before capital letters.
		nameFormatted = SplitByCapitalLetters(namePseudocode)
		if nameFormatted == "" {
			nameFormatted = namePseudocode
		}
	}
	return namePseudocode, nameFormatted, nil
}

// StringToCodeGenerationType converts a string to a CodeGenerationType
// enum value.
//
// Parameters:
//   - nodeTypeString: the string to convert
//
// Returns:
//   - the CodeGenerationType value
//   - an error if the string is invalid
func StringToCodeGenerationType(nodeTypeString string) (CodeGenerationType, error) {
	switch nodeTypeString {
	case "Function":
		return Function, nil
	case "Class":
		return Class, nil
	case "Method":
		return Method, nil
	case "Enum":
		return Enum, nil
	case "Module":
		return Module, nil
	default:
		return "", fmt.Errorf("invalid node type: %s", nodeTypeString)
	}
}
