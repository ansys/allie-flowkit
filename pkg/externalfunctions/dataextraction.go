package externalfunctions

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ansys/allie-flowkit/pkg/internalstates"
	"github.com/ansys/allie-flowkit/pkg/privatefunctions/codegeneration"
	"github.com/ansys/allie-flowkit/pkg/privatefunctions/milvus"
	"github.com/ansys/allie-sharedtypes/pkg/config"
	"github.com/ansys/allie-sharedtypes/pkg/logging"
	"github.com/ansys/allie-sharedtypes/pkg/sharedtypes"
	"github.com/google/go-github/v56/github"
	"github.com/google/uuid"
	"github.com/pandodao/tokenizer-go"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

// GetGithubFilesToExtract gets all files from github that need to be extracted.
//
// Tags:
//   - @displayName: List Github Files
//
// Parameters:
//   - githubRepoName: name of the github repository.
//   - githubRepoOwner: owner of the github repository.
//   - githubRepoBranch: branch of the github repository.
//   - githubAccessToken: access token for github.
//   - githubFileExtensions: github file extensions.
//   - githubFilteredDirectories: github filtered directories.
//   - githubExcludedDirectories: github excluded directories.
//
// Returns:
//   - githubFilesToExtract: github files to extract.
func GetGithubFilesToExtract(githubRepoName string, githubRepoOwner string,
	githubRepoBranch string, githubAccessToken string, githubFileExtensions []string,
	githubFilteredDirectories []string, githubExcludedDirectories []string) (githubFilesToExtract []string) {
	// If github repo name is empty, return empty list.
	if githubRepoName == "" {
		return githubFilesToExtract
	}

	client, ctx := dataExtractNewGithubClient(githubAccessToken)

	// Retrieve the specified branch SHA (commit hash) from the GitHub repository. This is used to identify the latest state of the branch.
	branch, _, err := client.Repositories.GetBranch(ctx, githubRepoOwner, githubRepoName, githubRepoBranch, 1)
	if err != nil {
		errMessage := fmt.Sprintf("Error getting branch %s: %v", githubRepoBranch, err)
		logging.Log.Error(internalstates.Ctx, errMessage)
		panic(errMessage)
	}

	// Extract the SHA from the branch information.
	sha := *branch.Commit.SHA

	// Retrieve the Git tree associated with the SHA. This tree represents the directory structure (files and subdirectories) of the repository at the specified SHA.
	tree, _, err := client.Git.GetTree(ctx, githubRepoOwner, githubRepoName, sha, true)
	if err != nil {
		errMessage := fmt.Sprintf("Error getting tree: %v", err)
		logging.Log.Error(internalstates.Ctx, errMessage)
		panic(errMessage)
	}

	// Extract the files that need to be extracted from the tree.
	githubFilesToExtract = dataExtractionFilterGithubTreeEntries(tree, githubFilteredDirectories, githubExcludedDirectories, githubFileExtensions)

	// Log the files that need to be extracted.
	for _, file := range githubFilesToExtract {
		logging.Log.Debugf(internalstates.Ctx, "Github file to extract: %s \n", file)
	}

	return githubFilesToExtract
}

// GetLocalFilesToExtract gets all files from local that need to be extracted.
//
// Tags:
//   - @displayName: List Local Files
//
// Parameters:
//   - localPath: path to the local directory.
//   - localFileExtensions: local file extensions.
//   - localFilteredDirectories: local filtered directories.
//   - localExcludedDirectories: local excluded directories.
//
// Returns:
//   - localFilesToExtract: local files to extract.
func GetLocalFilesToExtract(localPath string, localFileExtensions []string,
	localFilteredDirectories []string, localExcludedDirectories []string) (localFilesToExtract []string) {
	// If local path is empty, return empty list.
	if localPath == "" {
		return localFilesToExtract
	}

	// Check if the local path exists.
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		errMessage := fmt.Sprintf("Local path does not exist: %s", localPath)
		logging.Log.Error(internalstates.Ctx, errMessage)
		panic(errMessage)
	}

	localFiles := &[]string{}

	// Create a walker function that will be called for each file and directory found.
	walkFn := func(localPath string, f os.FileInfo, err error) error {
		return dataExtractionLocalFilepathExtractWalker(localPath, localFileExtensions, localFilteredDirectories, localExcludedDirectories,
			localFiles, f, err)
	}

	// Walk through all files and directories executing the walker function.
	err := filepath.Walk(localPath, walkFn)
	if err != nil {
		errMessage := fmt.Sprintf("Error walking through the files: %v", err)
		logging.Log.Error(internalstates.Ctx, errMessage)
		panic(errMessage)
	}

	// Log the files that need to be extracted.
	for _, file := range *localFiles {
		logging.Log.Debugf(internalstates.Ctx, "Local file to extract: %s \n", file)
	}

	return *localFiles
}

