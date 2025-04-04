package externalfunctions

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

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
	responseChannel := sendEmbeddingsRequest(input, llmHandlerEndpoint, false, nil)

	// Process the first response and close the channel
	var embedding32 []float32
	var err error
	for response := range responseChannel {
		// Check if the response is an error
		if response.Type == "error" {
			panic(response.Error)
		}

		// Log LLM response
		logging.Log.Debugf(&logging.ContextMap{}, "Received embeddings response.")

		// Get embedded vector array
		interfaceArray, ok := response.EmbeddedData.([]interface{})
		if !ok {
			errMessage := "error converting embedded data to interface array"
			logging.Log.Error(&logging.ContextMap{}, errMessage)
			panic(errMessage)
		}
		embedding32, err = convertToFloat32Slice(interfaceArray)
		if err != nil {
			errMessage := fmt.Sprintf("error converting embedded data to float32 slice: %v", err)
			logging.Log.Error(&logging.ContextMap{}, errMessage)
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

// PerformVectorEmbeddingRequestWithTokenLimitCatch performs a vector embedding request to LLM
// and catches the token limit error message
//
// Tags:
//   - @displayName: Embeddings with Token Limit Catch
//
// Parameters:
//   - input: the input string
//
// Returns:
//   - embeddedVector: the embedded vector in float32 format
func PerformVectorEmbeddingRequestWithTokenLimitCatch(input string, tokenLimitMessage string) (embeddedVector []float32, tokenLimitReached bool, responseMessage string) {
	// get the LLM handler endpoint
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send embeddings request
	responseChannel := sendEmbeddingsRequest(input, llmHandlerEndpoint, false, nil)

	// Process the first response and close the channel
	var embedding32 []float32
	var err error
	for response := range responseChannel {
		// Check if the response is an error
		if response.Type == "error" {
			if strings.Contains(response.Error.Message, "tokens") {
				return nil, true, tokenLimitMessage
			} else {
				panic(response.Error)
			}
		}

		// Log LLM response
		logging.Log.Debugf(&logging.ContextMap{}, "Received embeddings response.")

		// Get embedded vector array
		interfaceArray, ok := response.EmbeddedData.([]interface{})
		if !ok {
			errMessage := "error converting embedded data to interface array"
			logging.Log.Error(&logging.ContextMap{}, errMessage)
			panic(errMessage)
		}
		embedding32, err = convertToFloat32Slice(interfaceArray)
		if err != nil {
			errMessage := fmt.Sprintf("error converting embedded data to float32 slice: %v", err)
			logging.Log.Error(&logging.ContextMap{}, errMessage)
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

	return embedding32, false, ""
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
	responseChannel := sendEmbeddingsRequest(input, llmHandlerEndpoint, false, nil)

	// Process the first response and close the channel
	embedding32Array := make([][]float32, len(input))
	for response := range responseChannel {
		// Check if the response is an error
		if response.Type == "error" {
			panic(response.Error)
		}

		// Log LLM response
		logging.Log.Debugf(&logging.ContextMap{}, "Received batch embeddings response.")

		// Get embedded vector array
		interfaceArray, ok := response.EmbeddedData.([]interface{})
		if !ok {
			errMessage := "error converting embedded data to interface array"
			logging.Log.Error(&logging.ContextMap{}, errMessage)
			panic(errMessage)
		}

		for i, interfaceArrayElement := range interfaceArray {
			lowerInterfaceArray, ok := interfaceArrayElement.([]interface{})
			if !ok {
				errMessage := "error converting embedded data to interface array"
				logging.Log.Error(&logging.ContextMap{}, errMessage)
				panic(errMessage)
			}
			embedding32, err := convertToFloat32Slice(lowerInterfaceArray)
			if err != nil {
				errMessage := fmt.Sprintf("error converting embedded data to float32 slice: %v", err)
				logging.Log.Error(&logging.ContextMap{}, errMessage)
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

// PerformBatchHybridEmbeddingRequest performs a batch hybrid embedding request to LLM
// returning the sparse and dense embeddings
//
// Tags:
//   - @displayName: Batch Hybrid Embeddings
//
// Parameters:
//   - input: the input strings
//
// Returns:
//   - denseEmbeddings: the dense embeddings in float32 format
//   - sparseEmbeddings: the sparse embeddings in map format
func PerformBatchHybridEmbeddingRequest(input []string, maxBatchSize int) (denseEmbeddings [][]float32, sparseEmbeddings []map[uint]float32) {
	processedEmbeddings := 0

	// Process data in batches
	for i := 0; i < len(input); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(input) {
			end = len(input)
		}

		// Create a batch of data to send to LLM handler
		batchTextToEmbed := input[i:end]

		// Send http request
		batchDenseEmbeddings, batchLexicalWeights, err := llmHandlerPerformVectorEmbeddingRequest(batchTextToEmbed, true)
		if err != nil {
			errMessage := fmt.Sprintf("Error performing batch embedding request: %v", err)
			logging.Log.Error(&logging.ContextMap{}, errMessage)
			panic(errMessage)
		}

		// Add the embeddings to the list
		denseEmbeddings = append(denseEmbeddings, batchDenseEmbeddings...)
		sparseEmbeddings = append(sparseEmbeddings, batchLexicalWeights...)

		processedEmbeddings += len(batchTextToEmbed)
		logging.Log.Debugf(&logging.ContextMap{}, "Processed %d embeddings", processedEmbeddings)
	}

	return denseEmbeddings, sparseEmbeddings
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

	logging.Log.Debugf(&logging.ContextMap{}, "Received keywords response.")

	// Close the response channel
	close(responseChannel)

	// Unmarshal JSON data into the result variable
	err := json.Unmarshal([]byte(responseAsStr), &keywords)
	if err != nil {
		errMessage := fmt.Sprintf("Error unmarshalling keywords response from allie-llm: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
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

	logging.Log.Debugf(&logging.ContextMap{}, "Received summary response.")

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
	responseChannel := sendChatRequest(input, "general", history, 0, systemPrompt, llmHandlerEndpoint, nil, nil, nil)
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

// PerformGeneralRequestWithImages performs a general request to LLM with images
//
// Tags:
//   - @displayName: General LLM Request (with Images)
//
// Parameters:
//   - input: the user input
//   - history: the conversation history
//   - isStream: the flag to indicate whether the response should be streamed
//   - systemPrompt: the system prompt
//   - images: the images
//
// Returns:
//   - message: the response message
//   - stream: the stream channel
func PerformGeneralRequestWithImages(input string, history []sharedtypes.HistoricMessage, isStream bool, systemPrompt string, images []string) (message string, stream *chan string) {
	// get the LLM handler endpoint
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send chat request
	responseChannel := sendChatRequest(input, "general", history, 0, systemPrompt, llmHandlerEndpoint, nil, nil, images)
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

// PerformGeneralModelSpecificationRequest performs a specified request to LLM with a configured model and Systemprompt.
//
// Tags:
//   - @displayName: General LLM Request (Specified System Prompt)
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
func PerformGeneralModelSpecificationRequest(input string, history []sharedtypes.HistoricMessage, isStream bool, systemPrompt map[string]string, modelIds []string) (message string, stream *chan string) {
	// get the LLM handler endpoint
	fmt.Printf("[%s] type of alpsRequest inside modelspecification %T\n", time.Now().Format("2006-01-02 15:04:05.000"), systemPrompt)
	logging.Log.Infof(&logging.ContextMap{}, "[%s] type of alpsRequest inside modelspecification %T\n", time.Now().Format("2006-01-02 15:04:05.000"), systemPrompt)

	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT
	// Set up WebSocket connection with LLM and send chat request
	responseChannel := sendChatRequest(input, "general", history, 0, systemPrompt, llmHandlerEndpoint, modelIds, nil, nil)

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
	responseChannel := sendChatRequest(input, "general", history, 0, systemPrompt, llmHandlerEndpoint, modelIds, nil, nil)

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

// PerformGeneralRequestSpecificModel performs a general request to LLM with a specific model
//
// Tags:
//   - @displayName: General LLM Request (Specific Models & Model Options)
//
// Parameters:
//   - input: the user input
//   - history: the conversation history
//   - isStream: the flag to indicate whether the response should be streamed
//   - systemPrompt: the system prompt
//   - modelId: the model ID
//   - modelOptions: the model options
//
// Returns:
//   - message: the response message
//   - stream: the stream channel
func PerformGeneralRequestSpecificModelAndModelOptions(input string, history []sharedtypes.HistoricMessage, isStream bool, systemPrompt string, modelIds []string, modelOptions sharedtypes.ModelOptions) (message string, stream *chan string) {
	// get the LLM handler endpoint
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send chat request
	responseChannel := sendChatRequest(input, "general", history, 0, systemPrompt, llmHandlerEndpoint, modelIds, &modelOptions, nil)

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

// PerformGeneralRequestSpecificModelNoStreamWithOpenAiTokenOutput performs a general request to LLM with a specific model
// and returns the token count using OpenAI token count model. Does not stream the response.
//
// Tags:
//   - @displayName: General LLM Request (Specific Models, No Stream, OpenAI Token Output)
//
// Parameters:
//   - input: the user input
//   - history: the conversation history
//   - systemPrompt: the system prompt
//   - modelIds: the model IDs of the AI models to use
//   - tokenCountModelName: the model name to use for token count
//
// Returns:
//   - message: the response message
//   - tokenCount: the token count
func PerformGeneralRequestSpecificModelNoStreamWithOpenAiTokenOutput(input string, history []sharedtypes.HistoricMessage, systemPrompt string, modelIds []string, tokenCountModelName string) (message string, tokenCount int) {
	// get the LLM handler endpoint
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send chat request
	responseChannel := sendChatRequest(input, "general", history, 0, systemPrompt, llmHandlerEndpoint, modelIds, nil, nil)

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

	// get input token count
	totalTokenCount, err := openAiTokenCount(tokenCountModelName, input+systemPrompt)
	if err != nil {
		errorMessage := fmt.Sprintf("Error getting input token count: %v", err)
		logging.Log.Errorf(&logging.ContextMap{}, errorMessage)
		panic(errorMessage)
	}

	// get history token count
	for _, message := range history {
		historyTokenCount, err := openAiTokenCount(tokenCountModelName, message.Content)
		if err != nil {
			errorMessage := fmt.Sprintf("Error getting history token count: %v", err)
			logging.Log.Errorf(&logging.ContextMap{}, errorMessage)
			panic(errorMessage)
		}
		totalTokenCount += historyTokenCount
	}

	// get the output token count
	outputTokenCount, err := openAiTokenCount(tokenCountModelName, responseAsStr)
	if err != nil {
		errorMessage := fmt.Sprintf("Error getting output token count: %v", err)
		logging.Log.Errorf(&logging.ContextMap{}, errorMessage)
		panic(errorMessage)
	}
	totalTokenCount += outputTokenCount

	// log token count
	logging.Log.Debugf(&logging.ContextMap{}, "Total token count: %d", totalTokenCount)

	// Return the response
	return responseAsStr, totalTokenCount
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
	responseChannel := sendChatRequest(input, "code", history, 0, "", llmHandlerEndpoint, nil, nil, nil)

	// If isStream is true, create a stream channel and return asap
	if isStream {
		// Create a stream channel
		streamChannel := make(chan string, 400)

		// Start a goroutine to transfer the data from the response channel to the stream channel
		go transferDatafromResponseToStreamChannel(&responseChannel, &streamChannel, validateCode, false, "", 0, 0, "", "", false, "")

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
			logging.Log.Errorf(&logging.ContextMap{}, "Error extracting Python code: %v", err)
		} else {

			// Validate the Python code
			valid, warnings, err := validatePythonCode(pythonCode)
			if err != nil {
				logging.Log.Errorf(&logging.ContextMap{}, "Error validating Python code: %v", err)
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

// PerformGeneralRequestNoStreaming performs a general chat completion request to LLM without streaming
//
// Tags:
//   - @displayName: General LLM Request (no streaming)
//
// Parameters:
//   - input: the input string
//   - history: the conversation history
//   - systemPrompt: the system prompt
//
// Returns:
//   - message: the generated message
func PerformGeneralRequestNoStreaming(input string, history []sharedtypes.HistoricMessage, systemPrompt string) (message string) {
	// get the LLM handler endpoint
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send chat request
	responseString := sendChatRequestNoStreaming(input, "general", history, 0, systemPrompt, llmHandlerEndpoint, nil, nil, nil)

	// Return the response
	return responseString
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
		logging.Log.Warn(&logging.ContextMap{}, errMessage)
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
