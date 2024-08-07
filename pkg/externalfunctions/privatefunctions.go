package externalfunctions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ansys/allie-flowkit/pkg/internalstates"
	"github.com/ansys/allie-sharedtypes/pkg/config"
	"github.com/ansys/allie-sharedtypes/pkg/logging"
	"github.com/google/go-github/v56/github"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"nhooyr.io/websocket"
)

// transferDatafromResponseToStreamChannel transfers the data from the response channel to the stream channel
//
// Parameters:
//   - responseChannel: the response channel
//   - streamChannel: the stream channel
//   - validateCode: the flag to indicate whether the code should be validated
func transferDatafromResponseToStreamChannel(responseChannel *chan HandlerResponse, streamChannel *chan string, validateCode bool) {
	responseAsStr := ""
	for response := range *responseChannel {
		// Check if the response is an error
		if response.Type == "error" {
			*streamChannel <- response.Error.Message
			break
		}

		// append the response to the responseAsStr
		responseAsStr += *response.ChatData

		// send the response to the stream channel
		*streamChannel <- *response.ChatData

		// check for last response
		if *(response.IsLast) {

			// check for code validation
			if validateCode {
				// Extract the code from the response
				pythonCode, err := extractPythonCode(responseAsStr)
				if err != nil {
					logging.Log.Errorf(internalstates.Ctx, "Error extracting Python code: %v\n", err)
				} else {

					// Validate the Python code
					valid, warnings, err := validatePythonCode(pythonCode)
					if err != nil {
						logging.Log.Errorf(internalstates.Ctx, "Error validating Python code: %v\n", err)
					} else {
						if valid {
							if warnings {
								*streamChannel <- "Code has warnings."
							} else {
								*streamChannel <- "Code is valid."
							}
						} else {
							*streamChannel <- "Code is invalid."
						}
					}
				}
			}

			// exit the loop
			break
		}
	}
	close(*responseChannel)
	close(*streamChannel)
}

// sendChatRequestNoHistory sends a chat request to LLM without history
//
// Parameters:
//   - data: the input string
//   - chatRequestType: the chat request type
//   - sc: the session context
//
// Returns:
//   - chan HandlerResponse: the response channel
func sendChatRequestNoHistory(data string, chatRequestType string, maxKeywordsSearch uint32, llmHandlerEndpoint string, modelIds []string) chan HandlerResponse {
	return sendChatRequest(data, chatRequestType, nil, maxKeywordsSearch, "", llmHandlerEndpoint, modelIds)
}

// sendChatRequest sends a chat request to LLM
//
// Parameters:
//   - data: the input string
//   - chatRequestType: the chat request type
//   - history: the conversation history
//   - sc: the session context
//
// Returns:
//   - chan HandlerResponse: the response channel
func sendChatRequest(data string, chatRequestType string, history []HistoricMessage, maxKeywordsSearch uint32, systemPrompt string, llmHandlerEndpoint string, modelIds []string) chan HandlerResponse {
	// Initiate the channels
	requestChannelChat := make(chan []byte, 400)
	responseChannel := make(chan HandlerResponse) // Create a channel for responses

	c := initializeClient(llmHandlerEndpoint)
	go shutdownHandler(c)
	go listener(c, responseChannel)
	go writer(c, requestChannelChat, responseChannel)

	go sendRequest("chat", data, requestChannelChat, chatRequestType, "true", history, maxKeywordsSearch, systemPrompt, responseChannel, modelIds)

	return responseChannel // Return the response channel
}

// sendEmbeddingsRequest sends an embeddings request to LLM
//
// Parameters:
//   - data: the input string
//   - sc: the session context
//
// Returns:
//   - chan HandlerResponse: the response channel
func sendEmbeddingsRequest(data string, llmHandlerEndpoint string, modelIds []string) chan HandlerResponse {
	// Initiate the channels
	requestChannelEmbeddings := make(chan []byte, 400)
	responseChannel := make(chan HandlerResponse) // Create a channel for responses

	c := initializeClient(llmHandlerEndpoint)
	go shutdownHandler(c)
	go listener(c, responseChannel)
	go writer(c, requestChannelEmbeddings, responseChannel)

	go sendRequest("embeddings", data, requestChannelEmbeddings, "", "", nil, 0, "", responseChannel, modelIds)
	return responseChannel // Return the response channel
}

// initializeClient initializes the LLM Handler client
//
// Returns:
//   - *websocket.Conn: the websocket connection
func initializeClient(llmHandlerEndpoint string) *websocket.Conn {
	url := llmHandlerEndpoint

	c, _, err := websocket.Dial(context.Background(), url, nil)
	if err != nil {
		errMessage := fmt.Sprintf("failed to connect to allie-llm: %v", err)
		logging.Log.Error(internalstates.Ctx, errMessage)
		panic(errMessage)
	}

	// Send "testkey" for authentication
	err = c.Write(context.Background(), websocket.MessageText, []byte("testkey"))
	if err != nil {
		errMessage := fmt.Sprintf("failed to send authentication message to allie-llm: %v", err)
		logging.Log.Error(internalstates.Ctx, errMessage)
		panic(errMessage)
	}

	return c
}

