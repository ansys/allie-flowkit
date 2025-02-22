package azure

import (
	"encoding/json"

	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/ansys/allie-sharedtypes/pkg/config"
	"github.com/ansys/allie-sharedtypes/pkg/logging"
)

func Tool1() *azopenai.ChatCompletionsFunctionToolDefinitionFunction {
	ctx := &logging.ContextMap{}

	// Get the tool name from the configuration
	toolName, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_1_NAME"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_1_NAME not found in configuration")
	}

	// Get the tool description from the configuration
	toolDescription, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_1_DESCRIPTION"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_1_DESCRIPTION not found in configuration")
	}

	jsonBytes, err := json.Marshal(map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	})

	if err != nil {
		panic(err)
	}

	funcDef := &azopenai.ChatCompletionsFunctionToolDefinitionFunction{
		Name:        to.Ptr(toolName),
		Description: to.Ptr(toolDescription),
		Parameters:  jsonBytes,
	}

	return funcDef
}

func Tool2() *azopenai.ChatCompletionsFunctionToolDefinitionFunction {
	ctx := &logging.ContextMap{}

	// Get the tool name from the configuration
	toolName, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_2_NAME"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_2_NAME not found in configuration")
	}

	// Get the tool description from the configuration
	toolDescription, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_2_DESCRIPTION"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_2_DESCRIPTION not found in configuration")
	}

	// Get the tool properties from the configuration
	toolProperty1Name, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_PROPERTY_1_NAME"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_PROPERTY_1_NAME not found in configuration")
	}

	toolProperty1Type, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_PROPERTY_1_TYPE"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_PROPERTY_1_TYPE not found in configuration")
	}

	toolProperty1Description, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_PROPERTY_1_DESCRIPTION"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_PROPERTY_1_DESCRIPTION not found in configuration")
	}

	// Define the parameters for the function
	jsonBytes, err := json.Marshal(map[string]any{
		"required": []string{toolProperty1Name},
		"type":     "object",
		"properties": map[string]any{
			toolProperty1Name: map[string]any{
				"type":        toolProperty1Type,
				"description": toolProperty1Description,
			},
		},
	})

	if err != nil {
		panic(err)
	}

	funcDef := &azopenai.ChatCompletionsFunctionToolDefinitionFunction{
		Name:        to.Ptr(toolName),
		Description: to.Ptr(toolDescription),
		Parameters:  jsonBytes,
	}

	return funcDef
}

func Tool3() *azopenai.ChatCompletionsFunctionToolDefinitionFunction {
	ctx := &logging.ContextMap{}
	// Get the tool name from the configuration
	toolName, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_3_NAME"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_3_NAME not found in configuration")
	}

	// Get the tool description from the configuration
	toolDescription, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_3_DESCRIPTION"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_3_DESCRIPTION not found in configuration")
	}

	// Get the tool 1 property name from the configuration
	toolProperty1Name, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_PROPERTY_1_NAME"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_PROPERTY_1_NAME not found in configuration")
	}

	// Get the tool 1 property type from the configuration
	toolProperty1Type, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_PROPERTY_1_TYPE"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_PROPERTY_1_TYPE not found in configuration")
	}

	// Get the tool 1 property description from the configuration
	toolProperty1Description, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_PROPERTY_1_DESCRIPTION"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_PROPERTY_1_DESCRIPTION not found in configuration")
	}

	// Define the parameters for the function
	jsonBytes, err := json.Marshal(map[string]any{
		"required": []string{toolProperty1Name},
		"type":     "object",
		"properties": map[string]any{
			toolProperty1Name: map[string]any{
				"type":        toolProperty1Type,
				"description": toolProperty1Description,
			},
		},
	})

	if err != nil {
		panic(err)
	}

	funcDef := &azopenai.ChatCompletionsFunctionToolDefinitionFunction{
		Name:        to.Ptr(toolName),
		Description: to.Ptr(toolDescription),
		Parameters:  jsonBytes,
	}

	return funcDef
}

func Tool4() *azopenai.ChatCompletionsFunctionToolDefinitionFunction {
	// Get context
	ctx := &logging.ContextMap{}

	// Get the tool name from the configuration
	toolName, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_4_NAME"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_4_NAME not found in configuration")
	}

	// Get the tool description from the configuration
	toolDescription, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_4_DESCRIPTION"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_4_DESCRIPTION not found in configuration")
	}

	// Get the tool property 2 name from the configuration
	toolProperty2Name, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_PROPERTY_2_NAME"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_PROPERTY_2_NAME not found in configuration")
	}

	// Get the tool property 2 type from the configuration
	toolProperty2Type, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_PROPERTY_2_TYPE"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_PROPERTY_2_TYPE not found in configuration")
	}

	// Get the tool property 2 description 1 from the configuration
	toolProperty2Description1, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_PROPERTY_2_DESCRIPTION_1"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_PROPERTY_2_DESCRIPTION_1 not found in configuration")
	}

	// Define the parameters for the function
	jsonBytes, err := json.Marshal(map[string]any{
		"required": []string{toolProperty2Name},
		"type":     "object",
		"properties": map[string]any{
			toolProperty2Name: map[string]any{
				"type":        toolProperty2Type,
				"description": toolProperty2Description1,
			},
		},
	})

	if err != nil {
		panic(err)
	}

	funcDef := &azopenai.ChatCompletionsFunctionToolDefinitionFunction{
		Name:        to.Ptr(toolName),
		Description: to.Ptr(toolDescription),
		Parameters:  jsonBytes,
	}

	return funcDef
}

