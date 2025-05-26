package dispatcher

import (
	"fmt"
	"reflect"

	"github.com/ansys/aali-flowkit/pkg/externalfunctions"
)

// HandleMCPCall dispatches an MCP call to any registered function in ExternalFunctionsMap.
func HandleMCPCall(tool string, input map[string]interface{}) (map[string]interface{}, error) {
	fn, ok := externalfunctions.ExternalFunctionsMap[tool]
	if !ok {
		return nil, fmt.Errorf("unknown function: %s", tool)
	}

	fnVal := reflect.ValueOf(fn)
	fnType := fnVal.Type()

	// Validate input count
	if fnType.NumIn() != len(input) {
		return nil, fmt.Errorf("expected %d parameters, got %d", fnType.NumIn(), len(input))
	}

	// Build argument list in order
	args := make([]reflect.Value, fnType.NumIn())
	for i := 0; i < fnType.NumIn(); i++ {
		param := fmt.Sprintf("arg%d", i+1) // input keys: arg1, arg2, ...
		rawVal, ok := input[param]
		if !ok {
			return nil, fmt.Errorf("missing parameter: %s", param)
		}
		args[i] = reflect.ValueOf(rawVal)
	}

	// Call the function
	results := fnVal.Call(args)

	// Build response map
	output := make(map[string]interface{})
	for i, r := range results {
		output[fmt.Sprintf("out%d", i+1)] = r.Interface()
	}

	return output, nil
}
