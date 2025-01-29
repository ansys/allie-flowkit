package milvus

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

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

	logging.Log.Info(&logging.ContextMap{}, "Initialized Milvus")

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

	logging.Log.Infof(&logging.ContextMap{}, "Collections listed: %v", collections)

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

	logging.Log.Infof(&logging.ContextMap{}, "Collection loaded: %v", collectionName)

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
			logging.Log.Errorf(&logging.ContextMap{}, "Panic in CreateCollection: %v", r)
			funcError = r.(error)
			return
		}
	}()
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
		logging.Log.Infof(&logging.ContextMap{}, "Collection already exists: %s\n", schema.CollectionName)
		return nil
	}

	// Create collection if it does not exist
	err = milvusClient.CreateCollection(
		context.Background(), // ctx
		schema,
		2, // shardNum
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

	logging.Log.Infof(&logging.ContextMap{}, "Created collection: %s\n", schema.CollectionName)

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

	logging.Log.Info(&logging.ContextMap{}, "Scalar index created")

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
		case "[]string":
			fieldSchema.DataType = entity.FieldTypeArray
			fieldSchema.ElementType = entity.FieldTypeVarChar
			fieldSchema.TypeParams = map[string]string{"max_length": "40000", "max_capacity": "1000"}
		case "bool":
			fieldSchema.DataType = entity.FieldTypeBool
		case "map[string]string":
			fieldSchema.DataType = entity.FieldTypeJSON
		case "[]float32":
			fieldSchema.DataType = entity.FieldTypeFloatVector
		case "map[uint]float32":
			fieldSchema.DataType = entity.FieldTypeSparseVector
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
//
// Returns:
//   - error: Error if any issue occurs during sending the data to the Milvus DB.
func InsertData(collectionName string, dataToSend []interface{}) (funcError error) {
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

		// Send the data batch to Milvus
		err := sendInsertRequest(request)
		if err != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Error during sendInsertRequest: %v", err)
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
	err, _ := generic.CreatePayloadAndSendHttpRequest(url, request, &response)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error in CreatePayloadAndSendHttpRequest: %s", err)
		return err
	}

	// Check the response status code.
	if response.Code != http.StatusOK {
		logging.Log.Errorf(&logging.ContextMap{}, "Request failed with status code %d and message: %s\n", response.Code, response.Message)
		return errors.New(response.Message)
	}

	logging.Log.Infof(&logging.ContextMap{}, "Added %v entries to the DB.", response.InsertData.InsertCount)

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
	err, _ := generic.CreatePayloadAndSendHttpRequest(url, request, &response)
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
	err, _ := generic.CreatePayloadAndSendHttpRequest(url, request, &response)
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

func fromLexicalWeightsToSparseVector(lexicalWeights []map[uint]float32, dimension int) (sparseEmbeddings []entity.SparseEmbedding, err error) {
	sparsePositions := make([]uint32, 0)
	sparseValues := make([]float32, 0)
	var sparseEmbedding entity.SparseEmbedding

	for _, weightMap := range lexicalWeights {
		for key, value := range weightMap {
			sparsePositions = append(sparsePositions, uint32(key))
			sparseValues = append(sparseValues, value)
		}

		// Now create the sparse embedding
		sparseEmbedding, err = entity.NewSliceSparseEmbedding(sparsePositions, sparseValues)
		if err != nil {
			return nil, fmt.Errorf("failed to create sparse embedding: %w", err)
		}

		// Append the sparse embedding to the list
		sparseEmbeddings = append(sparseEmbeddings, sparseEmbedding)
	}

	return sparseEmbeddings, nil
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
