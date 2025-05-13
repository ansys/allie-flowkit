package externalfunctions

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
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
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ansys/aali-sharedtypes/pkg/config"
	"github.com/ansys/aali-sharedtypes/pkg/logging"
	"github.com/ansys/aali-sharedtypes/pkg/sharedtypes"
	"github.com/ansys/allie-flowkit/pkg/privatefunctions/codegeneration"

	"github.com/google/go-github/v56/github"
	"github.com/google/uuid"
	"github.com/tiktoken-go/tokenizer"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"golang.org/x/oauth2"
	"nhooyr.io/websocket"
)

// transferDatafromResponseToStreamChannel transfers the data from the response channel to the stream channel
//
// Parameters:
//   - responseChannel: the response channel
//   - streamChannel: the stream channel
//   - validateCode: the flag to indicate whether the code should be validated
func transferDatafromResponseToStreamChannel(
	responseChannel *chan sharedtypes.HandlerResponse,
	streamChannel *chan string,
	validateCode bool,
	sendTokenCount bool,
	tokenCountEndpoint string,
	previousInputTokenCount int,
	previousOutputTokenCount int,
	tokenCountModelName string,
	jwtToken string,
	userEmail string,
	sendContex bool,
	contex string) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic in transferDatafromResponseToStreamChannel: %v\n", r)
		}
	}()

	// Defer the closing of the channels
	defer close(*responseChannel)
	defer close(*streamChannel)

	// Loop through the response channel
	responseAsStr := ""
	for response := range *responseChannel {
		// Check if the response is an error
		if response.Type == "error" {
			logging.Log.Errorf(&logging.ContextMap{}, "Error in request %v: %v\n", response.InstructionGuid, response.Error.Message)
			// send the error message to the stream channel and exit function
			*streamChannel <- fmt.Sprintf("$&$error$&$:$&$%v$&$", response.Error.Message)
			return
		}

		// append the response to the responseAsStr
		responseAsStr += *response.ChatData

		// send the response to the stream channel
		*streamChannel <- *response.ChatData

		// check for last response
		if *(response.IsLast) {

			finalMessage := ""
			// check for token count
			if sendTokenCount {

				// get the output token count
				outputTokenCount, err := openAiTokenCount(tokenCountModelName, responseAsStr)
				if err != nil {
					logging.Log.Errorf(&logging.ContextMap{}, "Error getting token count: %v\n", err)
					// send the error message to the stream channel and exit function
					*streamChannel <- fmt.Sprintf("$&$error$&$:$&$Error getting token count: %v$&$", err)
				}

				// calculate the total token count
				totalInputTokenCount := previousInputTokenCount
				totalOuputTokenCount := previousOutputTokenCount + outputTokenCount

				// send the token count to the token count endpoint
				err = sendTokenCountToEndpoint(jwtToken, tokenCountEndpoint, totalInputTokenCount, totalOuputTokenCount)
				if err != nil {
					logging.Log.Errorf(&logging.ContextMap{}, "Error sending token count: %v\n", err)
					// send the error message to the stream channel and exit function
					*streamChannel <- fmt.Sprintf("$&$error$&$:$&$Error in updating token count: %v$&$", err)
				} else {
					// append the token count message to the final message
					finalMessage += fmt.Sprintf("$&$input_token_count$&$:$&$%d$&$;$&$output_token_count$&$:$&$%d$&$;", totalInputTokenCount, totalOuputTokenCount)
				}
			}

			// check for contex
			if sendContex {
				// append context to the final message
				finalMessage += fmt.Sprintf("$&$context$&$:$&$%s$&$;", contex)
			}

			// check for code validation
			if validateCode {
				// Extract the code from the response
				pythonCode, err := extractPythonCode(responseAsStr)
				if err != nil {
					logging.Log.Errorf(&logging.ContextMap{}, "Error extracting Python code: %v\n", err)
				} else {

					// Validate the Python code
					valid, warnings, err := validatePythonCode(pythonCode)
					if err != nil {
						logging.Log.Errorf(&logging.ContextMap{}, "Error validating Python code: %v\n", err)
					} else {
						if valid {
							if warnings {
								finalMessage += "$&$code_validation$&$:$&$warning$&$;"
							} else {
								finalMessage += "$&$code_validation$&$:$&$valid$&$;"
							}
						} else {
							finalMessage += "$&$code_validation$&$:$&$invalid$&$;"
						}
					}
				}
			}

			// send the final message to the stream channel
			if finalMessage != "" {
				*streamChannel <- finalMessage
			}

			// exit the function
			return
		}
	}
}

// sendTokenCount sends the token count to the token count endpoint
//
// Parameters:
// - userEmail: the email of the user
// - tokenCountEndpoint: the endpoint to send the token count to
// - inputTokenCount: the number of input tokens
// - ouputTokenCount: the number of output tokens
//
// Returns:
// - err: an error if the request fails
func sendTokenCountToEndpoint(jwtToken string, tokenCountEndpoint string, inputTokenCount int, ouputTokenCount int) (err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("panic in sendTokenCount: %v", r)
		}
	}()

	// verify that endpoint is filled
	if tokenCountEndpoint == "" {
		if config.GlobalConfig.ANSYS_AUTHORIZATION_URL == "" {
			return fmt.Errorf("no token count endpoint provided")
		} else {
			tokenCountEndpoint = config.GlobalConfig.ANSYS_AUTHORIZATION_URL + "/token_usage"
		}
	}

	// Create the request
	requestBody := TokenCountUpdateRequest{
		InputToken:  inputTokenCount,
		OutputToken: ouputTokenCount,
		Platform:    "Eng. Copilot",
	}

	// Convert payload to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %v", err)
	}

	// Create a new HTTP request
	request, err := http.NewRequest("POST", tokenCountEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+jwtToken)

	// Create an HTTP client and make the request
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != 200 {
		return fmt.Errorf("response status unequal 200 (%v) for request '%v' with jwt token '%v'", resp.Status, string(jsonData), jwtToken)
	}

	return nil
}

// sendChatRequestNoHistory sends a chat request to LLM without history
//
// Parameters:
//   - data: the input string
//   - chatRequestType: the chat request type
//   - maxKeywordsSearch: the maximum number of keywords to search for
//   - llmHandlerEndpoint: the LLM Handler endpoint
//   - modelIds: the model IDs
//   - options: the model options
//
// Returns:
//   - chan sharedtypes.HandlerResponse: the response channel
func sendChatRequestNoHistory(data string, chatRequestType string, maxKeywordsSearch uint32, llmHandlerEndpoint string, modelIds []string, options *sharedtypes.ModelOptions) chan sharedtypes.HandlerResponse {
	return sendChatRequest(data, chatRequestType, nil, maxKeywordsSearch, "", llmHandlerEndpoint, modelIds, options, nil)
}

