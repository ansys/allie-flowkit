# Adding New Functions, Types, and Categories

## 1. Adding a New Function

### Step 1: Define the Function
Define the new function in the appropriate package. If you were adding a function related to data extraction, you would need to add it to the `data_extraction` package.

Example:
```go
// File: allie-flowkit/pkg/externalfunctions/data_extraction.go

package externalfunctions

// TransformData processes input data and returns a transformed result.
// . . .
func TransformData(dataform string, depth int) (transformed string, err error) {
    // Function implementation
    return "transformed_data", nil
}
```

### Step 2: Incorperate the Function
Add the newly defined function to the `externalfunctions.go` file. Any newer functions unrelated to an existing file within `externalfunctions/` can be created and incorperated if necessary.

Now, we **must** add this newly defined method to externalfunctions.

Example (continued):
```go
// File: allie-flowkit/pkg/externalfunctions/externalfunctions.go
var ExternalFunctionsMap = map[string]interface{}{
	// llm handler
	"PerformVectorEmbeddingRequest":                                   PerformVectorEmbeddingRequest,
	// . . . 
    
	// data extraction
	"GetGithubFilesToExtract":                    GetGithubFilesToExtract,
	"GetLocalFilesToExtract":                     GetLocalFilesToExtract,
	"AppendStringSlices":                         AppendStringSlices,
	"DownloadGithubFileContent":                  DownloadGithubFileContent,
	"GetLocalFileContent":                        GetLocalFileContent,
	"GetDocumentType":                            GetDocumentType,
	"LangchainSplitter":                          LangchainSplitter,
	"GenerateDocumentTree":                       GenerateDocumentTree,
	"AddDataRequest":                             AddDataRequest,
	"CreateCollectionRequest":                    CreateCollectionRequest,
	"CreateGeneralDataExtractionDocumentObjects": CreateGeneralDataExtractionDocumentObjects,
    "TransformData":                              TransformData, // New function added here

    // CONTINUED
}
```

## 2. Adding a New Type

Define your new type in the appropriate package within `types.go`.

Example:
```go
// File: allie-flowkit/pkg/externalfunctions/types.go

package externalfunctions

// RLAgent is an reinforcement learning agent :O
type RLAgent struct {
    Critique string
    Reward float64
}
```

## 3. Adding a New Category
### Step 1: Make a New Category File

In the `pkg/externalfunctions/` directory, if necessary, make an entirely new Go file for your function(s), e.g., `sft.go`. Ensure to adhere to previous sections to add newly defined functions and types with a new category.

### Step 2: Update the Main File

Update the `main.go` file to include the new category (if necessary) with the corresponding file.

Example:
```go
// File: allie-flowkit/main.go

func main() {
    // . . .

    //go:embed pkg/externalfunctions/milvus.go
    var sftFile string // add the string declaration

    // . . .

    // Create file list
    files := map[string]string{
        "data_extraction":  dataExtractionFile,
        "generic":          genericFile,
        "knowledge_db":     knowledgeDBFile,
        "llm_handler":      llmHandlerFile,
        "ansys_gpt":        ansysGPTFile,
        "milvus":           milvusFile,
        "ansys_mesh_pilot": ansysMeshPilotFile,
        "auth":             authFile,
        "sft":              sftFile, // Add the new category file here
    }

    // CONTINUED
}
```
