package externalfunctions

import (
	"context"
	"fmt"
	"strings"

	"github.com/ansys/aali-sharedtypes/pkg/logging"
	qdrant_utils "github.com/ansys/allie-flowkit/pkg/privatefunctions/qdrant"
	"github.com/qdrant/go-client/qdrant"
)

// QdrantCreateCollection creates a collection in qdrant
//
// Tags:
//   - @displayName: Create Qdrant Collection
//
// Params:
//   - collectionName (string): The name of the collection
//   - vectorSize (uint64): The size of the vectors stored in this collection
//   - vectorDistance (string): The distance metric to use of vector similarity search (cosine, dot, euclid, manhattan)
func QdrantCreateCollection(collectionName string, vectorSize uint64, vectorDistance string) {
	client, err := qdrant_utils.QdrantClient()
	if err != nil {
		logPanic(nil, "unable to create qdrant client: %q", err)
	}

	ctx := context.TODO()

	err = client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: collectionName,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     vectorSize,
			Distance: qdrant_utils.VectorDistance(vectorDistance),
		}),
	})
	if err != nil {
		logPanic(nil, "failed to create collection: %q", err)
	}
}

// QdrantInsertData inserts data into a collection in qdrant
//
// Tags:
//   - @displayName: Insert Data into Qdrant
//
// Params:
//   - collectionName (string): The name of the collection
//   - data ([]map[string]any): The data points to insert
//   - idFieldName (string): The name of the field to use as the ID
//   - vectorFieldName (string): The name of the field to use as the vector
func QdrantInsertData(collectionName string, data []map[string]any, idFieldName string, vectorFieldName string) {
	points := make([]*qdrant.PointStruct, len(data))
	for i, d := range data {
		id := qdrant.NewIDUUID(d[idFieldName].(string))
		vector := qdrant.NewVectorsDense(d[vectorFieldName].([]float32))
		delete(d, idFieldName)
		delete(d, vectorFieldName)
		points[i] = &qdrant.PointStruct{
			Id:      id,
			Vectors: vector,
			Payload: qdrant.NewValueMap(d),
		}
	}

	client, err := qdrant_utils.QdrantClient()
	if err != nil {
		logPanic(nil, "unable to create qdrant client: %q", err)
	}

	ctx := context.TODO()

	resp, err := client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: collectionName,
		Points:         points,
	})

	if err != nil {
		logPanic(nil, "failed to insert data: %q", err)
	}
	logging.Log.Debugf(&logging.ContextMap{}, "successfully upserted %d points into qdrant collection %q: %q", len(points), collectionName, resp.GetStatus())
}

// QdrantCreateIndex creates a field index on a qdrant collection
//
// Tags:
//   - @displayName: Create Qdrant Index
//
// Params:
//   - collectionName (string): The name of the collection
//   - fieldName (string): The name of the payload field to create an index on
//   - fieldType (string): The qdrant type that the payload field is expected to be
func QdrantCreateIndex(collectionName string, fieldName string, fieldType string) {
	client, err := qdrant_utils.QdrantClient()
	if err != nil {
		logPanic(nil, "unable to create qdrant client: %q", err)
	}

	qdrantType, err := qdrantFieldType(fieldType)
	if err != nil {
		logPanic(nil, "could not create qdrant field type: %q", err)
	}

	request := qdrant.CreateFieldIndexCollection{
		CollectionName: collectionName,
		FieldName:      fieldName,
		FieldType:      qdrantType,
		// TODO: there is more customization here you can do, but specific to the field type
	}
	res, err := client.CreateFieldIndex(context.TODO(), &request)
	if err != nil {
		logPanic(nil, "failed to create index: %q", err)
	}
	logging.Log.Debugf(&logging.ContextMap{}, "successfully created index: %v", res.Status)
}

func qdrantFieldType(fieldType string) (*qdrant.FieldType, error) {
	switch strings.ToLower(fieldType) {
	case "keyword":
		return qdrant.FieldType_FieldTypeKeyword.Enum(), nil
	case "integer", "int":
		return qdrant.FieldType_FieldTypeInteger.Enum(), nil
	case "float":
		return qdrant.FieldType_FieldTypeFloat.Enum(), nil
	case "geo":
		return qdrant.FieldType_FieldTypeGeo.Enum(), nil
	case "text":
		return qdrant.FieldType_FieldTypeText.Enum(), nil
	case "bool", "boolean":
		return qdrant.FieldType_FieldTypeBool.Enum(), nil
	case "date", "time", "datetime":
		return qdrant.FieldType_FieldTypeDatetime.Enum(), nil
	case "uuid":
		return qdrant.FieldType_FieldTypeUuid.Enum(), nil
	default:
		return nil, fmt.Errorf("unknown qdrant field type %q", fieldType)
	}
}
