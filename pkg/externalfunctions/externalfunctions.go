package externalfunctions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/ansys/allie-flowkit/pkg/config"
	"github.com/schollz/closestmatch"
	"github.com/texttheater/golang-levenshtein/levenshtein"
)

var ExternalFunctionsMap = map[string]interface{}{
	"PerformVectorEmbeddingRequest":                 PerformVectorEmbeddingRequest,
	"PerformKeywordExtractionRequest":               PerformKeywordExtractionRequest,
	"PerformGeneralRequest":                         PerformGeneralRequest,
	"PerformCodeLLMRequest":                         PerformCodeLLMRequest,
	"BuildLibraryContext":                           BuildLibraryContext,
	"SendVectorsToKnowledgeDB":                      SendVectorsToKnowledgeDB,
	"GetListCollections":                            GetListCollections,
	"RetrieveDependencies":                          RetrieveDependencies,
	"GeneralNeo4jQuery":                             GeneralNeo4jQuery,
	"GeneralQuery":                                  GeneralQuery,
	"BuildFinalQueryForGeneralLLMRequest":           BuildFinalQueryForGeneralLLMRequest,
	"BuildFinalQueryForCodeLLMRequest":              BuildFinalQueryForCodeLLMRequest,
	"SimilaritySearch":                              SimilaritySearch,
	"CreateKeywordsDbFilter":                        CreateKeywordsDbFilter,
	"CreateTagsDbFilter":                            CreateTagsDbFilter,
	"CreateMetadataDbFilter":                        CreateMetadataDbFilter,
	"CreateDbFilter":                                CreateDbFilter,
	"AppendMessageHistory":                          AppendMessageHistory,
	"AnsysGPTCheckProhibitedWords":                  AnsysGPTCheckProhibitedWords,
	"AnsysGPTExtractFieldsFromQuery":                AnsysGPTExtractFieldsFromQuery,
	"AnsysGPTPerformLLMRephraseRequest":             AnsysGPTPerformLLMRephraseRequest,
	"AnsysGPTBuildFinalQuery":                       AnsysGPTBuildFinalQuery,
	"AnsysGPTPerformLLMRequest":                     AnsysGPTPerformLLMRequest,
	"AnsysGPTReturnIndexList":                       AnsysGPTReturnIndexList,
	"AnsysGPTACSSemanticHybridSearchs":              AnsysGPTACSSemanticHybridSearchs,
	"AnsysGPTRemoveNoneCitationsFromSearchResponse": AnsysGPTRemoveNoneCitationsFromSearchResponse,
	"AnsysGPTReorderSearchResponse":                 AnsysGPTReorderSearchResponse,
	"AnsysGPTGetSystemPrompt":                       AnsysGPTGetSystemPrompt,
}

// PerformVectorEmbeddingRequest performs a vector embedding request to LLM
//
// Parameters:
//   - input: the input string
//
// Returns:
//   - embeddedVector: the embedded vector in float32 format
func PerformVectorEmbeddingRequest(input string) (embeddedVector []float32) {
	// Log the request
	log.Println("Performing vector embedding request for demand:", input)

	// get the LLM handler endpoint
	llmHandlerEndpoint := *config.AllieFlowkitConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send embeddings request
	responseChannel := sendEmbeddingsRequest(input, llmHandlerEndpoint)

	// Process the first response and close the channel
	var embedding32 []float32
	for response := range responseChannel {
		// Log LLM response
		log.Println("Received embeddings response... Storing array.")

		// Get embedded vector array
		embedding32 = response.EmbeddedData

		// Mark that the first response has been received
		firstResponseReceived := true

		// Exit the loop after processing the first response
		if firstResponseReceived {
			break
		}
	}

	// Close the response channel
	close(responseChannel)

	return embedding32
}

