package internalstates

import "github.com/ansys/allie-flowkit/pkg/grpcdefinition"

// Global variables
var AvailableFunctions map[string]*grpcdefinition.FunctionDefinition

// InitializeInternalStates initializes the internal states of the agent
// This function should be called at the beginning of the agent
// to initialize the internal states of the agent
func InitializeInternalStates() {
	AvailableFunctions = map[string]*grpcdefinition.FunctionDefinition{}
}
