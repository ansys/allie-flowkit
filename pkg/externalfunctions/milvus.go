package externalfunctions

import (
	"github.com/ansys/allie-flowkit/pkg/privatefunctions/generic"
	"github.com/ansys/allie-flowkit/pkg/privatefunctions/milvus"
	"github.com/ansys/allie-sharedtypes/pkg/logging"
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
	// Initialize Milvus client
	client, err := milvus.Initialize()
	if err != nil {
		errorMessage := "Failed to initialize Milvus client: " + err.Error()
		logging.Log.Errorf(&logging.ContextMap{}, "%s", errorMessage)
		return
	}

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
	err = milvus.CreateCollection(milvusSchema, client)
	if err != nil {
		errorMessage := "Failed to create collection: " + err.Error()
		logging.Log.Errorf(&logging.ContextMap{}, "%s", errorMessage)
		panic(errorMessage)
	}
}

// MilvusInsertData inserts data into a collection in Milvus
//
// Tags:
//   - @displayName: Insert Data into Milvus Collection
//
// Params:
//   - collectionName (string): The name of the collection
//   - data ([]interface{}): The data to insert
//   - idFieldName (string): The name of the field to use as the ID
//   - idField (string): The ID field
func MilvusInsertData(collectionName string, data []interface{}, idFieldName string) {
	// Convert the snake_case idFieldName to PascalCase
	objectFieldName := generic.SnakeToCamel(idFieldName, true)

	// Insert data
	err := milvus.InsertData(collectionName, data, objectFieldName, idFieldName)
	if err != nil {
		errorMessage := "Failed to insert data: " + err.Error()
		logging.Log.Errorf(&logging.ContextMap{}, "%s", errorMessage)
		panic(errorMessage)
	}
}