// PerformSummaryRequest performs a keywords summary request to LLM
//
// Parameters:
//   - input: the input string
//   - maxKeywordsSearch: the maximum number of keywords to search for
//
// Returns:
//   - keywords: the keywords extracted from the input string as a slice of strings
func PerformKeywordExtractionRequest(input string, maxKeywordsSearch uint32) (keywords []string) {
	// get the LLM handler endpoint
	llmHandlerEndpoint := *config.AllieFlowkitConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send chat request
	responseChannel := sendChatRequestNoHistory(input, "keywords", maxKeywordsSearch, llmHandlerEndpoint)

	// Process all responses
	var responseAsStr string
	for response := range responseChannel {

		// Accumulate the responses
		responseAsStr += *(response.ChatData)

		// If we are at the last message, break the loop
		if *(response.IsLast) {
			break
		}
	}

	// Close the response channel
	close(responseChannel)

	// Unmarshal JSON data into the result variable
	err := json.Unmarshal([]byte(responseAsStr), &keywords)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON data: %s", err)
	}

	// Return the response
	return keywords
}

// PerformGeneralRequest performs a general chat completion request to LLM
//
// Parameters:
//   - input: the input string
//   - history: the conversation history
//   - isStream: the stream flag
//   - systemPrompt: the system prompt
//
// Returns:
//   - message: the generated message
//   - stream: the stream channel
func PerformGeneralRequest(input string, history []HistoricMessage, isStream bool, systemPrompt string) (message string, stream *chan string) {
	// get the LLM handler endpoint
	llmHandlerEndpoint := *config.AllieFlowkitConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send chat request
	responseChannel := sendChatRequest(input, "general", history, 0, systemPrompt, llmHandlerEndpoint)

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

// PerformCodeLLMRequest performs a code generation request to LLM
//
// Parameters:
//   - input: the input string
//   - history: the conversation history
//   - isStream: the stream flag
//
// Returns:
//   - message: the generated code
//   - stream: the stream channel
func PerformCodeLLMRequest(input string, history []HistoricMessage, isStream bool, validateCode bool) (message string, stream *chan string) {
	// get the LLM handler endpoint
	llmHandlerEndpoint := *config.AllieFlowkitConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send chat request
	responseChannel := sendChatRequest(input, "code", history, 0, "", llmHandlerEndpoint)

	// If isStream is true, create a stream channel and return asap
	if isStream {
		// Create a stream channel
		streamChannel := make(chan string, 400)

		// Start a goroutine to transfer the data from the response channel to the stream channel
		go transferDatafromResponseToStreamChannel(&responseChannel, &streamChannel, validateCode)

		// Return the stream channel
		return "", &streamChannel
	}

	// else Process all responses
	var responseAsStr string
	for response := range responseChannel {
		// Accumulate the responses
		responseAsStr += *(response.ChatData)

		// If we are at the last message, break the loop
		if *(response.IsLast) {
			break
		}
	}

	// Close the response channel
	close(responseChannel)

	// Code validation
	if validateCode {

		// Extract the code from the response
		pythonCode, err := extractPythonCode(responseAsStr)
		if err != nil {
			log.Printf("Error extracting Python code: %v\n", err)
		} else {

			// Validate the Python code
			valid, warnings, err := validatePythonCode(pythonCode)
			if err != nil {
				log.Printf("Error validating Python code: %v\n", err)
			} else {
				if valid {
					if warnings {
						responseAsStr += "\nCode has warnings."
					} else {
						responseAsStr += "\nCode is valid."
					}
				} else {
					responseAsStr += "\nCode is invalid."
				}
			}
		}
	}

	// Return the response
	return responseAsStr, nil
}

// BuildLibraryContext builds the context string for the query
//
// Parameters:
//   - message: the message string
//   - libraryContext: the library context string
//
// Returns:
//   - messageWithContext: the message with context
func BuildLibraryContext(message string, libraryContext string) (messageWithContext string) {
	// Check if "pyansys" is in the library context
	message = libraryContext + message

	return message
}

// SendVectorsToKnowledgeDB sends the given vector to the KnowledgeDB and
// returns the most relevant data. The number of results is specified in the
// config file. The keywords are used to filter the results. The min score
// filter is also specified in the config file. If it is not specified, the
// default value is used.
//
// The function returns the most relevant data.
//
// Parameters:
//   - vector: the vector to be sent to the KnowledgeDB
//   - keywords: the keywords to be used to filter the results
//   - keywordsSearch: the flag to enable the keywords search
//   - collection: the collection name
//   - similaritySearchResults: the number of results to be returned
//   - similaritySearchMinScore: the minimum score for the results
//
// Returns:
//   - databaseResponse: an array of the most relevant data
func SendVectorsToKnowledgeDB(vector []float32, keywords []string, keywordsSearch bool, collection string, similaritySearchResults int, similaritySearchMinScore float64) (databaseResponse []DbResponse) {
	// get the KnowledgeDB endpoint
	knowledgeDbEndpoint := *config.AllieFlowkitConfig.KNOWLEDGE_DB_ENDPOINT

	// Log the request
	log.Println("Connecting to the KnowledgeDB.")

	// Build filters
	var filters DbFilters

	// -- Add the keywords filter if needed
	if keywordsSearch {
		filters.KeywordsFilter = DbArrayFilter{
			NeedAll:    false,
			FilterData: keywords,
		}
	}

	// -- Add the level filter
	filters.LevelFilter = []string{"leaf"}

	// Create a new resource instance
	requestInput := similaritySearchInput{
		CollectionName:    collection,
		EmbeddedVector:    vector,
		MaxRetrievalCount: similaritySearchResults,
		Filters:           filters,
		MinScore:          similaritySearchMinScore,
		OutputFields: []string{
			"guid",
			"document_id",
			"document_name",
			"summary",
			"keywords",
			"text",
		},
	}

	// Convert the resource instance to JSON.
	jsonData, err := json.Marshal(requestInput)
	if err != nil {
		log.Fatal(err)
	}

	// Specify the target endpoint.
	requestURL := knowledgeDbEndpoint + "/similarity_search"

	// Create a new HTTP request with the JSON data.
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal(err)
	}

	// Set the appropriate content type for the request.
	req.Header.Set("Content-Type", "application/json")

	// Send the HTTP request using the default HTTP client.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Read and display the response body.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Log the similarity search response
	log.Println("Knowledge DB response:", string(body))
	log.Println("Knowledge DB response received!")

	// Unmarshal the response body to the appropriate struct.
	var response similaritySearchOutput
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Fatal(err)
	}

	var mostRelevantData []DbResponse
	var count int = 1
	for _, element := range response.SimilarityResult {
		// Log the result
		log.Printf("Result #%d:", count)
		log.Println("Similarity score:", element.Score)
		log.Println("Similarity file id:", element.Data.DocumentId)
		log.Println("Similarity file name:", element.Data.DocumentName)
		log.Println("Similarity summary:", element.Data.Summary)

		// Add the result to the list
		mostRelevantData = append(mostRelevantData, element.Data)

		// Check whether we have enough results
		if count >= similaritySearchResults {
			break
		} else {
			count++
		}
	}

	// Return the most relevant data
	return mostRelevantData
}

