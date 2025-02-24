package milvus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ansys/allie-flowkit/pkg/privatefunctions/generic"
	"github.com/ansys/allie-sharedtypes/pkg/config"
	"github.com/ansys/allie-sharedtypes/pkg/logging"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

// Global variables
var (
	CallsBatchSize          = 500
	DumpBatchSize           = 10000
	MilvusConnectionTimeout = 5 * time.Second
	MilvusConnectionRetries = 40
	MaxSearchRetrievalCount = 16384
)

// Initialize initializes the Milvus DB.
// The function first creates a Milvus client, then creates a collection with the specified name,
// creates indexes on the collection, and loads the collection.
//
// Parameters:
//   - collectionName: Name of the collection to be initialized.
//
// Returns:
//   - error: Error if any issue occurs during initializing the Milvus DB.
func Initialize() (milvusClient client.Client, funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic initialize: %v", r)
			funcError = r.(error)
			return
		}
	}()
	var err error

	// Create Milvus client
	milvusClient, err = newClient()
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "error during NewMilvusClient: %s", err.Error())
		return nil, err
	}

	// Load all existing collections
	collections, err := listCollections(milvusClient)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "error during ListCollections: %s", err.Error())
		return nil, err
	}

	for _, collection := range collections {
		err := loadCollection(collection, milvusClient)
		if err != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "error during LoadCollection: %s", err.Error())
			return nil, err
		}
	}

	logging.Log.Debugf(&logging.ContextMap{}, "Initialized Milvus")

	return milvusClient, nil
}

// newClient creates a new Milvus client.
//
// Returns:
//   - milvusClient: Milvus client.
//   - error: Error if any issue occurs during creating the Milvus client.
func newClient() (milvusClient client.Client, funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic in NewMilvusClient: %v", r)
			funcError = r.(error)
			return
		}
	}()
	// Set the Milvus connection timeout and retries
	milvusConnectionRetries := 3
	milvusConnectionTimeout := 10 * time.Second

	// Create the config for the Milvus client
	milvusConfig := client.Config{
		Address: fmt.Sprintf("%s:%s", config.GlobalConfig.MILVUS_HOST, config.GlobalConfig.MILVUS_PORT),
	}

	// Retry logic for timeout errors
	for retry := 0; retry < milvusConnectionRetries; retry++ {

		// Create a context with timeout
		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), milvusConnectionTimeout)

		// Create the Milvus client
		milvusClient, err := client.NewClient(
			ctxWithTimeout,
			milvusConfig,
		)

		cancel()

		if err == nil {
			return milvusClient, nil
		}

		// Check if the error is a timeout error
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			logging.Log.Warnf(&logging.ContextMap{}, "Timeout error during client.NewClient: %s (Retry %d/%d)", err.Error(), retry+1, milvusConnectionRetries)
			continue
		}

		// If the error is not a timeout error, log and return the error
		logging.Log.Errorf(&logging.ContextMap{}, "Error during client.NewClient: %s", err.Error())
		return nil, err
	}

	return nil, errors.New("unable to create Milvus client after maximum retries")
}

// listCollections lists all collections in the Milvus DB.
//
// Parameters:
//   - milvusClient: Milvus client.
//
// Returns:
//   - collections: List of collections in the Milvus DB.
//   - error: Error if any issue occurs during listing the collections.
func listCollections(milvusClient client.Client) (collections []string, funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic in ListCollections: %v", r)
			funcError = r.(error)
			return
		}
	}()
	listColl, err := milvusClient.ListCollections(
		context.Background(),
	)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "failed to list collections: %s", err.Error())
		return collections, err
	}

	// Create collection slice
	for _, collection := range listColl {
		if collection.Name != config.GlobalConfig.TEMP_COLLECTION_NAME {
			collections = append(collections, collection.Name)
		}
	}

	logging.Log.Debugf(&logging.ContextMap{}, "Collections listed: %v", collections)

	return collections, nil
}

// LoadCollection loads collection from disk to memory.
//
// Parameters:
//   - collectionName: Name of the collection to be loaded.
//   - milvusClient: Milvus client.
//
// Returns:
//   - error: Error if any issue occurs during loading the collection.
func loadCollection(collectionName string, milvusClient client.Client) (funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic in LoadCollection: %v", r)
			funcError = r.(error)
			return
		}
	}()

	err := milvusClient.LoadCollection(
		context.Background(), // ctx
		collectionName,       // CollectionName
		false,                // async
	)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "error during milvusClient.LoadCollection: %s", err.Error())
		return err
	}

	logging.Log.Debugf(&logging.ContextMap{}, "Collection loaded: %v", collectionName)

	return nil
}

