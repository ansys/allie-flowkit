package externalfunctions

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/ansys/allie-sharedtypes/pkg/config"
	"github.com/ansys/allie-sharedtypes/pkg/logging"

	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/ansys/allie-flowkit/pkg/meshpilot/azure"
)

// MeshPilotReAct decides which tool to use based on user input
//
// Tags:
//   - @displayName: MeshPilotReAct
//
// Parameters:
//   - instruction: the user query
//   - history: the chat history
//   - toolsHistory: the tool history
//   - reActStage: the reason and action stage
//
// Returns:
//   - stage: the reason and action
//   - toolId: the tool id
//   - toolName: the tool name
//   - toolArgument: the tool arguments
//   - result: the result
func MeshPilotReAct(instruction string,
	history []map[string]string,
	toolsHistory []map[string]string,
	reActStage string) (stage, toolId, toolName, toolArgument, result string) {

	stage = "terminated"

	ctx := &logging.ContextMap{}

	logging.Log.Debugf(ctx, "Performing Mesh Pilot ReAct request")

	if reActStage == "begin" {
		toolsHistory = []map[string]string{}
	}

	var azureOpenAIKey string
	var modelDeploymentID string
	var azureOpenAIEndpoint string

	if len(config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES) > 0 {
		// azure openai api key
		azureOpenAIKey = config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["AZURE_OPENAI_API_KEY"]
		// azure openai model name
		modelDeploymentID = config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["AZURE_OPENAI_CHAT_MODEL_NAME"]
		// azure openai endpoint
		azureOpenAIEndpoint = config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["AZURE_OPENAI_ENDPOINT"]
	} else {
		logging.Log.Fatal(ctx, "failed to load workflow config variables")
		return
	}

	if azureOpenAIKey == "" || modelDeploymentID == "" || azureOpenAIEndpoint == "" {
		logging.Log.Fatal(ctx, "environment variables missing")
		return
	}

	keyCredential := azcore.NewKeyCredential(azureOpenAIKey)

	client, err := azopenai.NewClientWithKeyCredential(azureOpenAIEndpoint, keyCredential, nil)

	if err != nil {
		logging.Log.Fatalf(ctx, "failed to create client: %v", err)
		return
	}

	logging.Log.Info(ctx, "MeshPilot ReAct...")
	logging.Log.Infof(ctx, "Beging stage: %q", reActStage)

	messages := []azopenai.ChatRequestMessageClassification{}

	// get system prompt from the configuration
	system_prompt, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["AZURE_OPENAI_SYSTEM_PROMPT"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load system prompt from the configuration")
		return
	}

	messages = append(messages, &azopenai.ChatRequestSystemMessage{Content: azopenai.NewChatRequestSystemMessageContent(system_prompt)})

	// populate history
	for _, message := range history {
		role := message["role"]
		content := message["content"]

		if role == "user" {
			messages = append(messages, &azopenai.ChatRequestUserMessage{Content: azopenai.NewChatRequestUserMessageContent(content)})
		} else if role == "assistant" {
			messages = append(messages, &azopenai.ChatRequestAssistantMessage{Content: azopenai.NewChatRequestAssistantMessageContent(content)})
		}
	}

	// user instruction
	if len(instruction) > 0 {
		messages = append(messages, &azopenai.ChatRequestUserMessage{Content: azopenai.NewChatRequestUserMessageContent(instruction)})
	}

	// populate tool history
	for _, tool := range toolsHistory {
		role := tool["role"]
		content := tool["content"]
		toolId := tool["toolId"]
		if role == "assistant" {
			toolName := tool["toolName"]
			toolArguments := tool["toolArguments"]
			messages = append(messages,
				&azopenai.ChatRequestAssistantMessage{
					Content: azopenai.NewChatRequestAssistantMessageContent(content),
					ToolCalls: []azopenai.ChatCompletionsToolCallClassification{
						&azopenai.ChatCompletionsFunctionToolCall{
							Function: &azopenai.FunctionCall{
								Arguments: &toolArguments,
								Name:      &toolName,
							},
							ID: &toolId,
						},
					},
				},
			)

		} else if role == "tool" {
			messages = append(messages, &azopenai.ChatRequestToolMessage{Content: azopenai.NewChatRequestToolMessageContent(content), ToolCallID: &toolId})
		}
	}

	logging.Log.Infof(ctx, "messages: %q", messages)

	resp, err := client.GetChatCompletions(context.TODO(), azopenai.ChatCompletionsOptions{
		DeploymentName: &modelDeploymentID,
		Messages:       messages,
		Tools: []azopenai.ChatCompletionsToolDefinitionClassification{
			&azopenai.ChatCompletionsFunctionToolDefinition{
				Function: azure.Tool1(),
			},
			&azopenai.ChatCompletionsFunctionToolDefinition{
				Function: azure.Tool2(),
			},
			&azopenai.ChatCompletionsFunctionToolDefinition{
				Function: azure.Tool3(),
			},
			&azopenai.ChatCompletionsFunctionToolDefinition{
				Function: azure.Tool4(),
			},
			&azopenai.ChatCompletionsFunctionToolDefinition{
				Function: azure.Tool5(),
			},
			&azopenai.ChatCompletionsFunctionToolDefinition{
				Function: azure.Tool6(),
			},
			&azopenai.ChatCompletionsFunctionToolDefinition{
				Function: azure.Tool7(),
			},
			&azopenai.ChatCompletionsFunctionToolDefinition{
				Function: azure.Tool8(),
			},
			&azopenai.ChatCompletionsFunctionToolDefinition{
				Function: azure.Tool9(),
			},
			&azopenai.ChatCompletionsFunctionToolDefinition{
				Function: azure.Tool10(),
			},
		},
		Temperature: to.Ptr[float32](0.0),
	}, nil)

	if err != nil {
		logging.Log.Fatalf(ctx, "failed to create chat completion: %v", err)
		return
	}

	if len(resp.Choices) == 0 {
		logging.Log.Fatalf(ctx, "No Response: %v", resp)
		return
	}

	choice := resp.Choices[0]

	finishReason := *choice.FinishReason
	logging.Log.Info(ctx, fmt.Sprintf("Finish Reason: %s", finishReason))

	if finishReason == "stop" {
		if choice.Message == nil {
			logging.Log.Fatal(ctx, "Finish Reason is stop but no Message")
			return
		}

		content := *choice.Message.Content
		stage = "stop"

		actions := []map[string]string{}
		finalResult := make(map[string]interface{})
		finalResult["Actions"] = actions
		finalResult["Message"] = content

		bytesStream, err := json.Marshal(finalResult)
		if err != nil {
			logging.Log.Fatalf(ctx, "Failed to marshal: %v", err)
		}

		result = string(bytesStream)
		logging.Log.Infof(ctx, "End stage: %q", stage)
		return
	}

	if len(choice.Message.ToolCalls) == 0 {
		logging.Log.Fatal(ctx, "No Tools")
		return
	}

	tool := choice.Message.ToolCalls[0]

	// Type assertion
	funcTool, ok := tool.(*azopenai.ChatCompletionsFunctionToolCall)

	if !ok {
		logging.Log.Fatal(ctx, "failed to convert to ChatCompletionsFunctionToolCall")
		return
	}

	toolId = *funcTool.ID
	toolCall := funcTool.Function
	toolName = *toolCall.Name
	logging.Log.Info(ctx, fmt.Sprintf("Function name: %q", toolName))

	toolArgument = *toolCall.Arguments
	logging.Log.Info(ctx, fmt.Sprintf("Function Arguments: %q", toolArgument))

	stage = "tool_call"
	logging.Log.Infof(ctx, "End stage: %q", stage)
	return
}