// GetListCollections retrieves the list of collections from the KnowledgeDB.
//
// The function returns the list of collections.
//
// Parameters:
//   - knowledgeDbEndpoint: the KnowledgeDB endpoint
//
// Returns:
//   - collectionsList: the list of collections
func GetListCollections() (collectionsList []string) {
	// get the KnowledgeDB endpoint
	knowledgeDbEndpoint := *config.AllieFlowkitConfig.KNOWLEDGE_DB_ENDPOINT

	// Specify the target endpoint.
	requestURL := knowledgeDbEndpoint + "/list_collections"

	// Create a new HTTP request with the JSON data.
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Set the appropriate content type for the request.
	req.Header.Set("Content-Type", "application/json")

	// Send the HTTP request using the default HTTP client.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Read and display the response body.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Unmarshal the response body to the appropriate struct.
	var response DBListCollectionsOutput
	err = json.Unmarshal(body, &response)

	// Log the result and return the list of collections
	if err != nil || !response.Success {
		log.Println("List collections retrieval failed!")
		return []string{}
	} else {
		log.Println("List collections response received!")
		log.Println("Collections:", response.Collections)
		return response.Collections
	}
}

// RetrieveDependencies retrieves the dependencies of the specified source node.
//
// The function returns the list of dependencies.
//
// Parameters:
//   - collectionName: the name of the collection to which the data objects will be added.
//   - relationshipName: the name of the relationship to retrieve dependencies for.
//   - relationshipDirection: the direction of the relationship to retrieve dependencies for.
//   - sourceDocumentId: the document ID of the source node.
//   - nodeTypesFilter: filter based on node types.
//   - maxHopsNumber: maximum number of hops to traverse.
//
// Returns:
//   - dependenciesIds: the list of dependencies
func RetrieveDependencies(
	collectionName string,
	relationshipName string,
	relationshipDirection string,
	sourceDocumentId string,
	nodeTypesFilter DbArrayFilter,
	maxHopsNumber int) (dependenciesIds []string) {
	// get the KnowledgeDB endpoint
	knowledgeDbEndpoint := *config.AllieFlowkitConfig.KNOWLEDGE_DB_ENDPOINT

	// Create the URL
	requestURL := knowledgeDbEndpoint + "/retrieve_dependencies"

	// Create the retrieveDependenciesInput object
	requestInput := retrieveDependenciesInput{
		CollectionName:        collectionName,
		RelationshipName:      relationshipName,
		RelationshipDirection: relationshipDirection,
		SourceDocumentId:      sourceDocumentId,
		NodeTypesFilter:       nodeTypesFilter,
		MaxHopsNumber:         maxHopsNumber,
	}

	// Convert the resource instance to JSON.
	jsonData, err := json.Marshal(requestInput)
	if err != nil {
		log.Fatal(err)
	}

	// Create a new HTTP request with the JSON data.
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal(err)
	}

	// Set the appropriate content type for the request.
	req.Header.Set("Content-Type", "application/json")

	// Send the HTTP request using the default HTTP client.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Read and display the response body.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Knowledge DB RetrieveDependencies response received!")

	// Unmarshal the response body to the appropriate struct.
	var response retrieveDependenciesOutput
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Fatal(err)
	}

	return response.DependenciesIds
}

