package milvus

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/ansys/allie-flowkit/pkg/internalstates"
	"github.com/ansys/allie-flowkit/pkg/privatefunctions/generic"
	"github.com/ansys/allie-sharedtypes/pkg/config"
	"github.com/ansys/allie-sharedtypes/pkg/logging"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

// Global variables
var (
	CallsBatchSize          = 10000
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
			logging.Log.Errorf(internalstates.Ctx, "Panic initialize: %v", r)
			funcError = r.(error)
			return
		}
	}()
	var err error

	// Create Milvus client
	milvusClient, err = newClient()
	if err != nil {
		logging.Log.Errorf(internalstates.Ctx, "error during NewMilvusClient: %s", err.Error())
		return nil, err
	}

	// Load all existing collections
	collections, err := listCollections(milvusClient)
	if err != nil {
		logging.Log.Errorf(internalstates.Ctx, "error during ListCollections: %s", err.Error())
		return nil, err
	}

	for _, collection := range collections {
		err := loadCollection(collection, milvusClient)
		if err != nil {
			logging.Log.Errorf(internalstates.Ctx, "error during LoadCollection: %s", err.Error())
			return nil, err
		}
	}

	logging.Log.Info(internalstates.Ctx, "Initialized Milvus")

	return milvusClient, nil
}

func newClient() (milvusClient client.Client, funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(internalstates.Ctx, "Panic in NewMilvusClient: %v", r)
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
			logging.Log.Warnf(internalstates.Ctx, "Timeout error during client.NewClient: %s (Retry %d/%d)", err.Error(), retry+1, milvusConnectionRetries)
			continue
		}

		// If the error is not a timeout error, log and return the error
		logging.Log.Errorf(internalstates.Ctx, "Error during client.NewClient: %s", err.Error())
		return nil, err
	}

	return nil, errors.New("unable to create Milvus client after maximum retries")
}

// listCollections
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
			logging.Log.Errorf(internalstates.Ctx, "Panic in ListCollections: %v", r)
			funcError = r.(error)
			return
		}
	}()
	listColl, err := milvusClient.ListCollections(
		context.Background(),
	)
	if err != nil {
		logging.Log.Errorf(internalstates.Ctx, "failed to list collections: %s", err.Error())
		return collections, err
	}

	// Create collection slice
	for _, collection := range listColl {
		if collection.Name != config.GlobalConfig.TEMP_COLLECTION_NAME {
			collections = append(collections, collection.Name)
		}
	}

	logging.Log.Infof(internalstates.Ctx, "Collections listed: %v", collections)

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
			logging.Log.Errorf(internalstates.Ctx, "Panic in LoadCollection: %v", r)
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
		logging.Log.Errorf(internalstates.Ctx, "error during milvusClient.LoadCollection: %s", err.Error())
		return err
	}

	logging.Log.Infof(internalstates.Ctx, "Collection loaded: %v", collectionName)

	return nil
}