// CreateCollection creates a collection in the Milvus DB.
//
// Parameters:
//   - schema: Schema of the collection to be created.
//
// Returns:
//   - error: Error if any issue occurs during creating the collection.
func CreateCollection(schema *entity.Schema) (funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic in CreateCollection: %v", r)
			funcError = r.(error)
			return
		}
	}()
	// Create Milvus client
	milvusClient, err := newClient()
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "error during NewMilvusClient: %s", err.Error())
		return err
	}

	// Check if the collection already exists
	hasColl, err := milvusClient.HasCollection(
		context.Background(),  // ctx
		schema.CollectionName, // CollectionName
	)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error during HasCollection: %v", err)
		return err
	}

	if hasColl {
		logging.Log.Debugf(&logging.ContextMap{}, "Collection already exists: %s\n", schema.CollectionName)
		// Load the collection
		err = loadCollection(schema.CollectionName, milvusClient)
		if err != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Error during LoadCollection: %v", err)
		}
		return nil
	}

	// Create collection if it does not exist
	err = milvusClient.CreateCollection(
		context.Background(), // ctx
		schema,
		2, // shardNum
		client.WithEnableDynamicSchema(true),
	)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error during CreateCollection: %v", err)
		return err
	}

	// Get the vector field name from the schema
	var denseVectorFieldName string
	for _, field := range schema.Fields {
		if field.DataType == entity.FieldTypeFloatVector || field.DataType == entity.FieldTypeBinaryVector {
			denseVectorFieldName = field.Name
			break
		}
	}

	// Get the sparse vector field name from the schema
	var sparseVectorFieldName string
	for _, field := range schema.Fields {
		if field.DataType == entity.FieldTypeSparseVector {
			sparseVectorFieldName = field.Name
			break
		}
	}

	// Create index on the collection
	err = CreateIndexes(schema.CollectionName, milvusClient, "guid", denseVectorFieldName, sparseVectorFieldName)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error during CreateIndexes: %v", err)
		return err
	}

	// Load the collection
	err = loadCollection(schema.CollectionName, milvusClient)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error during LoadCollection: %v", err)
		return err
	}

	logging.Log.Debugf(&logging.ContextMap{}, "Created collection: %s\n", schema.CollectionName)

	return nil
}

// CreateIndexes creates indexes for the Milvus DB.
// Firstly, a scalar index is created for the guid field. Secondly, a vector index is created for the embeddings field.
//
// Parameters:
//   - collectionName: Name of the collection to create the indexes for.
//   - milvusClient: Milvus client.
//
// Returns:
//   - error: Error if any issue occurs during creating the indexes.
func CreateIndexes(collectionName string, milvusClient client.Client, guidFieldName string, denseVectorFieldName string, sparseVectorFieldName string) (funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic in CreateIndexes: %v", r)
			funcError = r.(error)
			return
		}
	}()

	/////////////////////////////////////////
	// 1. Create Scalar Index
	/////////////////////////////////////////
	err := milvusClient.CreateIndex(
		context.Background(), // ctx
		collectionName,       // CollectionName
		guidFieldName,
		entity.NewScalarIndex(),
		false,
	)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "failed to create index: %s", err.Error())
		return err
	}

	logging.Log.Debugf(&logging.ContextMap{}, "Scalar index created")

	///////////////////////////////////////////
	// 2. Create Vector Index
	///////////////////////////////////////////

	var denseIdx entity.Index
	var metricType entity.MetricType

	// Determine the type of metric based on the configuration
	switch config.GlobalConfig.MILVUS_METRIC_TYPE {
	case "l2":
		metricType = entity.L2
	case "ip":
		metricType = entity.IP
	case "cosine":
		metricType = entity.COSINE
	default:
		metricType = entity.COSINE
	}

	// Determine the type of vector index based on the configuration
	switch config.GlobalConfig.MILVUS_INDEX_TYPE {
	case "flat":
		denseIdx, err = entity.NewIndexFlat(metricType)
	case "ivfFlat":
		denseIdx, err = entity.NewIndexIvfFlat(
			metricType, // metricType
			1024,       // ConstructParams
		)
	default:
		err = errors.New("unknown vector index")
	}

	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Failed to create %v index: %s", denseIdx.IndexType(), err.Error())
		return err
	}

	// Create a vector index for the denseVectorFieldName field
	err = milvusClient.CreateIndex(
		context.Background(), // ctx
		collectionName,       // CollectionName
		denseVectorFieldName,
		denseIdx,
		false,
	)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "failed to create index: %s", err.Error())
		return err
	}

	// Create a vector index for the sparseVectorFieldName field
	if sparseVectorFieldName != "" {
		sparseIdx, err := entity.NewIndexSparseInverted(entity.IP, 0)
		if err != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "failed to create index %v", err.Error())
			return err
		}
		err = milvusClient.CreateIndex(
			context.Background(),
			collectionName,
			sparseVectorFieldName,
			sparseIdx,
			false,
		)
		if err != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "failed to create index %v", err.Error())
			return err
		}
	}

	return nil
}

