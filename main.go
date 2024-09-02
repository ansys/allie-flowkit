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

func init() {
	// initialize config
	config.InitConfig([]string{"EXTERNALFUNCTIONS_GRPC_PORT", "LLM_HANDLER_ENDPOINT"}, map[string]interface{}{
		"VERSION":             "1.0",
		"STAGE":               "PROD",
		"ERROR_FILE_LOCATION": "error.log",
		"LOG_LEVEL":           "error",
		"LOCAL_LOGS_LOCATION": "logs.log",
		"DATADOG_SOURCE":      "nginx",
	})

	// initialize logging
	logging.InitLogger()
	logging.InitConfig(logging.Config{
		ErrorFileLocation: config.GlobalConfig.ERROR_FILE_LOCATION,
		LogLevel:          config.GlobalConfig.LOG_LEVEL,
		LocalLogs:         config.GlobalConfig.LOCAL_LOGS,
		LocalLogsLocation: config.GlobalConfig.LOCAL_LOGS_LOCATION,
		DatadogLogs:       config.GlobalConfig.DATADOG_LOGS,
		DatadogSource:     config.GlobalConfig.DATADOG_SOURCE,
		DatadogStage:      config.GlobalConfig.STAGE,
		DatadogVersion:    config.GlobalConfig.VERSION,
		DatadogService:    config.GlobalConfig.SERVICE_NAME,
		DatadogAPIKey:     config.GlobalConfig.LOGGING_API_KEY,
		DatadogLogsURL:    config.GlobalConfig.LOGGING_URL,
		DatadogMetrics:    config.GlobalConfig.DATADOG_METRICS,
		DatadogMetricsURL: config.GlobalConfig.METRICS_URL,
	})
}

func main() {
	// Initialize internal states
	internalstates.InitializeInternalStates()

	// Load function definitions
	err := functiondefinitions.ExtractFunctionDefinitionsFromPackage(externalFunctionsFile)
	if err != nil {
		logging.Log.Fatalf(internalstates.Ctx, "Error extracting function definitions from package: %v", err)
	}
	err = functiondefinitions.ExtractFunctionDefinitionsFromPackage(dataExtractionFile)
	if err != nil {
		logging.Log.Fatalf(internalstates.Ctx, "Error extracting function definitions from package: %v", err)
	}

	// Log the version of the system
	logging.Log.Info(internalstates.Ctx, "Launching Allie Flowkit")

	// start the gRPC server
	grpcserver.StartServer()
	logging.Log.Fatalf(internalstates.Ctx, "Error in gRPC server. Exiting application.")
}