// sendChatRequest sends a chat request to LLM
//
// Parameters:
//   - data: the input string
//   - chatRequestType: the chat request type
//   - history: the conversation history
//   - maxKeywordsSearch: the maximum number of keywords to search for
//   - systemPrompt: the system prompt
//   - llmHandlerEndpoint: the LLM Handler endpoint
//   - modelIds: the model IDs
//   - options: the model options
//
// Returns:
//   - chan sharedtypes.HandlerResponse: the response channel
func sendChatRequest(data string, chatRequestType string, history []sharedtypes.HistoricMessage, maxKeywordsSearch uint32, systemPrompt interface{}, llmHandlerEndpoint string, modelIds []string, options *sharedtypes.ModelOptions, images []string) chan sharedtypes.HandlerResponse {
	// Initiate the channels
	requestChannelChat := make(chan []byte, 400)
	responseChannel := make(chan sharedtypes.HandlerResponse) // Create a channel for responses

	c := initializeClient(llmHandlerEndpoint)
	go shutdownHandler(c)
	go listener(c, responseChannel, false)
	go writer(c, requestChannelChat, responseChannel)
	go sendRequest("chat", data, requestChannelChat, chatRequestType, "true", false, history, maxKeywordsSearch, systemPrompt, responseChannel, modelIds, options, images)

	return responseChannel // Return the response channel
}

// sendChatRequestNoStreaming sends a chat request to LLM without streaming
//
// Parameters:
//   - data: the input string
//   - chatRequestType: the chat request type
//   - history: the conversation history
//   - maxKeywordsSearch: the maximum number of keywords to search for
//   - systemPrompt: the system prompt
//   - llmHandlerEndpoint: the LLM Handler endpoint
//   - modelIds: the model IDs
//   - options: the model options
//
// Returns:
//   - string: the response
func sendChatRequestNoStreaming(data string, chatRequestType string, history []sharedtypes.HistoricMessage, maxKeywordsSearch uint32, systemPrompt string, llmHandlerEndpoint string, modelIds []string, options *sharedtypes.ModelOptions, images []string) string {
	// Initiate the channels
	requestChannelChat := make(chan []byte, 400)
	responseChannel := make(chan sharedtypes.HandlerResponse) // Create a channel for responses

	// Initialize the client, handlers and send the request
	c := initializeClient(llmHandlerEndpoint)
	go shutdownHandler(c)
	go listener(c, responseChannel, true)
	go writer(c, requestChannelChat, responseChannel)
	go sendRequest("chat", data, requestChannelChat, chatRequestType, "false", false, history, maxKeywordsSearch, systemPrompt, responseChannel, modelIds, options, images)

	// receive single answer from the response channel
	response := <-responseChannel

	// check for error
	if response.Type == "error" {
		logging.Log.Errorf(&logging.ContextMap{}, "Error in request %v: %v\n", response.InstructionGuid, response.Error.Message)
		panic(response.Error.Message)
	}

	return *response.ChatData
}

