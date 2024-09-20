package externalfunctions

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/ansys/allie-flowkit/pkg/internalstates"
	"github.com/ansys/allie-sharedtypes/pkg/config"
	"github.com/ansys/allie-sharedtypes/pkg/logging"
	"github.com/ansys/allie-sharedtypes/pkg/sharedtypes"
	"github.com/texttheater/golang-levenshtein/levenshtein"
)

// AnsysGPTCheckProhibitedWords checks the user query for prohibited words
//
// Tags:
//   - @displayName: Check Prohibited Words
//
// Parameters:
//   - query: the user query
//   - prohibitedWords: the list of prohibited words
//   - errorResponseMessage: the error response message
//
// Returns:
//   - foundProhibited: the flag indicating whether prohibited words were found
//   - responseMessage: the response message
func AnsysGPTCheckProhibitedWords(query string, prohibitedWords []string, errorResponseMessage string) (foundProhibited bool, responseMessage string) {
	// Check if all words in the value are present as whole words in the query
	queryLower := strings.ToLower(query)
	queryLower = strings.ReplaceAll(queryLower, ".", "")
	for _, prohibitedValue := range prohibitedWords {
		allWordsMatch := true
		for _, fieldWord := range strings.Fields(strings.ToLower(prohibitedValue)) {
			pattern := `\b` + regexp.QuoteMeta(fieldWord) + `\b`
			match, _ := regexp.MatchString(pattern, queryLower)
			if !match {
				allWordsMatch = false
				break
			}
		}
		if allWordsMatch {
			return true, errorResponseMessage
		}
	}

	// Check for prohibited words using fuzzy matching
	cutoff := 0.9
	for _, prohibitedValue := range prohibitedWords {
		wordMatchCount := 0
		for _, fieldWord := range strings.Fields(strings.ToLower(prohibitedValue)) {
			for _, word := range strings.Fields(queryLower) {
				distance := levenshtein.RatioForStrings([]rune(word), []rune(fieldWord), levenshtein.DefaultOptions)
				if distance >= cutoff {
					wordMatchCount++
					break
				}
			}
		}

		if wordMatchCount == len(strings.Fields(prohibitedValue)) {
			return true, errorResponseMessage
		}

		// If multiple words are present in the field , also check for the whole words without spaces
		if strings.Contains(prohibitedValue, " ") {
			for _, word := range strings.Fields(queryLower) {
				distance := levenshtein.RatioForStrings([]rune(word), []rune(prohibitedValue), levenshtein.DefaultOptions)
				if distance >= cutoff {
					return true, errorResponseMessage
				}
			}
		}
	}

	return false, ""
}

// AnsysGPTExtractFieldsFromQuery extracts the fields from the user query
//
// Tags:
//   - @displayName: Extract Fields
//
// Parameters:
//   - query: the user query
//   - fieldValues: the field values that the user query can contain
//   - defaultFields: the default fields that the user query can contain
//
// Returns:
//   - fields: the extracted fields
func AnsysGPTExtractFieldsFromQuery(query string, fieldValues map[string][]string, defaultFields []sharedtypes.AnsysGPTDefaultFields) (fields map[string]string) {
	// Initialize the fields map
	fields = make(map[string]string)

	// Check each field
	for field, values := range fieldValues {
		// Initializing the field with None
		fields[field] = ""

		// Sort the values by length in descending order
		sort.Slice(values, func(i, j int) bool {
			return len(values[i]) > len(values[j])
		})

		// Check if all words in the value are present as whole words in the query
		lowercaseQuery := strings.ToLower(query)
		for _, fieldValue := range values {
			allWordsMatch := true
			for _, fieldWord := range strings.Fields(strings.ToLower(fieldValue)) {
				pattern := `\b` + regexp.QuoteMeta(fieldWord) + `\b`
				match, _ := regexp.MatchString(pattern, lowercaseQuery)
				if !match {
					allWordsMatch = false
					break
				}
			}

			if allWordsMatch {
				fields[field] = fieldValue
				break
			}
		}

		// Split the query into words
		words := strings.Fields(lowercaseQuery)

		// If no exact match found, use fuzzy matching
		if fields[field] == "" {
			cutoff := 0.75
			for _, fieldValue := range values {
				for _, fieldWord := range strings.Fields(fieldValue) {
					for _, queryWord := range words {
						distance := levenshtein.RatioForStrings([]rune(fieldWord), []rune(queryWord), levenshtein.DefaultOptions)
						if distance >= cutoff {
							fields[field] = fieldValue
							break
						}
					}
				}
			}
		}
	}

	// If default value is found, use it
	for _, defaultField := range defaultFields {
		value, ok := fields[defaultField.FieldName]
		if ok && value == "" {
			if strings.Contains(strings.ToLower(query), strings.ToLower(defaultField.QueryWord)) {
				fields[defaultField.FieldName] = defaultField.FieldDefaultValue
			}
		}
	}

	return fields
}