// GeneralNeo4jQuery executes the given Neo4j query and returns the response.
//
// The function returns the neo4j response.
//
// Parameters:
//   - query: the Neo4j query to be executed.
//
// Returns:
//   - databaseResponse: the Neo4j response
func GeneralNeo4jQuery(query string) (databaseResponse neo4jResponse) {
	// get the KnowledgeDB endpoint
	knowledgeDbEndpoint := *config.AllieFlowkitConfig.KNOWLEDGE_DB_ENDPOINT

	// Create the URL
	requestURL := knowledgeDbEndpoint + "/general_neo4j_query"

	// Create the retrieveDependenciesInput object
	requestInput := GeneralNeo4jQueryInput{
		Query: query,
	}

	// Convert the resource instance to JSON.
	jsonData, err := json.Marshal(requestInput)
	if err != nil {
		log.Fatal(err)
	}

	// Create a new HTTP request with the JSON data.
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal(err)
	}

	// Set the appropriate content type for the request.
	req.Header.Set("Content-Type", "application/json")

	// Send the HTTP request using the default HTTP client.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Read and display the response body.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Knowledge DB GeneralNeo4jQuery response received!")

	// Unmarshal the response body to the appropriate struct.
	var response GeneralNeo4jQueryOutput
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Fatal(err)
	}

	return response.Response
}

