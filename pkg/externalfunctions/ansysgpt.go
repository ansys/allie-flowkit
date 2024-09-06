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

// AnsysGPTPerformLLMRephraseRequest performs a rephrase request to LLM
//
// Parameters:
//   - template: the template for the rephrase request
//   - query: the user query
//   - history: the conversation history
//
// Returns:
//   - rephrasedQuery: the rephrased query
func AnsysGPTPerformLLMRephraseRequest(template string, query string, history []sharedtypes.HistoricMessage) (rephrasedQuery string) {
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
	userTemplate := formatTemplate(template, dataMap)
	logging.Log.Debugf(internalstates.Ctx, "User template: %v", userTemplate)

	// Perform the general request
	rephrasedQuery, _, err := performGeneralRequest(userTemplate, nil, false, "You are AnsysGPT, a technical support assistant that is professional, friendly and multilingual that generates a clear and concise answer")
	if err != nil {
		panic(err)
	}

	logging.Log.Debugf(internalstates.Ctx, "Rephrased query: %v", rephrasedQuery)

	return rephrasedQuery
}

// AnsysGPTPerformLLMRephraseRequestOld performs a rephrase request to LLM
//
// Parameters:
//   - template: the template for the rephrase request
//   - query: the user query
//   - history: the conversation history
//
// Returns:
//   - rephrasedQuery: the rephrased query
func AnsysGPTPerformLLMRephraseRequestOld(template string, query string, history []sharedtypes.HistoricMessage) (rephrasedQuery string) {
	logging.Log.Debugf(internalstates.Ctx, "Performing LLM rephrase request")

	historyMessages := ""
	for _, entry := range history {
		switch entry.Role {
		case "user":
			historyMessages += "HumanMessage(content):" + entry.Content + "\n"
		case "assistant":
			historyMessages += "AIMessage(content):" + entry.Content + "\n"
		}
	}

	// Create map for the data to be used in the template
	dataMap := make(map[string]string)
	dataMap["query"] = query
	dataMap["chat_history"] = historyMessages

	// Format the template
	systemTemplate := formatTemplate(template, dataMap)
	logging.Log.Debugf(internalstates.Ctx, "System template: %v", systemTemplate)

	// Perform the general request
	rephrasedQuery, _, err := performGeneralRequest(query, nil, false, systemTemplate)
	if err != nil {
		panic(err)
	}

	logging.Log.Debugf(internalstates.Ctx, "Rephrased query: %v", rephrasedQuery)

	return rephrasedQuery
}

// AnsysGPTBuildFinalQuery builds the final query for Ansys GPT
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
			// indexList = append(indexList, "ibp-app-brief")
			// indexList = append(indexList, "pyansys_help_documentation")
			// indexList = append(indexList, "pyansys-examples")
		case "Ansys Semiconductor":
			// indexList = append(indexList, "ansysgpt-scbu")
		default:
			logging.Log.Warnf(internalstates.Ctx, "Invalid indexGroup: %v\n", indexGroup)
			return
		}
	}

	return indexList
}

// AnsysGPTACSSemanticHybridSearchs performs a semantic hybrid search in ACS
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
		return semanticSearchOutput[:topK]
	}

	return semanticSearchOutput
}