// CreateCustomSchema function to generate a Milvus schema based on user input
func CreateCustomSchema(collectionName string, fields []SchemaField, description string) (*entity.Schema, error) {
	var schemaFields []*entity.Field

	for _, field := range fields {
		fieldSchema := &entity.Field{
			Name:        field.Name,
			Description: field.Description,
		}

		// Convert field to Milvus schema field
		switch field.Type {
		case "int32":
			fieldSchema.DataType = entity.FieldTypeInt32
		case "int64":
			fieldSchema.DataType = entity.FieldTypeInt64
		case "float32":
			fieldSchema.DataType = entity.FieldTypeFloat
		case "float64":
			fieldSchema.DataType = entity.FieldTypeDouble
		case "string":
			fieldSchema.DataType = entity.FieldTypeVarChar
		case "bool":
			fieldSchema.DataType = entity.FieldTypeBool
		case "map[string]string":
			fieldSchema.DataType = entity.FieldTypeJSON
		case "[]float32":
			fieldSchema.DataType = entity.FieldTypeFloatVector
			field.Dimension = config.GlobalConfig.EMBEDDINGS_DIMENSIONS
		case "map[uint]float32":
			fieldSchema.DataType = entity.FieldTypeSparseVector
			field.Dimension = config.GlobalConfig.EMBEDDINGS_DIMENSIONS
		case "[]bool":
			fieldSchema.DataType = entity.FieldTypeBinaryVector
		default:
			return nil, fmt.Errorf("unsupported field type: %s", field.Type)
		}

		// *Note: Array of strings are not supported by Milvus schema, those fields will be added dynamically

		// Set primary key and auto ID options
		if field.PrimaryKey {
			fieldSchema.PrimaryKey = true
		}
		if field.AutoID {
			fieldSchema.AutoID = true
		}

		// Set maximum length for string fields
		maxLength := 40000
		if fieldSchema.DataType == entity.FieldTypeVarChar {
			fieldSchema.TypeParams = map[string]string{"max_length": strconv.Itoa(maxLength)}
		}

		// Set dimension for vector fields
		if fieldSchema.DataType == entity.FieldTypeFloatVector || fieldSchema.DataType == entity.FieldTypeBinaryVector {
			if field.Dimension <= 0 {
				return nil, fmt.Errorf("dimension must be greater than 0 for vector fields")
			}
			fieldSchema.TypeParams = map[string]string{"dim": strconv.Itoa(field.Dimension)}
		}

		schemaFields = append(schemaFields, fieldSchema)
	}

	// Add auto ID field if not already present
	schemaFields = append(schemaFields, &entity.Field{
		Name:        "id",
		DataType:    entity.FieldTypeInt64,
		Description: "Auto-generated ID field",
		PrimaryKey:  true,
		AutoID:      true,
	})

	schema := &entity.Schema{
		CollectionName: collectionName,
		Description:    description,
		Fields:         schemaFields,
	}

	return schema, nil
}

