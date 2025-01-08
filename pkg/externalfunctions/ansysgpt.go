package externalfunctions

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"

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
			cutoff := 0.76
			for _, fieldValue := range values {
				for _, queryWord := range words {
					distance := levenshtein.RatioForStrings([]rune(fieldValue), []rune(queryWord), levenshtein.DefaultOptions)
					if distance >= cutoff {
						fields[field] = fieldValue
						break
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
	logging.Log.Debugf(&logging.ContextMap{}, "Performing LLM rephrase request")

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
	logging.Log.Debugf(&logging.ContextMap{}, "User template: %v", userTemplate)

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
	rephrasedQuery, _, err := performGeneralRequest(userTemplate, exampleHistory, false, "You are a query rephrasing assistant. You receive a 'previous user query' as well as a 'current user query' and rephrase the 'current user query' to include any relevant information from the 'previous user query'.", nil)
	if err != nil {
		panic(err)
	}

	logging.Log.Debugf(&logging.ContextMap{}, "Rephrased query: %v", rephrasedQuery)

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
	logging.Log.Debugf(&logging.ContextMap{}, "Performing LLM rephrase request")

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
	logging.Log.Debugf(&logging.ContextMap{}, "User template for repharasing query: %v", userTemplate)

	// Perform the general request
	rephrasedQuery, _, err := performGeneralRequest(userTemplate, nil, false, systemPrompt, nil)
	if err != nil {
		panic(err)
	}

	logging.Log.Debugf(&logging.ContextMap{}, "Rephrased query: %v", rephrasedQuery)

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
	logging.Log.Debugf(&logging.ContextMap{}, "Building final query for Ansys GPT with context of length: %v", len(context))

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
	responseChannel := sendChatRequest(finalQuery, "general", history, 0, systemPrompt, llmHandlerEndpoint, nil, nil, nil)

	// If isStream is true, create a stream channel and return asap
	if isStream {
		// Create a stream channel
		streamChannel := make(chan string, 400)

		// Start a goroutine to transfer the data from the response channel to the stream channel
		go transferDatafromResponseToStreamChannel(&responseChannel, &streamChannel, false, false, "", 0, 0, "", "", false, "")

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
			// indexList = append(indexList, "ansysgpt-alh")
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
			logging.Log.Errorf(&logging.ContextMap{}, "Invalid indexGroup: %v\n", indexGroup)
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
	acsEndpoint string,
	acsApiKey string,
	acsApiVersion string,
	query string,
	embeddedQuery []float32,
	indexList []string,
	filter map[string]string,
	topK int) (output []sharedtypes.ACSSearchResponse) {

	output = make([]sharedtypes.ACSSearchResponse, 0)
	for _, indexName := range indexList {
		partOutput, err := ansysGPTACSSemanticHybridSearch(acsEndpoint, acsApiKey, acsApiVersion, query, embeddedQuery, indexName, filter, topK, false, nil)
		if err != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Error in semantic hybrid search: %v", err)
			panic(err)
		}
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
	logging.Log.Debugf(&logging.ContextMap{}, "Reordering search response of length %v based on reranker_score and returning only top %v results", len(semanticSearchOutput), topK)
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
	logging.Log.Debugf(&logging.ContextMap{}, "System prompt for final query: %v", systemTemplate)

	// return system prompt
	return systemTemplate
}

// AisPerformLLMRephraseRequest performs a rephrase request to LLM
//
// Tags:
//   - @displayName: AIS Rephrase Request
//
// Parameters:
//   - systemTemplate: the system template for the rephrase request
//   - userTemplate: the user template for the rephrase request
//   - query: the user query
//   - history: the conversation history
//
// Returns:
//   - rephrasedQuery: the rephrased query
func AisPerformLLMRephraseRequest(systemTemplate string, userTemplate string, query string, history []sharedtypes.HistoricMessage, tokenCountModelName string) (rephrasedQuery string, inputTokenCount int, outputTokenCount int) {
	logging.Log.Debugf(&logging.ContextMap{}, "Performing LLM rephrase request")

	// create "chat_history" string
	historyMessages := ""
	for _, message := range history {
		switch message.Role {
		case "user":
			historyMessages += "`HumanMessage`: `" + message.Content + "`\n"
		case "assistant":
			historyMessages += "`AIMessage`: `" + message.Content + "`\n"
		}
	}

	// Create map for the data to be used in the template
	dataMap := make(map[string]string)
	dataMap["query"] = query
	dataMap["chat_history"] = historyMessages

	// Format the user and system template
	userPrompt := formatTemplate(userTemplate, dataMap)
	logging.Log.Debugf(&logging.ContextMap{}, "User template for repharasing query: %v", userTemplate)
	systemPrompt := formatTemplate(systemTemplate, dataMap)
	logging.Log.Debugf(&logging.ContextMap{}, "System template for repharasing query: %v", systemTemplate)

	// create options
	var maxTokens int32 = 500
	var temperature float32 = 0.0
	options := &sharedtypes.ModelOptions{
		MaxTokens:   &maxTokens,
		Temperature: &temperature,
	}

	// Perform the general request
	rephrasedQuery, _, err := performGeneralRequest(userPrompt, nil, false, systemPrompt, options)
	if err != nil {
		panic(err)
	}

	// calculate input and output token count
	inputTokenCount, err = openAiTokenCount(tokenCountModelName, userPrompt+systemPrompt)
	if err != nil {
		panic(err)
	}
	outputTokenCount, err = openAiTokenCount(tokenCountModelName, rephrasedQuery)
	if err != nil {
		panic(err)
	}

	logging.Log.Debugf(&logging.ContextMap{}, "Rephrased query: %v", rephrasedQuery)

	return rephrasedQuery, inputTokenCount, outputTokenCount
}

// AisReturnIndexList returns the index list for AIS
//
// Tags:
//   - @displayName: Get AIS Index List
//
// Parameters:
//   - accessPoint: the access point
//
// Returns:
//   - indexList: the index list
func AisReturnIndexList(accessPoint string, physics []string) (indexList []string) {
	indexList = []string{}

	switch accessPoint {
	case "ansysgpt-general", "ais-embedded":
		if len(physics) == 1 && physics[0] == "scade" {
			// special case for Scade One
			indexList = append(indexList, "external-product-documentation-public")
		} else {
			// default ais case
			indexList = append(indexList,
				"granular-ansysgpt",
				"ansysgpt-documentation-2023r2",
				"lsdyna-documentation-r14",
				"scade-documentation-2023r2",
				"external-marketing",
				"external-product-documentation-public",
				"external-learning-hub",
				"external-crtech-thermal-desktop",
				"external-release-notes",
				"external-zemax-websites",
			)
		}
	case "ansysgpt-scbu":
		indexList = append(indexList,
			"ansysgpt-scbu",
			"external-scbu-learning-hub",
			"scbu-data-except-alh",
		)
	default:
		logging.Log.Errorf(&logging.ContextMap{}, "Invalid accessPoint: %v\n", accessPoint)
		return
	}

	return indexList
}

// AisAcsSemanticHybridSearchs performs a semantic hybrid search in ACS
//
// Tags:
//   - @displayName: AIS ACS Semantic Hybrid Search
//
// Parameters:
//   - query: the query string
//   - embeddedQuery: the embedded query
//   - indexList: the index list
//   - physics: the physics
//   - topK: the number of results to be returned
//
// Returns:
//   - output: the search results
func AisAcsSemanticHybridSearchs(
	acsEndpoint string,
	acsApiKey string,
	acsApiVersion string,
	query string,
	embeddedQuery []float32,
	indexList []string,
	physics []string,
	topK int) []sharedtypes.ACSSearchResponse {

	// Create a channel to collect results
	resultChan := make(chan []sharedtypes.ACSSearchResponse, len(indexList))

	// Create a WaitGroup to ensure all goroutines complete
	var wg sync.WaitGroup
	wg.Add(len(indexList))

	// Launch a goroutine for each index
	for _, indexName := range indexList {
		go func(idx string) {
			defer func() {
				r := recover()
				if r != nil {
					logging.Log.Errorf(&logging.ContextMap{}, "panic in paralell processing of ACS requests: %v", r)
				}
			}()
			defer wg.Done()
			// Run the search for this index
			result, err := ansysGPTACSSemanticHybridSearch(acsEndpoint, acsApiKey, acsApiVersion, query, embeddedQuery, idx, nil, topK, true, physics)
			if err != nil {
				logging.Log.Errorf(&logging.ContextMap{}, "Error in semantic hybrid search: %v", err)
				return
			}
			resultChan <- result
		}(indexName)
	}

	// Launch a goroutine to close the channel once all searches are complete
	go func() {
		defer func() {
			r := recover()
			if r != nil {
				logging.Log.Errorf(&logging.ContextMap{}, "panic in closing ACS result channel: %v", r)
			}
		}()
		wg.Wait()
		close(resultChan)
	}()

	// Collect all results
	var output []sharedtypes.ACSSearchResponse
	for results := range resultChan {
		output = append(output, results...)
	}

	return output
}

// AisChangeAcsResponsesByFactor changes the ACS responses by a factor
//
// Tags:
//   - @displayName: Change ACS Responses By Factor
//
// Parameters:
//   - factors: the factors
//   - semanticSearchOutput: the search response
//
// Returns:
//   - changedSemanticSearchOutput: the changed search response
func AisChangeAcsResponsesByFactor(factors map[string]float64, semanticSearchOutput []sharedtypes.ACSSearchResponse) (changedSemanticSearchOutput []sharedtypes.ACSSearchResponse) {

	// Iterate through the search response and change the 'weight' and '@search.reranker_score' based on the factors
	for i, value := range semanticSearchOutput {
		// Check if the document's 'typeOFasset' exists in the factors map
		factor, exists := factors[value.TypeOFasset]
		if exists {
			// Update 'weight'
			value.Weight = value.Weight * factor

			// Update '@search.reranker_score' if it is set, otherwise set it to 'weight'
			if value.SearchRerankerScore != 0 {
				value.SearchRerankerScore = value.SearchRerankerScore * factor
			} else {
				value.SearchRerankerScore = value.Weight
			}

			// assign the changed value to the output
			semanticSearchOutput[i] = value
		}
	}

	return semanticSearchOutput
}

// AisPerformLLMFinalRequest performs a final request to LLM
//
// Tags:
//   - @displayName: AIS Final Request
//
// Parameters:
//   - systemTemplate: the system template for the final request
//   - userTemplate: the user template for the final request
//   - query: the user query
//   - history: the conversation history
//   - prohibitedWords: the list of prohibited words
//   - errorList1: the list of error words
//   - errorList2: the list of error words
//
// Returns:
//   - stream: the stream channel
func AisPerformLLMFinalRequest(systemTemplate string,
	userTemplate string,
	query string,
	history []sharedtypes.HistoricMessage,
	context []sharedtypes.ACSSearchResponse,
	prohibitedWords []string,
	errorList1 []string,
	errorList2 []string,
	tokenCountEndpoint string,
	previousInputTokenCount int,
	previousOutputTokenCount int,
	tokenCountModelName string,
	isStream bool,
	userEmail string) (message string, stream *chan string) {

	logging.Log.Debugf(&logging.ContextMap{}, "Performing LLM final request")

	// create "chat_history" string
	historyMessages := ""
	for _, message := range history {
		switch message.Role {
		case "user":
			historyMessages += "`HumanMessage`: `" + message.Content + "`\n"
		case "assistant":
			historyMessages += "`AIMessage`: `" + message.Content + "`\n"
		}
	}

	// create json string from context
	contextString := ""
	if len(context) != 0 {
		contextString += "{"
		chunkNr := 1
		for _, example := range context {
			json, err := json.Marshal(example)
			if err != nil {
				logging.Log.Errorf(&logging.ContextMap{}, "Error marshalling context: %v", err)
				return "", nil
			}
			contextString += fmt.Sprintf("\"chunk %v\": %v", chunkNr, string(json)) + ", "
			chunkNr++
		}
		// remove last comma, then add closing bracket
		contextString = contextString[:len(contextString)-2]
		contextString += "}"
	}

	// create string from prohibited words
	prohibitedWordsString := ""
	for _, word := range prohibitedWords {
		prohibitedWordsString += word + ", "
	}

	// create string from error list 1
	errorList1String := ""
	for _, word := range errorList1 {
		errorList1String += word + ", "
	}

	// create string from error list 2
	errorList2String := ""
	for _, word := range errorList2 {
		errorList2String += word + ", "
	}

	// Create map for the data to be used in the template
	dataMap := make(map[string]string)
	dataMap["query"] = query
	dataMap["chat_history"] = historyMessages
	dataMap["context"] = contextString
	dataMap["prohibit_word_list"] = prohibitedWordsString
	dataMap["error_list_1"] = errorList1String
	dataMap["error_list_2"] = errorList2String

	// Format the user and system template
	userPrompt := formatTemplate(userTemplate, dataMap)
	logging.Log.Debugf(&logging.ContextMap{}, "User template for final query: %v", userPrompt)
	systemPrompt := formatTemplate(systemTemplate, dataMap)
	logging.Log.Debugf(&logging.ContextMap{}, "System template for final query: %v", systemPrompt)

	// create options
	var maxTokens int32 = 2000
	var temperature float32 = 0.0
	options := &sharedtypes.ModelOptions{
		MaxTokens:   &maxTokens,
		Temperature: &temperature,
	}

	// get the LLM handler endpoint.
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send chat request.
	responseChannel := sendChatRequest(userPrompt, "general", nil, 0, systemPrompt, llmHandlerEndpoint, nil, options, nil)

	// Create a stream channel
	streamChannel := make(chan string, 400)

	// calculate input token count
	inputTokenCount, err := openAiTokenCount(tokenCountModelName, userPrompt+systemPrompt)
	if err != nil {
		panic(err)
	}
	totalInputTokenCount := previousInputTokenCount + inputTokenCount

	// Start a goroutine to transfer the data from the response channel to the stream channel.
	go transferDatafromResponseToStreamChannel(&responseChannel, &streamChannel, false, true, tokenCountEndpoint, totalInputTokenCount, previousOutputTokenCount, tokenCountModelName, userEmail, true, contextString)

	return "", &streamChannel
}