// GeneralQuery performs a general query in the KnowledgeDB.
//
// The function returns the query results.
//
// Parameters:
//   - collectionName: the name of the collection to which the data objects will be added.
//   - maxRetrievalCount: the maximum number of results to be retrieved.
//   - outputFields: the fields to be included in the output.
//   - filters: the filter for the query.
//
// Returns:
//   - databaseResponse: the query results
func GeneralQuery(collectionName string, maxRetrievalCount int, outputFields []string, filters DbFilters) (databaseResponse []DbResponse) {
	// get the KnowledgeDB endpoint
	knowledgeDbEndpoint := *config.AllieFlowkitConfig.KNOWLEDGE_DB_ENDPOINT

	// Create the URL
	requestURL := knowledgeDbEndpoint + "/query"

	// Create the queryInput object
	requestInput := queryInput{
		CollectionName:    collectionName,
		MaxRetrievalCount: maxRetrievalCount,
		OutputFields:      outputFields,
		Filters:           filters,
	}

	// Convert the resource instance to JSON.
	jsonData, err := json.Marshal(requestInput)
	if err != nil {
		log.Fatal(err)
	}

	// Create a new HTTP request with the JSON data.
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal(err)
	}

	// Set the appropriate content type for the request.
	req.Header.Set("Content-Type", "application/json")

	// Send the HTTP request using the default HTTP client.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Read and display the response body.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Knowledge DB GeneralQuery response received!")

	// Unmarshal the response body to the appropriate struct.
	var response queryOutput
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Fatal(err)
	}

	return response.QueryResult
}

// BuildFinalQueryForGeneralLLMRequest builds the final query for a general
// request to LLM. The final query is a markdown string that contains the
// original request and the examples from the KnowledgeDB.
//
// Parameters:
//   - request: the original request
//   - knowledgedbResponse: the KnowledgeDB response
//
// Returns:
//   - finalQuery: the final query
func BuildFinalQueryForGeneralLLMRequest(request string, knowledgedbResponse []DbResponse) (finalQuery string) {

	// If there is no response from the KnowledgeDB, return the original request
	if len(knowledgedbResponse) == 0 {
		return request
	}

	// Build the final query using the KnowledgeDB response and the original request
	finalQuery = "Based on the following examples:\n\n--- INFO START ---\n"
	for _, example := range knowledgedbResponse {
		finalQuery += example.Text + "\n"
	}
	finalQuery += "--- INFO END ---\n\n" + request + "\n"

	// Return the final query
	return finalQuery
}

// BuildFinalQueryForCodeLLMRequest builds the final query for a code generation
// request to LLM. The final query is a markdown string that contains the
// original request and the code examples from the KnowledgeDB.
//
// Parameters:
//   - request: the original request
//   - knowledgedbResponse: the KnowledgeDB response
//
// Returns:
//   - finalQuery: the final query
func BuildFinalQueryForCodeLLMRequest(request string, knowledgedbResponse []DbResponse) (finalQuery string) {
	// Build the final query using the KnowledgeDB response and the original request
	// We have to use the text from the DB response and the original request.
	//
	// The prompt should be in the following format:
	//
	// ******************************************************************************
	// Based on the following examples:
	//
	// --- START EXAMPLE {response_n}---
	// >>> Summary:
	// {knowledge_db_response_n_summary}
	//
	// >>> Code snippet:
	// ```python
	// {knowledge_db_response_n_text}
	// ```
	// --- END EXAMPLE {response_n}---
	//
	// --- START EXAMPLE {response_n}---
	// ...
	// --- END EXAMPLE {response_n}---
	//
	// Generate the Python code for the following request:
	//
	// >>> Request:
	// {original_request}
	// ******************************************************************************

	// If there is no response from the KnowledgeDB, return the original request
	if len(knowledgedbResponse) > 0 {
		// Initial request
		finalQuery = "Based on the following examples:\n\n"

		for i, element := range knowledgedbResponse {
			// Add the example number
			finalQuery += "--- START EXAMPLE " + fmt.Sprint(i+1) + "---\n"
			finalQuery += ">>> Summary:\n" + element.Summary + "\n\n"
			finalQuery += ">>> Code snippet:\n```python\n" + element.Text + "\n```\n"
			finalQuery += "--- END EXAMPLE " + fmt.Sprint(i+1) + "---\n\n"
		}
	}

	// Pass in the original request
	finalQuery += "Generate the Python code for the following request:\n>>> Request:\n" + request + "\n"

	// Return the final query
	return finalQuery
}

