package azure

import (
	"encoding/json"

	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
)

func GetSolutionsToFixProblemToolDef() *azopenai.ChatCompletionsFunctionToolDefinitionFunction {

	jsonBytes, err := json.Marshal(map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	})

	if err != nil {
		panic(err)
	}

	funcDef := &azopenai.ChatCompletionsFunctionToolDefinitionFunction{
		Name:        to.Ptr("GetSolutionsToFixProblem"),
		Description: to.Ptr("This function return solutions to fix a problem"),
		Parameters:  jsonBytes,
	}

	return funcDef
}

func ExecuteUserSelectedSolutionToolDef() *azopenai.ChatCompletionsFunctionToolDefinitionFunction {

	jsonBytes, err := json.Marshal(map[string]any{
		"required": []string{"solution_description"},
		"type":     "object",
		"properties": map[string]any{
			"solution_description": map[string]any{
				"type":        "string",
				"description": "user selected solution description, the selected solution description is from one of the previous chat that listed solutions, send complete sentence (solution description that listed in previous chat) as an argument to this method",
			},
		},
	})

	if err != nil {
		panic(err)
	}

	funcDef := &azopenai.ChatCompletionsFunctionToolDefinitionFunction{
		Name:        to.Ptr("ExecuteUserSelectedSolution"),
		Description: to.Ptr("Get list of actions from the given solution description to execute the solution"),
		Parameters:  jsonBytes,
	}

	return funcDef
}

func ExplainExecutionOfUserSelectedSolutionToolDef() *azopenai.ChatCompletionsFunctionToolDefinitionFunction {

	jsonBytes, err := json.Marshal(map[string]any{
		"required": []string{"solution_description"},
		"type":     "object",
		"properties": map[string]any{
			"solution_description": map[string]any{
				"type":        "string",
				"description": "user selected solution description, the selected solution description is from one of the previous chat that listed solutions, send complete sentence (solution description that listed in previous chat) as an argument to this method",
			},
		},
	})

	if err != nil {
		panic(err)
	}

	funcDef := &azopenai.ChatCompletionsFunctionToolDefinitionFunction{
		Name:        to.Ptr("ExplainExecutionOfUserSelectedSolution"),
		Description: to.Ptr("From the selected solution describe how it is executed, share the details for selected solutionn"),
		Parameters:  jsonBytes,
	}

	return funcDef
}

func DeleteToolDef() *azopenai.ChatCompletionsFunctionToolDefinitionFunction {

	jsonBytes, err := json.Marshal(map[string]any{
		"required": []string{"instruction"},
		"type":     "object",
		"properties": map[string]any{
			"instruction": map[string]any{
				"type":        "string",
				"description": "Get list of actions to Delete either step/operation/control/outcome or label/part",
			},
		},
	})

	if err != nil {
		panic(err)
	}

	funcDef := &azopenai.ChatCompletionsFunctionToolDefinitionFunction{
		Name:        to.Ptr("Delete"),
		Description: to.Ptr("Delete instruction of any form as an argument to this method"),
		Parameters:  jsonBytes,
	}

	return funcDef
}

func CreateOrInsertOrAddToolDef() *azopenai.ChatCompletionsFunctionToolDefinitionFunction {

	jsonBytes, err := json.Marshal(map[string]any{
		"required": []string{"instruction"},
		"type":     "object",
		"properties": map[string]any{
			"instruction": map[string]any{
				"type":        "string",
				"description": "User provided an create instruction of any form as an argument to this method",
			},
		},
	})

	if err != nil {
		panic(err)
	}

	funcDef := &azopenai.ChatCompletionsFunctionToolDefinitionFunction{
		Name:        to.Ptr("CreateOrInsertOrAdd"),
		Description: to.Ptr("Get list of actions from the user provided instruction to create step/operation/control/outcome"),
		Parameters:  jsonBytes,
	}

	return funcDef
}

func UpdateOrSetToolDef() *azopenai.ChatCompletionsFunctionToolDefinitionFunction {

	jsonBytes, err := json.Marshal(map[string]any{
		"required": []string{"instruction"},
		"type":     "object",
		"properties": map[string]any{
			"instruction": map[string]any{
				"type":        "string",
				"description": "set/update instruction of any form as an argument to this method",
			},
		},
	})

	if err != nil {
		panic(err)
	}

	funcDef := &azopenai.ChatCompletionsFunctionToolDefinitionFunction{
		Name:        to.Ptr("UpdateOrSet"),
		Description: to.Ptr("Get list of actions to update/set properties of step/operation/control/outcome"),
		Parameters:  jsonBytes,
	}

	return funcDef
}

func ExecuteToolDef() *azopenai.ChatCompletionsFunctionToolDefinitionFunction {

	jsonBytes, err := json.Marshal(map[string]any{
		"required": []string{"instruction"},
		"type":     "object",
		"properties": map[string]any{
			"instruction": map[string]any{
				"type":        "string",
				"description": "execute instruction of any form as an argument to this method",
			},
		},
	})

	if err != nil {
		panic(err)
	}

	funcDef := &azopenai.ChatCompletionsFunctionToolDefinitionFunction{
		Name:        to.Ptr("Execute"),
		Description: to.Ptr("Get list of actions to execute step"),
		Parameters:  jsonBytes,
	}

	return funcDef
}

func RevertToolDef() *azopenai.ChatCompletionsFunctionToolDefinitionFunction {

	jsonBytes, err := json.Marshal(map[string]any{
		"required": []string{"instruction"},
		"type":     "object",
		"properties": map[string]any{
			"instruction": map[string]any{
				"type":        "string",
				"description": "revert instruction of any form as an argument to this method",
			},
		},
	})

	if err != nil {
		panic(err)
	}

	funcDef := &azopenai.ChatCompletionsFunctionToolDefinitionFunction{
		Name:        to.Ptr("Revert"),
		Description: to.Ptr("Get List of actions to Revert steps"),
		Parameters:  jsonBytes,
	}

	return funcDef
}

func UndoToolDef() *azopenai.ChatCompletionsFunctionToolDefinitionFunction {

	jsonBytes, err := json.Marshal(map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	})

	if err != nil {
		panic(err)
	}

	funcDef := &azopenai.ChatCompletionsFunctionToolDefinitionFunction{
		Name:        to.Ptr("Undo"),
		Description: to.Ptr("user can ask for either undo session or just undo, undo session changes that have done so far by this service"),
		Parameters:  jsonBytes,
	}

	return funcDef
}

func ConnectToolDef() *azopenai.ChatCompletionsFunctionToolDefinitionFunction {

	jsonBytes, err := json.Marshal(map[string]any{
		"required": []string{"instruction"},
		"type":     "object",
		"properties": map[string]any{
			"instruction": map[string]any{
				"type":        "string",
				"description": "Connect instruction of any form as an argument to this method",
			},
		},
	})

	if err != nil {
		panic(err)
	}

	funcDef := &azopenai.ChatCompletionsFunctionToolDefinitionFunction{
		Name:        to.Ptr("Connect"),
		Description: to.Ptr("Get list of actions to Connect either outcome to control or connect labels or parts"),
		Parameters:  jsonBytes,
	}

	return funcDef
}