// AppendStringSlices creates a new slice by appending all elements of the provided slices.
//
// Tags:
//   - @displayName: Append String Slices
//
// Parameters:
//   - slice1, slice2, slice3, slice4, slice5: slices to append.
//
// Returns:
//   - result: a new slice with all elements appended.
func AppendStringSlices(slice1, slice2, slice3, slice4, slice5 []string) []string {
	var result []string

	// Append all elements from each slice to the result slice
	result = append(result, slice1...)
	result = append(result, slice2...)
	result = append(result, slice3...)
	result = append(result, slice4...)
	result = append(result, slice5...)

	return result
}

// DownloadGithubFileContent downloads file content from github and returns checksum and content.
//
// Tags:
//   - @displayName: Download Github File Content
//
// Parameters:
//   - githubRepoName: name of the github repository.
//   - githubRepoOwner: owner of the github repository.
//   - githubRepoBranch: branch of the github repository.
//   - gihubFilePath: path to file in the github repository.
//   - githubAccessToken: access token for github.
//
// Returns:
//   - checksum: checksum of file.
//   - content: content of file.
func DownloadGithubFileContent(githubRepoName string, githubRepoOwner string,
	githubRepoBranch string, gihubFilePath string, githubAccessToken string) (checksum string, content []byte) {

	// Create a new GitHub client and context.
	client, ctx := dataExtractNewGithubClient(githubAccessToken)

	// Retrieve the file content from the GitHub repository.
	fileContent, _, _, err := client.Repositories.GetContents(ctx, githubRepoOwner, githubRepoName, gihubFilePath, &github.RepositoryContentGetOptions{Ref: githubRepoBranch})
	if err != nil {
		errMessage := fmt.Sprintf("Error getting file content from github file %v: %v", gihubFilePath, err)
		logging.Log.Error(internalstates.Ctx, errMessage)
		panic(errMessage)
	}

	// Extract the content from the file content.
	stringContent, err := fileContent.GetContent()
	if err != nil {
		errMessage := fmt.Sprintf("Error getting file content from github file %v: %v", gihubFilePath, err)
		logging.Log.Error(internalstates.Ctx, errMessage)
		panic(errMessage)
	}

	// Extract the checksum from the file content.
	checksum = fileContent.GetSHA()

	// Convert the content to a byte slice.
	content = []byte(stringContent)

	logging.Log.Debugf(internalstates.Ctx, "Got content from github file: %s", gihubFilePath)

	return checksum, content
}

// GetLocalFileContent reads local file and returns checksum and content.
//
// Tags:
//   - @displayName: Get Local File Content
//
// Parameters:
//   - localFilePath: path to file.
//
// Returns:
//   - checksum: checksum of file.
//   - content: content of file.
func GetLocalFileContent(localFilePath string) (checksum string, content []byte) {
	// Read file from local path.
	content, err := os.ReadFile(localFilePath)
	if err != nil {
		errMessage := fmt.Sprintf("Error getting local file content: %v", err)
		logging.Log.Error(internalstates.Ctx, errMessage)
		panic(errMessage)
	}

	// Calculate checksum from file content.
	hash := sha256.New()
	_, err = hash.Write(content)
	if err != nil {
		errMessage := fmt.Sprintf("Error getting local file content: %v", err)
		logging.Log.Error(internalstates.Ctx, errMessage)
		panic(errMessage)
	}

	// Convert checksum to a hexadecimal string.
	checksum = hex.EncodeToString(hash.Sum(nil))

	logging.Log.Debugf(internalstates.Ctx, "Got content from local file: %s", localFilePath)

	return checksum, content
}