// SimilaritySearch performs a similarity search in the KnowledgeDB.
//
// The function returns the similarity search results.
//
// Parameters:
//   - collectionName: the name of the collection to which the data objects will be added.
//   - embeddedVector: the embedded vector used for searching.
//   - maxRetrievalCount: the maximum number of results to be retrieved.
//   - outputFields: the fields to be included in the output.
//   - filters: the filter for the query.
//   - minScore: the minimum score filter.
//   - getLeafNodes: flag to indicate whether to retrieve all the leaf nodes in the result node branch.
//   - getSiblings: flag to indicate whether to retrieve the previous and next node to the result nodes.
//   - getParent: flag to indicate whether to retrieve the parent object.
//   - getChildren: flag to indicate whether to retrieve the children objects.
//
// Returns:
//   - databaseResponse: the similarity search results
func SimilaritySearch(
	collectionName string,
	embeddedVector []float32,
	maxRetrievalCount int,
	outputFields []string,
	filters DbFilters,
	minScore float64,
	getLeafNodes bool,
	getSiblings bool,
	getParent bool,
	getChildren bool) (databaseResponse []DbResponse) {
	// get the KnowledgeDB endpoint
	knowledgeDbEndpoint := *config.AllieFlowkitConfig.KNOWLEDGE_DB_ENDPOINT

	// Create the URL
	requestURL := knowledgeDbEndpoint + "/similarity_search"

	// Create the retrieveDependenciesInput object
	requestInput := similaritySearchInput{
		CollectionName:    collectionName,
		EmbeddedVector:    embeddedVector,
		MaxRetrievalCount: maxRetrievalCount,
		OutputFields:      outputFields,
		Filters:           filters,
		MinScore:          minScore,
		GetLeafNodes:      getLeafNodes,
		GetSiblings:       getSiblings,
		GetParent:         getParent,
		GetChildren:       getChildren,
	}

	// Convert the resource instance to JSON.
	jsonData, err := json.Marshal(requestInput)
	if err != nil {
		log.Fatal(err)
	}

	// Create a new HTTP request with the JSON data.
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal(err)
	}

	// Set the appropriate content type for the request.
	req.Header.Set("Content-Type", "application/json")

	// Send the HTTP request using the default HTTP client.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Read and display the response body.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Knowledge DB SimilaritySearch response received!")

	// Unmarshal the response body to the appropriate struct.
	var response similaritySearchOutput
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Fatal(err)
	}

	var similarityResults []DbResponse
	for _, element := range response.SimilarityResult {
		similarityResults = append(similarityResults, element.Data)
	}

	return similarityResults
}

// CreateKeywordsDbFilter creates a keywords filter for the KnowledgeDB.
//
// The function returns the keywords filter.
//
// Parameters:
//   - keywords: the keywords to be used for the filter
//   - needAll: flag to indicate whether all keywords are needed
//
// Returns:
//   - databaseFilter: the keywords filter
func CreateKeywordsDbFilter(keywords []string, needAll bool) (databaseFilter DbArrayFilter) {
	var keywordsFilters DbArrayFilter

	// -- Add the keywords filter if needed
	if len(keywords) > 0 {
		keywordsFilters = createDbArrayFilter(keywords, needAll)
	}

	return keywordsFilters
}

// CreateTagsDbFilter creates a tags filter for the KnowledgeDB.
//
// The function returns the tags filter.
//
// Parameters:
//   - tags: the tags to be used for the filter
//   - needAll: flag to indicate whether all tags are needed
//
// Returns:
//   - databaseFilter: the tags filter
func CreateTagsDbFilter(tags []string, needAll bool) (databaseFilter DbArrayFilter) {
	var tagsFilters DbArrayFilter

	// -- Add the tags filter if needed
	if len(tags) > 0 {
		tagsFilters = createDbArrayFilter(tags, needAll)
	}

	return tagsFilters
}

// CreateMetadataDbFilter creates a metadata filter for the KnowledgeDB.
//
// The function returns the metadata filter.
//
// Parameters:
//   - fieldName: the name of the field
//   - fieldType: the type of the field
//   - filterData: the filter data
//   - needAll: flag to indicate whether all data is needed
//
// Returns:
//   - databaseFilter: the metadata filter
func CreateMetadataDbFilter(fieldName string, fieldType string, filterData []string, needAll bool) (databaseFilter DbJsonFilter) {
	return createDbJsonFilter(fieldName, fieldType, filterData, needAll)
}

