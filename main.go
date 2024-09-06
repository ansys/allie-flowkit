package main

import (
	_ "embed"

	"github.com/ansys/allie-sharedtypes/pkg/config"
	"github.com/ansys/allie-sharedtypes/pkg/logging"

	"github.com/ansys/allie-flowkit/pkg/functiondefinitions"
	"github.com/ansys/allie-flowkit/pkg/grpcserver"
	"github.com/ansys/allie-flowkit/pkg/internalstates"
)

//go:embed pkg/externalfunctions/externalfunctions.go
var externalFunctionsFile string

//go:embed pkg/externalfunctions/dataextraction.go
var dataExtractionFile string

//go:embed pkg/externalfunctions/generic.go
var genericFile string

//go:embed pkg/externalfunctions/knowledgedb.go
var knowledgeDBFile string

//go:embed pkg/externalfunctions/llmhandler.go
var llmHandlerFile string

//go:embed pkg/externalfunctions/ansysgpt.go
var ansysGPTFile string

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
	files := []string{
		externalFunctionsFile,
		dataExtractionFile,
		genericFile,
		knowledgeDBFile,
		llmHandlerFile,
		ansysGPTFile,
	}

	// Load function definitions
	for _, file := range files {
		err := functiondefinitions.ExtractFunctionDefinitionsFromPackage(file)
		if err != nil {
			logging.Log.Fatalf(internalstates.Ctx, "Error extracting function definitions from package: %v", err)
		}
	}

	// Log the version of the system
	logging.Log.Info(internalstates.Ctx, "Launching Allie Flowkit")

	// start the gRPC server
	grpcserver.StartServer()
	logging.Log.Fatalf(internalstates.Ctx, "Error in gRPC server. Exiting application.")
}
