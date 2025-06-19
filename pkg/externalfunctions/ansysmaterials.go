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
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/ansys/aali-sharedtypes/pkg/config"
	"github.com/ansys/aali-sharedtypes/pkg/logging"
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
		panic("Failed to serialize suggested criteria into json: " + err.Error())
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

		if !exists {
			panic("Could not find attribute with name " + lowerAttrName)
		}

		criteriaWithGuids = append(criteriaWithGuids, sharedtypes.MaterialCriterionWithGuid{
			AttributeName: criterion.AttributeName,
			AttributeGuid: guid,
			Explanation:   criterion.Explanation,
			Confidence:    criterion.Confidence,
		})
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
		} else {
			logging.Log.Warnf(&logging.ContextMap{}, "Filtered out non existing attribute")
			logging.Log.Debugf(&logging.ContextMap{}, "Attribute name: %s", suggestion.AttributeName)
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
	if criteriaText == "" {
		logging.Log.Debugf(&logging.ContextMap{}, "No valid JSON found in LLM response: %s", llmResponse)
		return nil
	}

	logging.Log.Debugf(&logging.ContextMap{}, "Attempting to parse JSON:\n%s", criteriaText)

	var criteria LlmCriteria
	err := json.Unmarshal([]byte(criteriaText), &criteria)
	if err != nil {
		logging.Log.Debugf(&logging.ContextMap{}, "Failed to deserialize criteria JSON from LLM response: %v; Raw JSON: %s", err, criteriaText)
		return nil
	}

	if len(criteria.Criteria) == 0 {
		logging.Log.Infof(&logging.ContextMap{}, "Deserialized JSON successfully but found 0 criteria.")
		logging.Log.Debugf(&logging.ContextMap{}, "%+v", criteria)
	} else {
		logging.Log.Debugf(&logging.ContextMap{}, "Successfully extracted %d criteria.", len(criteria.Criteria))
	}
	return criteria.Criteria
}

// PerformMultipleGeneralRequestsAndExtractAttributesWithOpenAiTokenOutput performs multiple general LLM requests
// using specific models, extracts structured attributes (criteria) from the responses, and returns the total token count
// using the specified OpenAI token counting model. This version does not stream responses.
//
// Tags:
//   - @displayName: Multiple General LLM Requests (Specific Models, No Stream, Attribute Extraction, OpenAI Token Output)
//
// Parameters:
//   - input: the user input string
//   - history: the conversation history for context
//   - systemPrompt: the system prompt to guide the LLM
//   - modelIds: the model IDs of the LLMs to query
//   - tokenCountModelName: the model name used for token count calculation
//   - n: number of parallel requests to perform
//
// Returns:
//   - uniqueCriterion: a deduplicated list of extracted attributes (criteria) from all responses
//   - tokenCount: the total token count (input tokens × n + combined output tokens)
func PerformMultipleGeneralRequestsAndExtractAttributesWithOpenAiTokenOutput(input string, history []sharedtypes.HistoricMessage, systemPrompt string, modelIds []string, tokenCountModelName string, n int) (uniqueCriterion []sharedtypes.MaterialLlmCriterion, tokenCount int) {
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Helper function to send a request and get the response as string
	sendRequest := func() string {
		responseChannel := sendChatRequest(input, "general", history, 0, systemPrompt, llmHandlerEndpoint, modelIds, nil, nil)
		defer close(responseChannel)

		var responseStr string
		for response := range responseChannel {
			if response.Type == "error" {
				panic(response.Error)
			}
			responseStr += *(response.ChatData)
			if *(response.IsLast) {
				break
			}
		}
		return responseStr
	}

	logging.Log.Debugf(&logging.ContextMap{}, "System prompt: %s", systemPrompt)

	// Collect all responses
	allResponses := runRequestsInParallel(n, sendRequest)

	var allCriteria []sharedtypes.MaterialLlmCriterion
	for _, response := range allResponses {
		criteria := ExtractCriteriaSuggestions(response)
		if criteria != nil {
			allCriteria = append(allCriteria, criteria...)
		}
	}

	// get input token count
	inputTokenCount := getTokenCount(tokenCountModelName, input)

	// get the output token count
	var combinedResponseText string
	for _, response := range allResponses {
		combinedResponseText += response
	}
	outputTokenCount := getTokenCount(tokenCountModelName, combinedResponseText)

	var totalTokenCount = inputTokenCount*n + outputTokenCount
	logging.Log.Debugf(&logging.ContextMap{}, "Total token count: %d", totalTokenCount)

	if len(allCriteria) == 0 {
		logging.Log.Warnf(&logging.ContextMap{}, "No valid criteria found in any response")
		return []sharedtypes.MaterialLlmCriterion{}, outputTokenCount
	}

	// Only return unique duplicates
	uniqueCriterion = FilterOutDuplicateAttributes(allCriteria)

	return uniqueCriterion, totalTokenCount
}

func runRequestsInParallel(n int, sendRequest func() string) []string {
	responseChan := make(chan string, n)
	var wg sync.WaitGroup

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					logging.Log.Errorf(&logging.ContextMap{}, "Recovered from panic in LLM request: %v", r)
				}
			}()
			response := sendRequest()
			responseChan <- response
		}()
	}

	go func() {
		wg.Wait()
		close(responseChan)
	}()

	var allResponses []string
	for response := range responseChan {
		logging.Log.Debugf(&logging.ContextMap{}, "Raw LLM response: %s", response)
		allResponses = append(allResponses, response)
	}
	return allResponses
}

func getTokenCount(modelName, text string) int {
	count, err := openAiTokenCount(modelName, text)
	if err != nil {
		errorMessage := fmt.Sprintf("Error getting output token count: %v", err)
		logging.Log.Errorf(&logging.ContextMap{}, "%v", errorMessage)
		panic(errorMessage)
	}
	return count
}

func ExtractJson(text string) (json string) {
	// Remove Markdown code block markers
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```JSON")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	re := regexp.MustCompile("{[\\s\\S]*}")
	matches := re.FindStringSubmatch(text)
	if len(matches) >= 1 {
		return strings.TrimSpace(matches[0])
	}

	logging.Log.Warnf(&logging.ContextMap{}, "No valid JSON found in response")
	logging.Log.Debugf(&logging.ContextMap{}, "%s", text)
	return ""
}