// CreateDbFilter creates a filter for the KnowledgeDB.
//
// The function returns the filter.
//
// Parameters:
//   - guid: the guid filter
//   - documentId: the document ID filter
//   - documentName: the document name filter
//   - level: the level filter
//   - tags: the tags filter
//   - keywords: the keywords filter
//   - metadata: the metadata filter
//
// Returns:
//   - databaseFilter: the filter
func CreateDbFilter(
	guid []string,
	documentId []string,
	documentName []string,
	level []string,
	tags DbArrayFilter,
	keywords DbArrayFilter,
	metadata []DbJsonFilter) (databaseFilter DbFilters) {
	var filters DbFilters

	// -- Add the guid filter if needed
	if len(guid) > 0 {
		filters.GuidFilter = guid
	}

	// -- Add the document ID filter if needed
	if len(documentId) > 0 {
		filters.DocumentIdFilter = documentId
	}

	// -- Add the document name filter if needed
	if len(documentName) > 0 {
		filters.DocumentNameFilter = documentName
	}

	// -- Add the level filter if needed
	if len(level) > 0 {
		filters.LevelFilter = level
	}

	// -- Add the tags filter if needed
	if len(tags.FilterData) > 0 {
		filters.TagsFilter = tags
	}

	// -- Add the keywords filter if needed
	if len(keywords.FilterData) > 0 {
		filters.KeywordsFilter = keywords
	}

	// -- Add the metadata filter if needed
	if len(metadata) > 0 {
		filters.MetadataFilter = metadata
	}

	return filters
}

// AppendMessageHistoryInput represents the input for the AppendMessageHistory function.
type AppendMessageHistoryRole string

const (
	user      AppendMessageHistoryRole = "user"
	assistant AppendMessageHistoryRole = "assistant"
	system    AppendMessageHistoryRole = "system"
)