// AnsysGPTPerformLLMRephraseRequestNew performs a rephrase request to LLM
//
// Tags:
//   - @displayName: Rephrase Request New
//
// Parameters:
//   - template: the template for the rephrase request
//   - query: the user query
//   - history: the conversation history
//
// Returns:
//   - rephrasedQuery: the rephrased query
func AnsysGPTPerformLLMRephraseRequestNew(template string, query string, history []sharedtypes.HistoricMessage) (rephrasedQuery string) {
	logging.Log.Debugf(internalstates.Ctx, "Performing LLM rephrase request")

	historyMessages := ""

	if len(history) >= 1 {
		historyMessages += history[len(history)-2].Content
	} else {
		return query
	}

	// Create map for the data to be used in the template
	dataMap := make(map[string]string)
	dataMap["query"] = query
	dataMap["chat_history"] = historyMessages

	// Format the template
	userTemplate := formatTemplate(template, dataMap)
	logging.Log.Debugf(internalstates.Ctx, "User template: %v", userTemplate)

	// create example
	exampleHistory := make([]sharedtypes.HistoricMessage, 2)
	exampleHistory[0] = sharedtypes.HistoricMessage{
		Role:    "user",
		Content: "'previous user query': 'How to create a beam with Ansys Mechanical?'\n'current user query': 'How to make the beam larger?'",
	}
	exampleHistory[1] = sharedtypes.HistoricMessage{
		Role:    "assistant",
		Content: "How to make a beam larger in Ansys Mechanical?",
	}

	// Perform the general request
	rephrasedQuery, _, err := performGeneralRequest(userTemplate, exampleHistory, false, "You are a query rephrasing assistant. You receive a 'previous user query' as well as a 'current user query' and rephrase the 'current user query' to include any relevant information from the 'previous user query'.")
	if err != nil {
		panic(err)
	}

	logging.Log.Debugf(internalstates.Ctx, "Rephrased query: %v", rephrasedQuery)

	return rephrasedQuery
}

// AnsysGPTPerformLLMRephraseRequest performs a rephrase request to LLM
//
// Tags:
//   - @displayName: Rephrase Request
//
// Parameters:
//   - template: the template for the rephrase request
//   - query: the user query
//   - history: the conversation history
//
// Returns:
//   - rephrasedQuery: the rephrased query
func AnsysGPTPerformLLMRephraseRequest(userTemplate string, query string, history []sharedtypes.HistoricMessage, systemPrompt string) (rephrasedQuery string) {
	logging.Log.Debugf(internalstates.Ctx, "Performing LLM rephrase request")

	historyMessages := ""

	if len(history) >= 1 {
		historyMessages += "user:" + history[len(history)-2].Content + "\n"
	} else {
		return query
	}

	// Create map for the data to be used in the template
	dataMap := make(map[string]string)
	dataMap["query"] = query
	dataMap["chat_history"] = historyMessages

	// Format the template
	userTemplate = formatTemplate(userTemplate, dataMap)
	logging.Log.Debugf(internalstates.Ctx, "User template for repharasing query: %v", userTemplate)

	// Perform the general request
	rephrasedQuery, _, err := performGeneralRequest(userTemplate, nil, false, systemPrompt)
	if err != nil {
		panic(err)
	}

	logging.Log.Debugf(internalstates.Ctx, "Rephrased query: %v", rephrasedQuery)

	return rephrasedQuery
}