// GetDocumentType returns the document type of a file.
//
// Tags:
//   - @displayName: Get Document Type
//
// Parameters:
//   - filePath: path to file.
//
// Returns:
//   - documentType: file extension.
func GetDocumentType(filePath string) (documentType string) {
	// Extract the file extension from the file path and remove the leading period.
	fileExtension := filepath.Ext(filePath)
	documentType = strings.TrimPrefix(fileExtension, ".")

	return documentType
}

// LangchainSplitter splits content into chunks using langchain.
//
// Tags:
//   - @displayName: Split Content
//
// Parameters:
//   - content: content to split.
//   - documentType: type of document.
//   - chunkSize: size of the chunks.
//   - chunkOverlap: overlap of the chunks.
//
// Returns:
//   - output: chunks as an slice of strings.
func LangchainSplitter(bytesContent []byte, documentType string, chunkSize int, chunkOverlap int) (output []string) {
	output = []string{}
	var splittedChunks []schema.Document
	var err error

	// Creating a reader from the content of the file.
	reader := bytes.NewReader(bytesContent)

	// Creating a splitter with the chunk size and overlap specified in the config file.
	splitterOptions := []textsplitter.Option{}
	splitterOptions = append(splitterOptions, textsplitter.WithChunkSize(chunkSize))
	splitterOptions = append(splitterOptions, textsplitter.WithChunkOverlap(chunkOverlap))
	splitter := textsplitter.NewTokenSplitter(splitterOptions...)

	// Loading the content of the file and splitting it into chunks.
	switch documentType {
	case "html":
		htmlLoader := documentloaders.NewHTML(reader)
		splittedChunks, err = htmlLoader.LoadAndSplit(context.Background(), splitter)
		if err != nil {
			errMessage := fmt.Sprintf("Error getting file content from github: %v", err)
			logging.Log.Error(internalstates.Ctx, errMessage)
			panic(errMessage)
		}

		for _, chunk := range splittedChunks {
			output = append(output, chunk.PageContent)
		}

	case "py", "ipynb":
		output, err = dataExtractionPerformSplitterRequest(bytesContent, "py", chunkSize, chunkOverlap)
		if err != nil {
			errMessage := fmt.Sprintf("Error splitting python document: %v", err)
			logging.Log.Error(internalstates.Ctx, errMessage)
			panic(errMessage)
		}

	case "pdf":
		output, err = dataExtractionPerformSplitterRequest(bytesContent, "pdf", chunkSize, chunkOverlap)
		if err != nil {
			errMessage := fmt.Sprintf("Error splitting pdf document: %v", err)
			logging.Log.Error(internalstates.Ctx, errMessage)
			panic(errMessage)
		}

	case "pptx", "ppt":
		output, err = dataExtractionPerformSplitterRequest(bytesContent, "ppt", chunkSize, chunkOverlap)
		if err != nil {
			errMessage := fmt.Sprintf("Error splitting ppt document: %v", err)
			logging.Log.Error(internalstates.Ctx, errMessage)
			panic(errMessage)
		}

	default:
		// Default document type is text.
		txtLoader := documentloaders.NewText(reader)
		splittedChunks, err = txtLoader.LoadAndSplit(context.Background(), splitter)
		if err != nil {
			errMessage := fmt.Sprintf("Error getting file content from github: %v", err)
			logging.Log.Error(internalstates.Ctx, errMessage)
			panic(errMessage)
		}

		for _, chunk := range splittedChunks {
			output = append(output, chunk.PageContent)
		}
	}

	// Log number of chunks created.
	logging.Log.Debugf(internalstates.Ctx, "Splitted document in %v chunks \n", len(output))

	return output
}