func Tool5() *azopenai.ChatCompletionsFunctionToolDefinitionFunction {
	// Get context
	ctx := &logging.ContextMap{}

	// Get the tool name from the configuration
	toolName, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_5_NAME"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_5_NAME not found in configuration")
	}

	// Get the tool description from the configuration
	toolDescription, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_5_DESCRIPTION"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_5_DESCRIPTION not found in configuration")
	}

	// Get the tool property 2 name from the configuration
	toolProperty2Name, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_PROPERTY_2_NAME"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_PROPERTY_2_NAME not found in configuration")
	}

	// Get the tool property 2 type from the configuration
	toolProperty2Type, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_PROPERTY_2_TYPE"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_PROPERTY_2_TYPE not found in configuration")
	}

	// Get the tool property 2 description 2 from the configuration
	toolProperty2Description2, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_PROPERTY_2_DESCRIPTION_2"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_PROPERTY_2_DESCRIPTION_2 not found in configuration")
	}

	// Define the parameters for the function
	jsonBytes, err := json.Marshal(map[string]any{
		"required": []string{toolProperty2Name},
		"type":     "object",
		"properties": map[string]any{
			toolProperty2Name: map[string]any{
				"type":        toolProperty2Type,
				"description": toolProperty2Description2,
			},
		},
	})

	if err != nil {
		panic(err)
	}

	funcDef := &azopenai.ChatCompletionsFunctionToolDefinitionFunction{
		Name:        to.Ptr(toolName),
		Description: to.Ptr(toolDescription),
		Parameters:  jsonBytes,
	}

	return funcDef
}

func Tool6() *azopenai.ChatCompletionsFunctionToolDefinitionFunction {
	// Get context
	ctx := &logging.ContextMap{}

	// Get the tool name from the configuration
	toolName, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_6_NAME"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_6_NAME not found in configuration")
	}

	// Get the tool description from the configuration
	toolDescription, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_6_DESCRIPTION"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_6_DESCRIPTION not found in configuration")
	}

	// Get the tool property 2 name from the configuration
	toolProperty2Name, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_PROPERTY_2_NAME"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_PROPERTY_2_NAME not found in configuration")
	}

	// Get the tool property 2 type from the configuration
	toolProperty2Type, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_PROPERTY_2_TYPE"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_PROPERTY_2_TYPE not found in configuration")
	}

	// Get the tool property 2 description 3 from the configuration
	toolProperty2Description3, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_PROPERTY_2_DESCRIPTION_3"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_PROPERTY_2_DESCRIPTION_3 not found in configuration")
	}

	jsonBytes, err := json.Marshal(map[string]any{
		"required": []string{toolProperty2Name},
		"type":     "object",
		"properties": map[string]any{
			toolProperty2Name: map[string]any{
				"type":        toolProperty2Type,
				"description": toolProperty2Description3,
			},
		},
	})

	if err != nil {
		panic(err)
	}

	funcDef := &azopenai.ChatCompletionsFunctionToolDefinitionFunction{
		Name:        to.Ptr(toolName),
		Description: to.Ptr(toolDescription),
		Parameters:  jsonBytes,
	}

	return funcDef
}

func Tool7() *azopenai.ChatCompletionsFunctionToolDefinitionFunction {
	// Get context
	ctx := &logging.ContextMap{}

	// Get the tool name from the configuration
	toolName, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_7_NAME"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_7_NAME not found in configuration")
	}

	// Get the tool description from the configuration
	toolDescription, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_7_DESCRIPTION"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_7_DESCRIPTION not found in configuration")
	}

	// Get the tool property 2 name from the configuration
	toolProperty2Name, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_PROPERTY_2_NAME"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_PROPERTY_2_NAME not found in configuration")
	}

	// Get the tool property 2 type from the configuration
	toolProperty2Type, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_PROPERTY_2_TYPE"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_PROPERTY_2_TYPE not found in configuration")
	}

	// Get the tool property 2 description 4 from the configuration
	toolProperty2Description4, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_PROPERTY_2_DESCRIPTION_4"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_PROPERTY_2_DESCRIPTION_4 not found in configuration")
	}

	// Define the parameters for the function
	jsonBytes, err := json.Marshal(map[string]any{
		"required": []string{toolProperty2Name},
		"type":     "object",
		"properties": map[string]any{
			toolProperty2Name: map[string]any{
				"type":        toolProperty2Type,
				"description": toolProperty2Description4,
			},
		},
	})

	if err != nil {
		panic(err)
	}

	funcDef := &azopenai.ChatCompletionsFunctionToolDefinitionFunction{
		Name:        to.Ptr(toolName),
		Description: to.Ptr(toolDescription),
		Parameters:  jsonBytes,
	}

	return funcDef
}