// SimilartitySearchOnPathDescriptions do similarity search on path description
//
// Tags:
//   - @displayName: SimilartitySearchOnPathDescriptions
//
// Parameters:
//   - instruction: the user query
//   - toolName: the tool name
//
// Returns:
//   - descriptions: the list of descriptions
func SimilartitySearchOnPathDescriptions(instruction string, toolName string) (descriptions []string) {
	descriptions = []string{}

	db_ctx := context.Background()
	ctx := &logging.ContextMap{}

	logging.Log.Infof(ctx, "Instruction: %s", instruction)

	url := config.GlobalConfig.NEO4J_URI
	username := config.GlobalConfig.NEO4J_USERNAME
	password := config.GlobalConfig.NEO4J_PASSWORD

	var api_key string
	var resource string
	var deployment string

	if len(config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES) > 0 {
		// azure openai api key
		api_key = config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["AZURE_OPENAI_API_KEY"]
		// azure openai model name
		resource = config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["AZURE_OPENAI_RESOURCE"]
		// azure openai endpoint
		deployment = config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["AZURE_OPENAI_DEPLOYMENT"]
	} else {
		logging.Log.Fatal(ctx, "failed to load workflow config variables")
		return
	}

	// get azure openai dimension from the configuration
	dimension := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["AZURE_OPENAI_EMBEDDINGS_DIMENSION"]

	// Create a driver instance
	driver, err := neo4j.NewDriverWithContext(url, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		logging.Log.Fatalf(ctx, "failed to create driver: %v", err)
		return
	}

	// Open a new session
	neo4jSession := driver.NewSession(db_ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})

	defer neo4jSession.Close(db_ctx)

	// get similarity search query from the configuration
	query, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_DATABASE_SIMILARITY_SEARCH_QUERY"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load similarity search query from the configuration")
		return
	}

	indexName, err := getIndexNameFromToolName(toolName)

	if err != nil {
		logging.Log.Errorf(ctx, "Error at SimilaritySearchOnPathDescriptions: %v", err)
		return
	}

	params := map[string]interface{}{
		"token":       api_key,
		"resource":    resource,
		"deployment":  deployment,
		"dimension":   dimension,
		"instruction": instruction,
		"index":       indexName,
		"topK":        5,
	}

	result, err := neo4jSession.Run(db_ctx, query, params)
	if err != nil {
		logging.Log.Fatalf(ctx, "Raised Exception at Get Solutions From Failure Codes: %v", err)
		return
	}

	// Get the query node name from the configuration
	queryNodeName, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_DATABASE_SIMILARITY_SEARCH_QUERY_NODE_NAME"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load query node name from the configuration")
		return
	}

	// Get the query node property name from the configuration
	queryNodePropertyName, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_DATABASE_SIMILARITY_SEARCH_QUERY_NODE_PROPERTY_NAME"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load query node property name from the configuration")
		return
	}

	// Get the query score name from the configuration
	queryScoreName, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_DATABASE_SIMILARITY_SEARCH_QUERY_SCORE_NAME"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load query score name from the configuration")
		return
	}

	// Get the descriptions
	for result.Next(db_ctx) {
		record := result.Record()
		score, _ := record.Get(queryScoreName)
		rawNode, _ := record.Get(queryNodeName)

		node := rawNode.(map[string]interface{})
		description, _ := node[queryNodePropertyName].(string)
		descriptionScore := score.(float64)
		if descriptionScore > 0.8 {
			descriptions = append(descriptions, description)
		}
	}

	if err = result.Err(); err != nil {
		logging.Log.Fatalf(ctx, "Raised Exception at Get Solutions From Failure Codes: %v\n", err)
		return
	}

	logging.Log.Infof(ctx, "Descriptions: %q", descriptions)
	return
}

