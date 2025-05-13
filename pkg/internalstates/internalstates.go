package internalstates

import (
	"github.com/ansys/aali-sharedtypes/pkg/aaliflowkitgrpc"
)

// Global variables
var AvailableFunctions map[string]*aaliflowkitgrpc.FunctionDefinition

// InitializeInternalStates initializes the internal states of the agent
// This function should be called at the beginning of the agent
// to initialize the internal states of the agent
func InitializeInternalStates() {
	AvailableFunctions = map[string]*aaliflowkitgrpc.FunctionDefinition{}
}