// InsertData sends data to the Milvus DB in order to populate the vector database.
// The function sends the data in batches of const "CallsBatchSize" entries. The data is sent via HTTP POST requests.
//
// Parameters:
//   - collectionName: Name of the collection to send the data to.
//   - dataToSend: Data to be sent to the Milvus DB.
//   - objectFieldName: Name of the field in the data go object to be used as the object field.
//   - idFieldName: Name of the field in the milvus schema to be used as the ID field.
//
// Returns:
//   - error: Error if any issue occurs during sending the data to the Milvus DB.
func InsertData(collectionName string, dataToSend []interface{}, objectFieldName string, idFieldName string) (funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic in InsertData: %v", r)
			funcError = r.(error)
			return
		}
	}()

	// Create a MilvusInsertRequest object
	request := MilvusRequest{
		CollectionName: collectionName,
		Data:           nil,
	}

	startIndex := 0
	stopIndex := 0

	// Send data to insert in batches of CallsBatchSize
	for stopIndex < len(dataToSend) {
		// Calculate the batch stop index, considering the array bounds
		stopIndex += CallsBatchSize
		if stopIndex > len(dataToSend) {
			stopIndex = len(dataToSend)
		}

		// Assign the data batch to the request object
		request.Data = dataToSend[startIndex:stopIndex]

		// Remove duplicates inside milvus so updated data can be added
		entriesToRemove, err := QueryDuplicates(collectionName, dataToSend[startIndex:stopIndex], objectFieldName, idFieldName)
		if err != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Error during QueryDuplicates: %v", err)
			return err
		}

		// Send the data batch to Milvus
		err = sendInsertRequest(request)
		if err != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Error during sendInsertRequest: %v. Adding removed data.", err)
			return err
		}

		// If data was successfully inserted, remove the duplicates from the Milvus database
		err = RemoveDuplicates(collectionName, entriesToRemove)
		if err != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Error during RemoveDuplicates: %v", err)
			return err
		}

		logging.Log.Debugf(&logging.ContextMap{}, "Updated %d entries in Milvus", len(entriesToRemove))

		// Move the start index to the next batch
		startIndex = stopIndex
	}

	return nil
}

// Query queries the Milvus database.
//
// This function queries the Milvus database with an HTTP POST request.
//
// Parameters:
//   - collectionName: Name of the collection to query.
//   - responseLimit: Maximum number of responses.
//   - outputFields: Fields to return in the response.
//   - filterExpression: Filters to apply to the query.
//
// Returns:
//   - response: Response from the Milvus database.
//   - error: Error if any issue occurs while querying the Milvus database.
func Query(collectionName string, responseLimit int, outputFields []string, filterExpression string) (response MilvusQueryResponse, funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic in Query: %v", r)
			funcError = r.(error)
			return
		}
	}()

	// Create a MilvusQueryRequest object
	query := newMilvusQueryRequest(collectionName, outputFields, filterExpression, responseLimit, 0)

	// Send the Milvus query request and receive the response.
	response, err := sendQueryRequest(query)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error during sendQueryRequest: %v", err)
		return response, err
	}

	return response, nil
}

// QueryDuplicates queries the Milvus database for the provided data and retrieves the duplicates.
//
// Parameters:
//   - collectionName: Name of the collection in the Milvus database.
//   - data: Data to send to the Milvus database.
//   - objectFieldName: Name of the field in the data go object to be used as the object field.
//   - idFieldName: Name of the field in the milvus schema to be used as the ID field.
//
// Returns:
//   - removedData: Data that was removed from the Milvus database.
//   - error: Error if any issue occurs during querying the Milvus database.
func QueryDuplicates(collectionName string, data []interface{}, objectFieldName string, idFieldName string) (duplicatedEntriesId []int64, funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic in QueryDuplicates: %v", r)
			funcError = r.(error)
			return
		}
	}()

	dataIds := make([]string, 0, len(data))
	for _, entry := range data {
		idStr, err := generic.ExtractStringFieldFromStruct(entry, objectFieldName)
		if err != nil {
			panic(fmt.Sprintf("Error extracting field: %v", err))
		}

		dataIds = append(dataIds, idStr)
	}

	// Query Milvus for entries in the current batch
	filter, err := createStringFilterExpression(idFieldName, dataIds)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error during createStringFilterExpression: %v", err)
		return nil, err
	}

	// Send the Milvus query request and receive the response.
	response, err := Query(collectionName, 0, []string{"id"}, filter)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error during Query: %v", err)
		return nil, err
	}

	// Create a list of IDs for the duplicated entries
	for _, item := range response.Data {
		if obj, ok := item.(map[string]interface{}); ok {
			// get the id field from the object
			if idNum, ok := obj["id"].(json.Number); ok {
				id, err := strconv.ParseInt(idNum.String(), 10, 64)
				if err != nil {
					logging.Log.Errorf(&logging.ContextMap{}, "Error parsing ID: %v", err)
					continue
				}
				duplicatedEntriesId = append(duplicatedEntriesId, id)
			}
		}
	}
	return duplicatedEntriesId, nil
}