// FindRelevantPathDescriptionByPrompt finds the relevant description by prompting
//
// Tags:
//   - @displayName: FindRelevantPathDescriptionByPrompt
//
// Parameters:
//   - descriptions: the list of descriptions
//   - instruction: the user instruction
//
// Returns:
//   - relevantDescription: the relevant desctiption
func FindRelevantPathDescriptionByPrompt(descriptions []string, instruction string) (relevantDescription string) {

	relevantDescription = ""
	ctx := &logging.ContextMap{}

	if len(descriptions) == 0 {
		logging.Log.Fatalf(ctx, "no descriptions provided to this function")
		return
	}

	if len(descriptions) == 1 {
		relevantDescription = descriptions[0]
		return
	}

	var azureOpenAIKey string
	var modelDeploymentID string
	var azureOpenAIEndpoint string

	if len(config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES) > 0 {
		// azure openai api key
		azureOpenAIKey = config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["AZURE_OPENAI_API_KEY"]
		// azure openai model name
		modelDeploymentID = config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["AZURE_OPENAI_CHAT_MODEL_NAME"]
		// azure openai endpoint
		azureOpenAIEndpoint = config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["AZURE_OPENAI_ENDPOINT"]
	} else {
		logging.Log.Fatal(ctx, "failed to load workflow config variables")
		return
	}

	if azureOpenAIKey == "" || modelDeploymentID == "" || azureOpenAIEndpoint == "" {
		logging.Log.Fatalf(ctx, "environment variables missing")
		return
	}

	keyCredential := azcore.NewKeyCredential(azureOpenAIKey)

	client, err := azopenai.NewClientWithKeyCredential(azureOpenAIEndpoint, keyCredential, nil)

	if err != nil {
		log.Fatalf("failed to create client: %v", err)
		return
	}

	// get the prompt template from the configuration
	prompt_template, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_PROMPT_TEMPLATE_1"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load prompt template from the configuration")
		return
	}

	prompt := fmt.Sprintf(prompt_template, descriptions, instruction)

	resp, err := client.GetChatCompletions(context.TODO(), azopenai.ChatCompletionsOptions{
		DeploymentName: &modelDeploymentID,
		Messages: []azopenai.ChatRequestMessageClassification{
			&azopenai.ChatRequestUserMessage{
				Content: azopenai.NewChatRequestUserMessageContent(prompt),
			},
		},
		Temperature: to.Ptr[float32](0.0),
	}, nil)

	if err != nil {
		logging.Log.Fatalf(ctx, "error occur during chat completion %v", err)
		return
	}

	if len(resp.Choices) == 0 {
		logging.Log.Fatalf(ctx, "the response from azure is empty")
		return
	}

	message := resp.Choices[0].Message

	if message == nil {
		logging.Log.Fatalf(ctx, "no message found from the choice")
		return
	}

	var output *struct {
		Index int `json:"index"`
	}

	err = json.Unmarshal([]byte(*message.Content), &output)

	if err != nil {
		logging.Log.Fatalf(ctx, "failed to un marshal index output")
		return
	}

	logging.Log.Infof(ctx, "The Index: %d", output.Index)

	if output.Index < len(descriptions) && output.Index >= 0 {
		relevantDescription = descriptions[output.Index]
	} else {
		logging.Log.Errorf(ctx, "Output Index: %d, out of range( 0, %d )", output.Index, len(descriptions))
		return
	}

	logging.Log.Infof(ctx, "The relevant description: %s", relevantDescription)

	return
}

