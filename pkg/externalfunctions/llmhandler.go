package externalfunctions

import (
	"encoding/json"
	"fmt"

	"github.com/ansys/allie-flowkit/pkg/internalstates"
	"github.com/ansys/allie-sharedtypes/pkg/config"
	"github.com/ansys/allie-sharedtypes/pkg/logging"
	"github.com/ansys/allie-sharedtypes/pkg/sharedtypes"
)

// PerformVectorEmbeddingRequest performs a vector embedding request to LLM
//
// Tags:
//   - @displayName: Embeddings
//
// Parameters:
//   - input: the input string
//
// Returns:
//   - embeddedVector: the embedded vector in float32 format
func PerformVectorEmbeddingRequest(input string) (embeddedVector []float32) {
	// get the LLM handler endpoint
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send embeddings request
	responseChannel := sendEmbeddingsRequest(input, llmHandlerEndpoint, nil)

	// Process the first response and close the channel
	var embedding32 []float32
	var err error
	for response := range responseChannel {
		// Check if the response is an error
		if response.Type == "error" {
			panic(response.Error)
		}

		// Log LLM response
		logging.Log.Debugf(internalstates.Ctx, "Received embeddings response.")

		// Get embedded vector array
		interfaceArray, ok := response.EmbeddedData.([]interface{})
		if !ok {
			errMessage := "error converting embedded data to interface array"
			logging.Log.Error(internalstates.Ctx, errMessage)
			panic(errMessage)
		}
		embedding32, err = convertToFloat32Slice(interfaceArray)
		if err != nil {
			errMessage := fmt.Sprintf("error converting embedded data to float32 slice: %v", err)
			logging.Log.Error(internalstates.Ctx, errMessage)
			panic(errMessage)
		}

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

// PerformBatchEmbeddingRequest performs a batch vector embedding request to LLM
//
// Tags:
//   - @displayName: Batch Embeddings
//
// Parameters:
//   - input: the input strings
//
// Returns:
//   - embeddedVectors: the embedded vectors in float32 format
func PerformBatchEmbeddingRequest(input []string) (embeddedVectors [][]float32) {
	// get the LLM handler endpoint
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send embeddings request
	responseChannel := sendEmbeddingsRequest(input, llmHandlerEndpoint, nil)

	// Process the first response and close the channel
	embedding32Array := make([][]float32, len(input))
	for response := range responseChannel {
		// Check if the response is an error
		if response.Type == "error" {
			panic(response.Error)
		}

		// Log LLM response
		logging.Log.Debugf(internalstates.Ctx, "Received batch embeddings response.")

		// Get embedded vector array
		interfaceArray, ok := response.EmbeddedData.([]interface{})
		if !ok {
			errMessage := "error converting embedded data to interface array"
			logging.Log.Error(internalstates.Ctx, errMessage)
			panic(errMessage)
		}

		for i, interfaceArrayElement := range interfaceArray {
			lowerInterfaceArray, ok := interfaceArrayElement.([]interface{})
			if !ok {
				errMessage := "error converting embedded data to interface array"
				logging.Log.Error(internalstates.Ctx, errMessage)
				panic(errMessage)
			}
			embedding32, err := convertToFloat32Slice(lowerInterfaceArray)
			if err != nil {
				errMessage := fmt.Sprintf("error converting embedded data to float32 slice: %v", err)
				logging.Log.Error(internalstates.Ctx, errMessage)
				panic(errMessage)
			}
			embedding32Array[i] = embedding32
		}

		// Mark that the first response has been received
		firstResponseReceived := true

		// Exit the loop after processing the first response
		if firstResponseReceived {
			break
		}
	}

	// Close the response channel
	close(responseChannel)

	return embedding32Array
}

// PerformKeywordExtractionRequest performs a keywords extraction request to LLM
//
// Tags:
//   - @displayName: Keyword Extraction
//
// Parameters:
//   - input: the input string
//   - maxKeywordsSearch: the maximum number of keywords to search for
//
// Returns:
//   - keywords: the keywords extracted from the input string as a slice of strings
func PerformKeywordExtractionRequest(input string, maxKeywordsSearch uint32) (keywords []string) {
	// get the LLM handler endpoint
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send chat request
	responseChannel := sendChatRequestNoHistory(input, "keywords", maxKeywordsSearch, llmHandlerEndpoint, nil, nil)

	// Process all responses
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

	logging.Log.Debugf(internalstates.Ctx, "Received keywords response.")

	// Close the response channel
	close(responseChannel)

	// Unmarshal JSON data into the result variable
	err := json.Unmarshal([]byte(responseAsStr), &keywords)
	if err != nil {
		errMessage := fmt.Sprintf("Error unmarshalling keywords response from allie-llm: %v", err)
		logging.Log.Error(internalstates.Ctx, errMessage)
		panic(errMessage)
	}

	// Return the response
	return keywords
}

// PerformSummaryRequest performs a summary request to LLM
//
// Tags:
//   - @displayName: Summary
//
// Parameters:
//   - input: the input string
//
// Returns:
//   - summary: the summary extracted from the input string
func PerformSummaryRequest(input string) (summary string) {
	// get the LLM handler endpoint
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send chat request
	responseChannel := sendChatRequestNoHistory(input, "summary", 1, llmHandlerEndpoint, nil, nil)

	// Process all responses
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

	logging.Log.Debugf(internalstates.Ctx, "Received summary response.")

	// Close the response channel
	close(responseChannel)

	// Return the response
	return responseAsStr
}

// PerformGeneralRequest performs a general chat completion request to LLM
//
// Tags:
//   - @displayName: General LLM Request
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
func PerformGeneralRequest(input string, history []sharedtypes.HistoricMessage, isStream bool, systemPrompt string) (message string, stream *chan string) {
	// get the LLM handler endpoint
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send chat request
	responseChannel := sendChatRequest(input, "general", history, 0, systemPrompt, llmHandlerEndpoint, nil, nil)
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

// PerformGeneralRequestSpecificModel performs a general request to LLM with a specific model
//
// Tags:
//   - @displayName: General LLM Request (Specific Models)
//
// Parameters:
//   - input: the user input
//   - history: the conversation history
//   - isStream: the flag to indicate whether the response should be streamed
//   - systemPrompt: the system prompt
//   - modelId: the model ID
//
// Returns:
//   - message: the response message
//   - stream: the stream channel
func PerformGeneralRequestSpecificModel(input string, history []sharedtypes.HistoricMessage, isStream bool, systemPrompt string, modelIds []string) (message string, stream *chan string) {
	// get the LLM handler endpoint
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send chat request
	responseChannel := sendChatRequest(input, "general", history, 0, systemPrompt, llmHandlerEndpoint, modelIds, nil)

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

// PerformCodeLLMRequest performs a code generation request to LLM
//
// Tags:
//   - @displayName: Code LLM Request
//
// Parameters:
//   - input: the input string
//   - history: the conversation history
//   - isStream: the stream flag
//
// Returns:
//   - message: the generated code
//   - stream: the stream channel
func PerformCodeLLMRequest(input string, history []sharedtypes.HistoricMessage, isStream bool, validateCode bool) (message string, stream *chan string) {
	// get the LLM handler endpoint
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send chat request
	responseChannel := sendChatRequest(input, "code", history, 0, "", llmHandlerEndpoint, nil, nil)

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

	// Code validation
	if validateCode {

		// Extract the code from the response
		pythonCode, err := extractPythonCode(responseAsStr)
		if err != nil {
			logging.Log.Errorf(internalstates.Ctx, "Error extracting Python code: %v", err)
		} else {

			// Validate the Python code
			valid, warnings, err := validatePythonCode(pythonCode)
			if err != nil {
				logging.Log.Errorf(internalstates.Ctx, "Error validating Python code: %v", err)
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
// Tags:
//   - @displayName: Library Context
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

// BuildFinalQueryForGeneralLLMRequest builds the final query for a general
// request to LLM. The final query is a markdown string that contains the
// original request and the examples from the KnowledgeDB.
//
// Tags:
//   - @displayName: Final Query (General LLM Request)
//
// Parameters:
//   - request: the original request
//   - knowledgedbResponse: the KnowledgeDB response
//
// Returns:
//   - finalQuery: the final query
func BuildFinalQueryForGeneralLLMRequest(request string, knowledgedbResponse []sharedtypes.DbResponse) (finalQuery string) {

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
// Tags:
//   - @displayName: Final Query (Code LLM Request)
//
// Parameters:
//   - request: the original request
//   - knowledgedbResponse: the KnowledgeDB response
//
// Returns:
//   - finalQuery: the final query
func BuildFinalQueryForCodeLLMRequest(request string, knowledgedbResponse []sharedtypes.DbResponse) (finalQuery string) {
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

type AppendMessageHistoryRole string

const (
	user      AppendMessageHistoryRole = "user"
	assistant AppendMessageHistoryRole = "assistant"
	system    AppendMessageHistoryRole = "system"
)

// AppendMessageHistory appends a new message to the conversation history
//
// Tags:
//   - @displayName: Append Message History
//
// Parameters:
//   - newMessage: the new message
//   - role: the role of the message
//   - history: the conversation history
//
// Returns:
//   - updatedHistory: the updated conversation history
func AppendMessageHistory(newMessage string, role AppendMessageHistoryRole, history []sharedtypes.HistoricMessage) (updatedHistory []sharedtypes.HistoricMessage) {
	switch role {
	case user:
	case assistant:
	case system:
	default:
		errMessage := fmt.Sprintf("Invalid role used for 'AppendMessageHistory': %v", role)
		logging.Log.Warn(internalstates.Ctx, errMessage)
		panic(errMessage)
	}

	// skip for empty messages
	if newMessage == "" {
		return history
	}

	// Create a new HistoricMessage
	newMessageHistory := sharedtypes.HistoricMessage{
		Role:    string(role),
		Content: newMessage,
	}

	// Append the new message to the history
	history = append(history, newMessageHistory)

	return history
}

// ShortenMessageHistory shortens the conversation history to a maximum length.
// It will retain only the most recent messages and older messages will be
// removed.
//
// Tags:
//   - @displayName: Shorten History
//
// Parameters:
//   - history: the conversation history
//   - maxLength: the maximum length of the conversation history
//
// Returns:
//   - updatedHistory: the updated conversation history
func ShortenMessageHistory(history []sharedtypes.HistoricMessage, maxLength int) (updatedHistory []sharedtypes.HistoricMessage) {
	if len(history) <= maxLength {
		return history
	}

	return history[len(history)-maxLength:]
}