// RemoveDuplicates removes duplicates from the Milvus database.
//
// Parameters:
//   - collectionName: Name of the collection in the Milvus database.
//   - dataIds: IDs of the data to remove from the Milvus database.
//
// Returns:
//   - error: Error if any issue occurs during querying the Milvus database.
func RemoveDuplicates(collectionName string, dataIds []int64) (funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic in RemoveDuplicates: %v", r)
			funcError = r.(error)
			return
		}
	}()
	// If there are no data IDs, return
	if len(dataIds) == 0 {
		return nil
	}

	// Query Milvus for entries in the current batch
	filter, err := createNumberFilterExpression("id", dataIds)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error during createStringFilterExpression: %v", err)
		return err
	}

	response, err := Query(collectionName, 0, []string{"*"}, filter)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error during Query: %v", err)
		return err
	}

	if len(response.Data) > 0 {
		// Remove duplicates from the milvus db
		err = sendRemoveRequest(collectionName, filter)
		if err != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Error during sendRemoveRequest: %v", err)
			return err
		}
	}

	return nil
}

// sendRemoveRequest sends an HTTP POST request to the Milvus database for removing entries.
//
// Parameters:
//   - collectionName: Name of the collection to remove entries from.
//   - filter: Filter to apply to the remove request.
//
// Returns:
//   - error: Error if any issue occurs while sending the request to the Milvus database.
func sendRemoveRequest(collectionName string, filter string) (funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic in sendRemoveRequest: %v", r)
			funcError = r.(error)
			return
		}
	}()

	// Create the URL for the Milvus delete request
	url := fmt.Sprintf("http://%s:%s/v2/vectordb/entities/delete", config.GlobalConfig.MILVUS_HOST, config.GlobalConfig.MILVUS_PORT)

	// Create MilvusDeleteRequest object
	request := MilvusRequest{
		CollectionName: collectionName,
		Filter:         &filter,
	}

	// Create a MilvusDeleteResponse object
	var response MilvusDeleteResponse

	// Send the Milvus delete request and receive the response.
	err, _ := generic.CreatePayloadAndSendHttpRequest(url, "POST", request, &response)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error in CreatePayloadAndSendHttpRequest: %s", err)
		return err
	}

	// Check the response status code.
	if response.Code != 0 {
		logging.Log.Errorf(&logging.ContextMap{}, "Request failed with status code %d and message: %s\n", response.Code, response.Message)
		return errors.New(response.Message)
	}

	return nil
}

// sendInsertRequest sends an HTTP POST request to the Milvus DB.
// The function sends the request via HTTP POST request. It's used only for insert requests.
//
// Parameters:
//   - request: Request to be sent to the Milvus DB.
//
// Returns:
//   - error: Error if any issue occurs during sending the request to the Milvus DB.
func sendInsertRequest(request MilvusRequest) (funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic in sendInsertRequest: %v", r)
			funcError = r.(error)
			return
		}
	}()

	// Create a MilvusInsertResponse object
	var response MilvusInsertResponse

	// Create the URL for the Milvus insert request
	url := fmt.Sprintf("http://%s:%s/v1/vector/insert", config.GlobalConfig.MILVUS_HOST, config.GlobalConfig.MILVUS_PORT)

	// Send the Milvus insert request and receive the response.
	err, _ := generic.CreatePayloadAndSendHttpRequest(url, "POST", request, &response)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error in CreatePayloadAndSendHttpRequest: %s", err)
		return err
	}

	// Check the response status code.
	if response.Code != http.StatusOK {
		logging.Log.Errorf(&logging.ContextMap{}, "Request failed with status code %d and message: %s\n", response.Code, response.Message)
		return errors.New(response.Message)
	}

	logging.Log.Debugf(&logging.ContextMap{}, "Added %v entries to the DB.", response.InsertData.InsertCount)

	return nil
}

// sendQueryRequest sends an HTTP POST request to the Milvus DB.
// The function sends the request via HTTP POST request. It's used for query requests.
//
// Parameters:
//   - request: Request to be sent to the Milvus DB.
//
// Returns:
//   - response: Response from the Milvus DB.
//   - error: Error if any issue occurs during sending the request to the Milvus DB.
func sendQueryRequest(request MilvusRequest) (response MilvusQueryResponse, funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic in sendQueryRequest: %v", r)
			funcError = r.(error)
			return
		}
	}()

	// Create the URL for the Milvus query request
	url := fmt.Sprintf("http://%s:%s/v1/vector/query", config.GlobalConfig.MILVUS_HOST, config.GlobalConfig.MILVUS_PORT)

	// Send the Milvus query request and receive the response.
	err, _ := generic.CreatePayloadAndSendHttpRequest(url, "POST", request, &response)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error in CreatePayloadAndSendHttpRequest: %s", err)
		return response, err
	}

	// Check the response status code.
	if response.Code != http.StatusOK {
		logging.Log.Errorf(&logging.ContextMap{}, "Request failed with status code %d and message: %s\n", response.Code, response.Message)
		return response, errors.New(response.Message)
	}

	return response, nil
}