// FetchPropertiesFromPathDescription get properties from path description
//
// Tags:
//   - @displayName: FetchPropertiesFromPathDescription
//
// Parameters:
//   - description: the desctiption of path
//
// Returns:
//   - properties: the list of descriptions
func FetchPropertiesFromPathDescription(description string) (properties []string) {

	ctx := &logging.ContextMap{}

	logging.Log.Infof(ctx, "Fetching Properties From Path Descriptions...")

	// Get environment variables
	url := config.GlobalConfig.NEO4J_URI
	username := config.GlobalConfig.NEO4J_USERNAME
	password := config.GlobalConfig.NEO4J_PASSWORD

	db_ctx := context.Background()

	// Create a driver instance
	driver, err := neo4j.NewDriverWithContext(url, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		logging.Log.Fatalf(ctx, "failed to create driver: %v", err)
		return
	}

	// Open a new session
	neo4jSession := driver.NewSession(db_ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})

	defer neo4jSession.Close(db_ctx)

	// Get the query from the configuration
	query, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_DATABASE_FIND_NODE_QUERY"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load query from the configuration")
		return
	}

	params := map[string]interface{}{
		"description": description,
	}

	result, err := neo4jSession.Run(db_ctx, query, params)
	if err != nil {
		logging.Log.Fatalf(ctx, "Raised Exception to fetch path node from description: %v\n", err)
		return
	}

	// Get the query node name from the configuration
	queryNodeName, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_DATABASE_FIND_NODE_QUERY_NODE_NAME"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load query node name from the configuration")
		return
	}

	// Get the query node property name from the configuration
	queryNodePropertyName, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_DATABASE_FIND_NODE_QUERY_NODE_PROPERTY_NAME"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load query node property name from the configuration")
		return
	}

	for result.Next(db_ctx) {
		record := result.Record()
		node, _ := record.Get(queryNodeName)

		descNode := node.(neo4j.Node)
		props := descNode.Props[queryNodePropertyName].([]interface{})

		for _, property := range props {
			properties = append(properties, property.(string))
		}
	}

	if err = result.Err(); err != nil {
		logging.Log.Fatalf(ctx, "failed to fetch database records: %v", err)
		return
	}

	logging.Log.Infof(ctx, "Propetries: %q\n", properties)

	return
}

// FetchNodeDescriptionsFromPathDescription get node descriptions from path description
//
// Tags:
//   - @displayName: FetchNodeDescriptionsFromPathDescription
//
// Parameters:
//   - description: the desctiption of path
//
// Returns:
//   - actionDescriptions: action descriptions
func FetchNodeDescriptionsFromPathDescription(description string) (actionDescriptions string) {

	ctx := &logging.ContextMap{}

	logging.Log.Infof(ctx, "Fetching Node Descriptions From Path Descriptions...")

	url := config.GlobalConfig.NEO4J_URI
	username := config.GlobalConfig.NEO4J_USERNAME
	password := config.GlobalConfig.NEO4J_PASSWORD

	db_ctx := context.Background()

	// Create a driver instance
	driver, err := neo4j.NewDriverWithContext(url, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		logging.Log.Fatalf(ctx, "failed to create driver: %v", err)
		return
	}

	// Open a new session
	neo4jSession := driver.NewSession(db_ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})

	defer neo4jSession.Close(db_ctx)

	// Get the query from the configuration
	query, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_DATABASE_FETCH_PATH_NODES_QUERY"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load query from the configuration")
		return
	}

	params := map[string]interface{}{
		"description": description,
	}

	result, err := neo4jSession.Run(db_ctx, query, params)
	if err != nil {
		logging.Log.Fatalf(ctx, "Raised Exception to fetch path node from description: %v\n", err)
		return
	}

	// Get the query node name from the configuration
	queryNodeName, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_DATABASE_FETCH_PATH_NODES_QUERY_NODE_NAME"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load query node name from the configuration")
		return
	}

	// Get the query node property name from the configuration
	queryNodePropertyName, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_DATABASE_FETCH_PATH_NODES_QUERY_NODE_PROPERTY_NAME"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load query node property name from the configuration")
		return
	}

	nodeDescriptions := []string{}
	for result.Next(db_ctx) {
		record := result.Record()
		middleNodes, _ := record.Get(queryNodeName)

		nodes := middleNodes.([]interface{})

		for _, node := range nodes {
			node := node.(neo4j.Node)
			props := node.Props

			for key, value := range props {
				if key == queryNodePropertyName {
					nodeDescriptions = append(nodeDescriptions, value.(string))
				}
			}
		}
	}

	if err = result.Err(); err != nil {
		logging.Log.Fatalf(ctx, "failed to fetch database records: %v", err)
		return
	}

	logging.Log.Infof(ctx, "Node Descriptions: %q\n", nodeDescriptions)

	byteStream, err := json.Marshal(nodeDescriptions)

	if err != nil {
		logging.Log.Fatalf(ctx, "Failed to Marshal: %v", err)
		return
	}

	actionDescriptions = string(byteStream)
	return
}