// AnsysGPTGetSystemPrompt returns the system prompt for Ansys GPT
//
// Returns:
//   - systemPrompt: the system prompt
func AnsysGPTGetSystemPrompt(rephrasedQuery string) string {
	return `Orders: You are AnsysGPT, a technical support assistant that is professional, friendly and multilingual that generates a clear and concise answer to the user question adhering to these strict guidelines: \n
            You must always answer user queries using the provided 'context' and 'chat_history' only. If you cannot find an answer in the 'context' or the 'chat_history', never use your base knowledge to generate a response. \n

            You are a multilingual expert that will *always reply the user in the same language as that of their 'query' in ` + rephrasedQuery + `*. If the 'query' is in Japanese, your response must be in Japanese. If the 'query' is in Cantonese, your response must be in Cantonese. If the 'query' is in English, your response must be in English. You *must always* be consistent in your multilingual ability. \n

            You have the capability to learn or *remember information from past three interactions* with the user. \n

            You are a smart Technical support assistant that can distingush between a fresh independent query and a follow-up query based on 'chat_history'. \n

            If you find the user's 'query' to be a follow-up question, consider the 'chat_history' while generating responses. Use the information from the 'chat_history' to provide contextually relevant responses. When answering follow-up questions that can be answered using the 'chat_history' alone, do not provide any references. \n

            *Always* your answer must include the 'content', 'sourceURL_lvl3' of all the chunks in 'context' that are relevant to the user's query in 'query'. But, never cite 'sourceURL_lvl3' under the heading 'References'. \n

            The 'content' and 'sourceURL_lvl3' must be included together in your answer, with the 'sourceTitle_lvl2', 'sourceURL_lvl2' and '@search.reranker_score' serving as a citation for the 'content'. Include 'sourceURL_lvl3' directly in the answer in-line with the source, not in the references section. \n

            In your response follow a style of citation where each source is assigned a number, for example '[1]', that corresponds to the 'sourceURL_lvl3', 'sourceTitle_lvl2' and 'sourceURL_lvl2' in the 'context'. \n

            Make sure you always provide 'URL: Extract the value of 'sourceURL_lvl3'' in line with every source in your answer. For example 'You will learn to find the total drag and lift on a solar car in Ansys Fluent in this course. URL: [1] https://courses.ansys.com/index.php/courses/aerodynamics-of-a-solar-car/'. \n

            Never mention the position of chunk in your response for example 'chunk 1 / chunk 4'/ first chunk / third chunk'. \n

            **Always** aim to make your responses conversational and engaging, while still providing accurate and helpful information. \n

            If the user greets you, you must *always* reply them in a polite and friendly manner. You *must never* reply "I'm sorry, could you please provide more details or ask a different question?" in this case. \n

            If the user acknowledges you, you must *always* reply them in a polite and friendly manner. You *must never* reply "I'm sorry, could you please provide more details or ask a different question?" in this case. \n

            If the user asks about your purpose, you must *always* reply them in a polite and friendly manner. You *must never* reply "I'm sorry, could you please provide more details or ask a different question?" in this case. \n

            If the user asks who are you?, you must *always* reply them in a polite and friendly manner. You *must never* reply "I'm sorry, could you please provide more details or ask a different question?" in this case. \n

            When providing information from a source, try to introduce it in a *conversational manner*. For example, instead of saying 'In the chunk titled...', you could say 'I found a great resource titled... that explains...'. \n

            If a chunk has empty fields in it's 'sourceTitle_lvl2' and 'sourceURL_lvl2', you *must never* cite that chunk under references in your response. \n

            You must never provide JSON format in your answer and never cite references in JSON format.\n

            Strictly provide your response everytime in the below format:

            Your answer
            Always provide 'URL: Extract the value of 'sourceURL_lvl3'' *inline right next to each source* and *not at the end of your answer*.
            References:
            [1] Title: Extract the value of 'sourceTitle_lvl2', URL: Extract the value of 'sourceURL_lvl2', Relevance: Extract the value of '@search.reranker_score' /4.0.
            *Always* provide References for all the chunks in 'context'.
            Do not provide 'sourceTitle_lvl3' in your response.
            When answering follow-up questions that can be answered using the 'chat_history' alone, *do not provide any references*.
            **Never** cite chunk that has empty fields in it's 'sourceTitle_lvl2' and 'sourceURL_lvl2' under References.
            **Never** provide the JSON format in your response and References.

            Only provide a reference if it was found in the "context". Under no circumstances should you create your own references from your base knowledge or the internet. \n

            Here's an example of how you should structure your response: \n

                Designing an antenna involves several steps, and Ansys provides a variety of tools to assist you in this process. \n
                The Ansys HFSS Antenna Toolkit, for instance, can automatically create the geometry of your antenna design with boundaries and excitations assigned. It also sets up the solution and generates post-processing reports for several popular antenna elements. Over 60 standard antenna topologies are available in the toolkit, and all the antenna models generated are ready to simulate. You can run a quick analysis of any antenna of your choosing [1]. URL: [1] https://www.youtube.com/embed/mhM6U2xn0Q0?start=25&end=123  \n
                In another example, a rectangular edge fed patch antenna is created using the HFSS antenna toolkit. The antenna is synthesized for 3.5 GHz and the geometry model is already created for you. After analyzing the model, you can view the results generated from the toolkit. The goal is to fold or bend the antenna so that it fits onto the sidewall of a smartphone. After folding the antenna and reanalyzing, you can view the results such as return loss, input impedance, and total radiated power of the antenna [2]. URL: [2] https://www.youtube.com/embed/h0QttEmQ88E?start=94&end=186  \n
                Lastly, Ansys Electronics Desktop integrates rigorous electromagnetic analysis with system and circuit simulation in a comprehensive, easy-to-use design platform. This platform is used to automatically create antenna geometries with materials, boundaries, excitations, solution setups, and post-processing reports [3]. URL: [3] https://ansyskm.ansys.com/forums/topic/ansys-hfss-antenna-synthesis-from-hfss-antenna-toolkit-part-2/  \n
                I hope this helps you in your antenna design process. If you have any more questions, feel free to ask! \n
                References:
                [1] Title: "ANSYS HFSS: Antenna Synthesis from HFSS Antenna Toolkit - Part 2", URL: https://ansyskm.ansys.com/forums/topic/ansys-hfss-antenna-synthesis-from-hfss-antenna-toolkit-part-2/, Relevance: 3.53/4.0
                [2] Title: "Cosimulation Using Ansys HFSS and Circuit - Lesson 2 - ANSYS Innovation Courses", URL: https://courses.ansys.com/index.php/courses/cosimulation-using-ansys-hfss/lessons/cosimulation-using-ansys-hfss-and-circuit-lesson-2/, Relevance: 2.54/4.0`
}