// GenerateDocumentTree generates a tree structure from the document chunks.
//
// Tags:
//   - @displayName: Document Tree
//
// Parameters:
//   - documentName: name of the document.
//   - documentId: id of the document.
//   - documentChunks: chunks of the document.
//   - embeddingsDimensions: dimensions of the embeddings.
//   - getSummary: whether to get summary.
//   - getKeywords: whether to get keywords.
//   - numKeywords: number of keywords.
//   - chunkSize: size of the chunks.
//   - numLlmWorkers: number of llm workers.
//
// Returns:
//   - documentData: tree structure of the document.
func GenerateDocumentTree(documentName string, documentId string, documentChunks []string,
	embeddingsDimensions int, getSummary bool, getKeywords bool, numKeywords int, chunkSize int, numLlmWorkers int) (returnedDocumentData []sharedtypes.DbData) {

	logging.Log.Debugf(internalstates.Ctx, "Processing document: %s with %v leaf chunks \n", documentName, len(documentChunks))

	// Create llm handler input channel and wait group.
	llmHandlerInputChannel := make(chan *DataExtractionLLMInputChannelItem, 40)
	llmHandlerWaitGroup := sync.WaitGroup{}
	errorChannel := make(chan error, 1)

	// Start LLM Handler workers.
	for i := 0; i < numLlmWorkers; i++ {
		llmHandlerWaitGroup.Add(1)
		go dataExtractionLLMHandlerWorker(&llmHandlerWaitGroup, llmHandlerInputChannel, errorChannel, embeddingsDimensions)
	}

	// Create root data object.
	rootData := &sharedtypes.DbData{
		Guid:         "d" + strings.ReplaceAll(uuid.New().String(), "-", ""),
		DocumentId:   documentId,
		DocumentName: documentName,
		Text:         "",
		Summary:      "",
		Embedding:    make([]float32, embeddingsDimensions),
		ChildIds:     make([]string, 0, len(documentChunks)),
		Level:        "root",
	}

	// Assign non zero value to embedding so databae does not ignore the node.
	for i := range rootData.Embedding {
		rootData.Embedding[i] = 0.5
	}

	// Add root data object to document data.
	documentData := []*sharedtypes.DbData{rootData}

	// Create child data objects.
	orderedChildDataObjects, err := dataExtractionDocumentLevelHandler(llmHandlerInputChannel, errorChannel, documentChunks, documentId, documentName, getSummary, getKeywords, uint32(numKeywords))
	if err != nil {
		panic(err.Error())
	}

	// If summary is disabled -> flat structure, only iterate over chunks.
	if !getSummary {
		for _, childData := range orderedChildDataObjects {
			rootData.ChildIds = append(rootData.ChildIds, childData.Guid)
			childData.ParentId = rootData.Guid
			childData.Level = "leaf"
			documentData = append(documentData, childData)
		}

		// Assign first and last child ids to root data object.
		if len(orderedChildDataObjects) > 0 {
			rootData.FirstChildId = orderedChildDataObjects[0].Guid
			rootData.LastChildId = orderedChildDataObjects[len(orderedChildDataObjects)-1].Guid
		}
	}

	// If summary is enabled -> create summary and iterate over branches.
	if getSummary {
		// Prepare leaf data as not part of loop
		for _, childData := range orderedChildDataObjects {
			rootData.ChildIds = append(rootData.ChildIds, childData.Guid)
			childData.Level = "leaf"
			documentData = append(documentData, childData)
		}

		for {
			// Concatenate all summaries.
			branches := []*DataExtractionBranch{}
			branch := &DataExtractionBranch{
				Text:             "",
				ChildDataObjects: []*sharedtypes.DbData{},
			}
			branches = append(branches, branch)

			// Create branches from orderedChildDataObjects (based on summary length).
			for _, data := range orderedChildDataObjects {
				// Check whether summary is longer than allowed chunk length if yes, create new branch.
				branchTokenLength := tokenizer.MustCalToken(branch.Text)
				chunkSummaryTokenLength := tokenizer.MustCalToken(data.Summary)
				if branchTokenLength+chunkSummaryTokenLength > chunkSize {
					branch = &DataExtractionBranch{
						Text:             "",
						ChildDataObjects: []*sharedtypes.DbData{},
						ChildDataIds:     []string{},
					}
					branches = append(branches, branch)
				}

				branch.Text += data.Summary
				branch.ChildDataObjects = append(branch.ChildDataObjects, data)
				branch.ChildDataIds = append(branch.ChildDataIds, data.Guid)
			}

			// Text chunks are text parts from branches.
			textChunks := make([]string, 0, len(branches))
			for _, branch := range branches {
				textChunks = append(textChunks, branch.Text)
			}

			orderedChildDataObjectsFromBranches, err := dataExtractionDocumentLevelHandler(llmHandlerInputChannel, errorChannel, textChunks, documentId, documentName, getSummary, getKeywords, uint32(numKeywords))
			if err != nil {
				panic(err.Error())
			}

			// Exit if only one -> assign details to root.
			if len(orderedChildDataObjectsFromBranches) == 1 {
				// If root text has a title, append the child summaries to it.
				if rootData.Text != "" {
					rootData.Text += "\n" + orderedChildDataObjectsFromBranches[0].Text
				} else {
					rootData.Text = orderedChildDataObjectsFromBranches[0].Text
				}

				rootData.Summary = orderedChildDataObjectsFromBranches[0].Summary
				rootData.Embedding = orderedChildDataObjectsFromBranches[0].Embedding
				rootData.Keywords = orderedChildDataObjectsFromBranches[0].Keywords
				rootData.ChildIds = branches[0].ChildDataIds

				// Assign parent id to child data objects.
				for _, childData := range branches[0].ChildDataObjects {
					childData.ParentId = rootData.Guid
				}

				// Assign first and last child ids to root data object.
				if len(branches[0].ChildDataIds) > 0 {
					rootData.FirstChildId = branches[0].ChildDataIds[0]
					rootData.LastChildId = branches[0].ChildDataIds[len(branches[0].ChildDataIds)-1]
				}

				// Exit loop because top of the tree has been reached.
				break
			}

			// Assign details to parent data objects.
			for branchIdx, branch := range branches {
				parentData := orderedChildDataObjectsFromBranches[branchIdx]
				parentData.ChildIds = branch.ChildDataIds
				parentData.Level = "internal"

				// Assign first and last child ids to parent data object.
				if len(branch.ChildDataIds) > 0 {
					parentData.FirstChildId = branch.ChildDataIds[0]
					parentData.LastChildId = branch.ChildDataIds[len(branch.ChildDataIds)-1]
				}

				// Assign parent id to child data objects.
				for _, childData := range branch.ChildDataObjects {
					childData.ParentId = parentData.Guid
				}

				// Add parent data object to document data.
				documentData = append(documentData, parentData)
			}

			orderedChildDataObjects = orderedChildDataObjectsFromBranches
		}
	}

	// Send batch embedding request to LLM handler. Set max batch size to 1000.
	maxBatchSize := 100
	err = dataExtractionProcessBatchEmbeddings(documentData, maxBatchSize)
	if err != nil {
		errMessage := fmt.Sprintf("Error in dataExtractionProcessBatchEmbeddings: %v", err)
		logging.Log.Error(internalstates.Ctx, errMessage)
		panic(errMessage)
	}

	logging.Log.Debugf(internalstates.Ctx, "Finished processing document: %s \n", documentName)

	// Copy document data to returned document data
	returnedDocumentData = make([]sharedtypes.DbData, len(documentData))
	for i, data := range documentData {
		returnedDocumentData[i] = *data
	}

	// Close llm handler input channel and wait for all workers to finish.
	close(llmHandlerInputChannel)
	llmHandlerWaitGroup.Wait()

	return returnedDocumentData
}

