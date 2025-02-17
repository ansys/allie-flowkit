package externalfunctions

import (
	"github.com/ansys/allie-flowkit/pkg/privatefunctions/milvus"
	"github.com/ansys/allie-sharedtypes/pkg/logging"
)

func MilvusCreateCollection(collectionName string, schema []interface{}) {
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
				Name: field.(map[string]interface{})["name"].(string),
				Type: field.(map[string]interface{})["type"].(string),
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
