package milvus

import "github.com/ansys/allie-flowkit/pkg/privatefunctions/codegeneration"

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
	Data []codegeneration.VectorDatabaseElement `json:"data,omitempty"`

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
