// Copyright (C) 2025 ANSYS, Inc. and/or its affiliates.
// SPDX-License-Identifier: MIT
//
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package externalfunctions

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/ansys/aali-sharedtypes/pkg/sharedtypes"
)

type Response struct {
	Criteria []sharedtypes.MaterialCriterionWithGuid
	Tokens   int
}

type LlmCriteria struct {
	Criteria []sharedtypes.MaterialLlmCriterion
}

// SerializeResponse formats the criteria to a response suitable for the UI clients in string format
//
// Tags:
//   - @displayName: Serialize response for clients
//
// Parameters:
//   - criteriaSuggestions: the list of criteria with their identities
//   - tokens: tokens consumed by the request
//
// Returns:
//   - result: string representation of the response in JSON format
func SerializeResponse(criteriaSuggestions []sharedtypes.MaterialCriterionWithGuid, tokens int) (result string) {
	response := Response{Criteria: criteriaSuggestions, Tokens: tokens}

	responseJson, err := json.Marshal(response)
	if err != nil {
		panic("failed to serialize suggested criteria into json: " + err.Error())
	}

	return string(responseJson)
}

// AddGuidsToAttributes adds GUIDs to the attributes in the criteria
//
// Tags:
//   - @displayName: Add GUIDs to criteria suggestions
//
// Parameters:
//   - criteriaSuggestions: the list of criteria without identities
//   - availableAttributes: the list of available attributes with their identities
//
// Returns:
//   - criteriaWithGuids: the list of criteria with their identities
func AddGuidsToAttributes(criteriaSuggestions []sharedtypes.MaterialLlmCriterion, availableAttributes []sharedtypes.MaterialAttribute) (criteriaWithGuids []sharedtypes.MaterialCriterionWithGuid) {
	attributeMap := make(map[string]string)
	for _, attr := range availableAttributes {
		attributeMap[strings.ToLower(attr.Name)] = attr.Guid
	}

	for _, criterion := range criteriaSuggestions {
		lowerAttrName := strings.ToLower(criterion.AttributeName)
		guid, exists := attributeMap[lowerAttrName]
		if exists {
			criteriaWithGuids = append(criteriaWithGuids, sharedtypes.MaterialCriterionWithGuid{
				AttributeName: criterion.AttributeName,
				AttributeGuid: guid,
				Explanation:   criterion.Explanation,
				Confidence:    criterion.Confidence,
			})
		}
	}

	return criteriaWithGuids
}

// FilterOutNonExistingAttributes filters out criteria suggestions that do not match any of the available attributes based on their names
//
// Tags:
//   - @displayName: Filter out non-existing attributes
//
// Parameters:
//   - criteriaSuggestions: current list of criteria suggestions
//   - availableAttributes: the list of available attributes
//
// Returns:
//   - filtered: the list of criteria suggestions excluding those that do not match any of the available attributes
func FilterOutNonExistingAttributes(criteriaSuggestions []sharedtypes.MaterialLlmCriterion, availableAttributes []sharedtypes.MaterialAttribute) (filtered []sharedtypes.MaterialLlmCriterion) {
	attributeMap := make(map[string]bool)
	for _, attr := range availableAttributes {
		attributeMap[strings.ToLower(attr.Name)] = true
	}

	var filteredCriteria []sharedtypes.MaterialLlmCriterion
	for _, suggestion := range criteriaSuggestions {
		if attributeMap[strings.ToLower(suggestion.AttributeName)] {
			filteredCriteria = append(filteredCriteria, suggestion)
		}
	}

	return filteredCriteria
}

// FilterOutDuplicateAttributes filters out duplicate attributes from the criteria suggestions based on their names
//
// Tags:
//   - @displayName: Filter out duplicate attributes
//
// Parameters:
//   - criteriaSuggestions: current list of criteria suggestions
//
// Returns:
//   - filtered: the list of criteria suggestions excluding duplicates based on attribute names
func FilterOutDuplicateAttributes(criteriaSuggestions []sharedtypes.MaterialLlmCriterion) (filtered []sharedtypes.MaterialLlmCriterion) {
	seen := make(map[string]bool)

	for _, suggestion := range criteriaSuggestions {
		lowerAttrName := strings.ToLower(suggestion.AttributeName)
		if !seen[lowerAttrName] {
			seen[lowerAttrName] = true
			filtered = append(filtered, suggestion)
		}
	}

	return filtered
}

// ExtractCriteriaSuggestions extracts criteria suggestions from the LLM response text
//
// Tags:
//   - @displayName: Extract criteria suggestions from LLM response
//
// Parameters:
//   - llmResponse: the text response from the LLM containing JSON with criteria suggestions
//
// Returns:
//   - criteriaSuggestions: the list of criteria suggestions extracted from the LLM response
func ExtractCriteriaSuggestions(llmResponse string) (criteriaSuggestions []sharedtypes.MaterialLlmCriterion) {
	criteriaText := ExtractJson(llmResponse)
	var criteria LlmCriteria
	err := json.Unmarshal([]byte(criteriaText), &criteria)
	if err != nil {
		panic("failed to deserialize criteria JSON from the LLM response: " + err.Error())
	}

	return criteria.Criteria
}

func ExtractJson(text string) (json string) {
	re := regexp.MustCompile("{[\\s\\S]*}")
	matches := re.FindStringSubmatch(text)
	if len(matches) >= 1 {
		return strings.TrimSpace(matches[0])
	}
	panic("no JSON found in the input text")
}
