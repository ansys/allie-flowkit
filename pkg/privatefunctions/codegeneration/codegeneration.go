package codegeneration

import (
	"fmt"
	"regexp"
	"strings"
)

var MechanicalInstancesReplaceDict = map[string]string{
	"Model.Mesh":                      "Ansys.ACT.Automation.Mechanical.MeshControls.Mesh",
	"Model.CoordinateSystems":         "Ansys.ACT.Automation.Mechanical.CoordinateSystems",
	"Model.Analyses.AnalysisSettings": "Ansys.ACT.Automation.Mechanical.AnalysisSettings.ANSYSAnalysisSettings",
	"Model.Analyses.Solution":         "Ansys.ACT.Automation.Mechanical.Solution",
	"Model.Analyses":                  "Ansys.ACT.Automation.Mechanical.Analysis",
	"DataModel.Project.Model":         "Ansys.ACT.Automation.Mechanical.Model",
	"Model":                           "Ansys.ACT.Automation.Mechanical.Model",
	"ExtAPI.DataModel":                "Ansys.ACT.Automation.Mechanical",
	"ExtAPI.Application":              "Ansys.ACT.Interfaces.Mechanical.IMechanicalApplication",
}

var ReplacementPriorityList = []string{
	"Model.Mesh",
	"Model.CoordinateSystems",
	"Model.Analyses.AnalysisSettings",
	"Model.Analyses.Solution",
	"Model.Analyses",
	"DataModel.Project.Model",
	"Model",
	"ExtAPI.DataModel",
	"ExtAPI.Application",
}

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

func SplitByCapitalLetters(s string) string {
	re := regexp.MustCompile(`[A-Z]+[a-z]*|[A-Z]+`)
	words := re.FindAllString(s, -1)
	return strings.Join(words, " ")
}

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
