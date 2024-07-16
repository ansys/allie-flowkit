package externalfunctions

// HandlerRequest represents the client request for a specific chat or embeddings operation.
type HandlerRequest struct {
	Adapter             string            `json:"adapter"` // "chat", "embeddings"
	InstructionGuid     string            `json:"instructionGuid"`
	Data                string            `json:"data"`
	ChatRequestType     string            `json:"chatRequestType"`        // "summary", "code", "keywords", "general"; only relevant if "adapter" is "chat"
	DataStream          bool              `json:"dataStream"`             // only relevant if "adapter" is "chat"
	MaxNumberOfKeywords uint32            `json:"maxNumberOfKeywords"`    // only relevant if "chatRequestType" is "keywords"
	IsConversation      bool              `json:"isConversation"`         // only relevant if "chatRequestType" is "code"
	ConversationHistory []HistoricMessage `json:"conversationHistory"`    // only relevant if "isConversation" is true
	GeneralContext      string            `json:"generalContext"`         // any added context you might need
	MsgContext          string            `json:"msgContext"`             // any added context you might need
	SystemPrompt        string            `json:"systemPrompt"`           // only relevant if "chatRequestType" is "general"
	ModelOptions        ModelOptions      `json:"modelOptions,omitempty"` // only relevant if "adapter" is "chat"
	ClientGuid          string
}

// HandlerResponse represents the LLM Handler response for a specific request.
type HandlerResponse struct {
	// Common properties
	InstructionGuid string `json:"instructionGuid"`
	Type            string `json:"type"` // "info", "error", "chat", "embeddings"
	// Chat properties
	IsLast   *bool   `json:"isLast,omitempty"`
	Position *uint32 `json:"position,omitempty"`
	ChatData *string `json:"chatData,omitempty"`
	// Embeddings properties
	EmbeddedData []float32 `json:"embeddedData,omitempty"`
	// Error properties
	Error *ErrorResponse `json:"error,omitempty"`
	// Info properties
	InfoMessage *string `json:"infoMessage,omitempty"`
}

// ErrorResponse represents the error response sent to the client when something fails during the processing of the request.
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// TransferDetails holds communication channels for the websocket listener and writer.
type TransferDetails struct {
	ResponseChannel chan HandlerResponse
	RequestChannel  chan HandlerRequest
}

// HistoricMessage represents a past chat messages.
type HistoricMessage struct {
	Role    string `json:"role"` // "user", "assistant", "system"
	Content string `json:"content"`
}

// OpenAIOption represents an option for an OpenAI API call.
type ModelOptions struct {
	FrequencyPenalty *float32 `json:"frequencyPenalty,omitempty"`
	MaxTokens        *int32   `json:"maxTokens,omitempty"`
	PresencePenalty  *float32 `json:"presencePenalty,omitempty"`
	Stop             []string `json:"stop,omitempty"`
	Temperature      *float32 `json:"temperature,omitempty"`
	TopP             *float32 `json:"topP,omitempty"`
}

type DbFilters struct {
	// Filters for string fields
	GuidFilter         []string `json:"guid,omitempty"`
	DocumentIdFilter   []string `json:"document_id,omitempty"`
	DocumentNameFilter []string `json:"document_name,omitempty"`
	LevelFilter        []string `json:"level,omitempty"`

	// Filters for array fields
	TagsFilter     DbArrayFilter `json:"tags,omitempty"`
	KeywordsFilter DbArrayFilter `json:"keywords,omitempty"`

	// Filters for JSON fields
	MetadataFilter []DbJsonFilter `json:"metadata,omitempty"`
}

// DbArrayFilter is used to filter array fields in the database.
type DbArrayFilter struct {
	NeedAll    bool     `json:"needAll"`
	FilterData []string `json:"filterData"`
}

