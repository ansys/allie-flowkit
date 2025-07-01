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
	"PerformGeneralRequestSpecificModelModelOptionsAndImages":                        PerformGeneralRequestSpecificModelModelOptionsAndImages,
	"PerformGeneralRequestSpecificModelNoStreamWithOpenAiTokenOutput":                PerformGeneralRequestSpecificModelNoStreamWithOpenAiTokenOutput,
	"PerformGeneralRequestSpecificModelAndModelOptionsNoStreamWithOpenAiTokenOutput": PerformGeneralRequestSpecificModelAndModelOptionsNoStreamWithOpenAiTokenOutput,
	"PerformCodeLLMRequest":                                                          PerformCodeLLMRequest,
	"PerformGeneralRequestNoStreaming":                                               PerformGeneralRequestNoStreaming,
	"BuildLibraryContext":                                                            BuildLibraryContext,
	"BuildFinalQueryForGeneralLLMRequest":                                            BuildFinalQueryForGeneralLLMRequest,
	"BuildFinalQueryForCodeLLMRequest":                                               BuildFinalQueryForCodeLLMRequest,
	"AppendMessageHistory":                                                           AppendMessageHistory,
	"ShortenMessageHistory":                                                          ShortenMessageHistory,
	"CheckTokenLimitReached":                                                         CheckTokenLimitReached,

	// knowledge db
	"SendVectorsToKnowledgeDB": SendVectorsToKnowledgeDB,
	"GetListCollections":       GetListCollections,
	"RetrieveDependencies":     RetrieveDependencies,
	"GeneralGraphDbQuery":      GeneralGraphDbQuery,
	"AddGraphDbParameter":      AddGraphDbParameter,
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
	"AecGetContextFromRetrieverModule":               AecGetContextFromRetrieverModule,
	"AecPerformLLMFinalRequest":                      AecPerformLLMFinalRequest,

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
	"AssignStringToString":   AssignStringToString,
	"SendRestAPICall":        SendRestAPICall,
	"PrintFeedback":          PrintFeedback,
	"ExtractJSONStringField": ExtractJSONStringField,
	"GenerateUUID":           GenerateUUID,
	"JsonPath":               JsonPath,
	"StringConcat":           StringConcat,
	"StringFormat":           StringFormat,

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
	"SimilartitySearchOnPathDescriptions":       SimilartitySearchOnPathDescriptions,
	"FindRelevantPathDescriptionByPrompt":       FindRelevantPathDescriptionByPrompt,
	"FetchPropertiesFromPathDescription":        FetchPropertiesFromPathDescription,
	"FetchNodeDescriptionsFromPathDescription":  FetchNodeDescriptionsFromPathDescription,
	"FetchActionsPathFromPathDescription":       FetchActionsPathFromPathDescription,
	"SynthesizeActions":                         SynthesizeActions,
	"FinalizeResult":                            FinalizeResult,
	"GetSolutionsToFixProblem":                  GetSolutionsToFixProblem,
	"GetSelectedSolution":                       GetSelectedSolution,
	"AppendToolHistory":                         AppendToolHistory,
	"AppendMeshPilotHistory":                    AppendMeshPilotHistory,
	"GetActionsFromConfig":                      GetActionsFromConfig,
	"ParseHistory":                              ParseHistory,
	"SynthesizeActionsTool4":                    SynthesizeActionsTool4,
	"SynthesizeActionsTool13":                   SynthesizeActionsTool13,
	"SynthesizeActionsTool14":                   SynthesizeActionsTool14,
	"SimilartitySearchOnPathDescriptionsQdrant": SimilartitySearchOnPathDescriptionsQdrant,
	"GenerateSubWorkflowPrompt":                 GenerateSubWorkflowPrompt,
	"ProcessSubworkflowIdentificationOutput":    ProcessSubworkflowIdentificationOutput,
	"MarkdownToHTML":                            MarkdownToHTML,
	"ParseHistoryToHistoricMessages":            ParseHistoryToHistoricMessages,
	"FinalizeMessage":                           FinalizeMessage,
	"GenerateUserPrompt":                        GenerateUserPrompt,
	"GenerateUserPromptWithList":                GenerateUserPromptWithList,

	// qdrant
	"QdrantCreateCollection": QdrantCreateCollection,
	"QdrantInsertData":       QdrantInsertData,

	// auth
	"CheckApiKeyAuthMongoDb":                        CheckApiKeyAuthMongoDb,
	"CheckCreateUserIdMongoDb":                      CheckCreateUserIdMongoDb,
	"UpdateTotalTokenCountForCustomerMongoDb":       UpdateTotalTokenCountForCustomerMongoDb,
	"UpdateTotalTokenCountForUserIdMongoDb":         UpdateTotalTokenCountForUserIdMongoDb,
	"DenyCustomerAccessAndSendWarningMongoDb":       DenyCustomerAccessAndSendWarningMongoDb,
	"DenyCustomerAccessAndSendWarningMongoDbUserId": DenyCustomerAccessAndSendWarningMongoDbUserId,
	"SendLogicAppNotificationEmail":                 SendLogicAppNotificationEmail,
	"CreateMessageWithVariable":                     CreateMessageWithVariable,

	// mcp
	"ListAll":         ListAll,
	"ExecuteTool":     ExecuteTool,
	"GetResource":     GetResource,
	"GetSystemPrompt": GetSystemPrompt,

	// materials
	"SerializeResponse":                SerializeResponse,
	"AddGuidsToAttributes":             AddGuidsToAttributes,
	"FilterOutNonExistingAttributes":   FilterOutNonExistingAttributes,
	"FilterOutDuplicateAttributes":     FilterOutDuplicateAttributes,
	"ExtractCriteriaSuggestions":       ExtractCriteriaSuggestions,
	"LogRequestSuccess":                LogRequestSuccess,
	"LogRequestFailed":                 LogRequestFailed,
	"LogRequestFailedDebugWithMessage": LogRequestFailedDebugWithMessage,
	"PerformMultipleGeneralRequestsAndExtractAttributesWithOpenAiTokenOutput": PerformMultipleGeneralRequestsAndExtractAttributesWithOpenAiTokenOutput,

	// rhsc
	"SetCopilotGenerateRequestJsonBody": SetCopilotGenerateRequestJsonBody,
}
