package externalfunctions

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

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
					log.Printf("Error extracting Python code: %v\n", err)
				} else {

					// Validate the Python code
					valid, warnings, err := validatePythonCode(pythonCode)
					if err != nil {
						log.Printf("Error validating Python code: %v\n", err)
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
	go writer(c, requestChannelChat)

	go sendRequest("chat", data, requestChannelChat, chatRequestType, "true", history, maxKeywordsSearch, systemPrompt)

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
	go writer(c, requestChannelEmbeddings)

	go sendRequest("embeddings", data, requestChannelEmbeddings, "", "", nil, 0, "")

	return responseChannel // Return the response channel
}

// initializeClient initializes the LLM Handler client
//
// Returns:
//   - *websocket.Conn: the websocket connection
func initializeClient(llmHandlerEndpoint string) *websocket.Conn {
	defer func() {
		r := recover()
		if r != nil {
			log.Printf("Panic occured in initializeClient: %v\n", r)
		}
	}()

	url := llmHandlerEndpoint

	c, _, err := websocket.Dial(context.Background(), url, nil)
	if err != nil {
		log.Printf("Failed to connect to the websocket: %v\n", err)
	}

	// Send "testkey" for authentication
	err = c.Write(context.Background(), websocket.MessageText, []byte("testkey"))
	if err != nil {
		log.Printf("Failed to send authentication message: %v\n", err)
	}

	return c
}

// listener listens for messages from the LLM Handler
//
// Parameters:
//   - c: the websocket connection
//   - responseChannel: the response channel
func listener(c *websocket.Conn, responseChannel chan HandlerResponse) {
	defer func() {
		r := recover()
		if r != nil {
			log.Printf("Panic occured in listener: %v\n", r)
		}
	}()

	// Close the connection when the function returns
	defer c.Close(websocket.StatusNormalClosure, "")

	// Boolean flag to stop the listener (and close the connection)
	var stopListener bool

	for {
		// By default, stop the listener after receiving a message (most of them will be single messages)
		stopListener = true
		typ, message, err := c.Read(context.Background())
		if err != nil {
			log.Printf("Failed to read from the websocket: %v", err)
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
					log.Printf("Failed to unmarshal the message: %v\n", err)
					log.Printf("Failure message (as string): %v\n", msgAsStr)
					return
				}
			}

			if response.Type == "error" {
				log.Printf("Error in request %v: %v (%v)\n", response.InstructionGuid, response.Error.Code, response.Error.Message)
				return
			} else {
				switch response.Type {
				case "chat":
					if !*(response.IsLast) {
						// If it is not the last message, continue listening
						stopListener = false
					} else {
						// If it is the last message, stop listening
						log.Println("Chat response completely received from LLM.")
					}
				case "embeddings":
					log.Println("Embeddings received from LLM.")
				case "info":
					log.Printf("Info %v: %v\n", response.InstructionGuid, *response.InfoMessage)
					continue
				default:
					log.Println("Not supported adapter type.")
				}
				// Send the response to the channel
				responseChannel <- response
			}
		default:
			log.Printf("Unhandled message type: %v\n", typ)
		}

		// If stopListener is true, stop the listener
		// This will happen when:
		// - the chat response is the last one
		// - the embeddings response is received
		// - an unsupported adapter type is received
		if stopListener {
			log.Println("Stopping listener for LLM Handler request.")
			return
		}
	}
}

// writer writes messages to the LLM Handler
//
// Parameters:
//   - c: the websocket connection
//   - RequestChannel: the request channel
func writer(c *websocket.Conn, RequestChannel chan []byte) {
	defer func() {
		r := recover()
		if r != nil {
			log.Printf("Panic occured in writer: %v\n", r)
		}
	}()
	for {
		requestJSON := <-RequestChannel

		err := c.Write(context.Background(), websocket.MessageBinary, requestJSON)
		if err != nil {
			log.Printf("Failed to send message: %v\n", err)
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
func sendRequest(adapter string, data string, RequestChannel chan []byte, chatRequestType string, dataStream string, history []HistoricMessage, maxKeywordsSearch uint32, systemPrompt string) {
	defer func() {
		r := recover()
		if r != nil {
			log.Printf("Panic occured in SendRequest: %v\n", r)
		}
	}()

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
			log.Println("ChatRequestType is required for chat adapter")
		}
		request.ChatRequestType = chatRequestType

		if dataStream == "" {
			log.Println("DataStream is required for chat adapter")
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
		log.Printf("Failed to marshal the message: %v\n", err)
	}

	RequestChannel <- requestJSON
}

// shutdownHandler handles the shutdown of the LLM Handler
//
// Parameters:
//   - c: the websocket connection
//   - RequestChannel: the request channel
func shutdownHandler(c *websocket.Conn) {
	defer func() {
		r := recover()
		if r != nil {
			i := fmt.Sprintf("%v", r)
			log.Printf("Panic in shutdownHandler: %v\n", i)
		}
	}()
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

func formatTemplate(template string, data map[string]string) string {
	for key, value := range data {
		template = strings.ReplaceAll(template, `{`+key+`}`, value)
	}
	return template
}