func Tool8() *azopenai.ChatCompletionsFunctionToolDefinitionFunction {
	// Get context
	ctx := &logging.ContextMap{}

	// Get the tool name from the configuration
	toolName, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_8_NAME"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_8_NAME not found in configuration")
	}

	// Get the tool description from the configuration
	toolDescription, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_8_DESCRIPTION"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_8_DESCRIPTION not found in configuration")
	}

	// Get the tool property 2 name from the configuration
	toolProperty2Name, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_PROPERTY_2_NAME"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_PROPERTY_2_NAME not found in configuration")
	}

	// Get the tool property 2 type from the configuration
	toolProperty2Type, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_PROPERTY_2_TYPE"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_PROPERTY_2_TYPE not found in configuration")
	}

	// Get the tool property 2 description 5 from the configuration
	toolProperty2Description5, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_PROPERTY_2_DESCRIPTION_5"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_PROPERTY_2_DESCRIPTION_5 not found in configuration")
	}

	// Define the parameters for the function
	jsonBytes, err := json.Marshal(map[string]any{
		"required": []string{toolProperty2Name},
		"type":     "object",
		"properties": map[string]any{
			toolProperty2Name: map[string]any{
				"type":        toolProperty2Type,
				"description": toolProperty2Description5,
			},
		},
	})

	if err != nil {
		panic(err)
	}

	funcDef := &azopenai.ChatCompletionsFunctionToolDefinitionFunction{
		Name:        to.Ptr(toolName),
		Description: to.Ptr(toolDescription),
		Parameters:  jsonBytes,
	}

	return funcDef
}

func Tool9() *azopenai.ChatCompletionsFunctionToolDefinitionFunction {
	// Get context
	ctx := &logging.ContextMap{}

	// Get the tool name from the configuration
	toolName, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_9_NAME"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_9_NAME not found in configuration")
	}

	// Get the tool description from the configuration
	toolDescription, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_9_DESCRIPTION"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_9_DESCRIPTION not found in configuration")
	}

	// Define the parameters for the function
	jsonBytes, err := json.Marshal(map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	})

	if err != nil {
		panic(err)
	}

	funcDef := &azopenai.ChatCompletionsFunctionToolDefinitionFunction{
		Name:        to.Ptr(toolName),
		Description: to.Ptr(toolDescription),
		Parameters:  jsonBytes,
	}

	return funcDef
}

func Tool10() *azopenai.ChatCompletionsFunctionToolDefinitionFunction {
	// Get context
	ctx := &logging.ContextMap{}

	// Get the tool name from the configuration
	toolName, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_10_NAME"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_10_NAME not found in configuration")
	}

	// Get the tool description from the configuration
	toolDescription, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_10_DESCRIPTION"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_10_DESCRIPTION not found in configuration")
	}

	// Get the tool property 2 name from the configuration
	toolProperty2Name, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_PROPERTY_2_NAME"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_PROPERTY_2_NAME not found in configuration")
	}

	// Get the tool property 2 type from the configuration
	toolProperty2Type, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_PROPERTY_2_TYPE"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_PROPERTY_2_TYPE not found in configuration")
	}

	// Get the tool property 2 description 6 from the configuration
	toolProperty2Description6, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_PROPERTY_2_DESCRIPTION_6"]
	if !exists {
		logging.Log.Fatal(ctx, "APP_TOOL_PROPERTY_2_DESCRIPTION_6 not found in configuration")
	}

	// Define the parameters for the function
	jsonBytes, err := json.Marshal(map[string]any{
		"required": []string{toolProperty2Name},
		"type":     "object",
		"properties": map[string]any{
			toolProperty2Name: map[string]any{
				"type":        toolProperty2Type,
				"description": toolProperty2Description6,
			},
		},
	})

	if err != nil {
		panic(err)
	}

	funcDef := &azopenai.ChatCompletionsFunctionToolDefinitionFunction{
		Name:        to.Ptr(toolName),
		Description: to.Ptr(toolDescription),
		Parameters:  jsonBytes,
	}

	return funcDef
}