// DbJsonFilter is used to filter JSON fields in the database.
type DbJsonFilter struct {
	FieldName  string   `json:"fieldName"`
	FieldType  string   `json:"fieldType" description:"Can be either string or array."` // "string" or "array"
	FilterData []string `json:"filterData"`
	NeedAll    bool     `json:"needAll" description:"Only needed if the FieldType is array."` // only needed for array fields
}

// similarityElement represents a single element in the similarity search result.
type similarityElement struct {
	Score float64    `json:"distance"`
	Data  DbResponse `json:"data"`
}

// similaritySearchInput represents the input for the similarity search function.
type similaritySearchInput struct {
	CollectionName    string    `json:"collection_name" description:"Name of the collection to which the data objects will be added. Required for adding data." required:"true"`
	EmbeddedVector    []float32 `json:"embedded_vector" description:"Embedded vector used for searching. Required for retrieval." required:"true"`
	MaxRetrievalCount int       `json:"max_retrieval_count" description:"Maximum number of results to be retrieved. If it is not specified, the default value is milvus.MaxSearchRetrievalCount. Optional." required:"false"`
	OutputFields      []string  `json:"output_fields" description:"Fields to be included in the output. If not specified all fields will be retrieved.Optional" required:"false"`
	Filters           DbFilters `json:"filters" description:"Filter for the query. Optional." required:"false"`
	MinScore          float64   `json:"min_score" description:"Filter objects with a score higher than the specified minimum. Optional." required:"false"`
	GetLeafNodes      bool      `json:"get_leaf_nodes" description:"Flag to indicate whether to retrieve all the leaf nodes in the result node branch. Set to true to include the leaf nodes. Optional." required:"false"`
	GetSiblings       bool      `json:"get_siblings" description:"Flag to indicate whether to retrieve the previous and next node to the result nodes. Set to true to include the siblings. Optional." required:"false"`
	GetParent         bool      `json:"get_parent" description:"Flag to indicate whether to retrieve the parent object. Set to true to include the parent object. Optional." required:"false"`
	GetChildren       bool      `json:"get_children" description:"Flag to indicate whether to retrieve the children objects. Set to true to include the children objects. Optional." required:"false"`
}

// similaritySearchOutput represents the output for the similarity search function.
type similaritySearchOutput struct {
	SimilarityResult []similarityElement `json:"similarity_result" description:"Similarity Result"`
}

// queryInput represents the input for the query function.
type queryInput struct {
	CollectionName    string    `json:"collection_name" description:"Name of the collection to which the data objects will be added. Required for adding data." required:"true"`
	MaxRetrievalCount int       `json:"max_retrieval_count" description:"Maximum number of results to be retrieved. If not specified, the default value is retrieve all database. If the number of results is too big for the database, the request will be cancelled. Optional." required:"false"`
	OutputFields      []string  `json:"output_fields" description:"Fields to be included in the output. If not specified all fields will be retrieved.Optional" required:"false"`
	Filters           DbFilters `json:"filters" description:"Filter for the query. At least one filter must be defined." required:"true"`
}

// queryOutput represents the output for the query function.
type queryOutput struct {
	QueryResult []DbResponse `json:"queryResult" description:"Returns the results of the query."`
}

// DbData represents the data structure for the database.
type DbData struct {
	Guid              string                 `json:"guid"`
	DocumentId        string                 `json:"document_id"`
	DocumentName      string                 `json:"document_name"`
	Text              string                 `json:"text"`
	Keywords          []string               `json:"keywords"`
	Summary           string                 `json:"summary"`
	Embedding         []float32              `json:"embeddings"`
	Tags              []string               `json:"tags"`
	Metadata          map[string]interface{} `json:"metadata"`
	ParentId          string                 `json:"parent_id"`
	ChildIds          []string               `json:"child_ids"`
	PreviousSiblingId string                 `json:"previous_sibling_id"`
	NextSiblingId     string                 `json:"next_sibling_id"`
	LastChildId       string                 `json:"last_child_id"`
	FirstChildId      string                 `json:"first_child_id"`
	Level             string                 `json:"level"`
	HasNeo4jEntry     bool                   `json:"has_neo4j_entry"`
}