// AnsysGPTBuildFinalQuery builds the final query for Ansys GPT
//
// Tags:
//   - @displayName: Build Final Query
//
// Parameters:
//   - refrasedQuery: the refrased query
//   - context: the context
//
// Returns:
//   - finalQuery: the final query
func AnsysGPTBuildFinalQuery(refrasedQuery string, context []sharedtypes.ACSSearchResponse) (finalQuery string, errorResponse string, displayFixedMessageToUser bool) {
	logging.Log.Debugf(internalstates.Ctx, "Building final query for Ansys GPT with context of length: %v", len(context))

	// check if there is no context
	if len(context) == 0 {
		errorResponse = "Sorry, I could not find any knowledge from Ansys that can answer your question. Please try and revise your query by asking in a different way or adding more details."
		return "", errorResponse, true
	}

	// Build the final query using the KnowledgeDB response and the original request
	finalQuery = "Based on the following examples:\n\n--- INFO START ---\n"
	for _, example := range context {
		finalQuery += fmt.Sprintf("%v", example) + "\n"
	}
	finalQuery += "--- INFO END ---\n\n" + refrasedQuery + "\n"

	return finalQuery, "", false
}

// AnsysGPTPerformLLMRequest performs a request to Ansys GPT
//
// Tags:
//   - @displayName: LLM Request
//
// Parameters:
//   - finalQuery: the final query
//   - history: the conversation history
//   - systemPrompt: the system prompt
//
// Returns:
//   - stream: the stream channel
func AnsysGPTPerformLLMRequest(finalQuery string, history []sharedtypes.HistoricMessage, systemPrompt string, isStream bool) (message string, stream *chan string) {
	// get the LLM handler endpoint
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send chat request
	responseChannel := sendChatRequest(finalQuery, "general", history, 0, systemPrompt, llmHandlerEndpoint, nil)

	// If isStream is true, create a stream channel and return asap
	if isStream {
		// Create a stream channel
		streamChannel := make(chan string, 400)

		// Start a goroutine to transfer the data from the response channel to the stream channel
		go transferDatafromResponseToStreamChannel(&responseChannel, &streamChannel, false)

		// Return the stream channel
		return "", &streamChannel
	}

	// else Process all responses
	var responseAsStr string
	for response := range responseChannel {
		// Check if the response is an error
		if response.Type == "error" {
			panic(response.Error)
		}

		// Accumulate the responses
		responseAsStr += *(response.ChatData)

		// If we are at the last message, break the loop
		if *(response.IsLast) {
			break
		}
	}

	// Close the response channel
	close(responseChannel)

	// Return the response
	return responseAsStr, nil
}

// AnsysGPTReturnIndexList returns the index list for Ansys GPT
//
// Tags:
//   - @displayName: List Indexes
//
// Parameters:
//   - indexGroups: the index groups
//
// Returns:
//   - indexList: the index list
func AnsysGPTReturnIndexList(indexGroups []string) (indexList []string) {
	indexList = make([]string, 0)
	// iterate through indexGroups and append to indexList
	for _, indexGroup := range indexGroups {
		switch indexGroup {
		case "Ansys Learning":
			indexList = append(indexList, "granular-ansysgpt")
			indexList = append(indexList, "ansysgpt-alh")
		case "Ansys Products":
			indexList = append(indexList, "lsdyna-documentation-r14")
			indexList = append(indexList, "ansysgpt-documentation-2023r2")
			indexList = append(indexList, "scade-documentation-2023r2")
			indexList = append(indexList, "ansys-dot-com-marketing")
			indexList = append(indexList, "external-crtech-thermal-desktop")
			// indexList = append(indexList, "ibp-app-brief")
			// indexList = append(indexList, "pyansys_help_documentation")
			// indexList = append(indexList, "pyansys-examples")
		case "Ansys Semiconductor":
			indexList = append(indexList, "ansysgpt-scbu")
		default:
			logging.Log.Errorf(internalstates.Ctx, "Invalid indexGroup: %v\n", indexGroup)
			return
		}
	}

	return indexList
}

