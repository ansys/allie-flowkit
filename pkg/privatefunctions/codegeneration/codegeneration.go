package codegeneration

import (
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
