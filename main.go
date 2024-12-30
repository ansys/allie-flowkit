package main

import (
	_ "embed"

	"github.com/ansys/allie-sharedtypes/pkg/config"
	"github.com/ansys/allie-sharedtypes/pkg/logging"

	"github.com/ansys/allie-flowkit/pkg/externalfunctions"
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
	files := map[string]string{
		"data_extraction": dataExtractionFile,
		"generic":         genericFile,
		"knowledge_db":    knowledgeDBFile,
		"llm_handler":     llmHandlerFile,
		"ansys_gpt":       ansysGPTFile,
	}

	// Load function definitions
	for category, file := range files {
		err := functiondefinitions.ExtractFunctionDefinitionsFromPackage(file, category)
		if err != nil {
			logging.Log.Fatalf(internalstates.Ctx, "Error extracting function definitions from package: %v", err)
		}
	}

	// Log the version of the system
	logging.Log.Info(internalstates.Ctx, "Launching Allie Flowkit")

	TestCodeGenElements()
	TestCodeGenExamples()
	TestCodeGenUserGuide()

	// start the gRPC server
	grpcserver.StartServer()
	logging.Log.Fatalf(internalstates.Ctx, "Error in gRPC server. Exiting application.")
}

func TestCodeGenElements() {
	path := "./mechanical_def_complete.xml"

	// Load mechanical object definitions
	e := externalfunctions.LoadObjectDefinitions(path)

	// functionPrompt := `Please focus, this is really important to me. I have a {type} with this specifications:
	// Signature: {name}
	// Summary: {summary}
	// Parameters: {parameters}
	// Example: {example}
	// ReturnType: {returnType}

	// I want you to generate a short description of the function.
	// `
	// parameterPrompt := `Please focus, this is really important to me. I have a {type} with this specifications:
	// Name: {name}
	// Summary: {summary}
	// Type: {returnType}

	// I want you to generate a short description of the parameter.
	// `

	// systemPrompt := `You are a really helpful assistant`

	// // generate pseudo code
	// functionDef := externalfunctions.GeneratePseudocodeFromCodeGenerationFunctions(e, functionPrompt, parameterPrompt, systemPrompt, 20)

	// store in database
	embeddingsBatchSize := 200
	externalfunctions.StoreElementsInVectorDatabase(e, "mechanical_elements_collection", embeddingsBatchSize)
	externalfunctions.StoreElementsInGraphDatabase(e)
}

func TestCodeGenExamples() {
	path := "./mechanical_def_complete.xml"

	// Load mechanical object definitions
	e := externalfunctions.LoadObjectDefinitions(path)

	// Load examples dependencies
	dependenciesPath := "./example_dependencies.json"
	dependencies, equivalencesMap := externalfunctions.LoadAndCheckExampleDependencies(dependenciesPath, e)

	// Get examples to extract
	pathToExamples := "./examples"
	documentType := "py"
	chunkSize := 500
	chunkOverlap := 40
	examplesToExtract := externalfunctions.GetLocalFilesToExtract(pathToExamples, []string{documentType}, []string{}, []string{})

	// Create the CodeGenerationExample objects
	codeGenerationExamples := externalfunctions.LoadCodeGenerationExamples(examplesToExtract, dependencies, equivalencesMap, chunkSize, chunkOverlap)

	// store in database
	embeddingsBatchSize := 200
	externalfunctions.StoreExamplesInVectorDatabase(codeGenerationExamples, "mechanical_examples_collection", embeddingsBatchSize)
	externalfunctions.StoreExamplesInGraphDatabase(codeGenerationExamples)
}

func TestCodeGenUserGuide() {
	path := "./user_guide_structured"

	// Get the files to extract
	paths := externalfunctions.GetLocalFilesToExtract(path, []string{"json"}, []string{}, []string{})

	// Load the sections for all the files
	sections := externalfunctions.LoadUserGuideSections(paths)

	// store in database
	embeddingsBatchSize := 200
	chunkSize := 500
	chunkOverlap := 40
	externalfunctions.StoreUserGuideSectionsInVectorDatabase(sections, "mechanical_user_guide_collection", embeddingsBatchSize, chunkSize, chunkOverlap)
	externalfunctions.StoreUserGuideSectionsInGraphDatabase(sections)
}
