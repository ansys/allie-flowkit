package main

import (
	_ "embed"
	"log"
	"os"

	"github.com/ansys/allie-flowkit/pkg/config"
	"github.com/ansys/allie-flowkit/pkg/functiondefinitions"
	"github.com/ansys/allie-flowkit/pkg/grpcserver"
	"github.com/ansys/allie-flowkit/pkg/internalstates"
)

//go:embed pkg/externalfunctions/externalfunctions.go
var externalFunctionsFile string

func main() {
	// Read configuration file...
	// 1st option: read from environment variable
	configFile := os.Getenv("ALLIE_CONFIG_PATH")
	if configFile == "" {
		log.Println("ALLIE_CONFIG_PATH environment variable not found...")
		log.Println("Searching for configuration file (config.yaml) at same level as the agent...")
		// 2nd option: read from default location... root directory
		configFile = "config.yaml"
	}

	log.Printf("Reading configuration from file %s...\n", configFile)
	err := config.LoadConfigFromFile(configFile)
	if err != nil {
		log.Fatal("Error loading configuration. Exiting application.")
	}

	// Initialize internal states
	internalstates.InitializeInternalStates()

	// Load function definitions
	err = functiondefinitions.ExtractFunctionDefinitionsFromPackage(externalFunctionsFile)
	if err != nil {
		log.Fatalf("Error extracting function definitions from package: %v", err)
	}

	// start the gRPC server
	grpcserver.StartServer()
	log.Fatalf("Error in gRPC server. Exiting application.")
}