// DbResponse represents the response structure for the database.
type DbResponse struct {
	Guid              string                 `json:"guid"`
	DocumentId        string                 `json:"document_id"`
	DocumentName      string                 `json:"document_name"`
	Text              string                 `json:"text"`
	Keywords          []string               `json:"keywords"`
	Summary           string                 `json:"summary"`
	Embedding         []float32              `json:"embeddings"`
	Tags              []string               `json:"tags"`
	Metadata          map[string]interface{} `json:"metadata"`
	ParentId          string                 `json:"parent_id"`
	ChildIds          []string               `json:"child_ids"`
	PreviousSiblingId string                 `json:"previous_sibling_id"`
	NextSiblingId     string                 `json:"next_sibling_id"`
	LastChildId       string                 `json:"last_child_id"`
	FirstChildId      string                 `json:"first_child_id"`
	Distance          float64                `json:"distance"`
	Level             string                 `json:"level"`
	HasNeo4jEntry     bool                   `json:"has_neo4j_entry"`

	// Siblings
	Parent    *DbData  `json:"parent,omitempty"`
	Children  []DbData `json:"children,omitempty"`
	LeafNodes []DbData `json:"leaf_nodes,omitempty"`
	Siblings  []DbData `json:"siblings,omitempty"`
}

// DBListCollectionsOutput represents the output for the listCollections function.
type DBListCollectionsOutput struct {
	Success     bool     `json:"success" description:"Returns true if the collections were listed successfully. Returns false or an error if not."`
	Collections []string `json:"collections" description:"A list of collection names."`
}

// retrieveDependenciesInput represents the input for the retrieveDependencies function.
type retrieveDependenciesInput struct {
	CollectionName        string        `json:"collection_name" description:"Name of the collection to which the data objects will be added. Required for adding data." required:"true"`
	RelationshipName      string        `json:"relationship_name" description:"Name of the relationship to retrieve dependencies for. Required for retrieving dependencies." required:"true"`
	RelationshipDirection string        `json:"relationship_direction" description:"Direction of the relationship to retrieve dependencies for. It can be either 'in', 'out' or 'both'. Required for retrieving dependencies." required:"true"`
	SourceDocumentId      string        `json:"source_document_id" description:"Document ID of the source node. Required for retrieving dependencies." required:"true"`
	NodeTypesFilter       DbArrayFilter `json:"node_types_filter" description:"Filter based on node types. Use MilvusArrayFilter for specifying node type filtering criteria. Optional." required:"false"`
	DocumentTypesFilter   []string      `json:"document_types_filter" description:"Filter based on document types. Use MilvusArrayFilter for specifying document type filtering criteria. Optional." required:"false"`
	MaxHopsNumber         int           `json:"max_hops_number" description:"Maximum number of hops to traverse. Optional." required:"true"`
}

// retrieveDependenciesOutput represents the output for the retrieveDependencies function.
type retrieveDependenciesOutput struct {
	Success         bool     `json:"success" description:"Returns true if the collections were listed successfully. Returns false or an error if not."`
	DependenciesIds []string `json:"dependencies_ids" description:"A list of document IDs that are dependencies of the specified source node."`
}

// GeneralNeo4jQueryInput represents the input for the generalNeo4jQuery function.
type GeneralNeo4jQueryInput struct {
	Query string `json:"query" description:"Neo4j query to be executed. Required for executing a query." required:"true"`
}

// GeneralNeo4jQueryOutput represents the output for the generalNeo4jQuery function.
type GeneralNeo4jQueryOutput struct {
	Success  bool          `json:"success" description:"Returns true if the query was executed successfully. Returns false or an error if not."`
	Response neo4jResponse `json:"response" description:"Summary and records of the query execution."`
}