// listener listens for messages from the LLM Handler
//
// Parameters:
//   - c: the websocket connection
//   - responseChannel: the response channel
func listener(c *websocket.Conn, responseChannel chan HandlerResponse) {

	// Close the connection when the function returns
	defer c.Close(websocket.StatusNormalClosure, "")

	// Boolean flag to stop the listener (and close the connection)
	var stopListener bool

	for {
		// By default, stop the listener after receiving a message (most of them will be single messages)
		stopListener = true
		typ, message, err := c.Read(context.Background())
		if err != nil {
			errMessage := fmt.Sprintf("failed to read message from allie-llm: %v", err)
			logging.Log.Error(internalstates.Ctx, errMessage)
			response := HandlerResponse{
				Type: "error",
				Error: &ErrorResponse{
					Code:    4,
					Message: errMessage,
				},
			}
			responseChannel <- response
			return
		}
		switch typ {
		case websocket.MessageText, websocket.MessageBinary:
			var response HandlerResponse

			err = json.Unmarshal(message, &response)
			if err != nil {
				// Check if it is the authentication message
				msgAsStr := string(message)
				if msgAsStr == "authentication successful" {
					logging.Log.Debugf(internalstates.Ctx, "Authentication to LLM was successful.")
					continue
				} else {
					errMessage := fmt.Sprintf("failed to unmarshal message from allie-llm: %v", err)
					logging.Log.Error(internalstates.Ctx, errMessage)
					response := HandlerResponse{
						Type: "error",
						Error: &ErrorResponse{
							Code:    4,
							Message: errMessage,
						},
					}
					responseChannel <- response
					return
				}
			}

			if response.Type == "error" {
				errMessage := fmt.Sprintf("error in request %v: %v (%v)\n", response.InstructionGuid, response.Error.Code, response.Error.Message)
				logging.Log.Error(internalstates.Ctx, errMessage)
				response := HandlerResponse{
					Type: "error",
					Error: &ErrorResponse{
						Code:    4,
						Message: errMessage,
					},
				}
				responseChannel <- response
				return
			} else {
				switch response.Type {
				case "chat":
					if !*(response.IsLast) {
						// If it is not the last message, continue listening
						stopListener = false
					} else {
						// If it is the last message, stop listening
						logging.Log.Debugf(internalstates.Ctx, "Chat response completely received from allie-llm.")
					}
				case "embeddings":
					logging.Log.Debugf(internalstates.Ctx, "Embeddings received from allie-llm.")
				case "info":
					logging.Log.Infof(internalstates.Ctx, "Info %v: %v\n", response.InstructionGuid, *response.InfoMessage)
				default:
					logging.Log.Warn(internalstates.Ctx, "Response with unsupported value for 'Type' property received from allie-llm. Ignoring...")
				}
				// Send the response to the channel
				responseChannel <- response
			}
		default:
			logging.Log.Warnf(internalstates.Ctx, "Response with unsupported message type '%v'received from allie-llm. Ignoring...\n", typ)
		}

		// If stopListener is true, stop the listener
		// This will happen when:
		// - the chat response is the last one
		// - the embeddings response is received
		// - an unsupported adapter type is received
		if stopListener {
			logging.Log.Debugf(internalstates.Ctx, "Stopping listener for allie-llm request.")
			return
		}
	}
}

// writer writes messages to the LLM Handler
//
// Parameters:
//   - c: the websocket connection
//   - RequestChannel: the request channel
func writer(c *websocket.Conn, RequestChannel chan []byte, responseChannel chan HandlerResponse) {
	for {
		requestJSON := <-RequestChannel

		err := c.Write(context.Background(), websocket.MessageBinary, requestJSON)
		if err != nil {
			errMessage := fmt.Sprintf("failed to write message to allie-llm: %v", err)
			logging.Log.Error(internalstates.Ctx, errMessage)
			response := HandlerResponse{
				Type: "error",
				Error: &ErrorResponse{
					Code:    4,
					Message: errMessage,
				},
			}
			responseChannel <- response
			return
		}
	}
}

// sendRequest sends a request to LLM
//
// Parameters:
//   - adapter: the adapter type. Types: "chat", "embeddings"
//   - data: the input string
//   - RequestChannel: the request channel
//   - chatRequestType: the chat request type. Types: "summary", "code", "keywords"
//   - dataStream: the data stream flag
//   - history: the conversation history
//   - sc: the session context
func sendRequest(adapter string, data string, RequestChannel chan []byte, chatRequestType string, dataStream string, history []HistoricMessage, maxKeywordsSearch uint32, systemPrompt string, responseChannel chan HandlerResponse, modelIds []string) {
	request := HandlerRequest{
		Adapter:         adapter,
		InstructionGuid: strings.Replace(uuid.New().String(), "-", "", -1),
		Data:            data,
	}

	// check for modelId
	if len(modelIds) > 0 {
		request.ModelIds = modelIds
	}

	// If history is not empty, set the IsConversation flag to true
	// and set the conversation history
	if len(history) > 0 {
		request.IsConversation = true
		request.ConversationHistory = history
	}

	if adapter == "chat" {
		if chatRequestType == "" {
			errMessage := "Property 'ChatRequestType' is required for 'Adapter' type 'chat' requests to allie-llm."
			logging.Log.Warn(internalstates.Ctx, errMessage)
			response := HandlerResponse{
				Type: "error",
				Error: &ErrorResponse{
					Code:    4,
					Message: errMessage,
				},
			}
			responseChannel <- response
			return
		}
		request.ChatRequestType = chatRequestType

		if dataStream == "" {
			errMessage := "Property 'DataStream' is required for for 'Adapter' type 'chat' requests to allie-llm."
			logging.Log.Warn(internalstates.Ctx, errMessage)
			response := HandlerResponse{
				Type: "error",
				Error: &ErrorResponse{
					Code:    4,
					Message: errMessage,
				},
			}
			responseChannel <- response
			return
		}

		if dataStream == "true" {
			request.DataStream = true
		} else {
			request.DataStream = false
		}

		if request.ChatRequestType == "keywords" {
			request.MaxNumberOfKeywords = maxKeywordsSearch
		}

		if request.ChatRequestType == "general" {
			// Define the system prompt
			request.SystemPrompt = systemPrompt
		}

	}

	requestJSON, err := json.Marshal(request)
	if err != nil {
		errMessage := fmt.Sprintf("failed to marshal request to allie-llm: %v", err)
		logging.Log.Error(internalstates.Ctx, errMessage)
		response := HandlerResponse{
			Type: "error",
			Error: &ErrorResponse{
				Code:    4,
				Message: errMessage,
			},
		}
		responseChannel <- response
		return
	}

	RequestChannel <- requestJSON
}

