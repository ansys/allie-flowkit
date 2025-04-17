package externalfunctions

var ExternalFunctionsMap = map[string]interface{}{
	// llm handler
	"PerformVectorEmbeddingRequest":                                                  PerformVectorEmbeddingRequest,
	"PerformVectorEmbeddingRequestWithTokenLimitCatch":                               PerformVectorEmbeddingRequestWithTokenLimitCatch,
	"PerformBatchEmbeddingRequest":                                                   PerformBatchEmbeddingRequest,
	"PerformBatchHybridEmbeddingRequest":                                             PerformBatchHybridEmbeddingRequest,
	"PerformKeywordExtractionRequest":                                                PerformKeywordExtractionRequest,
	"PerformGeneralRequest":                                                          PerformGeneralRequest,
	"PerformGeneralRequestWithImages":                                                PerformGeneralRequestWithImages,
	"PerformGeneralModelSpecificationRequest":                                        PerformGeneralModelSpecificationRequest,
	"PerformGeneralRequestSpecificModel":                                             PerformGeneralRequestSpecificModel,
	"PerformGeneralRequestSpecificModelAndModelOptions":                              PerformGeneralRequestSpecificModelAndModelOptions,
	"PerformGeneralRequestSpecificModelNoStreamWithOpenAiTokenOutput":                PerformGeneralRequestSpecificModelNoStreamWithOpenAiTokenOutput,
	"PerformGeneralRequestSpecificModelAndModelOptionsNoStreamWithOpenAiTokenOutput": PerformGeneralRequestSpecificModelAndModelOptionsNoStreamWithOpenAiTokenOutput,
	"PerformCodeLLMRequest":                                                          PerformCodeLLMRequest,
	"PerformGeneralRequestNoStreaming":                                               PerformGeneralRequestNoStreaming,
	"BuildLibraryContext":                                                            BuildLibraryContext,
	"BuildFinalQueryForGeneralLLMRequest":                                            BuildFinalQueryForGeneralLLMRequest,
	"BuildFinalQueryForCodeLLMRequest":                                               BuildFinalQueryForCodeLLMRequest,
	"AppendMessageHistory":                                                           AppendMessageHistory,
	"ShortenMessageHistory":                                                          ShortenMessageHistory,

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

	// generic
	"AssignStringToString": AssignStringToString,
	"SendRestAPICall":      SendRestAPICall,

	// code generation
	"LoadCodeGenerationElements":      LoadCodeGenerationElements,
	"LoadCodeGenerationExamples":      LoadCodeGenerationExamples,
	"LoadAndCheckExampleDependencies": LoadAndCheckExampleDependencies,
	"LoadUserGuideSections":           LoadUserGuideSections,

	"StoreElementsInVectorDatabase":          StoreElementsInVectorDatabase,
	"StoreElementsInGraphDatabase":           StoreElementsInGraphDatabase,
	"StoreExamplesInVectorDatabase":          StoreExamplesInVectorDatabase,
	"StoreExamplesInGraphDatabase":           StoreExamplesInGraphDatabase,
	"StoreUserGuideSectionsInVectorDatabase": StoreUserGuideSectionsInVectorDatabase,
	"StoreUserGuideSectionsInGraphDatabase":  StoreUserGuideSectionsInGraphDatabase,

	// ansys mesh pilot
	"MeshPilotReAct":                           MeshPilotReAct,
	"SimilartitySearchOnPathDescriptions":      SimilartitySearchOnPathDescriptions,
	"FindRelevantPathDescriptionByPrompt":      FindRelevantPathDescriptionByPrompt,
	"FetchPropertiesFromPathDescription":       FetchPropertiesFromPathDescription,
	"FetchNodeDescriptionsFromPathDescription": FetchNodeDescriptionsFromPathDescription,
	"FetchActionsPathFromPathDescription":      FetchActionsPathFromPathDescription,
	"SynthesizeActions":                        SynthesizeActions,
	"FinalizeResult":                           FinalizeResult,
	"GetSolutionsToFixProblem":                 GetSolutionsToFixProblem,
	"GetSelectedSolution":                      GetSelectedSolution,
	"AppendToolHistory":                        AppendToolHistory,
	"AppendMeshPilotHistory":                   AppendMeshPilotHistory,
	"GetActionsFromConfig":                     GetActionsFromConfig,
	"ParseHistory":                             ParseHistory,

	// milvus
	"MilvusCreateCollection": MilvusCreateCollection,
	"MilvusInsertData":       MilvusInsertData,

	// auth
	"CheckApiKeyAuthMongoDb":                        CheckApiKeyAuthMongoDb,
	"CheckCreateUserIdMongoDb":                      CheckCreateUserIdMongoDb,
	"UpdateTotalTokenCountForCustomerMongoDb":       UpdateTotalTokenCountForCustomerMongoDb,
	"UpdateTotalTokenCountForUserIdMongoDb":         UpdateTotalTokenCountForUserIdMongoDb,
	"DenyCustomerAccessAndSendWarningMongoDb":       DenyCustomerAccessAndSendWarningMongoDb,
	"DenyCustomerAccessAndSendWarningMongoDbUserId": DenyCustomerAccessAndSendWarningMongoDbUserId,
	"SendLogicAppNotificationEmail":                 SendLogicAppNotificationEmail,
	"CreateMessageWithVariable":                     CreateMessageWithVariable,
}
