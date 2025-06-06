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

package main

import (
	_ "embed"

	"github.com/ansys/aali-sharedtypes/pkg/config"
	"github.com/ansys/aali-sharedtypes/pkg/logging"

	"github.com/ansys/aali-flowkit/pkg/functiondefinitions"
	"github.com/ansys/aali-flowkit/pkg/grpcserver"
	"github.com/ansys/aali-flowkit/pkg/internalstates"
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

//go:embed pkg/externalfunctions/ansysmaterials.go
var ansysMaterialsFile string

//go:embed pkg/externalfunctions/auth.go
var authFile string

func init() {
	// initialize config
	config.InitConfig([]string{}, map[string]interface{}{
		"SERVICE_NAME":        "aali-flowkit",
		"VERSION":             "1.0",
		"STAGE":               "PROD",
		"LOG_LEVEL":           "error",
		"ERROR_FILE_LOCATION": "error.log",
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
		"ansys_materials":  ansysMaterialsFile,
		"auth":             authFile,
	}

	// Load function definitions
	for category, file := range files {
		err := functiondefinitions.ExtractFunctionDefinitionsFromPackage(file, category)
		if err != nil {
			logging.Log.Fatalf(&logging.ContextMap{}, "Error extracting function definitions from package: %v", err)
		}
	}

	// Start the gRPC server
	grpcserver.StartServer()
	logging.Log.Fatalf(&logging.ContextMap{}, "Error in gRPC server. Exiting application.")
}