// shutdownHandler handles the shutdown of the LLM Handler
//
// Parameters:
//   - c: the websocket connection
//   - RequestChannel: the request channel
func shutdownHandler(c *websocket.Conn) {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT)

	sig := <-signalCh
	logging.Log.Debugf(internalstates.Ctx, "Closing client. Received closing signal: %v\n", sig)

	// close connection
	c.Close(websocket.StatusNormalClosure, "Normal Closure")

	os.Exit(0)
}

// createDbArrayFilter creates an array filter for the KnowledgeDB.
//
// The function returns the array filter.
//
// Parameters:
//   - filterData: the filter data
//   - needAll: flag to indicate whether all data is needed
//
// Returns:
//   - databaseFilter: the array filter
func createDbArrayFilter(filterData []string, needAll bool) (databaseFilter DbArrayFilter) {
	return DbArrayFilter{
		NeedAll:    needAll,
		FilterData: filterData,
	}
}

// createDbJsonFilter creates a JSON filter for the KnowledgeDB.
//
// The function returns the JSON filter.
//
// Parameters:
//   - fieldName: the name of the field
//   - fieldType: the type of the field
//   - filterData: the filter data
//   - needAll: flag to indicate whether all data is needed
//
// Returns:
//   - databaseFilter: the JSON filter
func createDbJsonFilter(fieldName string, fieldType string, filterData []string, needAll bool) (databaseFilter DbJsonFilter) {
	return DbJsonFilter{
		FieldName:  fieldName,
		FieldType:  fieldType,
		FilterData: filterData,
		NeedAll:    needAll,
	}
}

// randomNameGenerator generates a random name for the temporary Python script file
//
// Returns:
//   - string: the name of the temporary Python script file
func randomNameGenerator() string {
	// Generate a random number to include in the Python script
	randomNumber := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(1000000)

	// Create a temporary Python script file
	return fmt.Sprintf("temp_python_script_%d.py", randomNumber)
}

// extractPythonCode extracts the Python code from a markdown string. If the
// string does not contain a code block, it is assumed that the string is
// Python code and is returned as is.
//
// Parameters:
//   - markdown: the markdown string
//
// Returns:
//   - string: the Python code
//   - error: error if any
func extractPythonCode(markdown string) (pythonCode string, error error) {
	// Define a regular expression pattern to match Python code blocks
	pattern := "```python\\s*\\n([\\s\\S]*?)\\n\\s*```"

	// Compile the regular expression
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}

	// Find the first match
	match := regex.FindStringSubmatch(markdown)

	if len(match) < 2 {
		// No match found meaning that it is just Python code, presumably
		return markdown, nil
	}

	// Extract and return the Python code
	pythonCode = match[1]
	return pythonCode, nil
}

// validatePythonCode validates the Python code using a Python code analysis tool (pyright)
// and returns several values to indicate the validity of the Python code.
//
// Parameters:
//   - pythonCode: the Python code to be validated
//
// Returns:
//   - bool: true if the Python code is valid, false otherwise
//   - bool: true if the Python code has potential errors, false otherwise
//   - error: an error message if the Python code is invalid
func validatePythonCode(pythonCode string) (bool, bool, error) {
	// Create a temporary Python script file
	tmpFileName := randomNameGenerator()
	file, err := os.Create(tmpFileName)
	if err != nil {
		return false, false, err
	}
	defer func() {
		_ = file.Close()
		_ = os.Remove(tmpFileName) // Delete the temporary file
	}()

	// Write the Python code to the temporary file
	_, err = file.WriteString(pythonCode)
	if err != nil {
		return false, false, err
	}

	// Run a Python code analysis tool (pyright) to check for API validity
	cmd := exec.Command("pyright", tmpFileName)
	output, err := cmd.CombinedOutput()

	// Check if the Python code is valid (no errors in the output)
	if err == nil {
		// Check for potential warnings in output
		outputAsStr := string(output)
		if !strings.Contains(outputAsStr, "0 warnings") {
			logging.Log.Warn(internalstates.Ctx, "Potential errors in Python code...")
			return true, true, nil
		} else {
			return true, false, nil
		}

	} else {
		// If there were errors in the output, return the error message
		return false, false, fmt.Errorf("code validation failed: %s", output)
	}
}

// formatTemplate formats a template string with the given data
//
// Parameters:
//   - template: the template string
//   - data: the data to be used for formatting
//
// Returns:
//   - string: the formatted template string
func formatTemplate(template string, data map[string]string) string {
	for key, value := range data {
		template = strings.ReplaceAll(template, `{`+key+`}`, value)
	}
	return template
}