// FetchActionsPathFromPathDescription fetch actions from path description
//
// Tags:
//   - @displayName: FetchActionsPathFromPathDescription
//
// Parameters:
//   - description: the desctiption of path
//   - nodeLabel: the label of the node
//
// Returns:
//   - actions: the list of actions to execute
func FetchActionsPathFromPathDescription(description, nodeLabel string) (actions []map[string]string) {
	db_ctx := context.Background()

	ctx := &logging.ContextMap{}
	url := config.GlobalConfig.NEO4J_URI
	username := config.GlobalConfig.NEO4J_USERNAME
	password := config.GlobalConfig.NEO4J_PASSWORD

	// Create a driver instance
	driver, err := neo4j.NewDriverWithContext(url, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		logging.Log.Fatalf(ctx, "failed to create driver: %v", err)
		return
	}

	// Open a new session
	neo4jSession := driver.NewSession(db_ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})

	defer neo4jSession.Close(db_ctx)

	// Get the node label 1 from the configuration
	nodeLabel1, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_DATABASE_FETCH_PATH_NODES_QUERY_NODE_LABEL_1"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load node label 1 from the configuration")
		return
	}

	// Get the node label 2 from the configuration
	nodeLabel2, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_DATABASE_FETCH_PATH_NODES_QUERY_NODE_LABEL_2"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load node label 2 from the configuration")
		return
	}

	var query string
	if nodeLabel == nodeLabel2 {
		// Get the query from the configuration
		query, exists = config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_DATABASE_FETCH_PATH_NODES_QUERY_2"]
		if !exists {
			logging.Log.Fatal(ctx, "failed to load query from the configuration")
			return
		}
	} else if nodeLabel == nodeLabel1 {
		// Get the query from the configuration
		query, exists = config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_DATABASE_FETCH_PATH_NODES_QUERY"]
		if !exists {
			logging.Log.Fatal(ctx, "failed to load query from the configuration")
			return
		}
	} else {
		logging.Log.Infof(ctx, "Invalid Node Label: %q", nodeLabel)
		return
	}

	params := map[string]interface{}{
		"description": description,
	}

	result, err := neo4jSession.Run(db_ctx, query, params)
	if err != nil {
		logging.Log.Fatalf(ctx, "Raised Exception to fetch path node from description: %v", err)
		return
	}

	// Get the query node name from the configuration
	queryNodeName, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_DATABASE_FETCH_PATH_NODES_QUERY_NODE_NAME"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load query node name from the configuration")
		return
	}

	for result.Next(db_ctx) {
		record := result.Record()
		middleNodes, _ := record.Get(queryNodeName)
		nodes := middleNodes.([]interface{})
		for _, node := range nodes {
			node := node.(neo4j.Node)
			props := node.Props
			action := make(map[string]string)

			for key, value := range props {
				action[key] = value.(string)
			}
			actions = append(actions, action)
		}
	}

	if err = result.Err(); err != nil {
		logging.Log.Fatalf(ctx, "failed to fetch database records: %v", err)
		return
	}

	logging.Log.Info(ctx, "successfully fetched actions from database")
	return
}

