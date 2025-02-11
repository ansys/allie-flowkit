package milvus

type SchemaField struct {
	Name        string
	Type        string
	PrimaryKey  bool
	AutoID      bool
	Dimension   int
	Description string
}

type MilvusRequest struct {
	// Common fields
	CollectionName string `json:"collectionName"`

	// Insert request
	Data []interface{} `json:"data,omitempty"`

	// Search and Query request
	OutputFields []string `json:"outputFields,omitempty"`
	Filter       *string  `json:"filter,omitempty"`
	Limit        *int     `json:"limit,omitempty"`
	Offset       *int     `json:"offset,omitempty"`

	// Search request
	DenseVector  []float32        `json:"dense_vector,omitempty"`
	SparseVector map[uint]float32 `json:"sparse_vector,omitempty"`

	// Delete entry request
	AutoIds []AutoIdFlex `json:"id,omitempty"`
}

type AutoIdFlex int64

type MilvusInsertResponse struct {
	Code       int              `json:"code"`
	Message    string           `json:"message"`
	InsertData MilvusInsertData `json:"data"`
}

type MilvusQueryResponse struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Data    []interface{} `json:"data"`
}

type MilvusSearchResponse struct {
	Code       int           `json:"code"`
	Message    string        `json:"message"`
	SearchData []interface{} `json:"data"`
}

type MilvusInsertData struct {
	InsertCount int `json:"insertCount"`
}

type ArrayFilter struct {
	NeedAll    bool     `json:"needAll"`
	FilterData []string `json:"filterData"`
}

type JsonFilter struct {
	FieldName  string   `json:"fieldName"`
	FieldType  string   `json:"fieldType" description:"Can be either string or array."` // "string" or "array"
	FilterData []string `json:"filterData"`
	NeedAll    bool     `json:"needAll" description:"Only needed if the FieldType is array."` // only needed for array fields
}

type MilvusDeleteResponse struct {
	Code    int              `json:"code"`
	Data    MilvusDeleteData `json:"data"`
	Message string           `json:"message"`
}

type MilvusDeleteData struct {
	DeleteCount int `json:"deleteCount"`
}