func LoadMechanicalObjectDefinitions(path string) (elements []codegeneration.CodeGenerationElement) {
	// Read file from local path.
	content, err := os.ReadFile(path)
	if err != nil {
		errMessage := fmt.Sprintf("Error getting local file content: %v", err)
		logging.Log.Error(internalstates.Ctx, errMessage)
		panic(errMessage)
	}

	// Create object definition document.
	objectDefinitionDoc := codegeneration.MechanicalObjectDefinitionDocument{}

	// Unmarshal the XML content into the object definition document.
	err = xml.Unmarshal([]byte(content), &objectDefinitionDoc)
	if err != nil {
		errMessage := fmt.Sprintf("Error unmarshalling object definition document: %v", err)
		logging.Log.Error(internalstates.Ctx, errMessage)
		panic(errMessage)
	}

	for _, objectDefinition := range objectDefinitionDoc.Members {
		// If object name contains `` ignore it.
		if strings.Contains(objectDefinition.Name, "``") {
			continue
		}

		// Extract the prefix and name from the object definition.
		prefix := strings.Split(objectDefinition.Name, ":")[0]
		name := strings.Split(objectDefinition.Name, ":")[1]

		// Create the code generation element.
		element := codegeneration.CodeGenerationElement{
			Guid:       "d" + strings.ReplaceAll(uuid.New().String(), "-", ""),
			Name:       name,
			Summary:    objectDefinition.Summary,
			ReturnType: objectDefinition.ReturnType,
			Example:    objectDefinition.Example,
			Parameters: objectDefinition.Params,
			Remarks:    objectDefinition.Remarks,
		}

		switch prefix {
		case "M":
			element.Type = codegeneration.CodeGenerationType(codegeneration.Method)

			// Extract dependencies for method.
			dependencies := strings.Split(element.Name, "(")
			dependencies = strings.Split(dependencies[0], ".")
			element.Dependencies = dependencies

		case "P":
			element.Type = codegeneration.CodeGenerationType(codegeneration.Parameter)

			// Extract dependencies for parameter.
			dependencies := strings.Split(element.Name, ".")
			dependencies = dependencies[:len(dependencies)-1]
			element.Dependencies = dependencies

		case "F":
			element.Type = codegeneration.CodeGenerationType(codegeneration.Function)

			// Extract dependencies for function.
			dependencies := strings.Split(element.Name, "(")
			dependencies = strings.Split(dependencies[0], ".")
			element.Dependencies = dependencies

		case "T":
			element.Type = codegeneration.CodeGenerationType(codegeneration.Class)

			// Extract dependencies for class.
			dependencies := strings.Split(element.Name, ".")
			dependencies = dependencies[:len(dependencies)-1]
			element.Dependencies = dependencies

		case "E":
			// Ignore for now.

		default:
			errMessage := fmt.Sprintf("Unknown prefix: %s", prefix)
			logging.Log.Error(internalstates.Ctx, errMessage)
			panic(errMessage)
		}

		elements = append(elements, element)
	}

	return elements
}