// ansysGPTACSSemanticHybridSearch performs a semantic hybrid search in ACS
//
// Parameters:
//   - query: the query string
//   - embeddedQuery: the embedded query
//   - indexName: the index name
//   - filter: string build in specific format (https://learn.microsoft.com/en-us/azure/search/search-filters)
//   - filterAfterVectorSearch: the flag to define the filter order (recommended true)
//   - returnedProperties: the properties to be returned
//   - topK: the number of results to be returned from vector search
//   - searchedEmbeddedFields: the ACS fields to be searched
//
// Returns:
//   - output: the search results
func ansysGPTACSSemanticHybridSearch(
	query string,
	embeddedQuery []float32,
	indexName string,
	filter map[string]string,
	topK int) (output []ACSSearchResponse) {

	// get credentials
	acsEndpoint := config.GlobalConfig.ACS_ENDPOINT
	acsApiKey := config.GlobalConfig.ACS_API_KEY
	acsApiVersion := config.GlobalConfig.ACS_API_VERSION

	// Create the URL
	url := fmt.Sprintf("https://%s.search.windows.net/indexes/%s/docs/search?api-version=%s", acsEndpoint, indexName, acsApiVersion)

	// Construct the filter query
	var filterData []string
	for key, value := range filter {
		if value != "" {
			filterData = append(filterData, fmt.Sprintf("%s eq '%s'", key, value))
		}
	}
	filterQuery := strings.Join(filterData, " or ")

	logging.Log.Debugf(internalstates.Ctx, "filter_data is : %s\n", filterQuery)

	// Get the searchedEmbeddedFields and returnedProperties
	searchedEmbeddedFields, returnedProperties := getFieldsAndReturnProperties(indexName)

	// Create the search request payload
	searchRequest := ACSSearchRequest{
		Search: query,
		VectorQueries: []ACSVectorQuery{
			{
				Kind:   "vector",
				K:      30,
				Vector: embeddedQuery,
				Fields: searchedEmbeddedFields,
			},
		},
		VectorFilterMode:      "postFilter",
		Filter:                filterQuery,
		QueryType:             "semantic",
		SemanticConfiguration: "my-semantic-config",
		Top:                   topK,
		Select:                returnedProperties,
		Count:                 true,
	}

	requestBody, err := json.Marshal(searchRequest)
	if err != nil {
		errMessage := fmt.Sprintf("failed to marshal search request to ACS: %v", err)
		logging.Log.Error(internalstates.Ctx, errMessage)
		panic(errMessage)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		errMessage := fmt.Sprintf("failed to create POST request for ACS: %v", err)
		logging.Log.Error(internalstates.Ctx, errMessage)
		panic(errMessage)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", acsApiKey)

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errMessage := fmt.Sprintf("failed to send POST request to ACS: %v", err)
		logging.Log.Error(internalstates.Ctx, errMessage)
		panic(errMessage)
	}
	defer resp.Body.Close()

	// Read and return the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errMessage := fmt.Sprintf("failed to read response body from ACS: %v", err)
		logging.Log.Error(internalstates.Ctx, errMessage)
		panic(errMessage)
	}

	// extract and convert the response
	output = extractAndConvertACSResponse(body, indexName)

	return output
}

// getFieldsAndReturnProperties returns the searchedEmbeddedFields and returnedProperties based on the index name
//
// Parameters:
//   - indexName: the index name
//
// Returns:
//   - searchedEmbeddedFields: the ACS fields to be searched
//   - returnedProperties: the properties to be returned
func getFieldsAndReturnProperties(indexName string) (searchedEmbeddedFields string, returnedProperties string) {
	switch indexName {
	case "granular-ansysgpt", "ansysgpt-documentation-2023r2", "scade-documentation-2023r2", "ansys-dot-com-marketing", "ibp-app-brief":
		searchedEmbeddedFields = "content_vctr, sourceTitle_lvl1_vctr, sourceTitle_lvl2_vctr, sourceTitle_lvl3_vctr"
		returnedProperties = "physics, sourceTitle_lvl3, sourceURL_lvl3, sourceTitle_lvl2, weight, sourceURL_lvl2, product, content, typeOFasset, version"
	case "ansysgpt-alh", "ansysgpt-scbu":
		searchedEmbeddedFields = "contentVector, sourcetitleSAPVector"
		returnedProperties = "token_size, physics, typeOFasset, product, version, weight, content, sourcetitleSAP, sourceURLSAP, sourcetitleDCB, sourceURLDCB"
	case "lsdyna-documentation-r14":
		searchedEmbeddedFields = "contentVector, titleVector"
		returnedProperties = "title, url, token_size, physics, typeOFasset, content, product"
	default:
		errMessage := fmt.Sprintf("Index name not found: %s", indexName)
		logging.Log.Error(internalstates.Ctx, errMessage)
		panic(errMessage)
	}
	return searchedEmbeddedFields, returnedProperties
}

// extractAndConvertACSResponse extracts and converts the ACS response to ACSSearchResponse
//
// Parameters:
//   - body: the response body
//   - indexName: the index name
//
// Returns:
//   - output: the search results
func extractAndConvertACSResponse(body []byte, indexName string) (output []ACSSearchResponse) {
	respObject := ACSSearchResponseStruct{}
	switch indexName {

	case "granular-ansysgpt", "ansysgpt-documentation-2023r2", "scade-documentation-2023r2", "ansys-dot-com-marketing", "ibp-app-brief":
		err := json.Unmarshal(body, &respObject)
		if err != nil {
			errMessage := fmt.Sprintf("failed to unmarshal response body from ACS to ACSSearchResponseStruct: %v", err)
			logging.Log.Error(internalstates.Ctx, errMessage)
			panic(errMessage)
		}

	case "ansysgpt-alh", "ansysgpt-scbu":
		respObjectAlh := ACSSearchResponseStructALH{}
		err := json.Unmarshal(body, &respObjectAlh)
		if err != nil {
			errMessage := fmt.Sprintf("failed to unmarshal response body from ACS to ACSSearchResponseStructALH: %v", err)
			logging.Log.Error(internalstates.Ctx, errMessage)
			panic(errMessage)
		}

		for _, item := range respObjectAlh.Value {
			output = append(output, ACSSearchResponse{
				SourceTitleLvl2:     item.SourcetitleSAP,
				SourceURLLvl2:       item.SourceURLSAP,
				SourceTitleLvl3:     item.SourcetitleDCB,
				SourceURLLvl3:       item.SourceURLDCB,
				Content:             item.Content,
				TypeOFasset:         item.TypeOFasset,
				Physics:             item.Physics,
				Product:             item.Product,
				Version:             item.Version,
				Weight:              item.Weight,
				TokenSize:           item.TokenSize,
				SearchScore:         item.SearchScore,
				SearchRerankerScore: item.SearchRerankerScore,
			})
		}

	case "lsdyna-documentation-r14":
		respObjectLsdyna := ACSSearchResponseStructLSdyna{}
		err := json.Unmarshal(body, &respObjectLsdyna)
		if err != nil {
			errMessage := fmt.Sprintf("failed to unmarshal response body from ACS to ACSSearchResponseStructLSdyna: %v", err)
			logging.Log.Error(internalstates.Ctx, errMessage)
			panic(errMessage)
		}

		for _, item := range respObjectLsdyna.Value {
			output = append(output, ACSSearchResponse{
				SourceTitleLvl2:     item.Title,
				SourceURLLvl2:       item.Url,
				Content:             item.Content,
				TypeOFasset:         item.TypeOFasset,
				Physics:             item.Physics,
				Product:             item.Product,
				TokenSize:           item.TokenSize,
				SearchScore:         item.SearchScore,
				SearchRerankerScore: item.SearchRerankerScore,
			})
		}

	default:
		errMessage := fmt.Sprintf("Index name not found: %s", indexName)
		logging.Log.Error(internalstates.Ctx, errMessage)
		panic(errMessage)
	}

	return respObject.Value
}