// newMilvusQueryRequest creates a Milvus request object for querying data.
//
// Parameters:
//   - collectionName: Name of the collection to query.
//   - outputFields: Fields to return in the response.
//   - filter: Filter to apply to the query.
//   - limit: Maximum number of responses.
//   - ofset: Offset for the query.
//
// Returns:
//   - request: Request sent to the Milvus database.
func newMilvusQueryRequest(collectionName string, outputFields []string, filter string, limit int, ofset int) MilvusRequest {
	return MilvusRequest{
		CollectionName: collectionName,
		OutputFields:   outputFields,
		Filter:         &filter,
		Limit:          &limit,
		Offset:         &ofset,
	}
}

// createFilterExpression creates a filter expression for the filtering option in the query and search requests.
//
// Parameters:
//   - filterType: Type of the filter.
//   - filters: Filters to apply.
//
// Returns:
//   - filterExpression: Filter expression for the provided filterType.
//   - error: Error if any issue occurs while creating the filter expression.
func createJsonFilterExpression(filterType string, filters []JsonFilter) (filterExpression string, funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic in createJsonFilterExpression: %v", r)
			funcError = r.(error)
			return
		}
	}()

	var expressions []string

	for _, field := range filters {
		switch field.FieldType {
		case "string":
			stringExpression := []string{}
			for _, filterData := range field.FilterData {
				expression := fmt.Sprintf("%s['%s'] == '%s'", filterType, field.FieldName, filterData)
				stringExpression = append(stringExpression, expression)
			}
			expressions = append(expressions, strings.Join(stringExpression, " || "))

		case "array":
			containsFunc := ""
			if field.NeedAll {
				containsFunc = "json_contains_all"
			} else {
				containsFunc = "json_contains_any"
			}

			filterData := strings.Join(field.FilterData, "','")
			expression := fmt.Sprintf("%s(%s['%s'], ['%s'])", containsFunc, filterType, field.FieldName, filterData)
			expressions = append(expressions, expression)
		}
	}

	filterExpression = strings.Join(expressions, " and ")

	return filterExpression, nil
}

// createArrayFilterExpression creates a filter expression for the array filtering option in the query and search requests.
//
// Parameters:
//   - filterType: Type of the filter.
//   - filter: Filter to apply.
//
// Returns:
//   - filterExpression: Filter expression for the provided filter types.
//   - error: Error if any issue occurs while creating the filter expression.
func createArrayFilterExpression(filterType string, filter ArrayFilter) (filterExpression string, funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic in createArrayFilterExpression: %v", r)
			funcError = r.(error)
			return
		}
	}()

	var containsFunc string

	if filter.NeedAll {
		containsFunc = "array_contains_all"
	} else {
		containsFunc = "array_contains_any"
	}

	filterData := strings.Join(filter.FilterData, "','")
	filterExpression = fmt.Sprintf("%s(%s, ['%s'])", containsFunc, filterType, filterData)

	return filterExpression, nil
}

// createStringFilterExpression creates a filter expression for the string filtering option in the query and search requests.
//
// Parameters:
//   - filterType: Type of the filter.
//   - filter: Filter to apply.
//
// Returns:
//   - filterExpression: Filter expression for the provided filter type.
//   - error: Error if any issue occurs while creating the filter expression.
func createStringFilterExpression(filterType string, filter []string) (filterExpression string, funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic in createStringFilterExpression: %v", r)
			funcError = r.(error)
			return
		}
	}()

	for i := range filter {
		// Escape double backslashes in the filter value.
		filter[i] = strings.ReplaceAll(filter[i], "\\", "\\\\")
		// Escape single quotes in the filter value
		filter[i] = strings.ReplaceAll(filter[i], "'", "\\'")
		// Escape new lines in the filter value
		filter[i] = strings.ReplaceAll(filter[i], "\n", "\\n")
		// Escape carriage returns in the filter value
		filter[i] = strings.ReplaceAll(filter[i], "\r", "\\r")
	}

	filterExpression = fmt.Sprintf("%s in ['%s'", filterType, filter[0])
	if len(filter) > 1 {
		filterExpression += fmt.Sprintf(", '%s'", strings.Join(filter[1:], "', '"))
	}

	return filterExpression + "]", nil
}