func GeneratePseudocodeFromCodeGenerationFunctions(functions []codegeneration.CodeGenerationElement, pseudoCodeGenPrompt string, systemPrompt string, workers int) (completeElementDefinitions []codegeneration.CodeGenerationElement) {
	llmChannel := make(chan codegeneration.CodeGenerationElement, len(functions)) // Channel for functions to process
	errorChannel := make(chan error, 1)
	llmWaitGroup := sync.WaitGroup{}

	// Start LLM Handler workers.
	for i := 0; i < workers; i++ {
		llmWaitGroup.Add(1)
		go func() {
			defer llmWaitGroup.Done()
			for function := range llmChannel {
				// If type is not "function" or "method", ignore it.
				if function.Type != codegeneration.Function && function.Type != codegeneration.Method {
					function.NamePseudocode = function.Name
					function.Description = function.Summary
					completeElementDefinitions = append(completeElementDefinitions, function)
					continue
				}

				// prompt formatting depending on function
				parametersJSON, err := json.Marshal(function.Parameters)
				if err != nil {
					errMessage := fmt.Sprintf("Error marshalling function parameters: %v", err)
					logging.Log.Error(internalstates.Ctx, errMessage)
					errorChannel <- fmt.Errorf(errMessage)
				}

				exampleJSON, err := json.Marshal(function.Example)
				if err != nil {
					errMessage := fmt.Sprintf("Error marshalling function example: %v", err)
					logging.Log.Error(internalstates.Ctx, errMessage)
					errorChannel <- fmt.Errorf(errMessage)
				}

				valuesToFormat := map[string]string{
					"name":       function.Name,
					"parameters": string(parametersJSON),
					"summary":    function.Summary,
					"example":    string(exampleJSON),
				}

				prompt := formatTemplate(pseudoCodeGenPrompt, valuesToFormat)

				response, _, err := performGeneralRequest(prompt, []sharedtypes.HistoricMessage{}, false, systemPrompt, &sharedtypes.ModelOptions{})
				if err != nil {
					errorChannel <- err // Report errors
				}

				jsonContent := response
				// // Extract the start and end positions of the JSON content
				// start := strings.Index(response, "```json")
				// end := strings.LastIndex(response, "```")
				// fmt.Println("Start: ", start)
				// fmt.Println("Response: " + response)

				// if start == -1 || end == -1 || start >= end {
				// 	errMessage := fmt.Sprintf("Error extracting JSON content from response: %v", err)
				// 	logging.Log.Error(internalstates.Ctx, errMessage)
				// 	errorChannel <- fmt.Errorf(errMessage)
				// }

				// // Extract JSON content between the ```json block
				// jsonContent := response[start+len("```json") : end]
				// jsonContent = codegeneration.RemoveEmptyLines(jsonContent)
				// fmt.Println("JSON Content: ")
				// fmt.Println(jsonContent)
				// fmt.Println("-----------------")

				// Unmarshal the extracted JSON content
				codeGenerationPseudocodeResponse := codegeneration.CodeGenerationPseudocodeResponse{}
				err = json.Unmarshal([]byte(jsonContent), &codeGenerationPseudocodeResponse)
				if err != nil {
					errMessage := fmt.Sprintf("Error unmarshalling response: %v", err)
					logging.Log.Error(internalstates.Ctx, errMessage)
					errorChannel <- fmt.Errorf(errMessage)
				}

				// assign new signature and description to function
				function.NamePseudocode = codeGenerationPseudocodeResponse.Signature
				function.Description = codeGenerationPseudocodeResponse.Description

				completeElementDefinitions = append(completeElementDefinitions, function)
			}
		}()
	}

	// Add all functions to the channel.
	for _, function := range functions {
		llmChannel <- function
	}
	close(llmChannel) // Close the channel to signal workers no more items are coming

	// Wait for all workers to finish.
	llmWaitGroup.Wait()

	// Check for errors if needed.
	close(errorChannel)
	for err := range errorChannel {
		errMessage := fmt.Sprintf("Error marshalling function parameters: %v", err)
		logging.Log.Error(internalstates.Ctx, errMessage)
		panic(errMessage)
	}

	logging.Log.Infof(internalstates.Ctx, "Finished generating pseudocode for functions \n")

	return completeElementDefinitions
}