// dataExtractionFilterGithubTreeEntries filters the Github tree entries based on the specified filters.
//
// Parameters:
//   - tree: the Github tree.
//   - githubFilteredDirectories: the Github filtered directories.
//   - githubExcludedDirectories: the Github excluded directories.
//   - githubFileExtensions: the Github file extensions.
//
// Returns:
//   - []string: the files to extract.
func dataExtractionFilterGithubTreeEntries(tree *github.Tree, githubFilteredDirectories, githubExcludedDirectories, githubFileExtensions []string) (githubFilesToExtract []string) {
	filteredGithubTreeEntries := []*github.TreeEntry{}

	// Normalize directory paths.
	for i, directory := range githubFilteredDirectories {
		githubFilteredDirectories[i] = strings.ReplaceAll(directory, "\\", "/")
	}
	for i, directory := range githubExcludedDirectories {
		githubExcludedDirectories[i] = strings.ReplaceAll(directory, "\\", "/")
	}

	// If filtered directories are specified, only get files from those directories.
	if len(githubFilteredDirectories) > 0 {
		for _, directory := range githubFilteredDirectories {

			// Check whether excluded directories are in filtered directories.
			excludedDirectoriesInFilteredDirectories := []string{}
			for _, excludedDirectory := range githubExcludedDirectories {
				if strings.HasPrefix(excludedDirectory, directory) {
					excludedDirectoriesInFilteredDirectories = append(excludedDirectoriesInFilteredDirectories, excludedDirectory)
				}
			}

			for _, treeEntry := range tree.Entries {
				if strings.HasPrefix(*treeEntry.Path, directory) {
					// Make sure that the directory is not in the excluded directories.
					isExcludedDirectory := false
					for _, excludedDirectory := range excludedDirectoriesInFilteredDirectories {
						if strings.HasPrefix(*treeEntry.Path, excludedDirectory) {
							isExcludedDirectory = true
							break
						}
					}

					// If directory is in excluded directories, skip it.
					if isExcludedDirectory {
						continue
					}

					// If directory is not in excluded directories, add it to the list of filtered directories.
					filteredGithubTreeEntries = append(filteredGithubTreeEntries, treeEntry)
				}
			}
		}
	}

	// If no filtered directories are specified, get all files and check for excluded directories.
	if len(githubFilteredDirectories) == 0 && len(githubExcludedDirectories) > 0 {
		for _, treeEntry := range tree.Entries {
			// Make sure that the directory is not in the excluded directories.
			isExcludedDirectory := false
			for _, excludedDirectory := range githubExcludedDirectories {
				if strings.HasPrefix(*treeEntry.Path, excludedDirectory) {
					isExcludedDirectory = true
					break
				}
			}

			// If directory is in excluded directories, skip it.
			if isExcludedDirectory {
				continue
			}

			// If directory is not in excluded directories, add it to the list of filtered directories.
			filteredGithubTreeEntries = append(filteredGithubTreeEntries, treeEntry)
		}
	}

	// If no filtered directories are specified and no excluded directories are specified, get all files.
	if len(githubFilteredDirectories) == 0 && len(githubExcludedDirectories) == 0 {
		filteredGithubTreeEntries = tree.Entries
	}

	// Make sure all fileExtensions are lower case.
	for i, fileExtension := range githubFileExtensions {
		githubFileExtensions[i] = strings.ToLower(fileExtension)
	}

	// Make sure only files are in list and filter by file extensions.
	for _, entry := range filteredGithubTreeEntries {
		if *entry.Type == "blob" {

			// Check file extension and add to list if it matches the file extensions in the extraction details.
			if len(githubFileExtensions) > 0 {
				for _, fileExtension := range githubFileExtensions {
					if strings.HasSuffix(strings.ToLower(*entry.Path), fileExtension) {
						githubFilesToExtract = append(githubFilesToExtract, *entry.Path)
					}
				}
			} else {
				// If no file extensions are specified, add all files to the list.
				githubFilesToExtract = append(githubFilesToExtract, *entry.Path)
			}
		}
	}

	for _, file := range githubFilesToExtract {
		logging.Log.Debugf(internalstates.Ctx, "Github file to extract: %s \n", file)
	}

	return githubFilesToExtract
}

// dataExtractNewGithubClient initializes a new GitHub client with the given access token.
//
// Parameters:
//   - githubAccessToken: the GitHub access token.
//
// Returns:
//   - *github.Client: the GitHub client.
//   - context.Context: the context.
func dataExtractNewGithubClient(githubAccessToken string) (client *github.Client, ctx context.Context) {
	ctx = context.Background()

	// Setup OAuth2 token source with the GitHub access token.
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubAccessToken},
	)

	// Create an OAuth2 client with the token source.
	tc := oauth2.NewClient(ctx, ts)

	// Initialize a new GitHub client with the OAuth2 client.
	client = github.NewClient(tc)

	return client, ctx
}

