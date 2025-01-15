package externalfunctions

var ExternalFunctionsMap = map[string]interface{}{
	// llm handler
	"PerformVectorEmbeddingRequest":                    PerformVectorEmbeddingRequest,
	"PerformVectorEmbeddingRequestWithTokenLimitCatch": PerformVectorEmbeddingRequestWithTokenLimitCatch,
	"PerformBatchEmbeddingRequest":                     PerformBatchEmbeddingRequest,
	"PerformKeywordExtractionRequest":                  PerformKeywordExtractionRequest,
	"PerformGeneralRequest":                            PerformGeneralRequest,
	"PerformGeneralRequestWithImages":                  PerformGeneralRequestWithImages,
	"PerformGeneralRequestSpecificModel":               PerformGeneralRequestSpecificModel,
	"PerformCodeLLMRequest":                            PerformCodeLLMRequest,
	"BuildLibraryContext":                              BuildLibraryContext,
	"BuildFinalQueryForGeneralLLMRequest":              BuildFinalQueryForGeneralLLMRequest,
	"BuildFinalQueryForCodeLLMRequest":                 BuildFinalQueryForCodeLLMRequest,
	"AppendMessageHistory":                             AppendMessageHistory,
	"ShortenMessageHistory":                            ShortenMessageHistory,

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
	"AnsysGPTPerformLLMRephraseRequestNew":           AnsysGPTPerformLLMRephraseRequestNew,
	"AnsysGPTBuildFinalQuery":                        AnsysGPTBuildFinalQuery,
	"AnsysGPTPerformLLMRequest":                      AnsysGPTPerformLLMRequest,
	"AnsysGPTReturnIndexList":                        AnsysGPTReturnIndexList,
	"AnsysGPTACSSemanticHybridSearchs":               AnsysGPTACSSemanticHybridSearchs,
	"AnsysGPTRemoveNoneCitationsFromSearchResponse":  AnsysGPTRemoveNoneCitationsFromSearchResponse,
	"AnsysGPTReorderSearchResponseAndReturnOnlyTopK": AnsysGPTReorderSearchResponseAndReturnOnlyTopK,
	"AnsysGPTGetSystemPrompt":                        AnsysGPTGetSystemPrompt,
	"AisPerformLLMRephraseRequest":                   AisPerformLLMRephraseRequest,
	"AisReturnIndexList":                             AisReturnIndexList,
	"AisAcsSemanticHybridSearchs":                    AisAcsSemanticHybridSearchs,
	"AisChangeAcsResponsesByFactor":                  AisChangeAcsResponsesByFactor,
	"AisPerformLLMFinalRequest":                      AisPerformLLMFinalRequest,

	// data extraction
	"GetGithubFilesToExtract":   GetGithubFilesToExtract,
	"GetLocalFilesToExtract":    GetLocalFilesToExtract,
	"AppendStringSlices":        AppendStringSlices,
	"DownloadGithubFileContent": DownloadGithubFileContent,
	"GetLocalFileContent":       GetLocalFileContent,
	"GetDocumentType":           GetDocumentType,
	"LangchainSplitter":         LangchainSplitter,
	"GenerateDocumentTree":      GenerateDocumentTree,
	"AddDataRequest":            AddDataRequest,
	"CreateCollectionRequest":   CreateCollectionRequest,

	// generic
	"AssignStringToString": AssignStringToString,
	"SendRestAPICall":      SendRestAPICall,
}