func StoreElementsInVectorDatabase(elements []codegeneration.CodeGenerationElement, elementsCollectionName string, batchSize int) error {
	// Set default batch size if not provided.
	if batchSize <= 0 {
		batchSize = 500
	}

	// Generate the embeddings for the elements
	embeddings, err := codeGenerationProcessBatchEmbeddings(elements, batchSize)
	if err != nil {
		return fmt.Errorf("failed to generate embeddings for elements: %w", err)
	}

	// Create the vector database objects.
	vectorElements := []codegeneration.VectorDatabaseElement{}
	for i, element := range elements {
		// Create a new vector database object.
		vectorElement := codegeneration.VectorDatabaseElement{
			Guid:           element.Guid,
			Vector:         embeddings[i],
			NamePseudocode: element.NamePseudocode,
			Description:    element.Description,
			Type:           string(element.Type),
		}

		// Add the new vector database object to the list.
		vectorElements = append(vectorElements, vectorElement)
	}

	// Initialize the vector database.
	milvusClient, err := milvus.Initialize()
	if err != nil {
		errMessage := "error initializing the vector database"
		logging.Log.Errorf(internalstates.Ctx, "%s: %v", errMessage, err)
		return fmt.Errorf("%s: %v", errMessage, err)
	}

	// Create the schema for this collection
	schemaFields := []milvus.SchemaField{
		{
			Name: "guid",
			Type: "string",
		},
		{
			Name:      "vector",
			Type:      "[]float32",
			Dimension: config.GlobalConfig.EMBEDDINGS_DIMENSIONS,
		},
		{
			Name: "type",
			Type: "string",
		},
		{
			Name: "name_pseudocode",
			Type: "string",
		},
		{
			Name: "description",
			Type: "string",
		},
	}

	schema, err := milvus.CreateCustomSchema(elementsCollectionName, schemaFields, "collection for code generation elements")
	if err != nil {
		errMessage := "error creating the schema"
		logging.Log.Errorf(internalstates.Ctx, "%s: %v", errMessage, err)
		return fmt.Errorf("%s: %v", errMessage, err)
	}

	// Create the collection.
	err = milvus.CreateCollection(schema, milvusClient)
	if err != nil {
		errMessage := "error creating the collection"
		logging.Log.Errorf(internalstates.Ctx, "%s: %v", errMessage, err)
		return fmt.Errorf("%s: %v", errMessage, err)
	}

	// Convert []VectorDatabaseElement to []interface{}
	// elementsAsInterface := make([]interface{}, len(vectorElements))
	// for i, v := range vectorElements {
	// 	elementsAsInterface[i] = v
	// }

	// Insert the elements into the vector database.
	err = milvus.InsertData(elementsCollectionName, vectorElements)
	if err != nil {
		errMessage := "error inserting data into the vector database"
		logging.Log.Errorf(internalstates.Ctx, "%s: %v", errMessage, err)
		return fmt.Errorf("%s: %v", errMessage, err)
	}

	return nil
}