// SynthesizeActions update action as per user instruction
//
// Tags:
//   - @displayName: SynthesizeActions
//
// Parameters:
//   - instruction: the user instruction
//   - properties: the list of properties
//   - actions: the list of actions
//
// Returns:
//   - updatedActions: the list of synthesized actions
func SynthesizeActions(instruction string, properties []string, actions []map[string]string) (updatedActions []map[string]string) {

	ctx := &logging.ContextMap{}

	updatedActions = actions

	if len(properties) == 0 {
		logging.Log.Infof(ctx, "No properties to synthesize actions")
		return
	}

	var azureOpenAIKey string
	var modelDeploymentID string
	var azureOpenAIEndpoint string

	if len(config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES) > 0 {
		// azure openai api key
		azureOpenAIKey = config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["AZURE_OPENAI_API_KEY"]
		// azure openai model name
		modelDeploymentID = config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["AZURE_OPENAI_CHAT_MODEL_NAME"]
		// azure openai endpoint
		azureOpenAIEndpoint = config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["AZURE_OPENAI_ENDPOINT"]
	} else {
		logging.Log.Fatal(ctx, "failed to load workflow config variables")
		return
	}

	if azureOpenAIKey == "" || modelDeploymentID == "" || azureOpenAIEndpoint == "" {
		logging.Log.Fatalf(ctx, "environment variables missing")
		return
	}

	keyCredential := azcore.NewKeyCredential(azureOpenAIKey)

	client, err := azopenai.NewClientWithKeyCredential(azureOpenAIEndpoint, keyCredential, nil)

	if err != nil {
		logging.Log.Fatalf(ctx, "failed to create client: %v\n", err)
		return
	}

	// get prompt template from the configuration
	prompt_template, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_PROMPT_TEMPLATE_SYNTHESIZE"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load prompt template from the configuration")
		return
	}

	prompt := fmt.Sprintf(prompt_template, properties, instruction)

	resp, err := client.GetChatCompletions(context.TODO(), azopenai.ChatCompletionsOptions{
		DeploymentName: &modelDeploymentID,
		Messages: []azopenai.ChatRequestMessageClassification{
			&azopenai.ChatRequestUserMessage{
				Content: azopenai.NewChatRequestUserMessageContent(prompt),
			},
		},
		Temperature: to.Ptr[float32](0.0),
	}, nil)

	if err != nil {
		logging.Log.Fatalf(ctx, "error occur during chat completion %v", err)
		return
	}

	if len(resp.Choices) == 0 {
		logging.Log.Fatalf(ctx, "the response from azure is empty")
		return
	}

	message := resp.Choices[0].Message

	if message == nil {
		logging.Log.Fatalf(ctx, "no message found from the choice")
		return
	}

	logging.Log.Infof(ctx, "The Message: %s\n", *message.Content)

	var output map[string]interface{}

	err = json.Unmarshal([]byte(*message.Content), &output)

	if err != nil {
		logging.Log.Fatal(ctx, "failed to un marshal synthesizing actions")
		return
	}

	// Get synthesize actions find key from configuration
	synthesizeActionsFindKey, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_PROMPT_TEMPLATE_SYNTHESIZE_ACTION_FIND_KEY"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load synthesize actions find key from the configuration")
		return
	}

	// Get synthesize actions replace key 1 from configuration
	synthesizeActionsReplaceKey1, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_PROMPT_TEMPLATE_SYNTHESIZE_ACTION_REPLACE_KEY_1"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load synthesize actions replace key 1 from the configuration")
		return
	}

	// Get synthesize actions replace key 2 from configuration
	synthesizeActionsReplaceKey2, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_PROMPT_TEMPLATE_SYNTHESIZE_ACTION_REPLACE_KEY_2"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load synthesize actions replace key 2 from the configuration")
		return
	}

	// Get synthesize output key 1 from configuration
	synthesizeOutputKey1, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_PROMPT_TEMPLATE_SYNTHESIZE_OUTPUT_KEY_1"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load synthesize output key 1 from the configuration")
		return
	}

	// Get synthesize output key 2 from configuration
	synthesizeOutputKey2, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_PROMPT_TEMPLATE_SYNTHESIZE_OUTPUT_KEY_2"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load synthesize output key 2 from the configuration")
		return
	}

	// Update the actions
	for key, value := range output {
		switch v := value.(type) {
		case string:
			updateMeshPilotActionProperty(updatedActions, synthesizeActionsFindKey, key, synthesizeActionsReplaceKey1, v)
		case int:
		case int64:
		case int32:
			convValue := fmt.Sprintf("%d", v)
			updateMeshPilotActionProperty(updatedActions, synthesizeActionsFindKey, key, synthesizeActionsReplaceKey1, convValue)
		case float32:
		case float64:
			convValue := fmt.Sprintf("%f", v)
			updateMeshPilotActionProperty(updatedActions, synthesizeActionsFindKey, key, synthesizeActionsReplaceKey1, convValue)
		case map[string]interface{}:
			for key, value := range v {
				switch key {
				case synthesizeOutputKey1:
					updateMeshPilotActionProperty(updatedActions, synthesizeActionsFindKey, key, synthesizeActionsReplaceKey1, value.(string))
				case synthesizeOutputKey2:
					updateMeshPilotActionProperty(updatedActions, synthesizeActionsFindKey, key, synthesizeActionsReplaceKey2, value.(string))
				}
			}
		default:
			logging.Log.Infof(ctx, "Key: %s, Value is of a different type: %T", key, v)
		}
	}

	return
}

