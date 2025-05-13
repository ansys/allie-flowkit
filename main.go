package main

import (
	_ "embed"

	"github.com/ansys/aali-sharedtypes/pkg/config"
	"github.com/ansys/aali-sharedtypes/pkg/logging"

	"github.com/ansys/allie-flowkit/pkg/functiondefinitions"
	"github.com/ansys/allie-flowkit/pkg/grpcserver"
	"github.com/ansys/allie-flowkit/pkg/internalstates"
)

//go:embed pkg/externalfunctions/dataextraction.go
var dataExtractionFile string

//go:embed pkg/externalfunctions/qdrant.go
var qdrantFile string

//go:embed pkg/externalfunctions/generic.go
var genericFile string

//go:embed pkg/externalfunctions/knowledgedb.go
var knowledgeDBFile string

//go:embed pkg/externalfunctions/llmhandler.go
var llmHandlerFile string

//go:embed pkg/externalfunctions/ansysgpt.go
var ansysGPTFile string

//go:embed pkg/externalfunctions/ansysmeshpilot.go
var ansysMeshPilotFile string

//go:embed pkg/externalfunctions/auth.go
var authFile string

func init() {
	// initialize config
	config.InitConfig([]string{"EXTERNALFUNCTIONS_GRPC_PORT", "LLM_HANDLER_ENDPOINT"}, map[string]interface{}{
		"SERVICE_NAME":        "allie-flowkit",
		"VERSION":             "1.0",
		"STAGE":               "PROD",
		"ERROR_FILE_LOCATION": "error.log",
		"LOG_LEVEL":           "error",
		"LOCAL_LOGS_LOCATION": "logs.log",
		"DATADOG_SOURCE":      "nginx",
	})

	// initialize logging
	logging.InitLogger(config.GlobalConfig)
}

func main() {
	// Initialize internal states
	internalstates.InitializeInternalStates()

	// Create file list
	files := map[string]string{
		"data_extraction":  dataExtractionFile,
		"generic":          genericFile,
		"knowledge_db":     knowledgeDBFile,
		"llm_handler":      llmHandlerFile,
		"ansys_gpt":        ansysGPTFile,
		"qdrant":           qdrantFile,
		"ansys_mesh_pilot": ansysMeshPilotFile,
		"auth":             authFile,
	}

	// Load function definitions
	for category, file := range files {
		err := functiondefinitions.ExtractFunctionDefinitionsFromPackage(file, category)
		if err != nil {
			logging.Log.Fatalf(&logging.ContextMap{}, "Error extracting function definitions from package: %v", err)
		}
	}

	// start the gRPC server
	grpcserver.StartServer()
	logging.Log.Fatalf(&logging.ContextMap{}, "Error in gRPC server. Exiting application.")
}
