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

import (
	"github.com/ansys/aali-flowkit/pkg/privatefunctions/milvus"
	"github.com/ansys/aali-sharedtypes/pkg/logging"
)

// MilvusCreateCollection creates a collection in Milvus
//
// Tags:
//   - @displayName: Create Milvus Collection
//
// Params:
//   - collectionName (string): The name of the collection
//   - schema (map[string]interface{}): The schema of the collection
func MilvusCreateCollection(collectionName string, schema []map[string]interface{}) {
	// From schema to field schema
	schemaObject := []milvus.SchemaField{}
	if len(schema) != 0 {
		for _, field := range schema {
			schemaField := milvus.SchemaField{
				Name: field["name"].(string),
				Type: field["type"].(string),
			}
			schemaObject = append(schemaObject, schemaField)
		}
	} else {
		// Create default schema
		schemaObject = []milvus.SchemaField{
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
				Name: "text",
				Type: "string",
			},
		}
	}

	// Create custom schema
	milvusSchema, err := milvus.CreateCustomSchema(collectionName, schemaObject, "")
	if err != nil {
		errorMessage := "Failed to create schema: " + err.Error()
		logging.Log.Errorf(&logging.ContextMap{}, "%s", errorMessage)
		panic(errorMessage)
	}

	// Create collection
	err = milvus.CreateCollection(milvusSchema)
	if err != nil {
		errorMessage := "Failed to create collection: " + err.Error()
		logging.Log.Errorf(&logging.ContextMap{}, "%s", errorMessage)
		panic(errorMessage)
	}
}

// MilvusInsertData inserts data into a collection in Milvus
//
// Tags:
//   - @displayName: Insert Data into Milvus
//
// Params:
//   - collectionName (string): The name of the collection
//   - data ([]interface{}): The data to insert
//   - idFieldName (string): The name of the field to use as the ID
//   - idField (string): The ID field
func MilvusInsertData(collectionName string, data []interface{}, idFieldName string) {
	// Insert data
	err := milvus.InsertData(collectionName, data, idFieldName, idFieldName)
	if err != nil {
		errorMessage := "Failed to insert data: " + err.Error()
		logging.Log.Errorf(&logging.ContextMap{}, "%s", errorMessage)
		panic(errorMessage)
	}
}
