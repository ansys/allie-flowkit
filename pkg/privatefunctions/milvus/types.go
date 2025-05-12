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