// AppendMessageHistory appends a new message to the conversation history
//
// Parameters:
//   - newMessage: the new message
//   - role: the role of the message
//   - history: the conversation history
//
// Returns:
//   - updatedHistory: the updated conversation history
func AppendMessageHistory(newMessage string, role AppendMessageHistoryRole, history []HistoricMessage) (updatedHistory []HistoricMessage) {
	switch role {
	case user:
	case assistant:
	case system:
	default:
		log.Printf("Invalid role: %v\n", role)
		return history
	}

	// skip for empty messages
	if newMessage == "" {
		return history
	}

	// Create a new HistoricMessage
	newMessageHistory := HistoricMessage{
		Role:    string(role),
		Content: newMessage,
	}

	// Append the new message to the history
	history = append(history, newMessageHistory)

	return history
}

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
	// Check each prohibited word for exact match ignoring case
	for _, prohibitedWord := range prohibitedWords {
		if strings.Contains(strings.ToLower(query), strings.ToLower(prohibitedWord)) {
			return true, errorResponseMessage
		}
	}

	// If no exact match found, use fuzzy matching
	cutoff := 0.9
	cm := closestmatch.New(prohibitedWords, []int{3})
	for _, word := range strings.Fields(query) {
		closestWord := cm.Closest(word)
		distance := levenshtein.RatioForStrings([]rune(word), []rune(closestWord), levenshtein.DefaultOptions)
		if distance >= cutoff {
			return true, errorResponseMessage
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
func AnsysGPTExtractFieldsFromQuery(query string, fieldValues map[string][]string, defaultFields []AnsysGPTDefaultFields) (fields map[string]string) {
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

		// Check each possible value for exact match ignoring case
		for _, value := range values {
			if strings.Contains(strings.ToLower(query), strings.ToLower(value)) {
				fields[field] = value
				fmt.Println("Exact match found for", field, ":", value)
				break
			}
		}

		// Split the query into words
		words := strings.Fields(query)

		// If no exact match found, use fuzzy matching
		if fields[field] == "" {
			cutoff := 0.75
			cm := closestmatch.New(words, []int{3})
			for _, value := range values {
				for _, word := range strings.Fields(value) {
					closestWord := cm.Closest(word)
					distance := levenshtein.RatioForStrings([]rune(word), []rune(closestWord), levenshtein.DefaultOptions)
					if distance >= cutoff {
						fields[field] = value
						fmt.Println("Fuzzy match found for", field, ":", distance)
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
				fmt.Println("Default value found for", defaultField.FieldName, ":", defaultField.FieldDefaultValue)
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
func AnsysGPTPerformLLMRephraseRequest(template string, query string, history []HistoricMessage) (rephrasedQuery string) {
	fmt.Println("Performing rephrase request...")
	// Append messages with conversation entries
	historyMessages := ""
	// for _, entry := range history {
	// 	switch entry.Role {
	// 	case "user":
	// 		historyMessages += "HumanMessage(content): " + entry.Content + "\n"
	// 	case "assistant":
	// 		historyMessages += "AIMessage(content): " + entry.Content + "\n"
	// 	}
	// }
	// adding a sample comment

	// last message from history
    if len(history) > 0 {
        lastEntry := history[len(history)-1]
        switch lastEntry.Role {
        case "user":
            historyMessages += "HumanMessage(content): " + lastEntry.Content + "\n"
        case "assistant":
            historyMessages += "AIMessage(content): " + lastEntry.Content + "\n"
        }
    }


	// Create map for the data to be used in the template
	dataMap := make(map[string]string)
	dataMap["query"] = query
	dataMap["chat_history"] = historyMessages

	// Format the template
	systemTemplate := formatTemplate(template, dataMap)

	// Perform the general request
	rephrasedQuery, _ = PerformGeneralRequest(query, nil, false, systemTemplate)

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
func AnsysGPTBuildFinalQuery(refrasedQuery string, context []ACSSearchResponse) (finalQuery string, errorResponse string, displayFixedMessageToUser bool) {

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
func AnsysGPTPerformLLMRequest(finalQuery string, history []HistoricMessage, systemPrompt string, isStream bool) (message string, stream *chan string) {
	// get the LLM handler endpoint
	llmHandlerEndpoint := *config.AllieFlowkitConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send chat request
	responseChannel := sendChatRequest(finalQuery, "general", history, 0, systemPrompt, llmHandlerEndpoint)

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
			// indexList = append(indexList, "ansysgpt-alh")
		case "Ansys Products":
			// indexList = append(indexList, "lsdyna-documentation-r14")
			indexList = append(indexList, "ansysgpt-documentation-2023r2")
			indexList = append(indexList, "scade-documentation-2023r2")
			indexList = append(indexList, "ansys-dot-com-marketing")
			// indexList = append(indexList, "ibp-app-brief")
			// indexList = append(indexList, "pyansys_help_documentation")
			// indexList = append(indexList, "pyansys-examples")
		case "Ansys Semiconductor":
			// indexList = append(indexList, "ansysgpt-scbu")
		default:
			log.Printf("Invalid indexGroup: %v\n", indexGroup)
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
	topK int) (output []ACSSearchResponse) {

	output = make([]ACSSearchResponse, 0)
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
func AnsysGPTRemoveNoneCitationsFromSearchResponse(semanticSearchOutput []ACSSearchResponse, citations []AnsysGPTCitation) (reducedSemanticSearchOutput []ACSSearchResponse) {
	// iterate throught search response and keep matches to citations
	reducedSemanticSearchOutput = make([]ACSSearchResponse, len(citations))
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

// AnsysGPTReorderSearchResponse reorders the search response
//
// Parameters:
//   - semanticSearchOutput: the search response
//
// Returns:
//   - reorderedSemanticSearchOutput: the reordered search response
func AnsysGPTReorderSearchResponse(semanticSearchOutput []ACSSearchResponse) (reorderedSemanticSearchOutput []ACSSearchResponse) {
	// Sorting by Weight * SearchRerankerScore in descending order
	sort.Slice(semanticSearchOutput, func(i, j int) bool {
		return semanticSearchOutput[i].Weight*semanticSearchOutput[i].SearchRerankerScore > semanticSearchOutput[j].Weight*semanticSearchOutput[j].SearchRerankerScore
	})

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