// createNumberFilterExpression creates a filter expression for the number filtering option in the query and search requests.
//
// Parameters:
//   - filterType: Type of the filter.
//   - filter: Filter to apply.
//
// Returns:
//   - filterExpression: Filter expression for the provided filter type.
//   - error: Error if any issue occurs while creating the filter expression.
func createNumberFilterExpression(filterType string, filter []int64) (filterExpression string, funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic in createStringFilterExpression: %v", r)
			funcError = r.(error)
			return
		}
	}()
	// Ensure there is at least one filter value
	if len(filter) == 0 {
		return "", nil
	}

	// Convert numbers to string and join them properly
	filterStrings := make([]string, len(filter))
	for i, num := range filter {
		filterStrings[i] = strconv.FormatInt(num, 10)
	}

	filterExpression = fmt.Sprintf("%s in [%s]", filterType, strings.Join(filterStrings, ", "))

	return filterExpression, nil
}

// combineFilterExpressions combines the filter expressions for the filtering option in the query and search requests.
// This function combines the filter expressions for the "guid", "keywords", "document_id" and "documents" fields.
//
// Parameters:
//   - arrayFilters: Array filters to apply.
//   - stringFilters: String filters to apply.
//   - jsonFilters: JSON filters to apply.
//
// Returns:
//   - filterExpression: Filter expression to use in the query and search requests.
//   - error: Error if any issue occurs while combining the filter expressions.
func combineFilterExpressions(arrayFilters map[string]ArrayFilter, stringFilters map[string][]string, jsonFilters map[string][]JsonFilter) (filterExpression string, funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic in combineFilterExpressions: %v", r)
			funcError = r.(error)
			return
		}
	}()

	var expressions []string

	for field, filter := range arrayFilters {
		if filter.FilterData != nil && len(filter.FilterData) > 0 {
			filterExpr, err := createArrayFilterExpression(field, filter)
			if err != nil {
				logging.Log.Errorf(&logging.ContextMap{}, "Error in createArrayFilterExpression: %v", err)
				return "", err
			}
			expressions = append(expressions, filterExpr)
		}
	}

	for field, filter := range stringFilters {
		if len(filter) > 0 {
			filterExpr, err := createStringFilterExpression(field, filter)
			if err != nil {
				logging.Log.Errorf(&logging.ContextMap{}, "Error in createStringFilterExpression: %v", err)
				return "", err
			}
			expressions = append(expressions, filterExpr)
		}
	}

	for field, filter := range jsonFilters {
		if len(filter) > 0 {
			filterExpr, err := createJsonFilterExpression(field, filter)
			if err != nil {
				logging.Log.Errorf(&logging.ContextMap{}, "Error in createJsonFilterExpression: %v", err)
				return "", err
			}
			expressions = append(expressions, filterExpr)
		}
	}

	// Combine the filter expressions with "and"
	filterExpression = strings.Join(expressions, " and ")

	return filterExpression, nil
}

// sendSearchRequest sends an HTTP POST request to the Milvus DB.
// The function sends the request via HTTP POST request. It's used for search requests.
//
// Parameters:
//   - request: Request to be sent to the Milvus DB.
//
// Returns:
//   - response: Response from the Milvus DB.
//   - error: Error if any issue occurs during sending the request to the Milvus DB.
func sendSearchRequest(request MilvusRequest) (response MilvusSearchResponse, funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic in sendSearchRequest: %v", r)
			funcError = r.(error)
			return
		}
	}()

	// Create the URL for the Milvus search request
	url := fmt.Sprintf("http://%s:%s/v1/vector/search", config.GlobalConfig.MILVUS_HOST, config.GlobalConfig.MILVUS_PORT)

	// Send the Milvus search request and receive the response.
	err, _ := generic.CreatePayloadAndSendHttpRequest(url, "POST", request, &response)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error in CreatePayloadAndSendHttpRequest: %s", err)
		return response, err
	}

	// Check the response status code.
	if response.Code != http.StatusOK {
		logging.Log.Errorf(&logging.ContextMap{}, "Request failed with status code %d and message: %s\n", response.Code, response.Message)
		return response, errors.New(response.Message)
	}

	return response, nil
}