// dataExtractionLocalFilepathExtractWalker is the walker function for the local file extraction.
//
// Parameters:
//   - localPath: the local path.
//   - localFileExtensions: the local file extensions.
//   - localFilteredDirectories: the local filtered directories.
//   - localExcludedDirectories: the local excluded directories.
//   - filesToExtract: the files to extract.
//   - f: the file info
//
// Returns:
//   - error: error that occured during execution.
func dataExtractionLocalFilepathExtractWalker(localPath string, localFileExtensions []string,
	localFilteredDirectories []string, localExcludedDirectories []string, filesToExtract *[]string, f os.FileInfo, err error) error {
	// Initialize to false the skip flag.
	skip := false

	// Make sure that path does not have \.
	localPath = strings.ReplaceAll(localPath, "\\", "/")

	// If path starts with ./ or .\ remove it.
	localPath = strings.TrimPrefix(localPath, "./")
	localPath = strings.TrimPrefix(localPath, ".\\")

	// If filtered directories are specified, only get files from those directories.
	if len(localFilteredDirectories) > 0 {
		isFiltered := false
		for _, directory := range localFilteredDirectories {
			// If directory starts with ./ or .\ remove it.
			directory = strings.TrimPrefix(directory, "./")
			directory = strings.TrimPrefix(directory, ".\\")
			directory = strings.ReplaceAll(directory, "\\", "/")

			if strings.HasPrefix(localPath, directory) || strings.HasPrefix(directory, localPath) {
				// Check whether excluded directories are in filtered directories.
				excludedDirectoriesInFilteredDirectories := []string{}
				for _, excludedDirectory := range localExcludedDirectories {
					if strings.HasPrefix(excludedDirectory, directory) {
						excludedDirectoriesInFilteredDirectories = append(excludedDirectoriesInFilteredDirectories, excludedDirectory)
					}
				}

				// make sure that the directory is not in the excluded directories.
				isExcludedDirectory := false
				for _, excludedDirectory := range excludedDirectoriesInFilteredDirectories {
					excludedDirectory = strings.TrimPrefix(excludedDirectory, "./")
					excludedDirectory = strings.TrimPrefix(excludedDirectory, ".\\")
					excludedDirectory = strings.ReplaceAll(excludedDirectory, "\\", "/")
					if strings.HasPrefix(localPath, excludedDirectory) {
						isExcludedDirectory = true
						break
					}
				}

				if isExcludedDirectory {
					skip = true
					break
				}

				// If directory is not in excluded directories, and it is a filtered directory or contained in a filtered directory, set isFiltered to true.
				isFiltered = true
				break
			}
		}

		// If path is not in filtered directories, skip it.
		if !isFiltered {
			skip = true
		}
	}

	// If no filtered directories are specified, get all files and check for excluded directories.
	if len(localFilteredDirectories) == 0 && len(localExcludedDirectories) > 0 {
		for _, excludedDirectory := range localExcludedDirectories {
			if strings.HasPrefix(localPath, excludedDirectory) {
				skip = true
				break
			}
		}
	}

	// Make sure all fileExtensions are lower case.
	for i, fileExtension := range localFileExtensions {
		localFileExtensions[i] = strings.ToLower(fileExtension)
	}

	// Differentiate between file and directory.
	switch f.IsDir() {
	case true:
		// If a directory, check the skip flag and if true, skip the directory.
		if skip {
			return filepath.SkipDir
		}

	case false:
		// If a file, check file extensions are specified.
		if !skip {
			if len(localFileExtensions) > 0 {
				// Check file extension and add to list if it matches the file extensions in the extraction details.
				for _, fileExtension := range localFileExtensions {
					if strings.HasSuffix(strings.ToLower(localPath), fileExtension) {
						// Create the document details object
						*filesToExtract = append(*filesToExtract, localPath)
						break
					}
				}
			} else {
				// If no file extensions are specified, add all files to the list.
				*filesToExtract = append(*filesToExtract, localPath)
			}
		}
	}

	return nil
}

// dataExtractionDocumentLevelHandler handles the data extraction at document level.
//
// Parameters:
//   - inputChannel: the input channel.
//   - chunks: the document chunks.
//   - documentId: the document ID.
//   - documentPath: the document path.
//   - getSummary: the flag to indicate whether to get the summary.
//   - getKeywords: the flag to indicate whether to get the keywords.
//   - numKeywords: the number of keywords.
//
// Returns:
//   - orderedChildDataObjects: the ordered child data objects.
func dataExtractionDocumentLevelHandler(inputChannel chan *DataExtractionLLMInputChannelItem, errorChannel chan error, chunks []string, documentId string, documentPath string, getSummary bool,
	getKeywords bool, numKeywords uint32) (orderedChildDataObjects []*DataExtractionDocumentData, err error) {
	instructionSequenceWaitGroup := &sync.WaitGroup{}
	orderedChildData := make([]*DataExtractionDocumentData, 0, len(chunks))

	for idx, chunk := range chunks {
		// Create data child object.
		childData := &DataExtractionDocumentData{
			Guid:         "d" + strings.ReplaceAll(uuid.New().String(), "-", ""),
			DocumentId:   documentId,
			DocumentName: documentPath,
			Text:         chunk,
		}

		// Assing previous and next sibling ids if necessary.
		if idx > 0 {
			orderedChildData[idx-1].NextSiblingId = childData.Guid
			childData.PreviousSiblingId = orderedChildData[idx-1].Guid
		}

		orderedChildData = append(orderedChildData, childData)

		if len(childData.Text) > 0 {
			// Create embedding for child.
			embeddingChannelItem := dataExtractionNewLlmInputChannelItem(childData, instructionSequenceWaitGroup, "embeddings", "", 0, &sync.Mutex{})
			instructionSequenceWaitGroup.Add(1)

			// Send embedding request to llm input channel.
			inputChannel <- embeddingChannelItem

			// Create summary for child if enabled.
			if getSummary {
				summaryChannelItem := dataExtractionNewLlmInputChannelItem(childData, instructionSequenceWaitGroup, "chat", "summary", 0, &sync.Mutex{})
				instructionSequenceWaitGroup.Add(1)

				// Send summary request to llm input channel.
				inputChannel <- summaryChannelItem
			}

			// Create keywords for child if enabled.
			if getKeywords {
				keywordsChannelItem := dataExtractionNewLlmInputChannelItem(childData, instructionSequenceWaitGroup, "chat", "keywords", numKeywords, &sync.Mutex{})
				instructionSequenceWaitGroup.Add(1)

				// Send keywords request to llm input channel.
				inputChannel <- keywordsChannelItem
			}
		}
	}

	// Separate goroutine to wait on the wait group and signal completion.
	doneChan := make(chan struct{})
	go func() {
		instructionSequenceWaitGroup.Wait()
		close(doneChan)
	}()

	// Main goroutine to listen on the error channel and done channel.
	for {
		select {
		case err := <-errorChannel:
			return nil, err

		case <-doneChan:
			return orderedChildData, nil
		}
	}
}