// FinalizeResult converts actions to json string to send back data
//
// Tags:
//   - @displayName: FinalizeResult
//
// Parameters:
//   - actions: the executable actions
//   - toolName: tool name to create customize messages
//
// Returns:
//   - result: the actions in json format
func FinalizeResult(actions []map[string]string, toolName string) (result string) {

	ctx := &logging.ContextMap{}

	var hasActions bool
	var message string

	if len(actions) > 0 {
		hasActions = true
	} else {
		hasActions = false
	}

	// Get tool 2 name from the configuration
	tool2Name, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_2_NAME"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load tool 2 name from the configuration")
		return
	}

	// Get tool 4 name from the configuration
	tool4Name, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_4_NAME"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load tool 4 name from the configuration")
		return
	}

	// Get tool 5 name from the configuration
	tool5Name, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_5_NAME"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load tool 5 name from the configuration")
		return
	}

	// Get tool 6 name from the configuration
	tool6Name, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_6_NAME"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load tool 6 name from the configuration")
		return
	}

	// Get tool 7 name from the configuration
	tool7Name, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_7_NAME"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load tool 7 name from the configuration")
		return
	}

	// Get tool 8 name from the configuration
	tool8Name, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_8_NAME"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load tool 8 name from the configuration")
		return
	}

	// Get tool 10 name from the configuration
	tool10Name, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_10_NAME"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load tool 10 name from the configuration")
		return
	}

	// Get tool action success message from configuration
	toolActionSuccessMessage, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_ACTION_SUCCESS_MESSAGE"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load tool action success message from the configuration")
		return
	}

	// Get tool 2 action message from configuration
	tool2ActionMessage, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_2_ACTION_MESSAGE"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load tool 2 action message from the configuration")
		return
	}

	// Get tool 2 no action message from configuration
	tool2NoActionMessage, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_2_NO_ACTION_MESSAGE"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load tool 2 no action message from the configuration")
		return
	}

	// Get tool 4 no action message from configuration
	tool4NoActionMessage, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_4_NO_ACTION_MESSAGE"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load tool 4 no action message from the configuration")
		return
	}

	// Get tool 5 no action message from configuration
	tool5NoActionMessage, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_5_NO_ACTION_MESSAGE"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load tool 5 no action message from the configuration")
		return
	}

	// Get tool 6 no action message from configuration
	tool6NoActionMessage, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_6_NO_ACTION_MESSAGE"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load tool 6 no action message from the configuration")
		return
	}

	// Get tool 7 no action message from configuration
	tool7NoActionMessage, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_7_NO_ACTION_MESSAGE"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load tool 7 no action message from the configuration")
		return
	}

	// Get tool 8 no action message from configuration
	tool8NoActionMessage, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_8_NO_ACTION_MESSAGE"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load tool 8 no action message from the configuration")
		return
	}

	// Get tool 10 no action message from configuration
	tool10NoActionMessage, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_10_NO_ACTION_MESSAGE"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load tool 10 no action message from the configuration")
		return
	}

	message = toolActionSuccessMessage
	if toolName == tool2Name {
		if hasActions {
			message = tool2ActionMessage
		} else {
			message = tool2NoActionMessage
		}
	} else if toolName == tool4Name {
		if !hasActions {
			message = tool4NoActionMessage
		}
	} else if toolName == tool5Name {
		if !hasActions {
			message = tool5NoActionMessage
		}
	} else if toolName == tool6Name {
		if !hasActions {
			message = tool6NoActionMessage
		}
	} else if toolName == tool7Name {
		if !hasActions {
			message = tool7NoActionMessage
		}
	} else if toolName == tool8Name {
		if !hasActions {
			message = tool8NoActionMessage
		}
	} else if toolName == tool10Name {
		if !hasActions {
			message = tool10NoActionMessage
		}
	} else {
		logging.Log.Errorf(ctx, "Invalid toolName %s", toolName)
		return
	}

	finalMessage := make(map[string]interface{})
	finalMessage["Message"] = message
	finalMessage["Actions"] = actions

	// Marshal the actions
	bytesStream, err := json.Marshal(finalMessage)

	if err != nil {
		logging.Log.Errorf(ctx, "failed to convert actions to json: %v", err)
		return
	}

	result = string(bytesStream)
	logging.Log.Info(ctx, "successfully converted actions to json")

	return
}