// AnsysGPTACSSemanticHybridSearchs performs a semantic hybrid search in ACS
//
// Tags:
//   - @displayName: ACS Semantic Hybrid Search
//
// Parameters:
//   - query: the query string
//   - embeddedQuery: the embedded query
//   - indexList: the index list
//   - typeOfAsset: the type of asset
//   - physics: the physics
//   - product: the product
//   - productMain: the main product
//   - filter: the filter
//   - filterAfterVectorSearch: the flag to define the filter order
//   - returnedProperties: the properties to be returned
//   - topK: the number of results to be returned from vector search
//   - searchedEmbeddedFields: the ACS fields to be searched
//
// Returns:
//   - output: the search results
func AnsysGPTACSSemanticHybridSearchs(
	query string,
	embeddedQuery []float32,
	indexList []string,
	filter map[string]string,
	topK int) (output []sharedtypes.ACSSearchResponse) {

	output = make([]sharedtypes.ACSSearchResponse, 0)
	for _, indexName := range indexList {
		partOutput := ansysGPTACSSemanticHybridSearch(query, embeddedQuery, indexName, filter, topK)
		output = append(output, partOutput...)
	}

	return output
}

// AnsysGPTRemoveNoneCitationsFromSearchResponse removes none citations from search response
//
// Tags:
//   - @displayName: Remove None Citations
//
// Parameters:
//   - semanticSearchOutput: the search response
//   - citations: the citations
//
// Returns:
//   - reducedSemanticSearchOutput: the reduced search response
func AnsysGPTRemoveNoneCitationsFromSearchResponse(semanticSearchOutput []sharedtypes.ACSSearchResponse, citations []sharedtypes.AnsysGPTCitation) (reducedSemanticSearchOutput []sharedtypes.ACSSearchResponse) {
	// iterate throught search response and keep matches to citations
	reducedSemanticSearchOutput = make([]sharedtypes.ACSSearchResponse, len(citations))
	for _, value := range semanticSearchOutput {
		for _, citation := range citations {
			if value.SourceURLLvl2 == citation.Title {
				reducedSemanticSearchOutput = append(reducedSemanticSearchOutput, value)
			} else if value.SourceURLLvl2 == citation.URL {
				reducedSemanticSearchOutput = append(reducedSemanticSearchOutput, value)
			} else if value.SearchRerankerScore == citation.Relevance {
				reducedSemanticSearchOutput = append(reducedSemanticSearchOutput, value)
			}
		}
	}

	return reducedSemanticSearchOutput
}

// AnsysGPTReorderSearchResponseAndReturnOnlyTopK reorders the search response
//
// Tags:
//   - @displayName: Reorder Search Response
//
// Parameters:
//   - semanticSearchOutput: the search response
//   - topK: the number of results to be returned
//
// Returns:
//   - reorderedSemanticSearchOutput: the reordered search response
func AnsysGPTReorderSearchResponseAndReturnOnlyTopK(semanticSearchOutput []sharedtypes.ACSSearchResponse, topK int) (reorderedSemanticSearchOutput []sharedtypes.ACSSearchResponse) {
	logging.Log.Debugf(internalstates.Ctx, "Reordering search response of length %v based on reranker_score and returning only top %v results", len(semanticSearchOutput), topK)
	// Sorting by Weight * SearchRerankerScore in descending order
	sort.Slice(semanticSearchOutput, func(i, j int) bool {
		return semanticSearchOutput[i].Weight*semanticSearchOutput[i].SearchRerankerScore > semanticSearchOutput[j].Weight*semanticSearchOutput[j].SearchRerankerScore
	})

	// Return only topK results
	if len(semanticSearchOutput) > topK {
		semanticSearchOutput = semanticSearchOutput[:topK]
	}

	return semanticSearchOutput
}

// AnsysGPTGetSystemPrompt returns the system prompt for Ansys GPT
//
// Tags:
//   - @displayName: Get System Prompt
//
// Parameters:
//   - rephrasedQuery: the rephrased query
//
// Returns:
//   - systemPrompt: the system prompt
func AnsysGPTGetSystemPrompt(query string, prohibitedWords []string, template string) (systemPrompt string) {

	// create string from prohibited words
	prohibitedWordsString := ""
	for _, word := range prohibitedWords {
		prohibitedWordsString += word + ", "
	}

	// Create map for the data to be used in the template
	dataMap := make(map[string]string)
	dataMap["query"] = query
	dataMap["prohibited_words"] = prohibitedWordsString

	// Format the template
	systemTemplate := formatTemplate(template, dataMap)
	logging.Log.Debugf(internalstates.Ctx, "System prompt for final query: %v", systemTemplate)

	// return system prompt
	return systemTemplate
}