// dataExtractionNewLlmInputChannelItem creates a new llm input channel item.
//
// Parameters:
//   - data: data.
//   - instructionSequenceWaitGroup: instruction sequence wait group.
//   - adapter: adapter.
//   - chatRequestType: chat request type.
//   - maxNumberOfKeywords: max number of keywords.
//   - lock: lock.
//
// Returns:
//   - llmInputChannelItem: llm input channel item.
func dataExtractionNewLlmInputChannelItem(data *DataExtractionDocumentData, instructionSequenceWaitGroup *sync.WaitGroup, adapter string, chatRequestType string, maxNumberOfKeywords uint32, lock *sync.Mutex) *DataExtractionLLMInputChannelItem {
	return &DataExtractionLLMInputChannelItem{
		Data:                         data,
		InstructionSequenceWaitGroup: instructionSequenceWaitGroup,
		Adapter:                      adapter,
		ChatRequestType:              chatRequestType,
		MaxNumberOfKeywords:          maxNumberOfKeywords,
		Lock:                         lock,
	}
}

// dataExtractionLLMHandlerWorker is a worker function for the LLM Handler requests during data extraction.
//
// Parameters:
//   - waitgroup: the wait group
//   - inputChannel: the input channel
//   - errorChannel: the error channel
//   - embeddingsDimensions: the embeddings dimensions
//
// Returns:
//   - error: an error if any
func dataExtractionLLMHandlerWorker(waitgroup *sync.WaitGroup, inputChannel chan *DataExtractionLLMInputChannelItem, errorChannel chan error, embeddingsDimensions int) {
	defer waitgroup.Done()
	// Listen to Input Channel
	for instruction := range inputChannel {
		// Check if text field for chunk is empty.
		if instruction.Data.Text == "" {
			logging.Log.Warnf(internalstates.Ctx, "Text field is empty for document %v \n", instruction.Data.DocumentName)

			//If adapter type is embedding, set embedding to empty slice of dimension embeddingsDimensions.
			if instruction.Adapter == "embeddings" {
				instruction.Lock.Lock()
				instruction.Data.Embedding = make([]float32, embeddingsDimensions)
				instruction.Lock.Unlock()
			}

			// Lower instruction sequence waitgroup counter and update processed instructions counter.
			instruction.InstructionSequenceWaitGroup.Done()
			continue
		}

		// If text field is not empty perform request to LLM Handler depending on adapter type.
		instruction.Lock.Lock()
		switch instruction.Adapter {
		case "embeddings":
			res, err := llmHandlerPerformVectorEmbeddingRequest(instruction.Data.Text)
			if err != nil {
				errorChannel <- err
			}
			instruction.Data.Embedding = res
		case "chat":
			if instruction.ChatRequestType == "summary" {
				res, err := llmHandlerPerformSummaryRequest(instruction.Data.Text)
				if err != nil {
					errorChannel <- err
				}
				instruction.Data.Summary = res
			} else if instruction.ChatRequestType == "keywords" {
				res, err := llmHandlerPerformKeywordExtractionRequest(instruction.Data.Text, instruction.MaxNumberOfKeywords)
				if err != nil {
					errorChannel <- err
				}
				instruction.Data.Keywords = res
			}
		}
		instruction.Lock.Unlock()

		// Lower instruction sequence waitgroup counter
		instruction.InstructionSequenceWaitGroup.Done()
	}

	logging.Log.Debugf(internalstates.Ctx, "LLM Handler Worker stopped.")
}

// llmHandlerPerformVectorEmbeddingRequest performs a vector embedding request to LLM Handler.
//
// Parameters:
//   - input: the input string.
//
// Returns:
//   - embeddedVector: the embedded vector.
//   - error: an error if any.
func llmHandlerPerformVectorEmbeddingRequest(input string) (embeddedVector []float32, err error) {
	// get the LLM handler endpoint
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send embeddings request.
	responseChannel := sendEmbeddingsRequest(input, llmHandlerEndpoint, nil)

	// Process the first response and close the channel.
	var embedding32 []float32
	for response := range responseChannel {
		// Check if the response is an error.
		if response.Type == "error" {
			return nil, fmt.Errorf("error in vector embedding request %v: %v (%v)", response.InstructionGuid, response.Error.Code, response.Error.Message)
		}

		// Get embedded vector array.
		embedding32 = response.EmbeddedData

		// Mark that the first response has been received.
		firstResponseReceived := true

		// Exit the loop after processing the first response.
		if firstResponseReceived {
			break
		}
	}

	// Close the response channel
	close(responseChannel)

	return embedding32, nil
}

