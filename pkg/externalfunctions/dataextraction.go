package externalfunctions

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ansys/allie-flowkit/pkg/internalstates"
	"github.com/ansys/allie-sharedtypes/pkg/logging"
	"github.com/google/go-github/v56/github"
	"github.com/google/uuid"
	"github.com/pandodao/tokenizer-go"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

// DataExtractionGetGithubFilesToExtract gets all files from github that need to be extracted.
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
func DataExtractionGetGithubFilesToExtract(githubRepoName string, githubRepoOwner string,
	githubRepoBranch string, githubAccessToken string, githubFileExtensions []string,
	githubFilteredDirectories []string, githubExcludedDirectories []string) (githubFilesToExtract []string) {

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

	return githubFilesToExtract
}

// DataExtractionGetLocalFilesToExtract gets all files from local that need to be extracted.
//
// Parameters:
//   - localPath: path to the local directory.
//   - localFileExtensions: local file extensions.
//   - localFilteredDirectories: local filtered directories.
//   - localExcludedDirectories: local excluded directories.
//
// Returns:
//   - localFilesToExtract: local files to extract.
func DataExtractionGetLocalFilesToExtract(localPath string, localFileExtensions []string,
	localFilteredDirectories []string, localExcludedDirectories []string) (localFilesToExtract []string) {

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

	for _, file := range *localFiles {
		logging.Log.Debugf(internalstates.Ctx, "Local file to extract: %s \n", file)
	}

	return *localFiles
}

// DataExtractionDownloadGithubFileContent downloads file content from github and returns checksum and content.
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
func DataExtractionDownloadGithubFileContent(githubRepoName string, githubRepoOwner string,
	githubRepoBranch string, gihubFilePath string, githubAccessToken string) (checksum string, content string) {

	// Create a new GitHub client and context.
	client, ctx := dataExtractNewGithubClient(githubAccessToken)

	// Retrieve the file content from the GitHub repository.
	fileContent, _, _, err := client.Repositories.GetContents(ctx, githubRepoOwner, githubRepoName, gihubFilePath, &github.RepositoryContentGetOptions{Ref: githubRepoBranch})
	if err != nil {
		errMessage := fmt.Sprintf("Error getting file content from github: %v", err)
		logging.Log.Error(internalstates.Ctx, errMessage)
		panic(errMessage)
	}

	// Extract the content from the file content.
	content, err = fileContent.GetContent()
	if err != nil {
		errMessage := fmt.Sprintf("Error getting file content from github: %v", err)
		logging.Log.Error(internalstates.Ctx, errMessage)
		panic(errMessage)
	}

	// Extract the checksum from the file content.
	checksum = fileContent.GetSHA()

	return checksum, content
}

// DataExtractionGetLocalFileContent reads local file and returns checksum and content.
// Parameters:
//   - localFilePath: path to file.
//
// Returns:
//   - checksum: checksum of file.
//   - content: content of file.
func DataExtractionGetLocalFileContent(localFilePath string) (checksum string, content string) {
	// Read file from local path.
	contentBytes, err := os.ReadFile(localFilePath)
	if err != nil {
		errMessage := fmt.Sprintf("Error getting local file content: %v", err)
		logging.Log.Error(internalstates.Ctx, errMessage)
		panic(errMessage)
	}

	// Calculate checksum from file content.
	hash := sha256.New()
	_, err = hash.Write(contentBytes)
	if err != nil {
		errMessage := fmt.Sprintf("Error getting local file content: %v", err)
		logging.Log.Error(internalstates.Ctx, errMessage)
		panic(errMessage)
	}

	// Convert checksum to a hexadecimal string.
	checksum = hex.EncodeToString(hash.Sum(nil))

	// Convert content to a string.
	content = string(contentBytes)

	return checksum, content
}

// DataExtractionGetDocumentType returns the document type of a file.
//
// Parameters:
//   - filePath: path to file.
//
// Returns:
//   - documentType: file extension.
func DataExtractionGetDocumentType(filePath string) (documentType string) {
	// Extract the file extension from the file path and remove the leading period.
	fileExtension := filepath.Ext(filePath)
	documentType = strings.TrimPrefix(fileExtension, ".")

	return documentType
}

// DataExtractionLangchainSplitter splits content into chunks using langchain.
//
// Parameters:
//   - content: content to split.
//   - documentType: type of document.
//   - chunkSize: size of the chunks.
//   - chunkOverlap: overlap of the chunks.
//
// Returns:
//   - output: chunks as an slice of strings.
func DataExtractionLangchainSplitter(content string, documentType string, chunkSize int, chunkOverlap int) (output []string) {
	output = []string{}
	var splittedChunks []schema.Document
	var err error

	// Creating a reader from the content of the file.
	bytesContent := []byte(content)
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

// DataExtractionGenerateDocumentTree generates a tree structure from the document chunks.
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
func DataExtractionGenerateDocumentTree(documentName string, documentId string, documentChunks []string,
	embeddingsDimensions int, getSummary bool, getKeywords bool, numKeywords int, chunkSize int, numLlmWorkers int) (returnedDocumentData []DataExtractionDocumentData) {

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
	rootData := &DataExtractionDocumentData{
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
	documentData := []*DataExtractionDocumentData{rootData}

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
				ChildDataObjects: []*DataExtractionDocumentData{},
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
						ChildDataObjects: []*DataExtractionDocumentData{},
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

	logging.Log.Debugf(internalstates.Ctx, "Finished processing document: %s \n", documentName)

	// Copy document data to returned document data
	returnedDocumentData = make([]DataExtractionDocumentData, len(documentData))
	for i, data := range documentData {
		returnedDocumentData[i] = *data
	}

	// Close llm handler input channel and wait for all workers to finish.
	close(llmHandlerInputChannel)
	llmHandlerWaitGroup.Wait()

	return returnedDocumentData
}
