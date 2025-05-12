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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ansys/aali-sharedtypes/pkg/config"
	"github.com/ansys/aali-sharedtypes/pkg/logging"
	"github.com/ansys/aali-sharedtypes/pkg/sharedtypes"
)

// SendVectorsToKnowledgeDB sends the given vector to the KnowledgeDB and
// returns the most relevant data. The number of results is specified in the
// config file. The keywords are used to filter the results. The min score
// filter is also specified in the config file. If it is not specified, the
// default value is used.
//
// The function returns the most relevant data.
//
// Tags:
//   - @displayName: Similarity Search
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
func SendVectorsToKnowledgeDB(vector []float32, keywords []string, keywordsSearch bool, collection string, similaritySearchResults int, similaritySearchMinScore float64) (databaseResponse []sharedtypes.DbResponse) {
	// get the KnowledgeDB endpoint
	knowledgeDbEndpoint := config.GlobalConfig.KNOWLEDGE_DB_ENDPOINT

	// Log the request
	logging.Log.Debugf(&logging.ContextMap{}, "Connecting to the KnowledgeDB.")

	// Build filters
	var filters sharedtypes.DbFilters

	// -- Add the keywords filter if needed
	if keywordsSearch {
		filters.KeywordsFilter = sharedtypes.DbArrayFilter{
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
		errMessage := fmt.Sprintf("Error marshalling JSON data of POST /similarity_search request for aali-db: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	// Specify the target endpoint.
	requestURL := knowledgeDbEndpoint + "/similarity_search"

	// Create a new HTTP request with the JSON data.
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonData))
	if err != nil {
		errMessage := fmt.Sprintf("Error creating POST /similarity_search request for aali-db: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	// Set the appropriate content type for the request.
	req.Header.Set("Content-Type", "application/json")

	// Send the HTTP request using the default HTTP client.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errMessage := fmt.Sprintf("Error sending POST /similarity_search request to aali-db: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}
	defer resp.Body.Close()

	// Read and display the response body.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errMessage := fmt.Sprintf("Error reading response body of POST /similarity_search request from aali-db: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	// Log the similarity search response
	logging.Log.Debugf(&logging.ContextMap{}, "Knowledge DB response: %v", string(body))
	logging.Log.Debugf(&logging.ContextMap{}, "Knowledge DB response received!")

	// Unmarshal the response body to the appropriate struct.
	var response similaritySearchOutput
	err = json.Unmarshal(body, &response)
	if err != nil {
		errMessage := fmt.Sprintf("Error unmarshalling JSON data of POST /similarity_search response from aali-db: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	var mostRelevantData []sharedtypes.DbResponse
	var count int = 1
	for _, element := range response.SimilarityResult {
		// Log the result
		logging.Log.Debugf(&logging.ContextMap{}, "Result #%d:", count)
		logging.Log.Debugf(&logging.ContextMap{}, "Similarity score: %v", element.Score)
		logging.Log.Debugf(&logging.ContextMap{}, "Similarity file id: %v", element.Data.DocumentId)
		logging.Log.Debugf(&logging.ContextMap{}, "Similarity file name: %v", element.Data.DocumentName)
		logging.Log.Debugf(&logging.ContextMap{}, "Similarity summary: %v", element.Data.Summary)

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
// Tags:
//   - @displayName: List Collections
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
	knowledgeDbEndpoint := config.GlobalConfig.KNOWLEDGE_DB_ENDPOINT

	// Specify the target endpoint.
	requestURL := knowledgeDbEndpoint + "/list_collections"

	// Create a new HTTP request with the JSON data.
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		errMessage := fmt.Sprintf("Error creating GET /list_collections request for aali-db: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	// Set the appropriate content type for the request.
	req.Header.Set("Content-Type", "application/json")

	// Send the HTTP request using the default HTTP client.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errMessage := fmt.Sprintf("Error sending GET /list_collections request to aali-db: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}
	defer resp.Body.Close()

	// Read and display the response body.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errMessage := fmt.Sprintf("Error reading response body of GET /list_collections request from aali-db: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	// Unmarshal the response body to the appropriate struct.
	var response sharedtypes.DBListCollectionsOutput
	err = json.Unmarshal(body, &response)
	if err != nil {
		errMessage := fmt.Sprintf("Error unmarshalling JSON data of GET /list_collections response from aali-db: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	// Log the result and return the list of collections
	if !response.Success {
		errMessage := "Failed to retrieve list of collections from aali-db"
		logging.Log.Warn(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	} else {
		logging.Log.Debugf(&logging.ContextMap{}, "List collections response received!")
		logging.Log.Debugf(&logging.ContextMap{}, "Collections: %v", response.Collections)
		return response.Collections
	}
}

// RetrieveDependencies retrieves the dependencies of the specified source node.
//
// The function returns the list of dependencies.
//
// Tags:
//   - @displayName: Retrieve Dependencies
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
	nodeTypesFilter sharedtypes.DbArrayFilter,
	maxHopsNumber int) (dependenciesIds []string) {
	// get the KnowledgeDB endpoint
	knowledgeDbEndpoint := config.GlobalConfig.KNOWLEDGE_DB_ENDPOINT

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
		errMessage := fmt.Sprintf("Error marshalling JSON data of POST /retrieve_dependencies request for aali-db: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	// Create a new HTTP request with the JSON data.
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonData))
	if err != nil {
		errMessage := fmt.Sprintf("Error creating POST /retrieve_dependencies request for aali-db: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	// Set the appropriate content type for the request.
	req.Header.Set("Content-Type", "application/json")

	// Send the HTTP request using the default HTTP client.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errMessage := fmt.Sprintf("Error sending POST /retrieve_dependencies request to aali-db: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}
	defer resp.Body.Close()

	// Read and display the response body.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errMessage := fmt.Sprintf("Error reading response body of POST /retrieve_dependencies request from aali-db: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	logging.Log.Debugf(&logging.ContextMap{}, "Knowledge DB RetrieveDependencies response received!")

	// Unmarshal the response body to the appropriate struct.
	var response retrieveDependenciesOutput
	err = json.Unmarshal(body, &response)
	if err != nil {
		errMessage := fmt.Sprintf("Error unmarshalling JSON data of POST /retrieve_dependencies response from aali-db: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	return response.DependenciesIds
}

// GeneralNeo4jQuery executes the given Neo4j query and returns the response.
//
// The function returns the neo4j response.
//
// Tags:
//   - @displayName: General Neo4J Query
//
// Parameters:
//   - query: the Neo4j query to be executed.
//
// Returns:
//   - databaseResponse: the Neo4j response
func GeneralNeo4jQuery(query string) (databaseResponse sharedtypes.Neo4jResponse) {
	// get the KnowledgeDB endpoint
	knowledgeDbEndpoint := config.GlobalConfig.KNOWLEDGE_DB_ENDPOINT

	// Create the URL
	requestURL := knowledgeDbEndpoint + "/general_neo4j_query"

	// Create the retrieveDependenciesInput object
	requestInput := sharedtypes.GeneralNeo4jQueryInput{
		Query: query,
	}

	// Convert the resource instance to JSON.
	jsonData, err := json.Marshal(requestInput)
	if err != nil {
		errMessage := fmt.Sprintf("Error marshalling JSON data of POST /general_neo4j_query request for aali-db: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	// Create a new HTTP request with the JSON data.
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonData))
	if err != nil {
		errMessage := fmt.Sprintf("Error creating POST /general_neo4j_query request for aali-db: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	// Set the appropriate content type for the request.
	req.Header.Set("Content-Type", "application/json")

	// Send the HTTP request using the default HTTP client.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errMessage := fmt.Sprintf("Error sending POST /general_neo4j_query request to aali-db: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}
	defer resp.Body.Close()

	// Read and display the response body.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errMessage := fmt.Sprintf("Error reading response body of POST /general_neo4j_query request from aali-db: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	logging.Log.Debugf(&logging.ContextMap{}, "Knowledge DB GeneralNeo4jQuery response received!")

	// Unmarshal the response body to the appropriate struct.
	var response sharedtypes.GeneralNeo4jQueryOutput
	err = json.Unmarshal(body, &response)
	if err != nil {
		errMessage := fmt.Sprintf("Error unmarshalling JSON data of POST /general_neo4j_query response from aali-db: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	return response.Response
}

// GeneralQuery performs a general query in the KnowledgeDB.
//
// The function returns the query results.
//
// Tags:
//   - @displayName: Query
//
// Parameters:
//   - collectionName: the name of the collection to which the data objects will be added.
//   - maxRetrievalCount: the maximum number of results to be retrieved.
//   - outputFields: the fields to be included in the output.
//   - filters: the filter for the query.
//
// Returns:
//   - databaseResponse: the query results
func GeneralQuery(collectionName string, maxRetrievalCount int, outputFields []string, filters sharedtypes.DbFilters) (databaseResponse []sharedtypes.DbResponse) {
	// get the KnowledgeDB endpoint
	knowledgeDbEndpoint := config.GlobalConfig.KNOWLEDGE_DB_ENDPOINT

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
		errMessage := fmt.Sprintf("Error marshalling JSON data of POST /query request for aali-db: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	// Create a new HTTP request with the JSON data.
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonData))
	if err != nil {
		errMessage := fmt.Sprintf("Error creating POST /query request for aali-db: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	// Set the appropriate content type for the request.
	req.Header.Set("Content-Type", "application/json")

	// Send the HTTP request using the default HTTP client.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errMessage := fmt.Sprintf("Error sending POST /query request to aali-db: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}
	defer resp.Body.Close()

	// Read and display the response body.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errMessage := fmt.Sprintf("Error reading response body of POST /query request from aali-db: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	logging.Log.Debugf(&logging.ContextMap{}, "Knowledge DB GeneralQuery response received!")

	// Unmarshal the response body to the appropriate struct.
	var response queryOutput
	err = json.Unmarshal(body, &response)
	if err != nil {
		errMessage := fmt.Sprintf("Error unmarshalling JSON data of POST /query response from aali-db: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	return response.QueryResult
}

// SimilaritySearch performs a similarity search in the KnowledgeDB.
//
// The function returns the similarity search results.
//
// Tags:
//   - @displayName: Similarity Search (Filtered)
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
	filters sharedtypes.DbFilters,
	minScore float64,
	getLeafNodes bool,
	getSiblings bool,
	getParent bool,
	getChildren bool) (databaseResponse []sharedtypes.DbResponse) {
	// get the KnowledgeDB endpoint
	knowledgeDbEndpoint := config.GlobalConfig.KNOWLEDGE_DB_ENDPOINT

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
		errMessage := fmt.Sprintf("Error marshalling JSON data of POST /similarity_search request for aali-db: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	// Create a new HTTP request with the JSON data.
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonData))
	if err != nil {
		errMessage := fmt.Sprintf("Error creating POST /similarity_search request for aali-db: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	// Set the appropriate content type for the request.
	req.Header.Set("Content-Type", "application/json")

	// Send the HTTP request using the default HTTP client.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errMessage := fmt.Sprintf("Error sending POST /similarity_search request to aali-db: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}
	defer resp.Body.Close()

	// Read and display the response body.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errMessage := fmt.Sprintf("Error reading response body of POST /similarity_search request from aali-db: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	logging.Log.Debugf(&logging.ContextMap{}, "Knowledge DB SimilaritySearch response received!")

	// Unmarshal the response body to the appropriate struct.
	var response similaritySearchOutput
	err = json.Unmarshal(body, &response)
	if err != nil {
		errMessage := fmt.Sprintf("Error unmarshalling JSON data of POST /similarity_search response from aali-db: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	var similarityResults []sharedtypes.DbResponse
	for _, element := range response.SimilarityResult {
		similarityResults = append(similarityResults, element.Data)
	}

	return similarityResults
}

// CreateKeywordsDbFilter creates a keywords filter for the KnowledgeDB.
//
// The function returns the keywords filter.
//
// Tags:
//   - @displayName: Keywords Filter
//
// Parameters:
//   - keywords: the keywords to be used for the filter
//   - needAll: flag to indicate whether all keywords are needed
//
// Returns:
//   - databaseFilter: the keywords filter
func CreateKeywordsDbFilter(keywords []string, needAll bool) (databaseFilter sharedtypes.DbArrayFilter) {
	var keywordsFilters sharedtypes.DbArrayFilter

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
// Tags:
//   - @displayName: Tags Filter
//
// Parameters:
//   - tags: the tags to be used for the filter
//   - needAll: flag to indicate whether all tags are needed
//
// Returns:
//   - databaseFilter: the tags filter
func CreateTagsDbFilter(tags []string, needAll bool) (databaseFilter sharedtypes.DbArrayFilter) {
	var tagsFilters sharedtypes.DbArrayFilter

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
// Tags:
//   - @displayName: Metadata Filter
//
// Parameters:
//   - fieldName: the name of the field
//   - fieldType: the type of the field
//   - filterData: the filter data
//   - needAll: flag to indicate whether all data is needed
//
// Returns:
//   - databaseFilter: the metadata filter
func CreateMetadataDbFilter(fieldName string, fieldType string, filterData []string, needAll bool) (databaseFilter sharedtypes.DbJsonFilter) {
	return createDbJsonFilter(fieldName, fieldType, filterData, needAll)
}

// CreateDbFilter creates a filter for the KnowledgeDB.
//
// The function returns the filter.
//
// Tags:
//   - @displayName: Create Filter
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
	tags sharedtypes.DbArrayFilter,
	keywords sharedtypes.DbArrayFilter,
	metadata []sharedtypes.DbJsonFilter) (databaseFilter sharedtypes.DbFilters) {
	var filters sharedtypes.DbFilters

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

// AddDataRequest sends a request to the add_data endpoint.
//
// Tags:
//   - @displayName: Add Data
//
// Parameters:
//   - collectionName: name of the collection the request is sent to.
//   - data: the data to add.
func AddDataRequest(collectionName string, documentData []sharedtypes.DbData) {
	// Create the AddDataInput object
	requestObject := sharedtypes.DbAddDataInput{
		CollectionName: collectionName,
		Data:           documentData,
	}

	// Create the URL
	url := fmt.Sprintf("%s/%s", config.GlobalConfig.KNOWLEDGE_DB_ENDPOINT, "add_data")

	// Send the HTTP POST request
	var response sharedtypes.DbAddDataOutput
	err, _ := createPayloadAndSendHttpRequest(url, requestObject, &response)
	if err != nil {
		errorMessage := fmt.Sprintf("Error sending request to add_data endpoint: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errorMessage)
		panic(errorMessage)
	}

	logging.Log.Debugf(&logging.ContextMap{}, "Added data to collection: %s \n", collectionName)
}

// CreateCollectionRequest sends a request to the collection endpoint.
//
// Tags:
//   - @displayName: Create Collection
//
// Parameters:
//   - collectionName: the name of the collection to create.
func CreateCollectionRequest(collectionName string) {
	// Create the CreateCollectionInput object
	requestObject := sharedtypes.DbCreateCollectionInput{
		CollectionName: collectionName,
	}

	// Create the URL
	url := fmt.Sprintf("%s/%s", config.GlobalConfig.KNOWLEDGE_DB_ENDPOINT, "create_collection")

	// Send the HTTP POST request
	var response sharedtypes.DbCreateCollectionOutput
	err, statusCode := createPayloadAndSendHttpRequest(url, requestObject, &response)
	if err != nil {
		if statusCode == 409 {
			logging.Log.Warnf(&logging.ContextMap{}, "Collection already exists %s \n", collectionName)
		} else {
			errorMessage := fmt.Sprintf("Error sending request to create_collection endpoint: %v", err)
			logging.Log.Error(&logging.ContextMap{}, errorMessage)
			panic(errorMessage)
		}
	}

	logging.Log.Debugf(&logging.ContextMap{}, "Created collection: %s \n", collectionName)
}
