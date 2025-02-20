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

	// system prompt
	messages = append(messages, &azopenai.ChatRequestSystemMessage{Content: azopenai.NewChatRequestSystemMessageContent("You are a helpful AI agent called MeshPilot, helps user based on provided tools only. Give highly specific answers based on the information you're provided. The response from AI agent has to be markdown strictly.")})

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
				Function: azure.GetSolutionsToFixProblemToolDef(),
			},
			&azopenai.ChatCompletionsFunctionToolDefinition{
				Function: azure.ExecuteUserSelectedSolutionToolDef(),
			},
			&azopenai.ChatCompletionsFunctionToolDefinition{
				Function: azure.ExplainExecutionOfUserSelectedSolutionToolDef(),
			},
			&azopenai.ChatCompletionsFunctionToolDefinition{
				Function: azure.DeleteToolDef(),
			},
			&azopenai.ChatCompletionsFunctionToolDefinition{
				Function: azure.CreateOrInsertOrAddToolDef(),
			},
			&azopenai.ChatCompletionsFunctionToolDefinition{
				Function: azure.UpdateOrSetToolDef(),
			},
			&azopenai.ChatCompletionsFunctionToolDefinition{
				Function: azure.ExecuteToolDef(),
			},
			&azopenai.ChatCompletionsFunctionToolDefinition{
				Function: azure.RevertToolDef(),
			},
			&azopenai.ChatCompletionsFunctionToolDefinition{
				Function: azure.UndoToolDef(),
			},
			&azopenai.ChatCompletionsFunctionToolDefinition{
				Function: azure.ConnectToolDef(),
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

	logging.Log.Infof(ctx, "Workflow Config Variables: %q", config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES)

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

	dimension := "1536"

	// Create a driver instance
	driver, err := neo4j.NewDriverWithContext(url, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		logging.Log.Fatalf(ctx, "failed to create driver: %v", err)
		return
	}

	// Open a new session
	neo4jSession := driver.NewSession(db_ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})

	defer neo4jSession.Close(db_ctx)

	query := `
		WITH genai.vector.encode(
			$instruction,
			"AzureOpenAI",
			{
			token: $token,
			resource: $resource,
			deployment: $deployment,
			dimension: $dimension
			}) AS instruction_embedding
			
		CALL db.index.vector.queryNodes(
			$index,
			$topK,
			instruction_embedding
			) YIELD node AS instruction, score
		RETURN instruction {.Description } AS instruction, score
	`

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

	for result.Next(db_ctx) {
		record := result.Record()
		score, _ := record.Get("score")
		instruction, _ := record.Get("instruction")

		node := instruction.(map[string]interface{})
		description, _ := node["Description"].(string)
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

	prompt_template := `
		You are given a dynamic list of descriptions and a user input. 
		Your task is to find the index of the description in the list that has the relevant meaning to the user input.
		The indexing of the list of description start from 0, if there is no relevant description then return -1.
		Return only the index of the relevant description in the following JSON format:
		{ "index": <index> }

		List of descriptions:
		%q

		User input: %s

		Return the index of the most relevant description:
	`

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

	query := `
		MATCH (desc:Description)
		WHERE 'ACTION_SEQUENCE_BEGIN' IN desc.node_category
		AND desc.Description = $description

		RETURN desc AS descNode
	`

	params := map[string]interface{}{
		"description": description,
	}

	result, err := neo4jSession.Run(db_ctx, query, params)
	if err != nil {
		logging.Log.Fatalf(ctx, "Raised Exception to fetch path node from description: %v\n", err)
		return
	}

	for result.Next(db_ctx) {
		record := result.Record()
		node, _ := record.Get("descNode")

		descNode := node.(neo4j.Node)
		props := descNode.Props["properties"].([]interface{})

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

	query := `
		MATCH (start:State)
		WHERE 'Error' IN start.object_state
		AND start.Description = $description

		MATCH path = (start)-[:NEXT*]->(end:State)
		WHERE 'Succeeded' IN end.object_state

		WITH nodes(path) AS allNodes
		WITH allNodes[1..-1] AS middleNodes

		RETURN DISTINCT middleNodes
	`

	params := map[string]interface{}{
		"description": description,
	}

	result, err := neo4jSession.Run(db_ctx, query, params)
	if err != nil {
		logging.Log.Fatalf(ctx, "Raised Exception to fetch path node from description: %v\n", err)
		return
	}

	nodeDescriptions := []string{}
	for result.Next(db_ctx) {
		record := result.Record()
		middleNodes, _ := record.Get("middleNodes")

		nodes := middleNodes.([]interface{})

		for _, node := range nodes {
			node := node.(neo4j.Node)
			props := node.Props

			for key, value := range props {
				if key == "Description" {
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

	var query string
	if nodeLabel == "Description" {
		query = `
			MATCH (start:Description)
			WHERE 'ACTION_SEQUENCE_BEGIN' IN start.node_category
			AND start.Description = $description

			MATCH path = (start)-[:NEXT*]->(end:Description)
			WHERE 'ACTION_SEQUENCE_END' IN end.node_category

			WITH nodes(path) AS allNodes
			WITH allNodes[1..-1] AS middleNodes

			RETURN DISTINCT middleNodes
		`
	} else if nodeLabel == "State" {
		query = `
			MATCH (start:State)
			WHERE 'Error' IN start.object_state
			AND start.Description = $description

			MATCH path = (start)-[:NEXT*]->(end:State)
			WHERE 'Succeeded' IN end.object_state

			WITH nodes(path) AS allNodes
			WITH allNodes[1..-1] AS middleNodes

			RETURN DISTINCT middleNodes
		`
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

	for result.Next(db_ctx) {
		record := result.Record()
		middleNodes, _ := record.Get("middleNodes")

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

// SynthesizeActions update action act api's as per user instruction
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

	prompt_template := `
		You are given a list of JSON keys and a user question. Follow the instructions and generate the desired JSON format as shown in the example.
		
		Instructions:
		
		1. Identify the keys from the list that are explicitly mentioned in the user question.
		2. Extract the corresponding values from the user question.
		3. If a value has units, create a nested dictionary with "value" and "units".
		4. For boolean values, assign "1" for true and "0" for false.
		5. Ensure the output is in JSON format.
		6. **Only include keys that are explicitly mentioned in the user question. Do not infer or assume any additional keys or values.**
		7. Do not confuse different properties. Each key should be matched exactly as mentioned in the user question.
		8. Do not include properties that are not explicitly mentioned in the user question, even if they are listed in the properties.

		Example:
		
		List of JSON keys: ["temperature", "Status", "humidity"]
		
		User question: "The Temperature is 25 degrees and the status is true."
		
		Desired JSON output:
		
		{
			"temperature": {
				"value": "25",
				"units": "deg"
			},
			"Status": "1"
		}

		List of JSON keys: ["temperature", "Status", "humidity"]

		User question: "The boiling temperature of water is 100 Celsius at a pressure of 1 bar."
		
		Desired JSON output:
		
		{
			"temperature": {
				"value": "100",
				"units": "Celsius"
			}
		}

		List of JSON keys: %q
		
		User question: %s

		Desired JSON output:
	`

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

	for key, value := range output {
		switch v := value.(type) {
		case string:
			updateMeshPilotActionProperty(updatedActions, "ActApi", key, "Argument", v)
		case int:
		case int64:
		case int32:
			convValue := fmt.Sprintf("%d", v)
			updateMeshPilotActionProperty(updatedActions, "ActApi", key, "Argument", convValue)
		case float32:
		case float64:
			convValue := fmt.Sprintf("%f", v)
			updateMeshPilotActionProperty(updatedActions, "ActApi", key, "Argument", convValue)
		case map[string]interface{}:
			argumentValue := v["value"].(string)
			argumentUnits := v["units"].(string)
			updateMeshPilotActionProperty(updatedActions, "ActApi", key, "Argument", argumentValue)
			updateMeshPilotActionProperty(updatedActions, "ActApi", key, "ArgumentUnits", argumentUnits)
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

	message = "Succesfully executed user instruction"
	if toolName == "ExecuteUserSelectedSolution" {
		if hasActions {
			message = "Executed the solution sucessfully"
		} else {
			message = "No actions found for the selected solution"
		}
	} else if toolName == "Delete" {
		if !hasActions {
			message = "No actions found for the given delete instruction"
		}
	} else if toolName == "CreateOrInsertOrAdd" {
		if !hasActions {
			message = "No actions found for the given insert instruction"
		}
	} else if toolName == "UpdateOrSet" {
		if !hasActions {
			message = "No actions found for the given update instruction"
		}
	} else if toolName == "Execute" {
		if !hasActions {
			message = "No actions found for the given execute instruction"
		}
	} else if toolName == "Revert" {
		if !hasActions {
			message = "No actions found for the given revert instruction"
		}
	} else if toolName == "Connect" {
		if !hasActions {
			message = "No actions found for the given connect instruction"
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

	query := `
		MATCH (state:State)
		WHERE 'Error' IN state.object_state
		AND state.fm_failure_code = $fm_failure_code
		AND state.prime_mesh_failure_code = $prime_mesh_failure_code
		RETURN state
	`

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
			state, _ := record.Get("state")
			node := state.(neo4j.Node)
			description, _ := node.Props["Description"].(string)
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