// sendEmbeddingsRequest sends an embeddings request to LLM
//
// Parameters:
//   - data: the input string
//   - llmHandlerEndpoint: the LLM Handler endpoint
//   - getSparseEmbeddings: the flag to indicate whether to get sparse embeddings
//   - modelIds: the model IDs
//
// Returns:
//   - chan sharedtypes.HandlerResponse: the response channel
func sendEmbeddingsRequest(data interface{}, llmHandlerEndpoint string, getSparseEmbeddings bool, modelIds []string) chan sharedtypes.HandlerResponse {
	// Initiate the channels
	requestChannelEmbeddings := make(chan []byte, 400)
	responseChannel := make(chan sharedtypes.HandlerResponse) // Create a channel for responses

	c := initializeClient(llmHandlerEndpoint)
	go shutdownHandler(c)
	go listener(c, responseChannel, false)
	go writer(c, requestChannelEmbeddings, responseChannel)

	go sendRequest("embeddings", data, requestChannelEmbeddings, "", "", getSparseEmbeddings, nil, 0, "", responseChannel, modelIds, nil, nil)
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
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}
	// Disable the read limit
	c.SetReadLimit(-1)

	// Get API key
	apiKey := config.GlobalConfig.LLM_API_KEY

	// Legacy authentication
	if apiKey == "" {
		apiKey = "testkey"
	}

	// Send apikey for authentication
	err = c.Write(context.Background(), websocket.MessageText, []byte(apiKey))
	if err != nil {
		errMessage := fmt.Sprintf("failed to send authentication message to allie-llm: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	return c
}

// listener listens for messages from the LLM Handler
//
// Parameters:
//   - c: the websocket connection
//   - responseChannel: the response channel
func listener(c *websocket.Conn, responseChannel chan sharedtypes.HandlerResponse, singleRequest bool) {

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
			logging.Log.Error(&logging.ContextMap{}, errMessage)
			response := sharedtypes.HandlerResponse{
				Type: "error",
				Error: &sharedtypes.ErrorResponse{
					Code:    4,
					Message: errMessage,
				},
			}
			responseChannel <- response
			return
		}
		switch typ {
		case websocket.MessageText, websocket.MessageBinary:
			var response sharedtypes.HandlerResponse

			err = json.Unmarshal(message, &response)
			if err != nil {
				// Check if it is the authentication message
				msgAsStr := string(message)
				if msgAsStr == "authentication successful" {
					logging.Log.Debugf(&logging.ContextMap{}, "Authentication to LLM was successful.")
					continue
				} else {
					errMessage := fmt.Sprintf("failed to unmarshal message from allie-llm: %v", err)
					logging.Log.Error(&logging.ContextMap{}, errMessage)
					response := sharedtypes.HandlerResponse{
						Type: "error",
						Error: &sharedtypes.ErrorResponse{
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
				logging.Log.Error(&logging.ContextMap{}, errMessage)
				response := sharedtypes.HandlerResponse{
					Type: "error",
					Error: &sharedtypes.ErrorResponse{
						Code:    4,
						Message: errMessage,
					},
				}
				responseChannel <- response
				return
			} else {
				switch response.Type {
				case "chat":
					if !singleRequest && !*(response.IsLast) {
						// If it is not the last message, continue listening
						stopListener = false
					} else {
						// If it is the last message, stop listening
						logging.Log.Debugf(&logging.ContextMap{}, "Chat response completely received from allie-llm.")
					}
				case "embeddings":
					logging.Log.Debugf(&logging.ContextMap{}, "Embeddings received from allie-llm.")
				case "info":
					logging.Log.Infof(&logging.ContextMap{}, "Info %v: %v\n", response.InstructionGuid, *response.InfoMessage)
					stopListener = false
					continue
				default:
					logging.Log.Warn(&logging.ContextMap{}, "Response with unsupported value for 'Type' property received from allie-llm. Ignoring...")
				}
				// Send the response to the channel
				responseChannel <- response
			}
		default:
			logging.Log.Warnf(&logging.ContextMap{}, "Response with unsupported message type '%v'received from allie-llm. Ignoring...\n", typ)
		}

		// If stopListener is true, stop the listener
		// This will happen when:
		// - the chat response is the last one
		// - the embeddings response is received
		// - an unsupported adapter type is received
		if stopListener {
			logging.Log.Debugf(&logging.ContextMap{}, "Stopping listener for allie-llm request.")
			return
		}
	}
}

// writer writes messages to the LLM Handler
//
// Parameters:
//   - c: the websocket connection
//   - RequestChannel: the request channel
func writer(c *websocket.Conn, RequestChannel chan []byte, responseChannel chan sharedtypes.HandlerResponse) {
	defer close(RequestChannel)
	for {
		requestJSON := <-RequestChannel

		err := c.Write(context.Background(), websocket.MessageBinary, requestJSON)
		if err != nil {
			errMessage := fmt.Sprintf("failed to write message to allie-llm: %v", err)
			logging.Log.Error(&logging.ContextMap{}, errMessage)
			response := sharedtypes.HandlerResponse{
				Type: "error",
				Error: &sharedtypes.ErrorResponse{
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
func sendRequest(adapter string, data interface{}, RequestChannel chan []byte, chatRequestType string, dataStream string, getSparseEmbeddings bool, history []sharedtypes.HistoricMessage, maxKeywordsSearch uint32, systemPrompt interface{}, responseChannel chan sharedtypes.HandlerResponse, modelIds []string, options *sharedtypes.ModelOptions, images []string) {
	request := sharedtypes.HandlerRequest{
		Adapter:         adapter,
		InstructionGuid: strings.Replace(uuid.New().String(), "-", "", -1),
		Data:            data,
		Images:          images,
		EmbeddingOptions: sharedtypes.EmbeddingOptions{
			ReturnSparse: &getSparseEmbeddings,
		},
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
			logging.Log.Warn(&logging.ContextMap{}, errMessage)
			response := sharedtypes.HandlerResponse{
				Type: "error",
				Error: &sharedtypes.ErrorResponse{
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
			logging.Log.Warn(&logging.ContextMap{}, errMessage)
			response := sharedtypes.HandlerResponse{
				Type: "error",
				Error: &sharedtypes.ErrorResponse{
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
			request.SystemPrompt = systemPrompt
		}

		if options != nil {
			request.ModelOptions = *options
		}
	}

	requestJSON, err := json.Marshal(request)
	if err != nil {
		errMessage := fmt.Sprintf("failed to marshal request to allie-llm: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		response := sharedtypes.HandlerResponse{
			Type: "error",
			Error: &sharedtypes.ErrorResponse{
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
	logging.Log.Debugf(&logging.ContextMap{}, "Closing client. Received closing signal: %v\n", sig)

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
func createDbArrayFilter(filterData []string, needAll bool) (databaseFilter sharedtypes.DbArrayFilter) {
	return sharedtypes.DbArrayFilter{
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
func createDbJsonFilter(fieldName string, fieldType string, filterData []string, needAll bool) (databaseFilter sharedtypes.DbJsonFilter) {
	return sharedtypes.DbJsonFilter{
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
			logging.Log.Warn(&logging.ContextMap{}, "Potential errors in Python code...")
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
	acsEndpoint string,
	acsApiKey string,
	acsApiVersion string,
	query string,
	embeddedQuery []float32,
	indexName string,
	filter map[string]string,
	topK int,
	isAis bool,
	physics []string) (output []sharedtypes.ACSSearchResponse, err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("panic occured in ansysGPTACSSemanticHybridSearch: %v", r)
		}
	}()

	// Create the URL
	url := fmt.Sprintf("https://%s.search.windows.net/indexes/%s/docs/search?api-version=%s", acsEndpoint, indexName, acsApiVersion)

	// Construct the filter query
	var filterData []string
	if !isAis {
		// create filter query
		for key, value := range filter {
			if value != "" {
				// check for granular-ansysgpt index and append (currently not used)
				// if indexName == "granular-ansysgpt" {
				// 	// append physics filter
				// 	if key == "physics" {
				// 		filterData = append(filterData, fmt.Sprintf("%s eq 'tbd'", key))
				// 		filterData = append(filterData, fmt.Sprintf("%s eq 'n/a'", key))
				// 		filterData = append(filterData, fmt.Sprintf("%s eq 'general'", key))
				// 	}

				// 	// append product filter
				// 	if key == "product" {
				// 		filterData = append(filterData, fmt.Sprintf("%s eq 'tbd'", key))
				// 		filterData = append(filterData, fmt.Sprintf("%s eq 'n/a'", key))
				// 		filterData = append(filterData, fmt.Sprintf("%s eq 'general'", key))
				// 	}
				// }

				// normally append filter
				filterData = append(filterData, fmt.Sprintf("%s eq '%s'", key, value))
			}
		}

		// reset filter query for granular-ansysgpt index
		if indexName == "granular-ansysgpt" {
			filterData = []string{}
		}
	} else {
		for _, value := range physics {
			filterData = append(filterData, fmt.Sprintf("physics eq '%s'", value))
		}
	}

	// special case for Scade ONE
	if len(physics) == 1 && physics[0] == "scade" {
		filterData = append(filterData, "product eq 'scade one'")
	}

	// append with 'n/a' filter
	if indexName == "external-marketing" && len(physics) > 0 {
		filterData = append(filterData, "physics eq 'n/a'")
	}

	// join filter data
	filterQuery := strings.Join(filterData, " or ")
	logging.Log.Debugf(&logging.ContextMap{}, "filter_data is : %s\n", filterQuery)

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

	// Marshal the search request
	requestBody, err := json.Marshal(searchRequest)
	if err != nil {
		errMessage := fmt.Errorf("failed to marshal search request to ACS: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		return nil, errMessage
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		errMessage := fmt.Errorf("failed to create POST request for ACS: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		return nil, errMessage
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", acsApiKey)

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errMessage := fmt.Errorf("failed to send POST request to ACS: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		return nil, errMessage
	}
	defer resp.Body.Close()

	// Read and return the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errMessage := fmt.Errorf("failed to read response body from ACS: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		return nil, errMessage
	}

	// check if the reponse is an error
	if resp.StatusCode != 200 {
		errMessage := fmt.Errorf("error in ACS semantic hybrid search for index %v: %s", indexName, string(body))
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		return nil, errMessage
	}

	// extract and convert the response
	output = extractAndConvertACSResponse(body, indexName)
	for _, item := range output {
		logging.Log.Debugf(&logging.ContextMap{}, "ACS topic returned for index %v: %v\n", indexName, item.SourceTitleLvl2)
	}

	// assign index name to the output
	for i := range output {
		output[i].IndexName = indexName
	}

	return output, nil
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
	case "granular-ansysgpt", "ansysgpt-documentation-2023r2", "ansys-dot-com-marketing", "ibp-app-brief":
		searchedEmbeddedFields = "content_vctr, sourceTitle_lvl1_vctr, sourceTitle_lvl2_vctr, sourceTitle_lvl3_vctr"
		returnedProperties = "token_size, physics, typeOFasset, product, version, weight, content, sourceTitle_lvl2, sourceURL_lvl2, sourceTitle_lvl3, sourceURL_lvl3"
	case "scade-documentation-2023r2":
		searchedEmbeddedFields = "content_vctr, sourceTitle_lvl3_vctr"
		returnedProperties = "token_size, physics, typeOFasset, product, version, weight, content, sourceTitle_lvl2, sourceURL_lvl2, sourceTitle_lvl3, sourceURL_lvl3"
	case "external-marketing":
		searchedEmbeddedFields = "content_vctr, sourceTitle_lvl1_vctr, sourceTitle_lvl2_vctr, sourceTitle_lvl3_vctr"
		returnedProperties = "token_size, physics, typeOFasset, product, industry, application, modelUsed, version, weight, content, sourceTitle_lvl2, sourceURL_lvl2, sourceTitle_lvl3, sourceURL_lvl3"
	case "ansysgpt-alh":
		searchedEmbeddedFields = "contentVector, sourcetitleSAPVector"
		returnedProperties = "token_size, physics, typeOFasset, product, version, weight, content, sourcetitleSAP, sourceURLSAP, sourcetitleDCB, sourceURLDCB"
	case "lsdyna-documentation-r14":
		searchedEmbeddedFields = "contentVector, titleVector"
		returnedProperties = "title, url, token_size, physics, typeOFasset, content, product"
	case "ansysgpt-scbu":
		searchedEmbeddedFields = "contentVector"
		returnedProperties = "token_size, physics, typeOFasset, product, version, weight, content, sourcetitleSAP, sourceURLSAP, sourcetitleDCB, sourceURLDCB"
	case "external-crtech-thermal-desktop":
		searchedEmbeddedFields = "contentVector, sourceTitle_lvl1_vctr, sourceTitle_lvl2_vctr, sourceTitle_lvl3_vctr"
		returnedProperties = "token_size, physics, typeOFasset, product, version, weight, bridge_id, content, sourceTitle_lvl2, sourceURL_lvl2, sourceTitle_lvl3, sourceURL_lvl3"
	case "external-product-documentation-public", "external-product-documentation-public-25r1", "external-learning-hub", "external-release-notes", "external-zemax-websites", "external-scbu-learning-hub":
		searchedEmbeddedFields = "contentVector, sourceTitle_lvl1_vctr, sourceTitle_lvl2_vctr, sourceTitle_lvl3_vctr"
		returnedProperties = "token_size, physics, typeOFasset, product, version, weight, content, sourceTitle_lvl2, sourceURL_lvl2, sourceTitle_lvl3, sourceURL_lvl3"
	case "scbu-data-except-alh":
		searchedEmbeddedFields = "content_vctr, sourceTitle_lvl1_vctr, sourceTitle_lvl2_vctr, sourceTitle_lvl3_vctr"
		returnedProperties = "token_size, physics, typeOFasset, product, index_connection_id, version, weight, content, sourceTitle_lvl2, sourceURL_lvl2, sourceTitle_lvl3, sourceURL_lvl3"
	default:
		errMessage := fmt.Sprintf("Index name not found: %s", indexName)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
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
func extractAndConvertACSResponse(body []byte, indexName string) (output []sharedtypes.ACSSearchResponse) {
	respObject := ACSSearchResponseStruct{}
	switch indexName {

	case "granular-ansysgpt", "ansysgpt-documentation-2023r2", "scade-documentation-2023r2", "ansys-dot-com-marketing", "external-marketing", "ibp-app-brief":
		err := json.Unmarshal(body, &respObject)
		if err != nil {
			errMessage := fmt.Sprintf("failed to unmarshal response body from ACS to ACSSearchResponseStruct: %v", err)
			logging.Log.Error(&logging.ContextMap{}, errMessage)
			panic(errMessage)
		}
		output = respObject.Value

	case "ansysgpt-alh", "ansysgpt-scbu":
		respObjectAlh := ACSSearchResponseStructALH{}
		err := json.Unmarshal(body, &respObjectAlh)
		if err != nil {
			errMessage := fmt.Sprintf("failed to unmarshal response body from ACS to ACSSearchResponseStructALH: %v", err)
			logging.Log.Error(&logging.ContextMap{}, errMessage)
			panic(errMessage)
		}

		for _, item := range respObjectAlh.Value {
			output = append(output, sharedtypes.ACSSearchResponse{
				SourceTitleLvl3:     item.SourcetitleSAP,
				SourceURLLvl3:       item.SourceURLSAP,
				SourceTitleLvl2:     item.SourcetitleSAP,
				SourceURLLvl2:       item.SourceURLSAP,
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
			logging.Log.Error(&logging.ContextMap{}, errMessage)
			panic(errMessage)
		}

		for _, item := range respObjectLsdyna.Value {
			output = append(output, sharedtypes.ACSSearchResponse{
				SourceTitleLvl2:     item.Title,
				SourceURLLvl2:       item.Url,
				SourceTitleLvl3:     item.Title,
				SourceURLLvl3:       item.Url,
				Content:             item.Content,
				TypeOFasset:         item.TypeOFasset,
				Physics:             item.Physics,
				Product:             item.Product,
				TokenSize:           item.TokenSize,
				SearchScore:         item.SearchScore,
				SearchRerankerScore: item.SearchRerankerScore,
			})
		}

	case "external-product-documentation-public", "external-product-documentation-public-25r1", "external-learning-hub", "external-crtech-thermal-desktop", "external-release-notes", "external-zemax-websites", "external-scbu-learning-hub", "scbu-data-except-alh":
		respObjectCrtech := ACSSearchResponseStructCrtech{}
		err := json.Unmarshal(body, &respObjectCrtech)
		if err != nil {
			errMessage := fmt.Sprintf("failed to unmarshal response body from ACS to ACSSearchResponseStructCrtech: %v", err)
			logging.Log.Error(&logging.ContextMap{}, errMessage)
			panic(errMessage)
		}

		for _, item := range respObjectCrtech.Value {
			output = append(output, sharedtypes.ACSSearchResponse{
				SourceTitleLvl2:     item.SourceTitleLvl2,
				SourceURLLvl2:       item.SourceURLLvl2,
				SourceTitleLvl3:     item.SourceTitleLvl3,
				SourceURLLvl3:       item.SourceURLLvl3,
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

	default:
		errMessage := fmt.Sprintf("Index name not found: %s", indexName)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	return output
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
		logging.Log.Debugf(&logging.ContextMap{}, "Github file to extract: %s \n", file)
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
	getKeywords bool, numKeywords uint32) (orderedChildDataObjects []*sharedtypes.DbData, err error) {
	instructionSequenceWaitGroup := &sync.WaitGroup{}
	orderedChildData := make([]*sharedtypes.DbData, 0, len(chunks))

	for idx, chunk := range chunks {
		// Create data child object.
		childData := &sharedtypes.DbData{
			Guid:         uuid.New(),
			DocumentId:   documentId,
			DocumentName: documentPath,
			Text:         chunk,
		}

		// Assing previous and next sibling ids if necessary.
		if idx > 0 {
			orderedChildData[idx-1].NextSiblingId = &childData.Guid
			childData.PreviousSiblingId = &orderedChildData[idx-1].Guid
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
func dataExtractionNewLlmInputChannelItem(data *sharedtypes.DbData, instructionSequenceWaitGroup *sync.WaitGroup, adapter string, chatRequestType string, maxNumberOfKeywords uint32, lock *sync.Mutex) *DataExtractionLLMInputChannelItem {
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
			logging.Log.Warnf(&logging.ContextMap{}, "Text field is empty for document %v \n", instruction.Data.DocumentName)

			// Lower instruction sequence waitgroup counter and update processed instructions counter.
			instruction.InstructionSequenceWaitGroup.Done()
			continue
		}

		// If text field is not empty perform request to LLM Handler depending on adapter type.
		instruction.Lock.Lock()
		switch instruction.Adapter {
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

	logging.Log.Debugf(&logging.ContextMap{}, "LLM Handler Worker stopped.")
}

// dataExtractionProcessBatchEmbeddings processes the data extraction batch embeddings.
//
// Parameters:
//   - documentData: the document data.
//   - maxBatchSize: the max batch size.
//
// Returns:
//   - error: an error if any
func dataExtractionProcessBatchEmbeddings(documentData []*sharedtypes.DbData, maxBatchSize int) error {
	// Remove empty chunks (including root node if applicable)
	nonEmptyDocumentData := make([]*sharedtypes.DbData, 0, len(documentData))
	for _, data := range documentData {
		if data.Text != "" {
			nonEmptyDocumentData = append(nonEmptyDocumentData, data)
		}
	}

	if len(nonEmptyDocumentData) == 0 {
		logging.Log.Error(&logging.ContextMap{}, "error in dataExtractionProcessBatchEmbeddings: documentData slice is empty")
		return fmt.Errorf("error in dataExtractionProcessBatchEmbeddings: documentData slice is empty")
	}

	// Process data in batches
	for i := 0; i < len(nonEmptyDocumentData); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(nonEmptyDocumentData) {
			end = len(nonEmptyDocumentData)
		}

		// Create a batch of data to send to LLM handler
		batchData := nonEmptyDocumentData[i:end]
		batchTextToEmbed := make([]string, len(batchData))
		for j, data := range batchData {
			batchTextToEmbed[j] = data.Text
		}

		// Perform vector embedding request to LLM handler
		batchEmbeddings, _, err := llmHandlerPerformVectorEmbeddingRequest(batchTextToEmbed, false)
		if err != nil {
			return fmt.Errorf("failed to perform vector embedding request: %w", err)
		}

		// Update document data with embeddings
		for j, embeddings := range batchEmbeddings {
			batchData[j].Embedding = embeddings
		}
	}

	return nil
}

// llmHandlerPerformVectorEmbeddingRequest performs a vector embedding request to LLM Handler.
//
// Parameters:
//   - input: slice of input strings.
//
// Returns:
//   - embeddedVector: the embedded vectors.
//   - error: an error if any.
func llmHandlerPerformVectorEmbeddingRequest(input []string, sparse bool) (embeddedVectors [][]float32, sparseEmbeddings []map[uint]float32, err error) {
	// get the LLM handler endpoint
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send embeddings request.
	responseChannel := sendEmbeddingsRequest(input, llmHandlerEndpoint, sparse, nil)

	// Process the first response and close the channel.
	embeddedVectors = make([][]float32, len(input))
	sparseEmbeddings = make([]map[uint]float32, len(input))
	for response := range responseChannel {
		// Check if the response is an error.
		if response.Type == "error" {
			return nil, nil, fmt.Errorf("error in vector embedding request %v: %v (%v)", response.InstructionGuid, response.Error.Code, response.Error.Message)
		}

		// Check if the response is an info message.
		if response.Type == "info" {
			logging.Log.Infof(&logging.ContextMap{}, "Received info message for batch embedding request: %v: %v", response.InstructionGuid, response.InfoMessage)
			continue
		}

		// Get embedded vector array
		interfaceArray, ok := response.EmbeddedData.([]interface{})
		if !ok {
			return nil, nil, fmt.Errorf("error converting embedded data to interface array")
		}
		for i, interfaceArrayElement := range interfaceArray {
			lowerInterfaceArray, ok := interfaceArrayElement.([]interface{})
			if !ok {
				return nil, nil, fmt.Errorf("error converting embedded data to interface array")
			}
			embedding32, err := convertToFloat32Slice(lowerInterfaceArray)
			if err != nil {
				return nil, nil, err
			}
			embeddedVectors[i] = embedding32
		}

		// If sparse embeddings are requested, get the sparse embeddings.
		if sparse {
			// Assert that response.LexicalWeights is []interface{}
			lexicalWeights, ok := response.LexicalWeights.([]interface{})
			if !ok {
				return nil, nil, fmt.Errorf("error converting lexical weights to []interface{}, got %T", response.LexicalWeights)
			}

			// Convert []interface{} to []map[uint]float32
			sparseEmbeddings = make([]map[uint]float32, len(lexicalWeights))
			for i, lw := range lexicalWeights {
				// Assert each element is map[string]interface{}
				rawMap, ok := lw.(map[string]interface{})
				if !ok {
					return nil, nil, fmt.Errorf("error converting lexical weight to map[string]interface{}, got %T", lw)
				}

				// Convert map[string]interface{} to map[uint]float32
				convertedMap := make(map[uint]float32)
				for key, value := range rawMap {
					// Convert key from string to uint
					keyUint, err := strconv.ParseUint(key, 10, 32)
					if err != nil {
						return nil, nil, fmt.Errorf("error converting key to uint, got %v", key)
					}

					// Assert value is float64 (common type for numbers in JSON)
					floatValue, ok := value.(float64)
					if !ok {
						return nil, nil, fmt.Errorf("error converting value to float64, got %T", value)
					}

					// Convert float64 to float32
					convertedMap[uint(keyUint)] = float32(floatValue)
				}

				// Add converted map to sparseEmbeddings
				sparseEmbeddings[i] = convertedMap
			}
		}

		// Mark that the first response has been received.
		firstResponseReceived := true

		// Exit the loop after processing the first response.
		if firstResponseReceived {
			break
		}
	}

	// Close the response channel
	close(responseChannel)

	return embeddedVectors, sparseEmbeddings, nil
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
	responseChannel := sendChatRequestNoHistory(input, "summary", 1, llmHandlerEndpoint, nil, nil)

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

	logging.Log.Debugf(&logging.ContextMap{}, "Received summary response.")

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
func performGeneralRequest(input string, history []sharedtypes.HistoricMessage, isStream bool, systemPrompt string, options *sharedtypes.ModelOptions) (message string, stream *chan string, err error) {
	// get the LLM handler endpoint.
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send chat request.
	responseChannel := sendChatRequest(input, "general", history, 0, systemPrompt, llmHandlerEndpoint, nil, options, nil)

	// If isStream is true, create a stream channel and return asap.
	if isStream {
		// Create a stream channel
		streamChannel := make(chan string, 400)

		// Start a goroutine to transfer the data from the response channel to the stream channel.
		go transferDatafromResponseToStreamChannel(&responseChannel, &streamChannel, false, false, "", 0, 0, "", "", "", false, "")

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

		if response.Type == "info" {
			logging.Log.Infof(&logging.ContextMap{}, "Received info message for general llm request: %v: %v", response.InstructionGuid, response.InfoMessage)
			continue
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
	responseChannel := sendChatRequestNoHistory(input, "keywords", numKeywords, llmHandlerEndpoint, nil, nil)

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

	logging.Log.Debugf(&logging.ContextMap{}, "Received keywords response.")

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
	client := &http.Client{}

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

// convertToFloat32Slice converts an interface slice to a float32 slice.
//
// Parameters:
//   - interfaceSlice: the interface slice.
//
// Returns:
//   - float32Slice: the float32 slice.
//   - error: an error if any.
func convertToFloat32Slice(interfaceSlice []interface{}) ([]float32, error) {
	float32Slice := make([]float32, 0, len(interfaceSlice))
	for _, v := range interfaceSlice {
		// Type assertion to float32
		f, ok := v.(float64)
		if !ok {
			// Type assertion failed, return an error
			return nil, fmt.Errorf("value %v is not of type float64", v)
		}
		// convert the float64 to float32
		f32 := float32(f)
		// Append the float32 value to the slice
		float32Slice = append(float32Slice, f32)
	}
	return float32Slice, nil
}

// createPayloadAndSendHttpRequest creates a JSON payload and sends an HTTP POST request.
//
// Parameters:
//   - url: the URL to send the request to.
//   - requestObject: the object to send in the request body.
//   - responsePtr: a pointer to the object to store the response in.
//
// Returns:
//   - error: the error returned by the function.
func createPayloadAndSendHttpRequest(url string, requestObject interface{}, responsePtr interface{}) (funcError error, statusCode int) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic in CreatePayloadAndSendHttpRequest: %v", r)
			funcError = r.(error)
			return
		}
	}()

	// Define the JSON payload.
	jsonPayload, err := json.Marshal(requestObject)
	if err != nil {
		return fmt.Errorf("error marshalling JSON payload: %v", err), 0
	}

	// Create a new HTTP POST request.
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err, 0
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	// Send the HTTP POST request.
	resp, err := client.Do(req)
	if err != nil {
		return err, resp.StatusCode
	}
	defer resp.Body.Close()

	// Decode the JSON response body into the 'data' struct.
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(responsePtr); err != nil {
		return err, 0
	}

	// Check the response status code.
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(resp.Status), resp.StatusCode
	}

	return nil, 0
}

// openAiTokenCount returns the number of tokens in a message for a given model.
//
// Parameters:
//   - modelName: the model name.
//   - message: the message.
//
// Returns:
//   - int: the number of tokens.
//   - error: an error if any.
func openAiTokenCount(modelName string, message string) (int, error) {
	// get model from model name
	var encoding tokenizer.Encoding
	switch modelName {
	case "gpt-4-turbo", "gpt-4", "gpt-3.5-turbo":
		encoding = tokenizer.Cl100kBase
	case "gpt-4o", "gpt-4o-mini":
		encoding = tokenizer.O200kBase
	default:
		return 0, fmt.Errorf("model %s not found", modelName)
	}

	// Load the tokenizer for the specified model
	tokenizer, err := tokenizer.Get(encoding)
	if err != nil {
		return 0, fmt.Errorf("failed to load tokenizer for model %s: %w", modelName, err)
	}

	// Tokenize the message
	tokens, _, err := tokenizer.Encode(message)
	if err != nil {
		return 0, fmt.Errorf("failed to tokenize message: %w", err)
	}

	// Return the number of tokens
	return len(tokens), nil
}

// codeGenerationProcessBatchEmbeddings processes the data extraction batch embeddings.
//
// Parameters:
//   - documentData: the document data.
//   - maxBatchSize: the max batch size.
//
// Returns:
//   - error: an error if any
func codeGenerationProcessBatchEmbeddings(elements []sharedtypes.CodeGenerationElement, maxBatchSize int) (elementEmbeddings [][]float32, err error) {
	// Process data in batches
	for i := 0; i < len(elements); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(elements) {
			end = len(elements)
		}

		// Create a batch of data to send to LLM handler
		batchData := elements[i:end]
		batchTextToEmbed := make([]string, len(batchData))
		for j, data := range batchData {
			batchTextToEmbed[j] = "Name: " + data.Name + "\nDescription: " + data.Description
		}

		// Perform vector embedding request to LLM handler
		batchEmbeddings, _, err := llmHandlerPerformVectorEmbeddingRequest(batchTextToEmbed, false)
		if err != nil {
			return nil, fmt.Errorf("failed to perform vector embedding request: %w", err)
		}

		// Add the embeddings to the list
		elementEmbeddings = append(elementEmbeddings, batchEmbeddings...)
	}

	logging.Log.Debugf(&logging.ContextMap{}, "Processed %d embeddings", len(elements))

	return elementEmbeddings, nil
}

// codeGenerationProcessHybridSearchEmbeddings processes the data extraction batch embeddings.
//
// Parameters:
//   - elements: the elements.
//   - maxBatchSize: the max batch size.
//
// Returns:
//   - error: an error if any
func codeGenerationProcessHybridSearchEmbeddings(elements []sharedtypes.CodeGenerationElement, maxBatchSize int) (denseEmbeddings [][]float32, lexicalWeights []map[uint]float32, err error) {
	processedEmbeddings := 0

	// Process data in batches
	for i := 0; i < len(elements); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(elements) {
			end = len(elements)
		}

		// Create a batch of data to send to LLM handler
		batchData := elements[i:end]
		batchTextToEmbed := make([]string, len(batchData))
		for j, data := range batchData {
			batchTextToEmbed[j] = data.NameFormatted + "\n" + data.NamePseudocode + "\n" + data.Summary + "\n" + strings.Join(data.Dependencies, " ") + "\n" + strings.Join(data.Dependencies, ".")
		}

		// Send http request
		batchDenseEmbeddings, batchLexicalWeights, err := llmHandlerPerformVectorEmbeddingRequest(batchTextToEmbed, true)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to perform vector embedding request: %w", err)
		}

		// Add the embeddings to the list
		denseEmbeddings = append(denseEmbeddings, batchDenseEmbeddings...)
		lexicalWeights = append(lexicalWeights, batchLexicalWeights...)

		processedEmbeddings += len(batchData)
		logging.Log.Debugf(&logging.ContextMap{}, "Processed %d embeddings", processedEmbeddings)
	}

	return denseEmbeddings, lexicalWeights, nil
}

// codeGenerationProcessHybridSearchEmbeddings processes the data extraction batch embeddings.
//
// Parameters:
//   - elements: the elements.
//   - maxBatchSize: the max batch size.
//
// Returns:
//   - error: an error if any
func codeGenerationProcessHybridSearchEmbeddingsForExamples(elements []codegeneration.VectorDatabaseExample, maxBatchSize int) (denseEmbeddings [][]float32, lexicalWeights []map[uint]float32, err error) {
	processedEmbeddings := 0

	// Process data in batches
	for i := 0; i < len(elements); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(elements) {
			end = len(elements)
		}

		// Create a batch of data to send to LLM handler
		batchData := elements[i:end]
		batchTextToEmbed := make([]string, len(batchData))
		for j, data := range batchData {
			batchTextToEmbed[j] = data.DocumentName + "\n" + data.Text
		}

		// Send http request
		batchDenseEmbeddings, batchLexicalWeights, err := llmHandlerPerformVectorEmbeddingRequest(batchTextToEmbed, true)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to perform vector embedding request: %w", err)
		}

		// Add the embeddings to the list
		denseEmbeddings = append(denseEmbeddings, batchDenseEmbeddings...)
		lexicalWeights = append(lexicalWeights, batchLexicalWeights...)

		processedEmbeddings += len(batchData)
		logging.Log.Debugf(&logging.ContextMap{}, "Processed %d embeddings", processedEmbeddings)
	}

	return denseEmbeddings, lexicalWeights, nil
}

func codeGenerationProcessHybridSearchEmbeddingsForUserGuideSections(sections []codegeneration.VectorDatabaseUserGuideSection, maxBatchSize int) (denseEmbeddings [][]float32, lexicalWeights []map[uint]float32, err error) {
	processedEmbeddings := 0

	// Process data in batches
	for i := 0; i < len(sections); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(sections) {
			end = len(sections)
		}

		// Create a batch of data to send to LLM handler
		batchData := sections[i:end]
		batchTextToEmbed := make([]string, len(batchData))
		for j, data := range batchData {
			batchTextToEmbed[j] = data.Title + "\n" + data.Text
		}

		// Send embedding request
		batchDenseEmbeddings, batchLexicalWeights, err := llmHandlerPerformVectorEmbeddingRequest(batchTextToEmbed, true)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to perform vector embedding request: %w", err)
		}

		// Add the embeddings to the list
		denseEmbeddings = append(denseEmbeddings, batchDenseEmbeddings...)
		lexicalWeights = append(lexicalWeights, batchLexicalWeights...)

		processedEmbeddings += len(batchData)
		logging.Log.Debugf(&logging.ContextMap{}, "Processed %d embeddings", processedEmbeddings)
	}

	return denseEmbeddings, lexicalWeights, nil
}

type pythonEmbeddingRequest struct {
	Passages          []string `json:"passages"`
	ReturnDense       bool     `json:"return_dense"`
	ReturnSparse      bool     `json:"return_sparse"`
	ReturnColbertVecs bool     `json:"return_colbert_vecs"`
	IsDocument        bool     `json:"is_document"`
}

type pythonEmbeddingResponse struct {
	ColbertVecs    [][][]float32      `json:"colbert_vecs"`
	LexicalWeights []map[uint]float32 `json:"lexical_weights"`
	DenseVecs      [][]float32        `json:"dense_vecs"`
}

func CreateEmbeddings(dense bool, sparse bool, colbert bool, isDocument bool, passages []string) (dense_vector [][]float32, lexical_weights []map[uint]float32, colbert_vecs [][][]float32, func_error error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic occured in CreateEmbeddings: %v", r)
			func_error = r.(error)
		}
	}()

	// create embeddings
	url := "http://localhost:8000/embedding"

	request := pythonEmbeddingRequest{
		Passages:          passages,
		ReturnDense:       dense,
		ReturnSparse:      sparse,
		ReturnColbertVecs: colbert,
		IsDocument:        isDocument,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error marshalling request: %v", err)
		return nil, nil, nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error sending request to python helper server extract-text: %v", err)
		return nil, nil, nil, err
	}
	defer resp.Body.Close()

	// read response
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)

	responseBody := buf.String()

	// parse response
	var response pythonEmbeddingResponse
	err = json.Unmarshal([]byte(responseBody), &response)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error unmarshalling response: %v", err)
		return nil, nil, nil, err
	}

	return response.DenseVecs, response.LexicalWeights, response.ColbertVecs, nil

}

func dataExtractionTextSplitter(input string, chunkSize int, chunkOverlap int) (chunks []string, err error) {
	var splittedChunks []schema.Document

	// Creating a reader from the content of the file.
	reader := bytes.NewReader([]byte(input))

	// Creating a splitter with the chunk size and overlap specified in the config file.
	splitterOptions := []textsplitter.Option{}
	splitterOptions = append(splitterOptions, textsplitter.WithChunkSize(chunkSize))
	splitterOptions = append(splitterOptions, textsplitter.WithChunkOverlap(chunkOverlap))
	splitter := textsplitter.NewTokenSplitter(splitterOptions...)

	txtLoader := documentloaders.NewText(reader)
	splittedChunks, err = txtLoader.LoadAndSplit(context.Background(), splitter)
	if err != nil {
		errMessage := fmt.Sprintf("Error getting file content from github: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		return nil, err
	}

	for _, chunk := range splittedChunks {
		chunks = append(chunks, chunk.PageContent)
	}

	return chunks, err
}

// getLocalFileContent reads local file and returns checksum and content.
//
// Parameters:
//   - localFilePath: path to file.
//
// Returns:
//   - checksum: checksum of file.
//   - content: content of file.
//   - error: error if any.
func getLocalFileContent(localFilePath string) (checksum string, content []byte, err error) {
	// Read file from local path.
	content, err = os.ReadFile(localFilePath)
	if err != nil {
		errMessage := fmt.Sprintf("Error getting local file content: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		return "", nil, err
	}

	// Calculate checksum from file content.
	hash := sha256.New()
	_, err = hash.Write(content)
	if err != nil {
		errMessage := fmt.Sprintf("Error getting local file content: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		return "", nil, err
	}

	// Convert checksum to a hexadecimal string.
	checksum = hex.EncodeToString(hash.Sum(nil))

	logging.Log.Debugf(&logging.ContextMap{}, "Got content from local file: %s", localFilePath)

	return checksum, content, err
}

// updateMeshPilotActionProperty update action property by given key-value pair.
//
// Parameters:
//   - list: list of actions
//   - findKey: key for search
//   - findValue: value for search
//   - assignKey: need update for a key
//   - assignValue: assign value to a assignKey
//
// Returns:
//   - updated: boolean returns true if updated else false.
func updateMeshPilotActionProperty(list []map[string]string, findKey, findValue, assignKey, assignValue string) (updated bool) {
	updated = false
	for _, item := range list {
		if v, ok := item[findKey]; ok && v == findValue {
			item[assignKey] = assignValue
			updated = true
			return
		}
	}
	return
}

// getIndexNameFromToolName index name by tool
//
// Parameters:
//   - toolName: path to file.
//
// Returns:
//   - indexName: index name.
//   - err: error if any.
func getIndexNameFromToolName(toolName string) (indexName string, err error) {
	err = nil
	if toolName == "ExecuteUserSelectedSolution" {
		indexName = "state_description_embeddings"
	} else if toolName == "ExplainExecutionOfUserSelectedSolution" {
		indexName = "state_description_embeddings"
	} else if toolName == "Delete" {
		indexName = "delete_description_embeddings"
	} else if toolName == "CreateOrInsertOrAdd" {
		indexName = "insert_description_embeddings"
	} else if toolName == "SetOrUpdate" {
		indexName = "update_description_embeddings"
	} else if toolName == "Execute" {
		indexName = "execute_description_embeddings"
	} else if toolName == "Revert" {
		indexName = "revert_description_embeddings"
	} else if toolName == "Connect" {
		indexName = "connect_description_embeddings"
	} else {
		err = fmt.Errorf("invalid toolName: %s", toolName)
	}

	return
}

// downloadGithubFileContent downloads file content from github and returns checksum and content.
//
// Parameters:
//   - githubRepoName: name of the github repository.
//   - githubRepoOwner: owner of the github repository.
//   - githubRepoBranch: branch of the github repository.
//   - gihubFilePath: path to file in the github repository.
//   - githubAccessToken: access token for github.
//
// Returns:
//   - checksum: checksum of file.
//   - content: content of file.
func downloadGithubFileContent(githubRepoName string, githubRepoOwner string,
	githubRepoBranch string, gihubFilePath string, githubAccessToken string) (checksum string, content []byte, err error) {

	// Create a new GitHub client and context.
	client, ctx := dataExtractNewGithubClient(githubAccessToken)

	// Retrieve the file content from the GitHub repository.
	fileContent, _, _, err := client.Repositories.GetContents(ctx, githubRepoOwner, githubRepoName, gihubFilePath, &github.RepositoryContentGetOptions{Ref: githubRepoBranch})
	if err != nil {
		errMessage := fmt.Sprintf("Error getting file content from github file %v: %v", gihubFilePath, err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		return "", nil, err
	}

	// Extract the content from the file content.
	stringContent, err := fileContent.GetContent()
	if err != nil {
		errMessage := fmt.Sprintf("Error getting file content from github file %v: %v", gihubFilePath, err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		return "", nil, err
	}

	// Extract the checksum from the file content.
	checksum = fileContent.GetSHA()

	// Convert the content to a byte slice.
	content = []byte(stringContent)

	logging.Log.Debugf(&logging.ContextMap{}, "Got content from github file: %s", gihubFilePath)

	return checksum, content, nil
}

// mongoDbInitializeClient initializes the mongodb client
// This function should be called at the beginning of the agent
// to initialize the mongodb client
//
// Parameters:
//   - mongoDbEndpoint: The MongoDB endpoint.
//   - databaseName: The name of the database.
//
// Returns:
//   - mongoDbClient: The MongoDB client.
//   - err: An error if any.
func mongoDbInitializeClient(mongoDbEndpoint string, databaseName string, collectionName string) (mongoDbContext *MongoDbContext, err error) {
	// Set the server API options
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(mongoDbEndpoint).SetServerAPIOptions(serverAPI)

	// create context
	mongoDbCtx := context.Background()

	// Create a new client and connect to the server
	client, err := mongo.Connect(mongoDbCtx, opts)
	if err != nil {
		return nil, fmt.Errorf("error in mongo.Connect: %v", err)
	}

	// Ping to verify connection
	err = client.Ping(mongoDbCtx, readpref.Primary())
	if err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %v", err)
	}

	// create database
	database := client.Database(databaseName)

	// check if collection exists
	exists, err := mongoDbCollectionExists(database, collectionName)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error checking if collection exists: %v", err)
		panic(err)
	}
	if !exists {
		logging.Log.Errorf(&logging.ContextMap{}, "Collection %s does not exist", collectionName)
		panic("Collection " + collectionName + " does not exist")
	}

	// create collection
	collection := database.Collection(collectionName)

	// create mongodb client
	mongoDbContext = &MongoDbContext{
		Client:     client,
		Database:   database,
		Collection: collection,
	}

	return mongoDbContext, nil
}

// mongoDbCollectionExists checks if a collection exists in the database
//
// Parameters:
//   - database: The MongoDB database.
//   - collectionName: The name of the collection.
//
// Returns:
//   - exists: True if the collection exists, false otherwise.
//   - err: An error if any.
func mongoDbCollectionExists(database *mongo.Database, collectionName string) (exists bool, err error) {
	// Get the list of collections in the database
	collections, err := database.ListCollectionNames(context.Background(), map[string]interface{}{})
	if err != nil {
		return false, err
	}

	// Check if the collection name exists in the list
	for _, name := range collections {
		if name == collectionName {
			return true, nil
		}
	}

	return false, nil
}

// mongoDbGetCustomerByApiKey retrieves
// the customer object from the database using the API key
//
// Parameters:
//   - mongoDbContext: The MongoDB context.
//   - apiKey: The API key.
//
// Returns:
//   - exists: True if the customer exists, false otherwise.
//   - customer: The customer object.
//   - err: An error if any.
func mongoDbGetCustomerByApiKey(mongoDbContext *MongoDbContext, apiKey string) (exists bool, customer *MongoDbCustomerObject, err error) {
	// Create filter for API key
	filter := bson.M{"api_key": apiKey}

	// Find one document
	err = mongoDbContext.Collection.FindOne(context.Background(), filter).Decode(&customer)
	if err != nil {
		// No matching document found
		if err == mongo.ErrNoDocuments {
			return false, customer, nil
		}
		// other error
		return false, customer, err
	}

	return true, customer, nil
}

// mongoDbGetCreateCustomerByUserId retrieves or creates a customer object by user ID.
//
// Parameters:
//   - mongoDbContext: The MongoDB context.
//   - userId: The user ID.
//   - tokenLimitForNewUsers: The token limit for new users.
//
// Returns:
//   - err: An error if any.
func mongoDbGetCreateCustomerByUserId(mongoDbContext *MongoDbContext, userId string, tokenLimitForNewUsers int) (existingUser bool, customer *MongoDbCustomerObjectDisco, err error) {
	// Create filter for API key
	filter := bson.M{"user_id": userId}

	// Find one document
	existingUser = true
	err = mongoDbContext.Collection.FindOne(context.Background(), filter).Decode(&customer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// No matching document found
			existingUser = false
		} else {
			// other error
			return false, customer, err
		}
	}

	// if customer does not exist, create it
	if !existingUser {
		customer = &MongoDbCustomerObjectDisco{
			UserId:          userId,
			AccessDenied:    false,
			TotalTokenCount: 0,
			TokenLimit:      tokenLimitForNewUsers,
			WarningSent:     false,
		}

		// Insert the new customer document
		_, err = mongoDbContext.Collection.InsertOne(context.Background(), customer)
		if err != nil {
			return false, customer, fmt.Errorf("failed to insert new customer: %v", err)
		}
	}

	return existingUser, customer, nil
}

// mongoDbAddToTotalTokenCount increments the total token count for a customer
//
// Parameters:
//   - mongoDbContext: The MongoDB context.
//   - apiKey: The API key.
//   - additionalTokenCount: The number of tokens to add.
//
// Returns:
//   - err: An error if any.
func mongoDbAddToTotalTokenCount(mongoDbContext *MongoDbContext, indetificationKey string, indetificationValue string, additionalTokenCount int) (err error) {
	// Create filter for API key & update for total token count
	filter := bson.M{indetificationKey: indetificationValue}
	update := bson.M{
		"$inc": bson.M{
			"total_token_usage": additionalTokenCount,
		},
	}

	// Update the document
	result, err := mongoDbContext.Collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return fmt.Errorf("failed to update token usage: %v", err)
	}

	// Check if the document was updated
	if result.MatchedCount == 0 {
		return fmt.Errorf("no customer found with id: %s", indetificationValue)
	}

	return nil
}

// mongoDbUpdateAccessAndWarning updates the access_denied and warning_sent fields
//
// Parameters:
//   - mongoDbContext: The MongoDB context.
//   - apiKey: The API key.
//
// Returns:
//   - err: An error if any.
func mongoDbUpdateAccessAndWarning(mongoDbContext *MongoDbContext, indetificationKey string, indetificationValue string) (err error) {
	// Create filter for API key & update access_denied and warning_sent
	filter := bson.M{indetificationKey: indetificationValue}
	update := bson.M{
		"$set": bson.M{
			"access_denied": true,
			"warning_sent":  true,
		},
	}

	// Update the document
	result, err := mongoDbContext.Collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return fmt.Errorf("failed to update token usage: %v", err)
	}

	// Check if the document was updated
	if result.MatchedCount == 0 {
		return fmt.Errorf("no customer found with id: %s", indetificationValue)
	}

	return nil
}

func logPanic(ctx *logging.ContextMap, msg string, args ...any) {
	errMsg := fmt.Sprintf(msg, args...)
	var logCtx *logging.ContextMap
	if ctx == nil {
		logCtx = &logging.ContextMap{}
	} else {
		logCtx = ctx
	}
	logging.Log.Error(logCtx, errMsg)
	panic(errMsg)
}
