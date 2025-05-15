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

package azure

import (
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/ansys/aali-sharedtypes/pkg/config"
	"github.com/ansys/aali-sharedtypes/pkg/logging"
)

func mustCfg(ctx *logging.ContextMap, key string) string {
	toolVal, ok := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES[key]

	if !ok {
		err := fmt.Sprintf("%s not found in configuration", key)
		logging.Log.Error(ctx, err)
		panic(err)
	}

	return toolVal
}

func Tool1() *azopenai.ChatCompletionsFunctionToolDefinitionFunction {
	ctx := &logging.ContextMap{}

	// Get the tool name from the configuration
	toolName := mustCfg(ctx, "APP_TOOL_1_NAME")
	toolDescription := mustCfg(ctx, "APP_TOOL_1_DESCRIPTION")

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
	toolName := mustCfg(ctx, "APP_TOOL_2_NAME")
	toolDescription := mustCfg(ctx, "APP_TOOL_2_DESCRIPTION")
	toolProperty1Name := mustCfg(ctx, "APP_TOOL_PROPERTY_1_NAME")
	toolProperty1Type := mustCfg(ctx, "APP_TOOL_PROPERTY_1_TYPE")
	toolProperty1Description := mustCfg(ctx, "APP_TOOL_PROPERTY_1_DESCRIPTION")

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
	toolName := mustCfg(ctx, "APP_TOOL_3_NAME")
	toolDescription := mustCfg(ctx, "APP_TOOL_3_DESCRIPTION")
	toolProperty1Name := mustCfg(ctx, "APP_TOOL_PROPERTY_1_NAME")
	toolProperty1Type := mustCfg(ctx, "APP_TOOL_PROPERTY_1_TYPE")
	toolProperty1Description := mustCfg(ctx, "APP_TOOL_PROPERTY_1_DESCRIPTION")

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
	toolName := mustCfg(ctx, "APP_TOOL_4_NAME")
	toolDescription := mustCfg(ctx, "APP_TOOL_4_DESCRIPTION")
	toolProperty2Name := mustCfg(ctx, "APP_TOOL_PROPERTY_2_NAME")
	toolProperty2Type := mustCfg(ctx, "APP_TOOL_PROPERTY_2_TYPE")
	toolProperty2Description1 := mustCfg(ctx, "APP_TOOL_PROPERTY_2_DESCRIPTION_1")

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
	toolName := mustCfg(ctx, "APP_TOOL_5_NAME")
	toolDescription := mustCfg(ctx, "APP_TOOL_5_DESCRIPTION")
	toolProperty2Name := mustCfg(ctx, "APP_TOOL_PROPERTY_2_NAME")
	toolProperty2Type := mustCfg(ctx, "APP_TOOL_PROPERTY_2_TYPE")
	toolProperty2Description2 := mustCfg(ctx, "APP_TOOL_PROPERTY_2_DESCRIPTION_2")

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
	toolName := mustCfg(ctx, "APP_TOOL_6_NAME")
	toolDescription := mustCfg(ctx, "APP_TOOL_6_DESCRIPTION")
	toolProperty2Name := mustCfg(ctx, "APP_TOOL_PROPERTY_2_NAME")
	toolProperty2Type := mustCfg(ctx, "APP_TOOL_PROPERTY_2_TYPE")
	toolProperty2Description3 := mustCfg(ctx, "APP_TOOL_PROPERTY_2_DESCRIPTION_3")

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
	toolName := mustCfg(ctx, "APP_TOOL_7_NAME")
	toolDescription := mustCfg(ctx, "APP_TOOL_7_DESCRIPTION")
	toolProperty2Name := mustCfg(ctx, "APP_TOOL_PROPERTY_2_NAME")
	toolProperty2Type := mustCfg(ctx, "APP_TOOL_PROPERTY_2_TYPE")
	toolProperty2Description4 := mustCfg(ctx, "APP_TOOL_PROPERTY_2_DESCRIPTION_4")

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
	toolName := mustCfg(ctx, "APP_TOOL_8_NAME")
	toolDescription := mustCfg(ctx, "APP_TOOL_8_DESCRIPTION")
	toolProperty2Name := mustCfg(ctx, "APP_TOOL_PROPERTY_2_NAME")
	toolProperty2Type := mustCfg(ctx, "APP_TOOL_PROPERTY_2_TYPE")
	toolProperty2Description5 := mustCfg(ctx, "APP_TOOL_PROPERTY_2_DESCRIPTION_5")

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
	toolName := mustCfg(ctx, "APP_TOOL_9_NAME")
	toolDescription := mustCfg(ctx, "APP_TOOL_9_DESCRIPTION")

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
	toolName := mustCfg(ctx, "APP_TOOL_10_NAME")
	toolDescription := mustCfg(ctx, "APP_TOOL_10_DESCRIPTION")
	toolProperty2Name := mustCfg(ctx, "APP_TOOL_PROPERTY_2_NAME")
	toolProperty2Type := mustCfg(ctx, "APP_TOOL_PROPERTY_2_TYPE")
	toolProperty2Description6 := mustCfg(ctx, "APP_TOOL_PROPERTY_2_DESCRIPTION_6")

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

func Tool11() *azopenai.ChatCompletionsFunctionToolDefinitionFunction {
	// Get context
	ctx := &logging.ContextMap{}

	// Get the tool name and description from the configuration
	toolName := mustCfg(ctx, "APP_TOOL_11_NAME")
	toolDescription := mustCfg(ctx, "APP_TOOL_11_DESCRIPTION")

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

func Tool12() *azopenai.ChatCompletionsFunctionToolDefinitionFunction {
	// Get context
	ctx := &logging.ContextMap{}

	// Get the tool name and description from the configuration
	toolName := mustCfg(ctx, "APP_TOOL_12_NAME")
	toolDescription := mustCfg(ctx, "APP_TOOL_12_DESCRIPTION")

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

func Tool13() *azopenai.ChatCompletionsFunctionToolDefinitionFunction {
	// Get context
	ctx := &logging.ContextMap{}

	// Get the tool name from the configuration
	toolName := mustCfg(ctx, "APP_TOOL_13_NAME")
	toolDescription := mustCfg(ctx, "APP_TOOL_13_DESCRIPTION")
	toolProperty2Name := mustCfg(ctx, "APP_TOOL_PROPERTY_2_NAME")
	toolProperty2Type := mustCfg(ctx, "APP_TOOL_PROPERTY_2_TYPE")
	toolProperty2Description1 := mustCfg(ctx, "APP_TOOL_PROPERTY_2_DESCRIPTION_7")

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