func removeFieldFromData(data []interface{}, fieldToRemove string) []interface{} {
	for i, item := range data {
		if obj, ok := item.(map[string]interface{}); ok {
			delete(obj, fieldToRemove)
			data[i] = obj
		}
	}
	return data
}

// // Returns nil if threshhold is not reached
// func SearchMilvusCollection(request MilvusRequest, milvusClient client.Client, minScore float32, numResults int, denseVectorWeight float64, sparseVectorWeight float64) (response []client.SearchResult, func_error error) {
// 	defer func() {
// 		r := recover()
// 		if r != nil {
// 			logging.Log.Errorf(&logging.ContextMap{}, "Panic occured in SearchMilvusCollection: %v", r)
// 			func_error = r.(error)
// 		}
// 	}()

// 	sub_requests := []*client.ANNSearchRequest{}

// 	// Define Sparse Vector Search Request (if defined)
// 	if request.SparseVector != nil {
// 		positions := make([]uint32, 0, len(request.SparseVector))
// 		values := make([]float32, 0, len(request.SparseVector))

// 		for key, value := range request.SparseVector {
// 			positions = append(positions, uint32(key))
// 			values = append(values, value)
// 		}

// 		sparse_vector_formatted, err := entity.NewSliceSparseEmbedding(positions, values)
// 		if err != nil {
// 			logging.Log.Errorf(&logging.ContextMap{}, "failed to create sparse embedding %v", err.Error())
// 			return nil, err
// 		}
// 		sparse_search_params, err := entity.NewIndexSparseInvertedSearchParam(0)
// 		if err != nil {
// 			logging.Log.Errorf(&logging.ContextMap{}, "failed to create search params %v", err.Error())
// 			return nil, err
// 		}

// 		sparse_search_request := client.NewANNSearchRequest("sparse_vector", entity.IP, "", []entity.Vector{sparse_vector_formatted}, sparse_search_params, 100)

// 		sub_requests = append(sub_requests, sparse_search_request)
// 	}

// 	// Define Dense Vector Search Request (if defined)
// 	if request.DenseVector != nil {
// 		dense_vectors := []entity.Vector{entity.FloatVector(request.DenseVector)}
// 		dense_search_params, err := entity.NewIndexFlatSearchParam()
// 		if err != nil {
// 			logging.Log.Errorf(&logging.ContextMap{}, "failed to create search params %v", err.Error())
// 			return nil, err
// 		}
// 		dense_vector_search_request := client.NewANNSearchRequest("dense_vector", entity.COSINE, *request.Filter, dense_vectors, dense_search_params, 100)

// 		sub_requests = append(sub_requests, dense_vector_search_request)
// 	}

// 	// create reranker
// 	reranker_weights := []float64{sparseVectorWeight, denseVectorWeight}
// 	if request.SparseVector == nil || request.DenseVector == nil {
// 		reranker_weights = []float64{1.0}
// 	}
// 	reranker := client.NewWeightedReranker(reranker_weights)

// 	opt := client.SearchQueryOptionFunc(func(option *client.SearchQueryOption) {
// 		option.Offset = 0
// 	})

// 	response, err := milvusClient.HybridSearch(context.TODO(), request.CollectionName, []string{}, numResults, request.OutputFields, reranker, sub_requests, opt)
// 	if err != nil {
// 		logging.Log.Errorf(&logging.ContextMap{}, "failed to search collection %v", err.Error())
// 		return nil, err
// 	}

// 	return response, nil
// }

// // Returns nil if threshhold is not reached
// func QueryUserGuideByName(sectionName string, collectionName string, milvusClient client.Client, maxResults int) (response client.ResultSet, func_error error) {
// 	defer func() {
// 		r := recover()
// 		if r != nil {
// 			logging.Log.Errorf(&logging.ContextMap{}, "Panic occured in QueryUserGuideByName: %v", r)
// 			func_error = r.(error)
// 		}
// 	}()

// 	filter := "section_name == '" + sectionName + "'"
// 	outputFields := []string{"document_name", "section_name", "previous_chunk", "next_chunk", "text", "level", "parent_section_name", "guid"}

// 	response, err := milvusClient.Query(context.TODO(), collectionName, nil, filter, outputFields)
// 	if err != nil {
// 		logging.Log.Errorf(&logging.ContextMap{}, "failed to search collection %v", err.Error())
// 		return nil, err
// 	}

// 	return response, nil
// }