// GetSolutionsToFixProblem do similarity search on path description
//
// Tags:
//   - @displayName: GetSolutionsToFixProblem
//
// Parameters:
//   - fmFailureCode: FM failure Code
//   - primeMeshFailureCode: Prime Mesh Failure Code
//
// Returns:
//   - solutions: the list of solutions in json
func GetSolutionsToFixProblem(fmFailureCode, primeMeshFailureCode string) (solutions string) {

	ctx := &logging.ContextMap{}

	logging.Log.Info(ctx, "mesh pilot get solutions to fix problem...")

	solutions = ""

	url := config.GlobalConfig.NEO4J_URI
	username := config.GlobalConfig.NEO4J_USERNAME
	password := config.GlobalConfig.NEO4J_PASSWORD

	solutionsVec := []string{}

	db_ctx := context.Background()
	// Create a driver instance
	driver, err := neo4j.NewDriverWithContext(url, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		logging.Log.Fatalf(ctx, "failed to create driver: %v", err)
		return
	}

	// Open a new session
	session := driver.NewSession(db_ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(db_ctx)

	logging.Log.Info(ctx, fmt.Sprintf("GetSolutionsFromFailureCodes: %s, %s\n", fmFailureCode, primeMeshFailureCode))

	// Get the query from the configuration
	query, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_DATABASE_GET_SOLUTIONS_QUERY"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load query from the configuration")
		return
	}

	// Get the query node name from the configuration
	queryNodeName, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_DATABASE_GET_SOLUTIONS_QUERY_NODE_NAME"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load query node name from the configuration")
		return
	}

	// Get the query node property name from the configuration
	queryNodePropertyName, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_DATABASE_GET_SOLUTIONS_QUERY_NODE_PROPERTY_NAME"]
	if !exists {
		logging.Log.Fatal(ctx, "failed to load query node property name from the configuration")
		return
	}

	params := map[string]interface{}{
		"fm_failure_code":         strings.TrimSpace(fmFailureCode),
		"prime_mesh_failure_code": strings.TrimSpace(primeMeshFailureCode),
	}

	_, err = session.ExecuteRead(db_ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		result, err := transaction.Run(db_ctx, query, params)

		if err != nil {
			logging.Log.Errorf(ctx, "Error during transaction.Run: %v", err)
			return nil, err
		}

		for result.Next(db_ctx) {
			record := result.Record()
			state, _ := record.Get(queryNodeName)
			node := state.(neo4j.Node)
			description, _ := node.Props[queryNodePropertyName].(string)
			solutionsVec = append(solutionsVec, description)
		}
		return true, nil
	})

	if err != nil {
		logging.Log.Errorf(ctx, "Error during session.ExecuteRead: %v", err)
		return
	}

	byteStream, err := json.Marshal(solutionsVec)
	if err != nil {
		logging.Log.Fatalf(ctx, "Error marshalling solutions: %v\n", err)
		return
	}

	solutions = string(byteStream)
	logging.Log.Info(ctx, "found solutions to fix problem...")
	return
}

// GetSelectedSolution get user selected solutions from the options provided
//
// Tags:
//   - @displayName: GetSelectedSolution
//
// Parameters:
//   - arguments: these are the arguments ReAct found based on user choice
//
// Returns:
//   - solution: the selected solution
func GetSelectedSolution(arguments string) (solution string) {

	ctx := &logging.ContextMap{}

	var output struct {
		Solution string `json:"solution_description"`
	}

	err := json.Unmarshal([]byte(arguments), &output)

	if err != nil {
		logging.Log.Fatalf(ctx, "failed to un marshal index output")
		return
	}

	solution = output.Solution
	logging.Log.Infof(ctx, "Selected Solution: %s", solution)
	return
}

// AppendToolHistory this function append tool history
//
// Tags:
//   - @displayName: AppendToolHistory
//
// Parameters:
//   - toolHistory: the tool history
//   - toolId: the tool id
//   - toolName: the tool name
//   - toolArguments: the tool arguments
//   - toolResponse: the tool response
//
// Returns:
//   - updatedToolHistory: the updated tool history
func AppendToolHistory(toolHistory []map[string]string, toolId, toolName, toolArguments, toolResponse string) (updatedToolHistory []map[string]string) {
	ctx := &logging.ContextMap{}

	// populate tool history
	for _, tool := range toolHistory {
		toolId := tool["toolId"]
		content := tool["content"]
		tool := map[string]string{
			"toolId":  toolId,
			"content": content,
		}
		updatedToolHistory = append(updatedToolHistory, tool)
	}

	// populate current tool
	if len(toolId) > 0 && len(toolResponse) > 0 {
		// append assistant tool call
		tool := map[string]string{
			"role":          "assistant",
			"content":       toolResponse,
			"toolId":        toolId,
			"toolName":      toolName,
			"toolArguments": toolArguments,
		}
		updatedToolHistory = append(updatedToolHistory, tool)

		// append tool
		tool = map[string]string{
			"role":    "tool",
			"content": toolResponse,
			"toolId":  toolId,
		}
		updatedToolHistory = append(updatedToolHistory, tool)
	}

	logging.Log.Info(ctx, fmt.Sprintf("Updated Tool History: %q", updatedToolHistory))
	return
}

// AppendMeshPilotHistory this function append mesh pilot history
//
// Tags:
//   - @displayName: AppendMeshPilotHistory
//
// Parameters:
//   - history: the tool history
//   - role: the tool id
//   - content: the tool name
//
// Returns:
//   - updatedHistory: the updated mesh pilot history
func AppendMeshPilotHistory(history []map[string]string, role, content string) (updatedHistory []map[string]string) {
	ctx := &logging.ContextMap{}

	updatedHistory = []map[string]string{}

	for _, item := range history {
		updatedHistory = append(updatedHistory, item)
	}

	updatedHistory = append(updatedHistory, map[string]string{
		"role":    role,
		"content": content,
	})

	logging.Log.Infof(ctx, "Updated history: %q", updatedHistory)
	return
}
