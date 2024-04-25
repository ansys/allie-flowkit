package externalfunctions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// PerformVectorEmbeddingRequest performs a vector embedding request to LLM
//
// Parameters:
//   - input: the input string
//
// Returns:
//   - embeddedVector: the embedded vector in float32 format
func PerformVectorEmbeddingRequest(input string, llmHandlerEndpoint string) (embeddedVector []float32) {
	// Log the request
	log.Println("Performing vector embedding request for demand:", input)

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
func PerformKeywordExtractionRequest(input string, maxKeywordsSearch uint32, llmHandlerEndpoint string) (keywords []string) {
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
func PerformGeneralRequest(input string, history []HistoricMessage, isStream bool, systemPrompt string, llmHandlerEndpoint string) (message string, stream *chan string) {
	// Set up WebSocket connection with LLM and send chat request
	responseChannel := sendChatRequest(input, "general", history, 0, systemPrompt, llmHandlerEndpoint)

	// If isStream is true, create a stream channel and return asap
	if isStream {
		// Create a stream channel
		streamChannel := make(chan string, 400)

		// Start a goroutine to transfer the data from the response channel to the stream channel
		go transferDatafromResponseToStreamChannel(&responseChannel, &streamChannel)

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
func PerformCodeLLMRequest(input string, history []HistoricMessage, isStream bool, llmHandlerEndpoint string) (message string, stream *chan string) {
	// Set up WebSocket connection with LLM and send chat request
	responseChannel := sendChatRequest(input, "code", history, 0, "", llmHandlerEndpoint)

	// If isStream is true, create a stream channel and return asap
	if isStream {
		// Create a stream channel
		streamChannel := make(chan string, 400)

		// Start a goroutine to transfer the data from the response channel to the stream channel
		go transferDatafromResponseToStreamChannel(&responseChannel, &streamChannel)

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
func SendVectorsToKnowledgeDB(vector []float32, keywords []string, keywordsSearch bool, collection string, similaritySearchResults int, similaritySearchMinScore float64, knowledgeDbEndpoint string) (databaseResponse []DbResponse) {

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
func GetListCollections(knowledgeDbEndpoint string) (collectionsList []string) {
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
	maxHopsNumber int,
	knowledgeDbEndpoint string) (dependenciesIds []string) {

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
func GeneralNeo4jQuery(query string, knowledgeDbEndpoint string) (databaseResponse neo4jResponse) {
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
func GeneralQuery(collectionName string, maxRetrievalCount int, outputFields []string, filters DbFilters, knowledgeDbEndpoint string) (databaseResponse []DbResponse) {
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

	// Initial request
	finalQuery = "Based on the following examples:\n\n"

	for i, element := range knowledgedbResponse {
		// Add the example number
		finalQuery += "--- START EXAMPLE " + fmt.Sprint(i+1) + "---\n"
		finalQuery += ">>> Summary:\n" + element.Summary + "\n\n"
		finalQuery += ">>> Code snippet:\n```python\n" + element.Text + "\n```\n"
		finalQuery += "--- END EXAMPLE " + fmt.Sprint(i+1) + "---\n\n"
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
	getChildren bool,
	knowledgeDbEndpoint string) (databaseResponse []DbResponse) {
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
