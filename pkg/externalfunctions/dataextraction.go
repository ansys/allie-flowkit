package externalfunctions

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ansys/allie-flowkit/pkg/privatefunctions/codegeneration"
	"github.com/ansys/allie-flowkit/pkg/privatefunctions/milvus"
	"github.com/ansys/allie-flowkit/pkg/privatefunctions/neo4j"
	"github.com/ansys/allie-sharedtypes/pkg/config"
	"github.com/ansys/allie-sharedtypes/pkg/logging"
	"github.com/ansys/allie-sharedtypes/pkg/sharedtypes"
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
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	// Extract the SHA from the branch information.
	sha := *branch.Commit.SHA

	// Retrieve the Git tree associated with the SHA. This tree represents the directory structure (files and subdirectories) of the repository at the specified SHA.
	tree, _, err := client.Git.GetTree(ctx, githubRepoOwner, githubRepoName, sha, true)
	if err != nil {
		errMessage := fmt.Sprintf("Error getting tree: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	// Extract the files that need to be extracted from the tree.
	githubFilesToExtract = dataExtractionFilterGithubTreeEntries(tree, githubFilteredDirectories, githubExcludedDirectories, githubFileExtensions)

	// Log the files that need to be extracted.
	for _, file := range githubFilesToExtract {
		logging.Log.Debugf(&logging.ContextMap{}, "Github file to extract: %s \n", file)
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
		logging.Log.Error(&logging.ContextMap{}, errMessage)
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
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	// Log the files that need to be extracted.
	for _, file := range *localFiles {
		logging.Log.Debugf(&logging.ContextMap{}, "Local file to extract: %s \n", file)
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

	checksum, content, err := downloadGithubFileContent(githubRepoName, githubRepoOwner, githubRepoBranch, gihubFilePath, githubAccessToken)
	if err != nil {
		errMessage := fmt.Sprintf("Error getting file content from github: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	return checksum, content
}

// DownloadGithubFilesContent downloads file content from github and returns checksum and content.
//
// Tags:
//   - @displayName: Download Github Files Content
//
// Parameters:
//   - githubRepoName: name of the github repository.
//   - githubRepoOwner: owner of the github repository.
//   - githubRepoBranch: branch of the github repository.
//   - gihubFilePath: path to file in the github repository.
//   - githubAccessToken: access token for github.
//
// Returns:
//   - filesMap: map of file paths to file content.
func DownloadGithubFilesContent(githubRepoName string, githubRepoOwner string,
	githubRepoBranch string, gihubFilePaths []string, githubAccessToken string) (filesMap map[string][]byte) {
	filesMap = make(map[string][]byte)

	for _, gihubFilePath := range gihubFilePaths {
		_, content, err := downloadGithubFileContent(githubRepoName, githubRepoOwner, githubRepoBranch, gihubFilePath, githubAccessToken)
		if err != nil {
			errMessage := fmt.Sprintf("Error getting file content from github: %v", err)
			logging.Log.Error(&logging.ContextMap{}, errMessage)
			panic(errMessage)
		}

		filesMap[gihubFilePath] = content
	}

	return filesMap
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
	// Get the checksum and content of the local file.
	checksum, content, err := getLocalFileContent(localFilePath)
	if err != nil {
		errMessage := fmt.Sprintf("Error getting file content from local: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	return checksum, content
}

// GetLocalFilesContent reads local files and returns content.
//
// Tags:
//   - @displayName: Get Local Files Content
//
// Parameters:
//   - localFilePaths: paths to files.
//
// Returns:
//   - filesMap: map of file paths to file content.
func GetLocalFilesContent(localFilePaths []string) (filesMap map[string][]byte) {
	filesMap = make(map[string][]byte)

	for _, localFilePath := range localFilePaths {
		_, content, err := getLocalFileContent(localFilePath)
		if err != nil {
			errMessage := fmt.Sprintf("Error getting file content from local: %v", err)
			logging.Log.Error(&logging.ContextMap{}, errMessage)
			panic(errMessage)
		}

		filesMap[localFilePath] = content
	}

	return filesMap
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
			logging.Log.Error(&logging.ContextMap{}, errMessage)
			panic(errMessage)
		}

		for _, chunk := range splittedChunks {
			output = append(output, chunk.PageContent)
		}

	case "py", "ipynb":
		output, err = dataExtractionPerformSplitterRequest(bytesContent, "py", chunkSize, chunkOverlap)
		if err != nil {
			errMessage := fmt.Sprintf("Error splitting python document: %v", err)
			logging.Log.Error(&logging.ContextMap{}, errMessage)
			panic(errMessage)
		}

	case "pdf":
		output, err = dataExtractionPerformSplitterRequest(bytesContent, "pdf", chunkSize, chunkOverlap)
		if err != nil {
			errMessage := fmt.Sprintf("Error splitting pdf document: %v", err)
			logging.Log.Error(&logging.ContextMap{}, errMessage)
			panic(errMessage)
		}

	case "pptx", "ppt":
		output, err = dataExtractionPerformSplitterRequest(bytesContent, "ppt", chunkSize, chunkOverlap)
		if err != nil {
			errMessage := fmt.Sprintf("Error splitting ppt document: %v", err)
			logging.Log.Error(&logging.ContextMap{}, errMessage)
			panic(errMessage)
		}

	default:
		// Default document type is text.
		txtLoader := documentloaders.NewText(reader)
		splittedChunks, err = txtLoader.LoadAndSplit(context.Background(), splitter)
		if err != nil {
			errMessage := fmt.Sprintf("Error getting file content from github: %v", err)
			logging.Log.Error(&logging.ContextMap{}, errMessage)
			panic(errMessage)
		}

		for _, chunk := range splittedChunks {
			output = append(output, chunk.PageContent)
		}
	}

	// Log number of chunks created.
	logging.Log.Debugf(&logging.ContextMap{}, "Splitted document in %v chunks \n", len(output))

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

	logging.Log.Debugf(&logging.ContextMap{}, "Processing document: %s with %v leaf chunks \n", documentName, len(documentChunks))

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
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	logging.Log.Debugf(&logging.ContextMap{}, "Finished processing document: %s \n", documentName)

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

// LoadCodeGenerationElements loads code generation elements from an xml or json file.
//
// Tags:
//   - @displayName: Load Code Generation Elements
//
// Parameters:
//   - content: content of the file in []byte format.
//   - elementsFilePath: path to the file.
//
// Returns:
//   - elements: code generation elements.
func LoadCodeGenerationElements(content []byte, elementsFilePath string) (elements []sharedtypes.CodeGenerationElement) {
	// Get the file extension.
	fileExtension := filepath.Ext(elementsFilePath)

	// Create object definition document.
	objectDefinitionDoc := codegeneration.XMLObjectDefinitionDocument{}

	var err error
	switch fileExtension {
	case ".xml":
		// Unmarshal the XML content into the object definition document.
		err = xml.Unmarshal([]byte(content), &objectDefinitionDoc)
		if err != nil {
			errMessage := fmt.Sprintf("Error unmarshalling object definition document: %v", err)
			logging.Log.Error(&logging.ContextMap{}, errMessage)
			panic(errMessage)
		}
	case ".json":
		// Unmarshal the JSON content into a list of assembly members.
		err = json.Unmarshal(content, &objectDefinitionDoc.Members)
		if err != nil {
			errMessage := fmt.Sprintf("Error unmarshalling object definition document: %v", err)
			logging.Log.Error(&logging.ContextMap{}, errMessage)
			panic(errMessage)
		}

	default:
		errMessage := fmt.Sprintf("Unknown file extension: %s", fileExtension)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	for _, objectDefinition := range objectDefinitionDoc.Members {
		// If object name contains `` ignore it.
		if strings.Contains(objectDefinition.Name, "``") {
			continue
		}

		// Extract the prefix and name from the object definition.
		prefix := strings.SplitN(objectDefinition.Name, ":", 2)[0]
		name := ""
		if parts := strings.SplitN(objectDefinition.Name, ":", 2); len(parts) > 1 {
			name = parts[1]
		}

		// Create the code generation element.
		element := sharedtypes.CodeGenerationElement{
			Guid:              "d" + strings.ReplaceAll(uuid.New().String(), "-", ""),
			Name:              name,
			Summary:           objectDefinition.Summary,
			ReturnType:        objectDefinition.ReturnType,
			Example:           objectDefinition.Example,
			Parameters:        objectDefinition.Params,
			Remarks:           objectDefinition.Remarks,
			ReturnDescription: objectDefinition.Returns,
		}

		// Create a list with all the return types of the element.
		element.ReturnElementList, err = codegeneration.CreateReturnList(objectDefinition.ReturnType)
		if err != nil {
			errMessage := fmt.Sprintf("Error creating return element list: %v", err)
			logging.Log.Error(&logging.ContextMap{}, errMessage)
			panic(errMessage)
		}

		switch prefix {
		case "M":
			element.Type = sharedtypes.CodeGenerationType(sharedtypes.Method)

			// Extract dependencies for method.
			dependencies := strings.Split(element.Name, "(")
			dependencies = strings.Split(dependencies[0], ".")
			dependencies = dependencies[:len(dependencies)-1]
			element.Dependencies = dependencies

		case "P":
			element.Type = sharedtypes.CodeGenerationType(sharedtypes.Parameter)

			// Extract dependencies for parameter.
			dependencies := strings.Split(element.Name, ".")
			dependencies = dependencies[:len(dependencies)-1]
			element.Dependencies = dependencies

		case "F":
			element.Type = sharedtypes.CodeGenerationType(sharedtypes.Function)

			// Extract dependencies for function.
			dependencies := strings.Split(element.Name, "(")
			dependencies = strings.Split(dependencies[0], ".")
			dependencies = dependencies[:len(dependencies)-1]
			element.Dependencies = dependencies

		case "T":
			element.Type = sharedtypes.CodeGenerationType(sharedtypes.Class)

			// Extract dependencies for class.
			dependencies := strings.Split(element.Name, ".")
			dependencies = dependencies[:len(dependencies)-1]
			element.Dependencies = dependencies

		case "E":
			element.Type = sharedtypes.CodeGenerationType(sharedtypes.Enum)

			// Extract dependencies for enum.
			dependencies := strings.Split(element.Name, ".")
			dependencies = dependencies[:len(dependencies)-1]
			element.Dependencies = dependencies

			// Extract enum values.
			cleaned := strings.Trim(objectDefinition.EnumValues, "[]")
			element.EnumValues = strings.Split(cleaned, ",")

		case "MOD":
			element.Type = sharedtypes.CodeGenerationType(sharedtypes.Module)

			// Extract dependencies for class.
			dependencies := strings.Split(element.Name, ".")
			dependencies = dependencies[:len(dependencies)-1]
			element.Dependencies = dependencies

		default:
			errMessage := fmt.Sprintf("Unknown prefix: %s", prefix)
			logging.Log.Error(&logging.ContextMap{}, errMessage)
			panic(errMessage)
		}

		// Get name pseudocode and formatted name.
		element.NamePseudocode, element.NameFormatted, err = codegeneration.ProcessElementName(element.Name, element.Dependencies)
		if err != nil {
			errMessage := fmt.Sprintf("Error processing element name: %v", err)
			logging.Log.Error(&logging.ContextMap{}, errMessage)
			panic(errMessage)
		}

		elements = append(elements, element)
	}

	logging.Log.Debugf(&logging.ContextMap{}, "Loaded %v code generation elements from file: %s", len(elements), elementsFilePath)

	return elements
}

// StoreElementsInVectorDatabase stores elements in the vector database.
//
// Tags:
//   - @displayName: Store Elements in Vector Database
//
// Parameters:
//   - elements: code generation elements.
//   - elementsCollectionName: name of the collection.
//   - batchSize: batch size for embeddings.
func StoreElementsInVectorDatabase(elements []sharedtypes.CodeGenerationElement, elementsCollectionName string, batchSize int) {
	// Set default batch size if not provided.
	if batchSize <= 0 {
		batchSize = 2
	}

	logging.Log.Debugf(&logging.ContextMap{}, "Storing %v code generation elements in the vector database", len(elements))

	// Generate dense and sparse embeddings
	denseEmbeddings, sparseEmbeddings, err := codeGenerationProcessHybridSearchEmbeddings(elements, batchSize)
	if err != nil {
		errMessage := fmt.Sprintf("Error generating embeddings for elements: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	// Create the vector database objects.
	vectorElements := []codegeneration.VectorDatabaseElement{}
	for i, element := range elements {
		// Create a new vector database object.
		vectorElement := codegeneration.VectorDatabaseElement{
			Guid:           element.Guid,
			DenseVector:    denseEmbeddings[i],
			SparseVector:   sparseEmbeddings[i],
			Name:           element.Name,
			NamePseudocode: element.NamePseudocode,
			NameFormatted:  element.NameFormatted,
			Type:           string(element.Type),
			ParentClass:    strings.Join(element.Dependencies, "."),
		}

		// Add the new vector database object to the list.
		vectorElements = append(vectorElements, vectorElement)
	}

	// Create the schema for this collection
	schemaFields := []milvus.SchemaField{
		{
			Name: "guid",
			Type: "string",
		},
		{
			Name: "dense_vector",
			Type: "[]float32",
		},
		{
			Name: "sparse_vector",
			Type: "map[uint]float32",
		},
		{
			Name: "type",
			Type: "string",
		},
		{
			Name: "name",
			Type: "string",
		},
		{
			Name: "name_pseudocode",
			Type: "string",
		},
		{
			Name: "name_formatted",
			Type: "string",
		},
		{
			Name: "parent_class",
			Type: "string",
		},
	}

	schema, err := milvus.CreateCustomSchema(elementsCollectionName, schemaFields, "collection for code generation elements")
	if err != nil {
		errMessage := fmt.Sprintf("Error creating the schema for the collection: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	// Create the collection.
	err = milvus.CreateCollection(schema)
	if err != nil {
		errMessage := fmt.Sprintf("Error creating the collection: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	// Convert []VectorDatabaseElement to []interface{}
	elementsAsInterface := make([]interface{}, len(vectorElements))
	for i, v := range vectorElements {
		elementsAsInterface[i] = v
	}

	// Insert the elements into the vector database.
	err = milvus.InsertData(elementsCollectionName, elementsAsInterface, "Name", "name")
	if err != nil {
		errMessage := fmt.Sprintf("Error inserting data into the vector database: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}
}

// StoreElementsInGraphDatabase stores elements in the graph database.
//
// Tags:
//   - @displayName: Store Elements in Graph Database
//
// Parameters:
//   - elements: code generation elements.
func StoreElementsInGraphDatabase(elements []sharedtypes.CodeGenerationElement) {
	// Initialize the graph database.
	neo4j.Initialize(config.GlobalConfig.NEO4J_URI, config.GlobalConfig.NEO4J_USERNAME, config.GlobalConfig.NEO4J_PASSWORD)

	// Add the elements to the graph database.
	neo4j.Neo4j_Driver.AddCodeGenerationElementNodes(elements)

	// Add the dependencies to the graph database.
	neo4j.Neo4j_Driver.CreateCodeGenerationRelationships(elements)
}

// LoadAndCheckExampleDependencies loads and checks the dependencies of the examples.
//
// Tags:
//   - @displayName: Load and Check Example Dependencies
//
// Parameters:
//   - dependenciesContent: content of the dependencies file in []byte format.
//   - elements: code generation elements.
//   - instancesReplacementDict: dictionary of instances replacements.
//   - InstancesReplacementPriorityList: list of instances replacement priority.
//
// Returns:
//   - checkedDependenciesMap: checked dependencies.
//   - equivalencesMap: equivalences.
func LoadAndCheckExampleDependencies(
	dependenciesContent []byte,
	elements []sharedtypes.CodeGenerationElement,
	instancesReplacementDict map[string]string,
	InstancesReplacementPriorityList []string,
) (checkedDependenciesMap map[string][]string, equivalencesMap map[string]map[string]string) {
	// Unmarshal the JSON content into the dependencies map.
	var dependenciesMap map[string][]string
	err := json.Unmarshal(dependenciesContent, &dependenciesMap)
	if err != nil {
		errMessage := fmt.Sprintf("Error unmarshalling dependencies: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	// Initialize maps.
	checkedDependenciesMap = make(map[string][]string)
	equivalencesMap = make(map[string]map[string]string)

	// Function to replace ExtAPI dependencies.
	replaceExtAPI := func(dependencies []string) ([]string, map[string]string) {
		updatedDependencies := make([]string, 0, len(dependencies))
		equivalences := make(map[string]string)
		for _, dependency := range dependencies {
			original := dependency
			for _, key := range InstancesReplacementPriorityList { // Iterate over keys in the desired priority order
				value := instancesReplacementDict[key]
				if strings.HasPrefix(dependency, key) {
					dependency = strings.Replace(dependency, key, value, 1) // Replace only the prefix
					break                                                   // Stop after the first match since keys are prefixes
				}
			}
			if original != dependency {
				equivalences[dependency] = original
			}
			updatedDependencies = append(updatedDependencies, dependency)
		}
		return updatedDependencies, equivalences
	}

	// Process dependencies.
	for key, dependencies := range dependenciesMap {
		updatedDependencies, equivalences := replaceExtAPI(dependencies)

		// Filter checked dependencies and populate the equivalences map accordingly.
		checkedDependencies := []string{}
		checkedEquivalences := make(map[string]string)
		for _, dependency := range updatedDependencies {
			matchFound := false

			// Check if the exact dependency exists in functions.
			for _, function := range elements {
				functionNameNoParams := strings.Split(function.Name, "(")[0]
				if functionNameNoParams == dependency {
					checkedDependencies = append(checkedDependencies, function.Name)
					if original, ok := equivalences[dependency]; ok {
						checkedEquivalences[dependency] = original
					}
					matchFound = true
					break
				}
			}

			// If no match, check for dependency without the last `.whatever` part.
			if !matchFound {
				lastDotIndex := strings.LastIndex(dependency, ".")
				if lastDotIndex != -1 {
					truncatedDependency := dependency[:lastDotIndex]
					for _, function := range elements {
						functionNameNoParams := strings.Split(function.Name, "(")[0]
						if functionNameNoParams == truncatedDependency {
							// Update dependency and equivalences.
							checkedDependencies = append(checkedDependencies, function.Name)
							if original, ok := equivalences[dependency]; ok {
								checkedEquivalences[truncatedDependency] = original[:strings.LastIndex(original, ".")]
							}
							matchFound = true
							break
						}
					}
				}
			}

			// If still no match, dependency remains unvalidated.
			if !matchFound {
				continue
			}
		}

		checkedDependenciesMap[key] = checkedDependencies
		equivalencesMap[key] = checkedEquivalences
	}

	// Final Step: Remove duplicates from both maps.
	deduplicate := func(slice []string) []string {
		unique := make(map[string]bool)
		result := []string{}
		for _, item := range slice {
			if !unique[item] {
				unique[item] = true
				result = append(result, item)
			}
		}
		return result
	}

	for key := range checkedDependenciesMap {
		checkedDependenciesMap[key] = deduplicate(checkedDependenciesMap[key])
	}

	for key, equivalences := range equivalencesMap {
		uniqueEquivalences := make(map[string]string)
		seen := make(map[string]bool)
		for newDep, original := range equivalences {
			if !seen[newDep] {
				seen[newDep] = true
				uniqueEquivalences[newDep] = original
			}
		}
		equivalencesMap[key] = uniqueEquivalences
	}

	return checkedDependenciesMap, equivalencesMap
}

// LoadCodeGenerationExamples loads code generation examples from the provided paths.
//
// Tags:
//   - @displayName: Load Code Generation Examples
//
// Parameters:
//   - source: source of the examples (local or github).
//   - examplesToExtract: paths to the examples.
//   - githubRepoName: name of the github repository.
//   - githubRepoOwner: owner of the github repository.
//   - githubRepoBranch: branch of the github repository.
//   - githubAccessToken: access token for the github repository.
//   - dependencies: dependencies of the examples.
//   - equivalencesMap: equivalences of the examples.
//   - chunkSize: size of the chunks.
//   - chunkOverlap: overlap of the chunks.
//
// Returns:
//   - examples: code generation examples.
func LoadCodeGenerationExamples(
	source string,
	examplesToExtract []string,
	githubRepoName string,
	githubRepoOwner string,
	githubRepoBranch string,
	githubAccessToken string,
	dependencies map[string][]string,
	equivalencesMap map[string]map[string]string,
	chunkSize int,
	chunkOverlap int) (examples []sharedtypes.CodeGenerationExample) {
	// Initialize the examples slice.
	examples = []sharedtypes.CodeGenerationExample{}

	// Load the examples from the provided paths.
	for _, examplePath := range examplesToExtract {
		// Read file from either local or github source.
		var content []byte
		var err error
		switch source {
		case "local":
			_, content, err = getLocalFileContent(examplePath)
			if err != nil {
				errMessage := fmt.Sprintf("Error getting local file content: %v", err)
				logging.Log.Error(&logging.ContextMap{}, errMessage)
				panic(errMessage)
			}
		case "github":
			_, content, err = downloadGithubFileContent(githubRepoName, githubRepoOwner, githubRepoBranch, examplePath, githubAccessToken)
			if err != nil {
				errMessage := fmt.Sprintf("Error getting github file content: %v", err)
				logging.Log.Error(&logging.ContextMap{}, errMessage)
				panic(errMessage)
			}
		default:
			errMessage := fmt.Sprintf("Unknown data source: %s", source)
			logging.Log.Error(&logging.ContextMap{}, errMessage)
			panic(errMessage)
		}

		// Create the chunks for the current element.
		chunks, err := dataExtractionTextSplitter(string(content), chunkSize, chunkOverlap)
		if err != nil {
			errMessage := fmt.Sprintf("Error splitting text into chunks: %v", err)
			logging.Log.Error(&logging.ContextMap{}, errMessage)
			panic(errMessage)
		}

		// The name should be only the file name
		fileName := filepath.Base(examplePath)

		// Check if the example has dependencies.
		exampleDependencies := []string{}
		if dependencies[fileName] != nil {
			exampleDependencies = dependencies[fileName]
		}

		// Create the object
		example := sharedtypes.CodeGenerationExample{
			Chunks:                 chunks,
			Name:                   fileName,
			Dependencies:           exampleDependencies,
			DependencyEquivalences: equivalencesMap[fileName],
		}

		// Add the example to the examples slice.
		examples = append(examples, example)
	}

	return examples
}

// StoreExamplesInVectorDatabase stores examples in the vector database.
//
// Tags:
//   - @displayName: Store Examples in Vector Database
//
// Parameters:
//   - examples: code generation examples.
//   - examplesCollectionName: name of the collection.
//   - batchSize: batch size for embeddings.
func StoreExamplesInVectorDatabase(examples []sharedtypes.CodeGenerationExample, examplesCollectionName string, batchSize int) {
	// Set default batch size if not provided.
	if batchSize <= 0 {
		batchSize = 2
	}

	// Create the schema for this collection
	schemaFields := []milvus.SchemaField{
		{
			Name: "guid",
			Type: "string",
		},
		{
			Name: "document_name",
			Type: "string",
		},
		{
			Name: "previous_chunk",
			Type: "string",
		},
		{
			Name: "next_chunk",
			Type: "string",
		},
		{
			Name: "dense_vector",
			Type: "[]float32",
		},
		{
			Name: "sparse_vector",
			Type: "map[uint]float32",
		},
		{
			Name: "text",
			Type: "string",
		},
		{
			Name: "dependency_equivalences",
			Type: "map[string]string",
		},
	}

	schema, err := milvus.CreateCustomSchema(examplesCollectionName, schemaFields, "collection for code generation examples")
	if err != nil {
		errMessage := "error creating the schema for the collection"
		logging.Log.Errorf(&logging.ContextMap{}, "%s: %v", errMessage, err)
		panic(fmt.Errorf("%s: %v", errMessage, err))
	}

	// Create the collection.
	err = milvus.CreateCollection(schema)
	if err != nil {
		errMessage := "error creating the collection"
		logging.Log.Errorf(&logging.ContextMap{}, "%s: %v", errMessage, err)
		panic(fmt.Errorf("%s: %v", errMessage, err))
	}

	// Create the vector database objects.
	vectorExamples := []codegeneration.VectorDatabaseExample{}
	for _, element := range examples {
		chunkGuids := make([]string, len(element.Chunks)) // Track GUIDs for all chunks in the current element

		// Generate GUIDs for each chunk in advance.
		for j := 0; j < len(element.Chunks); j++ {
			guid := "d" + strings.ReplaceAll(uuid.New().String(), "-", "")
			chunkGuids[j] = guid
		}

		// Create vector database objects and assign PreviousChunk and NextChunk.
		for j := 0; j < len(element.Chunks); j++ {
			vectorExample := codegeneration.VectorDatabaseExample{
				Guid:                   chunkGuids[j], // Current chunk's GUID
				DocumentName:           element.Name,
				PreviousChunk:          "", // Default empty
				NextChunk:              "", // Default empty
				Dependencies:           element.Dependencies,
				DependencyEquivalences: element.DependencyEquivalences,
				Text:                   element.Chunks[j],
			}

			// Assign PreviousChunk and NextChunk GUIDs.
			if j > 0 {
				vectorExample.PreviousChunk = chunkGuids[j-1]
			}
			if j < len(element.Chunks)-1 {
				vectorExample.NextChunk = chunkGuids[j+1]
			}

			// Add the new vector database object to the list.
			vectorExamples = append(vectorExamples, vectorExample)
		}
	}

	// Generate dense and sparse embeddings
	denseEmbeddings, sparseEmbeddings, err := codeGenerationProcessHybridSearchEmbeddingsForExamples(vectorExamples, batchSize)
	if err != nil {
		errMessage := fmt.Sprintf("Error generating embeddings for examples: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	// Assign embeddings to the vector database objects.
	for i := range vectorExamples {
		vectorExamples[i].DenseVector = denseEmbeddings[i]
		vectorExamples[i].SparseVector = sparseEmbeddings[i]
	}

	// Convert []VectorDatabaseElement to []interface{}
	elementsAsInterface := make([]interface{}, len(vectorExamples))
	for i, v := range vectorExamples {
		elementsAsInterface[i] = v
	}

	// Insert the elements into the vector database.
	err = milvus.InsertData(examplesCollectionName, elementsAsInterface, "DocumentName", "document_name")
	if err != nil {
		errMessage := fmt.Sprintf("Error inserting data into the vector database: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}
}

// StoreExamplesInGraphDatabase stores examples in the graph database.
//
// Tags:
//   - @displayName: Store Examples in Graph Database
//
// Parameters:
//   - examples: code generation examples.
func StoreExamplesInGraphDatabase(examples []sharedtypes.CodeGenerationExample) {
	// Initialize the graph database.
	neo4j.Initialize(config.GlobalConfig.NEO4J_URI, config.GlobalConfig.NEO4J_USERNAME, config.GlobalConfig.NEO4J_PASSWORD)

	// Add the elements to the graph database.
	neo4j.Neo4j_Driver.AddCodeGenerationExampleNodes(examples)

	// Add the dependencies to the graph database.
	neo4j.Neo4j_Driver.CreateCodeGenerationExampleRelationships(examples)
}

// LoadUserGuideSections loads user guide sections from the provided paths.
//
// Tags:
//   - @displayName: Load User Guide Sections
//
// Parameters:
//   - source: source of the sections (local or github).
//   - sectionFilePaths: paths to the sections.
//   - githubRepoName: name of the github repository.
//   - githubRepoOwner: owner of the github repository.
//   - githubRepoBranch: branch of the github repository.
//   - githubAccessToken: access token for the github repository.
//
// Returns:
//   - sections: user guide sections.
func LoadUserGuideSections(source string, sectionFilePaths []string, githubRepoName string, githubRepoOwner string,
	githubRepoBranch string, githubAccessToken string) (sections []sharedtypes.CodeGenerationUserGuideSection) {
	// Initialize the sections.
	sections = []sharedtypes.CodeGenerationUserGuideSection{}

	for _, path := range sectionFilePaths {
		// Read file from either local or github source.
		var content []byte
		var err error
		switch source {
		case "local":
			_, content, err = getLocalFileContent(path)
			if err != nil {
				errMessage := fmt.Sprintf("Error getting local file content: %v", err)
				logging.Log.Error(&logging.ContextMap{}, errMessage)
				panic(errMessage)
			}
		case "github":
			_, content, err = downloadGithubFileContent(githubRepoName, githubRepoOwner, githubRepoBranch, path, githubAccessToken)
			if err != nil {
				errMessage := fmt.Sprintf("Error getting github file content: %v", err)
				logging.Log.Error(&logging.ContextMap{}, errMessage)
				panic(errMessage)
			}
		default:
			errMessage := fmt.Sprintf("Unknown data source: %s", source)
			logging.Log.Error(&logging.ContextMap{}, errMessage)
			panic(errMessage)
		}

		// Initialize the sections.
		newSections := []sharedtypes.CodeGenerationUserGuideSection{}

		// Unmarshal the JSON content into the sections.
		err = json.Unmarshal(content, &newSections)
		if err != nil {
			errMessage := fmt.Sprintf("Error unmarshalling user guide sections: %v", err)
			logging.Log.Error(&logging.ContextMap{}, errMessage)
			panic(errMessage)
		}

		// Add the new sections to the sections.
		sections = append(sections, newSections...)
	}

	logging.Log.Debugf(&logging.ContextMap{}, "Loaded %v user guide sections \n", len(sections))

	return sections
}

// StoreUserGuideSectionsInVectorDatabase stores user guide sections in the vector database.
//
// Tags:
//   - @displayName: Store User Guide Sections in Vector Database
//
// Parameters:
//   - sections: user guide sections.
//   - userGuideCollectionName: name of the collection.
//   - batchSize: batch size for embeddings.
//   - chunkSize: size of the chunks.
//   - chunkOverlap: overlap of the chunks.
func StoreUserGuideSectionsInVectorDatabase(sections []sharedtypes.CodeGenerationUserGuideSection, userGuideCollectionName string, batchSize int, chunkSize int, chunkOverlap int) {
	// Set default batch size if not provided.
	if batchSize <= 0 {
		batchSize = 2
	}

	// Create the schema for this collection
	schemaFields := []milvus.SchemaField{
		{
			Name: "guid",
			Type: "string",
		},
		{
			Name: "section_name",
			Type: "string",
		},
		{
			Name: "level",
			Type: "string",
		},
		{
			Name: "document_name",
			Type: "string",
		},
		{
			Name: "parent_section_name",
			Type: "string",
		},
		{
			Name: "previous_chunk",
			Type: "string",
		},
		{
			Name: "next_chunk",
			Type: "string",
		},
		{
			Name: "dense_vector",
			Type: "[]float32",
		},
		{
			Name: "sparse_vector",
			Type: "map[uint]float32",
		},
		{
			Name: "text",
			Type: "string",
		},
	}

	schema, err := milvus.CreateCustomSchema(userGuideCollectionName, schemaFields, "collection for code generation user guide sections")
	if err != nil {
		errMessage := "error creating the schema for the collection"
		logging.Log.Errorf(&logging.ContextMap{}, "%s: %v", errMessage, err)
		panic(fmt.Errorf("%s: %v", errMessage, err))
	}

	// Create the collection.
	err = milvus.CreateCollection(schema)
	if err != nil {
		errMessage := "error creating the collection"
		logging.Log.Errorf(&logging.ContextMap{}, "%s: %v", errMessage, err)
		panic(fmt.Errorf("%s: %v", errMessage, err))
	}

	// Create the vector database objects.
	vectorUserGuideSectionChunks := []codegeneration.VectorDatabaseUserGuideSection{}
	for _, section := range sections {
		// Create the chunks for the current element.
		chunks, err := dataExtractionTextSplitter(section.Content, chunkSize, chunkOverlap)
		if err != nil {
			errMessage := fmt.Sprintf("Error splitting text into chunks: %v", err)
			logging.Log.Error(&logging.ContextMap{}, errMessage)
			panic(errMessage)
		}
		section.Chunks = chunks

		chunkGuids := make([]string, len(section.Chunks)) // Track GUIDs for all chunks in the current element

		// Generate GUIDs for each chunk in advance.
		for j := 0; j < len(section.Chunks); j++ {
			guid := "d" + strings.ReplaceAll(uuid.New().String(), "-", "")
			chunkGuids[j] = guid
		}

		// Create vector database objects and assign PreviousChunk and NextChunk.
		for j := 0; j < len(section.Chunks); j++ {
			vectorUserGuideSectionChunk := codegeneration.VectorDatabaseUserGuideSection{
				Guid:              chunkGuids[j], // Current chunk's GUID
				SectionName:       section.Name,
				DocumentName:      section.DocumentName,
				Title:             section.Title,
				ParentSectionName: section.Parent,
				Level:             section.Level,
				PreviousChunk:     "", // Default empty
				NextChunk:         "", // Default empty
				Text:              section.Chunks[j],
			}

			// Assign PreviousChunk and NextChunk GUIDs.
			if j > 0 {
				vectorUserGuideSectionChunk.PreviousChunk = chunkGuids[j-1]
			}
			if j < len(section.Chunks)-1 {
				vectorUserGuideSectionChunk.NextChunk = chunkGuids[j+1]
			}

			// Add the new vector database object to the list.
			vectorUserGuideSectionChunks = append(vectorUserGuideSectionChunks, vectorUserGuideSectionChunk)
		}
	}

	// Generate dense and sparse embeddings
	denseEmbeddings, sparseEmbeddings, err := codeGenerationProcessHybridSearchEmbeddingsForUserGuideSections(vectorUserGuideSectionChunks, batchSize)
	if err != nil {
		errMessage := fmt.Sprintf("Error generating embeddings for user guide sections: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	// Assign embeddings to the vector database objects.
	for i := range vectorUserGuideSectionChunks {
		vectorUserGuideSectionChunks[i].DenseVector = denseEmbeddings[i]
		vectorUserGuideSectionChunks[i].SparseVector = sparseEmbeddings[i]
	}

	// Convert []VectorDatabaseElement to []interface{}
	elementsAsInterface := make([]interface{}, len(vectorUserGuideSectionChunks))
	for i, v := range vectorUserGuideSectionChunks {
		elementsAsInterface[i] = v
	}

	// Insert the elements into the vector database.
	err = milvus.InsertData(userGuideCollectionName, elementsAsInterface, "SectionName", "section_name")
	if err != nil {
		errMessage := fmt.Sprintf("Error inserting data into the vector database: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}
}

// StoreUserGuideSectionsInGraphDatabase stores user guide sections in the graph database.
//
// Tags:
//   - @displayName: Store User Guide Sections in Graph Database
//
// Parameters:
//   - elements: user guide sections.
//   - label: label for the sections (UserGuide by default).
func StoreUserGuideSectionsInGraphDatabase(sections []sharedtypes.CodeGenerationUserGuideSection, label string) {
	// Initialize the graph database.
	neo4j.Initialize(config.GlobalConfig.NEO4J_URI, config.GlobalConfig.NEO4J_USERNAME, config.GlobalConfig.NEO4J_PASSWORD)

	// Add the elements to the graph database.
	neo4j.Neo4j_Driver.AddUserGuideSectionNodes(sections, label)

	// Add the dependencies to the graph database.
	neo4j.Neo4j_Driver.CreateUserGuideSectionRelationships(sections, label)
}

// CreateGeneralDataExtractionDocumentObjects creates general data extraction document objects from
// the provided document chunks, dense embeddings, and sparse embeddings.
//
// Tags:
//   - @displayName: Create General Data Extraction Document Objects
//
// Parameters:
//   - documentName: name of the document.
//   - documentChunks: chunks of the document.
//   - denseEmbeddings: dense embeddings of the document.
//   - sparseEmbeddings: sparse embeddings of the document.
//
// Returns:
//   - extractionData: general data extraction document objects in interface format.
func CreateGeneralDataExtractionDocumentObjects(documentName string,
	documentChunks []string,
	denseEmbeddings [][]float32,
	sparseEmbeddings []map[uint]float32,
) (extractionData []interface{}) {
	extractionDataObjects := []GeneralDataExtractionDocument{}

	// Generate GUIDs for each chunk in advance.
	chunkGuids := make([]string, len(documentChunks))
	for j := 0; j < len(documentChunks); j++ {
		guid := "d" + strings.ReplaceAll(uuid.New().String(), "-", "")
		chunkGuids[j] = guid
	}

	// Create vector database objects and assign PreviousChunk and NextChunk.
	for j := 0; j < len(documentChunks); j++ {
		documentChunkElement := GeneralDataExtractionDocument{
			Guid:          chunkGuids[j], // Current chunk's GUID
			DocumentName:  documentName,
			PreviousChunk: "", // Default empty
			NextChunk:     "", // Default empty
			Text:          documentChunks[j],
			DenseVector:   denseEmbeddings[j],
			SparseVector:  sparseEmbeddings[j],
		}

		// Assign PreviousChunk and NextChunk GUIDs.
		if j > 0 {
			documentChunkElement.PreviousChunk = chunkGuids[j-1]
		}
		if j < len(documentChunks)-1 {
			documentChunkElement.NextChunk = chunkGuids[j+1]
		}

		// Add the new vector database object to the list.
		extractionDataObjects = append(extractionDataObjects, documentChunkElement)
	}

	// Convert []VectorDatabaseElement to []interface{}
	extractionData = make([]interface{}, len(extractionDataObjects))
	for i, v := range extractionDataObjects {
		extractionData[i] = v
	}

	return extractionData
}