// neo4jResponse represents the response structure for the Neo4j query.
type neo4jResponse struct {
	Record          neo4jRecord     `json:"record"`
	SummaryCounters summaryCounters `json:"summaryCounters"`
}

// neo4jRecord represents the record structure for the Neo4j query.
type neo4jRecord []struct {
	Values []value `json:"Values"`
}

// value represents the value structure for the Neo4j query.
type value struct {
	Id        int      `json:"Id"`
	NodeTypes []string `json:"Labels"`
	Props     props    `json:"Props"`
}

// props represents the properties structure for the Neo4j query.
type props struct {
	CollectionName string   `json:"collectionName"`
	DocumentId     string   `json:"documentId"`
	DocumentTypes  []string `json:"documentTypes,omitempty"`
	Guid           string   `json:"guid,omitempty"`
}

// summaryCounters represents the summary counters structure for the Neo4j query.
type summaryCounters struct {
	NodesCreated         int `json:"nodes_created"`
	NodesDeleted         int `json:"nodes_deleted"`
	RelationshipsCreated int `json:"relationships_created"`
	RelationshipsDeleted int `json:"relationships_deleted"`
	PropertiesSet        int `json:"properties_set"`
	LabelsAdded          int `json:"labels_added"`
	LabelsRemoved        int `json:"labels_removed"`
	IndexesAdded         int `json:"indexes_added"`
	IndexesRemoved       int `json:"indexes_removed"`
	ConstraintsAdded     int `json:"constraints_added"`
	ConstraintsRemoved   int `json:"constraints_removed"`
}

// DefaultFields represents the default fields for the user query.
type AnsysGPTDefaultFields struct {
	QueryWord         string
	FieldName         string
	FieldDefaultValue string
}

// ACSRequest represents the request structure for the Azure Cognitive Search.
type ACSSearchRequest struct {
	Search                string           `json:"search"`
	VectorQueries         []ACSVectorQuery `json:"vectorQueries"`
	VectorFilterMode      string           `json:"vectorFilterMode"`
	Filter                string           `json:"filter"`
	QueryType             string           `json:"queryType"`
	SemanticConfiguration string           `json:"semanticConfiguration"`
	Top                   int              `json:"top"`
	Select                string           `json:"select"`
	Count                 bool             `json:"count"`
}

// ACSVectorQuery represents the vector query structure for the Azure Cognitive Search.
type ACSVectorQuery struct {
	Kind   string    `json:"kind"`
	K      int       `json:"k"`
	Vector []float32 `json:"vector"`
	Fields string    `json:"fields"`
}

// ACSSearchResponseStruct represents the response structure for the Azure Cognitive Search.
type ACSSearchResponseStruct struct {
	OdataContext string              `json:"@odata.context"`
	OdataCount   int                 `json:"@odata.count"`
	Value        []ACSSearchResponse `json:"value"`
}

// ACSSearchResponse represents the response structure for the Azure Cognitive Search.
type ACSSearchResponse struct {
	Physics             string  `json:"physics"`
	SourceTitleLvl3     string  `json:"sourceTitle_lvl3"`
	SourceURLLvl3       string  `json:"sourceURL_lvl3"`
	TokenSize           int     `json:"tokenSize"`
	SourceTitleLvl2     string  `json:"sourceTitle_lvl2"`
	Weight              float64 `json:"weight"`
	SourceURLLvl2       string  `json:"sourceURL_lvl2"`
	Product             string  `json:"product"`
	Content             string  `json:"content"`
	TypeOFasset         string  `json:"typeOFasset"`
	Version             string  `json:"version"`
	SearchScore         float64 `json:"@search.score"`
	SearchRerankerScore float64 `json:"@search.reranker_score"`
}

// AnsysGPTCitation represents the citation structure for the Ansys GPT.
type AnsysGPTCitation struct {
	Title     string  `json:"Title"`
	URL       string  `json:"URL"`
	Relevance float64 `json:"Relevance"`
}