// llmHandlerPerformSummaryRequest performs a summary request to LLM Handler.
//
// Parameters:
//   - input: the input string.
//
// Returns:
//   - summary: the summary.
//   - error: an error if any.
func llmHandlerPerformSummaryRequest(input string) (summary string, err error) {
	// get the LLM handler endpoint
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send chat request.
	responseChannel := sendChatRequestNoHistory(input, "summary", 1, llmHandlerEndpoint, nil)

	// Process all responses.
	var responseAsStr string
	for response := range responseChannel {
		// Check if the response is an error.
		if response.Type == "error" {
			return "", fmt.Errorf("error in summary request %v: %v (%v)", response.InstructionGuid, response.Error.Code, response.Error.Message)
		}

		// Accumulate the responses.
		responseAsStr += *(response.ChatData)

		// If we are at the last message, break the loop.
		if *(response.IsLast) {
			break
		}
	}

	logging.Log.Debugf(internalstates.Ctx, "Received summary response.")

	// Close the response channel.
	close(responseChannel)

	// Return the response.
	return responseAsStr, nil
}

// performGeneralRequest performs a general chat completion request to LLM.
//
// Parameters:
//   - input: the input string.
//   - history: the conversation history.
//   - isStream: the stream flag.
//   - systemPrompt: the system prompt.
//
// Returns:
//   - message: the generated message.
//   - stream: the stream channel.
//   - err: the error.
func performGeneralRequest(input string, history []HistoricMessage, isStream bool, systemPrompt string) (message string, stream *chan string, err error) {
	// get the LLM handler endpoint.
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send chat request.
	responseChannel := sendChatRequest(input, "general", history, 0, systemPrompt, llmHandlerEndpoint, nil)

	// If isStream is true, create a stream channel and return asap.
	if isStream {
		// Create a stream channel
		streamChannel := make(chan string, 400)

		// Start a goroutine to transfer the data from the response channel to the stream channel.
		go transferDatafromResponseToStreamChannel(&responseChannel, &streamChannel, false)

		// Return the stream channel.
		return "", &streamChannel, nil
	}

	// Process all responses.
	var responseAsStr string
	for response := range responseChannel {
		// Check if the response is an error.
		if response.Type == "error" {
			return "", nil, fmt.Errorf("error in general llm request %v: %v (%v)", response.InstructionGuid, response.Error.Code, response.Error.Message)
		}

		// Accumulate the responses.
		responseAsStr += *(response.ChatData)

		// If we are at the last message, break the loop.
		if *(response.IsLast) {
			break
		}
	}

	// Close the response channel
	close(responseChannel)

	// Return the response
	return responseAsStr, nil, nil
}

// llmHandlerPerformKeywordExtractionRequest performs a keyword extraction request to LLM Handler.
//
// Parameters:
//   - input: the input string.
//   - numKeywords: the number of keywords.
//
// Returns:
//   - keywords: the keywords.
//   - error: an error if any.
func llmHandlerPerformKeywordExtractionRequest(input string, numKeywords uint32) (keywords []string, err error) {
	// get the LLM handler endpoint.
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send chat request.
	responseChannel := sendChatRequestNoHistory(input, "keywords", numKeywords, llmHandlerEndpoint, nil)

	// Process all responses.
	var responseAsStr string
	for response := range responseChannel {
		// Check if the response is an error.
		if response.Type == "error" {
			return nil, fmt.Errorf("error in keyword extraction request %v: %v (%v)", response.InstructionGuid, response.Error.Code, response.Error.Message)
		}

		// Accumulate the responses.
		responseAsStr += *(response.ChatData)

		// If we are at the last message, break the loop.
		if *(response.IsLast) {
			break
		}
	}

	logging.Log.Debugf(internalstates.Ctx, "Received keywords response.")

	// Close the response channel.
	close(responseChannel)

	// Return the response.
	return strings.Split(responseAsStr, ","), nil
}

// dataExtractionPerformSplitterRequest performs a data extraction splitter request to the Python service.
//
// Parameters:
//   - content: the content.
//   - documentType: the document type.
//   - chunkSize: the chunk size.
//   - chunkOverlap: the chunk overlap.
//
// Returns:
//   - output: the output.
//   - error: an error if any.
func dataExtractionPerformSplitterRequest(content []byte, documentType string, chunkSize int, chunkOverlap int) (output []string, err error) {
	// Define the URL and headers.
	url := config.GlobalConfig.FLOWKIT_PYTHON_ENDPOINT + "/splitter/" + documentType
	headers := map[string]string{
		"Content-Type": "application/json",
		"api-key":      config.GlobalConfig.FLOWKIT_PYTHON_API_KEY,
	}
	splitterRequest := DataExtractionSplitterServiceRequest{
		DocumentContent: content,
		ChunkSize:       chunkSize,
		ChunkOverlap:    chunkOverlap,
	}

	// Marshal the request.
	body, err := json.Marshal(splitterRequest)
	if err != nil {
		return nil, err
	}

	// Send the request.
	response, err := httpRequest("POST", url, headers, body)
	if err != nil {
		return nil, err
	}

	// Unmarshall response to  DataExtractionSplitterServiceResponse.
	splitterResponse := DataExtractionSplitterServiceResponse{}
	err = json.Unmarshal(response, &splitterResponse)
	if err != nil {
		return nil, err
	}

	// Return the chunks.
	output = splitterResponse.Chunks

	return output, nil
}

// httpRequest is a general function for making HTTP requests.
//
// Parameters:
//   - method: HTTP method.
//   - url: URL to make the request to.
//   - headers: headers to include in the request.
//   - body: body of the request.
//
// Returns:
//   - response body.
//   - error.
func httpRequest(method string, url string, headers map[string]string, body []byte) ([]byte, error) {
	// Create a new request using http.
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	// Add headers.
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Create a new HTTP client and set timeout.
	client := &http.Client{
		Timeout: time.Second * 30,
	}

	// Send the request.
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body.
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Check if the status code is not 200 OK.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return respBody, fmt.Errorf("HTTP request failed with status code %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
