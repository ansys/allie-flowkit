package externalfunctions

var ExternalFunctionsMap = map[string]interface{}{
	// llm handler
	"PerformVectorEmbeddingRequest":       PerformVectorEmbeddingRequest,
	"PerformBatchEmbeddingRequest":        PerformBatchEmbeddingRequest,
	"PerformKeywordExtractionRequest":     PerformKeywordExtractionRequest,
	"PerformGeneralRequest":               PerformGeneralRequest,
	"PerformGeneralRequestSpecificModel":  PerformGeneralRequestSpecificModel,
	"PerformCodeLLMRequest":               PerformCodeLLMRequest,
	"BuildLibraryContext":                 BuildLibraryContext,
	"BuildFinalQueryForGeneralLLMRequest": BuildFinalQueryForGeneralLLMRequest,
	"BuildFinalQueryForCodeLLMRequest":    BuildFinalQueryForCodeLLMRequest,
	"AppendMessageHistory":                AppendMessageHistory,

	// knowledge db
	"SendVectorsToKnowledgeDB": SendVectorsToKnowledgeDB,
	"GetListCollections":       GetListCollections,
	"RetrieveDependencies":     RetrieveDependencies,
	"GeneralNeo4jQuery":        GeneralNeo4jQuery,
	"GeneralQuery":             GeneralQuery,
	"SimilaritySearch":         SimilaritySearch,
	"CreateKeywordsDbFilter":   CreateKeywordsDbFilter,
	"CreateTagsDbFilter":       CreateTagsDbFilter,
	"CreateMetadataDbFilter":   CreateMetadataDbFilter,
	"CreateDbFilter":           CreateDbFilter,

	// ansys gpt
	"AnsysGPTCheckProhibitedWords":                   AnsysGPTCheckProhibitedWords,
	"AnsysGPTExtractFieldsFromQuery":                 AnsysGPTExtractFieldsFromQuery,
	"AnsysGPTPerformLLMRephraseRequest":              AnsysGPTPerformLLMRephraseRequest,
	"AnsysGPTPerformLLMRephraseRequestOld":           AnsysGPTPerformLLMRephraseRequestOld,
	"AnsysGPTBuildFinalQuery":                        AnsysGPTBuildFinalQuery,
	"AnsysGPTPerformLLMRequest":                      AnsysGPTPerformLLMRequest,
	"AnsysGPTReturnIndexList":                        AnsysGPTReturnIndexList,
	"AnsysGPTACSSemanticHybridSearchs":               AnsysGPTACSSemanticHybridSearchs,
	"AnsysGPTRemoveNoneCitationsFromSearchResponse":  AnsysGPTRemoveNoneCitationsFromSearchResponse,
	"AnsysGPTReorderSearchResponseAndReturnOnlyTopK": AnsysGPTReorderSearchResponseAndReturnOnlyTopK,
	"AnsysGPTGetSystemPrompt":                        AnsysGPTGetSystemPrompt,

	// data extraction
	"DataExtractionGetGithubFilesToExtract":   DataExtractionGetGithubFilesToExtract,
	"DataExtractionGetLocalFilesToExtract":    DataExtractionGetLocalFilesToExtract,
	"DataExtractionAppendStringSlices":        DataExtractionAppendStringSlices,
	"DataExtractionDownloadGithubFileContent": DataExtractionDownloadGithubFileContent,
	"DataExtractionGetLocalFileContent":       DataExtractionGetLocalFileContent,
	"DataExtractionGetDocumentType":           DataExtractionGetDocumentType,
	"DataExtractionLangchainSplitter":         DataExtractionLangchainSplitter,
	"DataExtractionGenerateDocumentTree":      DataExtractionGenerateDocumentTree,
	"DataExtractionAddDataRequest":            DataExtractionAddDataRequest,
	"DataExtractionCreateCollectionRequest":   DataExtractionCreateCollectionRequest,

	// generic
	"AssignStringToString": AssignStringToString,
	"SendRestAPICall":      SendRestAPICall,
}
