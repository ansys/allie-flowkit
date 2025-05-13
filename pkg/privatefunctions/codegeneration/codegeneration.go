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

// CreateReturnList creates a list of return elements from a string
//
// Parameters:
//   - returnString: the string to create the list from
//
// Returns:
//   - the list of return elements
//   - an error if the string is empty
func CreateReturnList(returnString string) (returnElementList []string, err error) {
	if returnString == "" {
		return returnElementList, nil
	}

	// Regular expression to extract types inside brackets
	re := regexp.MustCompile(`([\w.]+|\[[\w.,\s]+\])`)

	// Clean and extract the types
	cleanedString := strings.ReplaceAll(returnString, " ", "")
	cleanedString = strings.Trim(cleanedString, "[]") // Remove outer brackets if any

	// Extract matches
	matches := re.FindAllString(cleanedString, -1)

	// Set subtypes to ignore
	ignoreTypes := []string{
		// Container types
		"list", "tuple", "dict", "set", "frozenset", "deque",
		"Union", "Optional", "List", "Dict", "Set", "FrozenSet", "Deque",

		// Built-in types
		"None", "str", "int", "float", "bool", "complex", "bytes", "bytearray",
		"memoryview", "range", "slice",

		// Function-related types
		"Callable", "Coroutine", "Generator", "AsyncGenerator", "Iterable", "Iterator",
		"AsyncIterable", "AsyncIterator",

		// Special typing types
		"Type", "Any", "Literal", "Final", "ClassVar", "NoReturn", "Self", "NewType",
	}

	for _, match := range matches {
		// Remove brackets from individual types
		match = strings.Trim(match, "[]")
		subTypes := strings.Split(match, ",")
		for _, subType := range subTypes {
			// Check if the type should be ignored
			ignore := false
			for _, ignoreType := range ignoreTypes {
				if subType == ignoreType {
					ignore = true
					break
				}
			}

			if ignore {
				continue
			}

			returnElementList = append(returnElementList, strings.TrimSpace(subType))
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