// CreateCollection creates a collection in the Milvus DB.
//
// Parameters:
//   - schema: Schema of the collection to be created.
//   - milvusClient: Milvus client.
//
// Returns:
//   - error: Error if any issue occurs during creating the collection.
func CreateCollection(schema *entity.Schema, milvusClient client.Client) (funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(internalstates.Ctx, "Panic in CreateCollection: %v", r)
			funcError = r.(error)
			return
		}
	}()
	err := milvusClient.CreateCollection(
		context.Background(), // ctx
		schema,
		2, // shardNum
	)
	if err != nil {
		logging.Log.Errorf(internalstates.Ctx, "Error during CreateCollection: %v", err)
		return err
	}

	logging.Log.Infof(internalstates.Ctx, "Created collection: %s\n", schema.CollectionName)

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
			fieldSchema.DataType = entity.FieldTypeString
		case "bool":
			fieldSchema.DataType = entity.FieldTypeBool
		case "float_vector":
			fieldSchema.DataType = entity.FieldTypeFloatVector
		case "binary_vector":
			fieldSchema.DataType = entity.FieldTypeBinaryVector
		default:
			return nil, fmt.Errorf("unsupported field type: %s", field.Type)
		}

		// Set primary key and auto ID options
		if field.PrimaryKey {
			fieldSchema.PrimaryKey = true
		}
		if field.AutoID {
			fieldSchema.AutoID = true
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
//
// Returns:
//   - error: Error if any issue occurs during sending the data to the Milvus DB.
func InsertData(collectionName string, dataToSend []interface{}) (funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(internalstates.Ctx, "Panic in InsertData: %v", r)
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

		// Send the data batch to Milvus
		err := sendInsertRequest(request)
		if err != nil {
			logging.Log.Errorf(internalstates.Ctx, "Error during sendInsertRequest: %v", err)
			return err
		}

		// Move the start index to the next batch
		startIndex = stopIndex
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
			logging.Log.Errorf(internalstates.Ctx, "Panic in sendInsertRequest: %v", r)
			funcError = r.(error)
			return
		}
	}()

	// Create a MilvusInsertResponse object
	var response MilvusInsertResponse

	// Create the URL for the Milvus insert request
	url := fmt.Sprintf("http://%s:%s/v1/vector/insert", config.GlobalConfig.MILVUS_HOST, config.GlobalConfig.MILVUS_PORT)

	// Send the Milvus insert request and receive the response.
	err, _ := generic.CreatePayloadAndSendHttpRequest(url, request, &response)
	if err != nil {
		logging.Log.Errorf(internalstates.Ctx, "Error in CreatePayloadAndSendHttpRequest: %s", err)
		return err
	}

	// Check the response status code.
	if response.Code != http.StatusOK {
		logging.Log.Errorf(internalstates.Ctx, "Request failed with status code %d and message: %s\n", response.Code, response.Message)
		return errors.New(response.Message)
	}

	logging.Log.Infof(internalstates.Ctx, "Added %v entries to the DB.", response.InsertData.InsertCount)

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
			logging.Log.Errorf(internalstates.Ctx, "Panic in sendQueryRequest: %v", r)
			funcError = r.(error)
			return
		}
	}()

	// Create the URL for the Milvus query request
	url := fmt.Sprintf("http://%s:%s/v1/vector/query", config.GlobalConfig.MILVUS_HOST, config.GlobalConfig.MILVUS_PORT)

	// Send the Milvus query request and receive the response.
	err, _ := generic.CreatePayloadAndSendHttpRequest(url, request, &response)
	if err != nil {
		logging.Log.Errorf(internalstates.Ctx, "Error in CreatePayloadAndSendHttpRequest: %s", err)
		return response, err
	}

	// Check the response status code.
	if response.Code != http.StatusOK {
		logging.Log.Errorf(internalstates.Ctx, "Request failed with status code %d and message: %s\n", response.Code, response.Message)
		return response, errors.New(response.Message)
	}

	return response, nil
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
			logging.Log.Errorf(internalstates.Ctx, "Panic in sendSearchRequest: %v", r)
			funcError = r.(error)
			return
		}
	}()

	// Create the URL for the Milvus search request
	url := fmt.Sprintf("http://%s:%s/v1/vector/search", config.GlobalConfig.MILVUS_HOST, config.GlobalConfig.MILVUS_PORT)

	// Send the Milvus search request and receive the response.
	err, _ := generic.CreatePayloadAndSendHttpRequest(url, request, &response)
	if err != nil {
		logging.Log.Errorf(internalstates.Ctx, "Error in CreatePayloadAndSendHttpRequest: %s", err)
		return response, err
	}

	// Check the response status code.
	if response.Code != http.StatusOK {
		logging.Log.Errorf(internalstates.Ctx, "Request failed with status code %d and message: %s\n", response.Code, response.Message)
		return response, errors.New(response.Message)
	}

	return response, nil
}
