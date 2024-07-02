package externalfunctions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/ansys/allie-flowkit/pkg/config"
	"github.com/google/uuid"
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
					errMessage := fmt.Sprintf("Error extracting Python code: %v\n", err)
					log.Println(errMessage)
					panic(errMessage)
				} else {

					// Validate the Python code
					valid, warnings, err := validatePythonCode(pythonCode)
					if err != nil {
						errMessage := fmt.Sprintf("Error validating Python code: %v\n", err)
						log.Println(errMessage)
						panic(errMessage)
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
func sendChatRequestNoHistory(data string, chatRequestType string, maxKeywordsSearch uint32, llmHandlerEndpoint string) chan HandlerResponse {
	return sendChatRequest(data, chatRequestType, nil, maxKeywordsSearch, "", llmHandlerEndpoint)
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
func sendChatRequest(data string, chatRequestType string, history []HistoricMessage, maxKeywordsSearch uint32, systemPrompt string, llmHandlerEndpoint string) chan HandlerResponse {
	// Initiate the channels
	requestChannelChat = make(chan []byte, 400)
	responseChannel := make(chan HandlerResponse) // Create a channel for responses

	c := initializeClient(llmHandlerEndpoint)
	go shutdownHandler(c)
	go listener(c, responseChannel)
	go writer(c, requestChannelChat, responseChannel)

	go sendRequest("chat", data, requestChannelChat, chatRequestType, "true", history, maxKeywordsSearch, systemPrompt, responseChannel)

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
func sendEmbeddingsRequest(data string, llmHandlerEndpoint string) chan HandlerResponse {
	// Initiate the channels
	requestChannelEmbeddings = make(chan []byte, 400)
	responseChannel := make(chan HandlerResponse) // Create a channel for responses

	c := initializeClient(llmHandlerEndpoint)
	go shutdownHandler(c)
	go listener(c, responseChannel)
	go writer(c, requestChannelEmbeddings, responseChannel)

	go sendRequest("embeddings", data, requestChannelEmbeddings, "", "", nil, 0, "", responseChannel)

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
		log.Println(errMessage)
		panic(errMessage)
	}

	// Send "testkey" for authentication
	err = c.Write(context.Background(), websocket.MessageText, []byte("testkey"))
	if err != nil {
		errMessage := fmt.Sprintf("failed to send authentication message to allie-llm: %v", err)
		log.Println(errMessage)
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
			log.Println(errMessage)
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
					log.Println("Authentication to LLM was successful.")
					continue
				} else {
					errMessage := fmt.Sprintf("failed to unmarshal message from allie-llm: %v", err)
					log.Println(errMessage)
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
				log.Println(errMessage)
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
						log.Println("Chat response completely received from allie-llm.")
					}
				case "embeddings":
					log.Println("Embeddings received from allie-llm.")
				case "info":
					log.Printf("Info %v: %v\n", response.InstructionGuid, *response.InfoMessage)
				default:
					log.Println("Response with unsupported value for 'Type' property received from allie-llm. Ignoring...")
				}
				// Send the response to the channel
				responseChannel <- response
			}
		default:
			log.Printf("Response with unsupported message type '%v'received from allie-llm. Ignoring...\n", typ)
		}

		// If stopListener is true, stop the listener
		// This will happen when:
		// - the chat response is the last one
		// - the embeddings response is received
		// - an unsupported adapter type is received
		if stopListener {
			log.Println("Stopping listener for allie-llm request.")
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
			log.Println(errMessage)
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
func sendRequest(adapter string, data string, RequestChannel chan []byte, chatRequestType string, dataStream string, history []HistoricMessage, maxKeywordsSearch uint32, systemPrompt string, responseChannel chan HandlerResponse) {
	request := HandlerRequest{
		Adapter:         adapter,
		InstructionGuid: strings.Replace(uuid.New().String(), "-", "", -1),
		Data:            data,
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
			log.Println(errMessage)
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
			log.Println(errMessage)
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
		log.Println(errMessage)
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
	log.Printf("Closing client. Received closing signal: %v\n", sig)

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
			log.Println("Potential errors in Python code...")
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
	acsEndpoint := config.AllieFlowkitConfig.ACS_ENDPOINT
	acsApiKey := config.AllieFlowkitConfig.ACS_API_KEY
	acsApiVersion := config.AllieFlowkitConfig.ACS_API_VERSION

	// define searchedProperties and returnedProperties
	searchedEmbeddedFields := "content_vctr, sourceTitle_lvl1_vctr, sourceTitle_lvl2_vctr, sourceTitle_lvl3_vctr"
	returnedProperties := "physics, sourceTitle_lvl3, sourceURL_lvl3, sourceTitle_lvl2, weight, sourceURL_lvl2, product, content, typeOFasset, version"

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

	log.Printf("filter_data is : %s\n", filterQuery)

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
		log.Println(errMessage)
		panic(errMessage)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		errMessage := fmt.Sprintf("failed to create POST request for ACS: %v", err)
		log.Println(errMessage)
		panic(errMessage)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", acsApiKey)

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errMessage := fmt.Sprintf("failed to send POST request to ACS: %v", err)
		log.Println(errMessage)
		panic(errMessage)
	}
	defer resp.Body.Close()

	// Read and return the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errMessage := fmt.Sprintf("failed to read response body from ACS: %v", err)
		log.Println(errMessage)
		panic(errMessage)
	}

	// conver body to []map[string]interface{}
	respObject := ACSSearchResponseStruct{}
	err = json.Unmarshal(body, &respObject)
	if err != nil {
		errMessage := fmt.Sprintf("failed to unmarshal response body from ACS: %v", err)
		log.Println(errMessage)
		panic(errMessage)
	}

	return respObject.Value
}
