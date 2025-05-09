package externalfunctions

import (
	"context"
	"fmt"

	"github.com/ansys/aali-sharedtypes/pkg/logging"
	"github.com/ansys/aali-sharedtypes/pkg/sharedtypes"
	"github.com/ansys/allie-flowkit/pkg/privatefunctions/graphdb"
	qdrant_utils "github.com/ansys/allie-flowkit/pkg/privatefunctions/qdrant"
	"github.com/qdrant/go-client/qdrant"
)

// SendVectorsToKnowledgeDB sends the given vector to the KnowledgeDB and
// returns the most relevant data. The number of results is specified in the
// config file. The keywords are used to filter the results. The min score
// filter is also specified in the config file. If it is not specified, the
// default value is used.
//
// The function returns the most relevant data.
//
// Tags:
//   - @displayName: Similarity Search
//
// Parameters:
//   - vector: the vector to be sent to the KnowledgeDB
//   - keywords: the keywords to be used to filter the results
//   - keywordsSearch: the flag to enable the keywords search
//   - collection: the collection name
//   - similaritySearchResults: the number of results to be returned
//   - similaritySearchMinScore: the minimum score for the results
//
// Returns:
//   - databaseResponse: an array of the most relevant data
func SendVectorsToKnowledgeDB(vector []float32, keywords []string, keywordsSearch bool, collection string, similaritySearchResults int, similaritySearchMinScore float64) (databaseResponse []sharedtypes.DbResponse) {
	logCtx := &logging.ContextMap{}
	client, err := qdrant_utils.QdrantClient()
	if err != nil {
		logPanic(logCtx, "unable to create qdrant client: %q", err)
	}

	// perform the qdrant query
	filter := qdrant.Filter{
		Must: []*qdrant.Condition{
			qdrant.NewMatch("level", "leaf"),
		},
	}
	if keywordsSearch {
		filter.Must = append(filter.Must, qdrant.NewMatchKeywords("keywords", keywords...))

	}
	limit := uint64(similaritySearchResults)
	scoreThreshold := float32(similaritySearchMinScore)
	query := qdrant.QueryPoints{
		CollectionName: collection,
		Query:          qdrant.NewQueryDense(vector),
		Limit:          &limit,
		ScoreThreshold: &scoreThreshold,
		Filter:         &filter,
		WithVectors:    qdrant.NewWithVectorsEnable(false),
		WithPayload:    qdrant.NewWithPayloadInclude("guid", "document_id", "document_name", "summary", "keywords", "text"),
	}
	scoredPoints, err := client.Query(context.TODO(), &query)
	if err != nil {
		logPanic(logCtx, "error in qdrant query: %q", err)
	}
	logging.Log.Debugf(logCtx, "Got %d points from qdrant query", len(scoredPoints))

	// transform qdrant result into allie type
	dbResponses := make([]sharedtypes.DbResponse, len(scoredPoints))
	for i, scoredPoint := range scoredPoints {
		logging.Log.Debugf(&logging.ContextMap{}, "Result #%d:", i)
		logging.Log.Debugf(&logging.ContextMap{}, "Similarity score: %v", scoredPoint.Score)
		dbResponse, err := qdrant_utils.TryIntoWithJson[map[string]*qdrant.Value, sharedtypes.DbResponse](scoredPoint.Payload)
		if err != nil {
			errMsg := fmt.Sprintf("error converting qdrant payload to dbResponse: %q", err)
			logging.Log.Errorf(logCtx, "%s", errMsg)
			panic(errMsg)
		}

		logging.Log.Debugf(&logging.ContextMap{}, "Similarity file id: %v", dbResponse.DocumentId)
		logging.Log.Debugf(&logging.ContextMap{}, "Similarity file name: %v", dbResponse.DocumentName)
		logging.Log.Debugf(&logging.ContextMap{}, "Similarity summary: %v", dbResponse.Summary)

		// Add the result to the list
		dbResponses[i] = dbResponse
	}
	return dbResponses
}

// GetListCollections retrieves the list of collections from the KnowledgeDB.
//
// Tags:
//   - @displayName: List Collections
//
// The function returns the list of collections.
//
// Parameters:
//   - knowledgeDbEndpoint: the KnowledgeDB endpoint
//
// Returns:
//   - collectionsList: the list of collections
func GetListCollections() (collectionsList []string) {
	logCtx := &logging.ContextMap{}
	client, err := qdrant_utils.QdrantClient()
	if err != nil {
		logPanic(logCtx, "unable to create qdrant client: %q", err)
	}

	collectionsList, err = client.ListCollections(context.TODO())
	if err != nil {
		logPanic(logCtx, "unable to list qdrant collections: %q", err)
	}
	return collectionsList
}

// RetrieveDependencies retrieves the dependencies of the specified source node.
//
// The function returns the list of dependencies.
//
// Tags:
//   - @displayName: Retrieve Dependencies
//
// Parameters:
//   - collectionName: the name of the collection to which the data objects will be added.
//   - relationshipName: the name of the relationship to retrieve dependencies for.
//   - relationshipDirection: the direction of the relationship to retrieve dependencies for.
//   - sourceDocumentId: the document ID of the source node.
//   - nodeTypesFilter: filter based on node types.
//   - maxHopsNumber: maximum number of hops to traverse.
//
// Returns:
//   - dependenciesIds: the list of dependencies
func RetrieveDependencies(
	collectionName string,
	relationshipName string,
	relationshipDirection string,
	sourceDocumentId string,
	nodeTypesFilter sharedtypes.DbArrayFilter,
	maxHopsNumber int) (dependenciesIds []string) {
	ctx := &logging.ContextMap{}
	dependenciesIds, err := graphdb.GraphDbDriver.RetrieveDependencies(
		ctx,
		collectionName,
		relationshipName,
		relationshipDirection,
		sourceDocumentId,
		nodeTypesFilter,
		[]string{},
		maxHopsNumber,
	)
	if err != nil {
		logPanic(nil, "unable to retrieve dependencies: %q", err)
	}
	return dependenciesIds
}

// GeneralGraphDbQuery executes the given Cypher query and returns the response.
//
// The function returns the neo4j response.
//
// Tags:
//   - @displayName: General Graph DB Query
//
// Parameters:
//   - query: the Neo4j query to be executed.
//
// Returns:
//   - databaseResponse: the Neo4j response
func GeneralGraphDbQuery(query string) []map[string]any {
	res, err := graphdb.GraphDbDriver.WriteCypherQuery(query)
	if err != nil {
		logPanic(nil, "error executing cypher query: %q", err)
	}
	return res
}

// GeneralQuery performs a general query in the KnowledgeDB.
//
// The function returns the query results.
//
// Tags:
//   - @displayName: Query
//
// Parameters:
//   - collectionName: the name of the collection to which the data objects will be added.
//   - maxRetrievalCount: the maximum number of results to be retrieved.
//   - outputFields: the fields to be included in the output.
//   - filters: the filter for the query.
//
// Returns:
//   - databaseResponse: the query results
func GeneralQuery(collectionName string, maxRetrievalCount int, outputFields []string, filters sharedtypes.DbFilters) (databaseResponse []sharedtypes.DbResponse) {
	logCtx := &logging.ContextMap{}
	client, err := qdrant_utils.QdrantClient()
	if err != nil {
		logPanic(logCtx, "unable to create qdrant client: %q", err)
	}

	// perform the qdrant query
	limit := uint64(maxRetrievalCount)
	filter := qdrant_utils.DbFiltersAsQdrant(filters)
	query := qdrant.QueryPoints{
		CollectionName: collectionName,
		Limit:          &limit,
		Filter:         filter,
		WithVectors:    qdrant.NewWithVectorsEnable(false),
		WithPayload:    qdrant.NewWithPayloadInclude(outputFields...),
	}
	scoredPoints, err := client.Query(context.TODO(), &query)
	if err != nil {
		logPanic(logCtx, "error in qdrant query: %q", err)
	}
	logging.Log.Debugf(logCtx, "Got %d points from qdrant query", len(scoredPoints))

	// convert to allie type
	databaseResponse = make([]sharedtypes.DbResponse, len(scoredPoints))
	for i, scoredPoint := range scoredPoints {

		dbResponse, err := qdrant_utils.TryIntoWithJson[map[string]*qdrant.Value, sharedtypes.DbResponse](scoredPoint.Payload)
		if err != nil {
			logPanic(logCtx, "error converting qdrant payload to dbResponse: %q", err)
		}
		databaseResponse[i] = dbResponse
	}
	return databaseResponse
}

// SimilaritySearch performs a similarity search in the KnowledgeDB.
//
// The function returns the similarity search results.
//
// Tags:
//   - @displayName: Similarity Search (Filtered)
//
// Parameters:
//   - collectionName: the name of the collection to which the data objects will be added.
//   - embeddedVector: the embedded vector used for searching.
//   - maxRetrievalCount: the maximum number of results to be retrieved.
//   - outputFields: the fields to be included in the output.
//   - filters: the filter for the query.
//   - minScore: the minimum score filter.
//   - getLeafNodes: flag to indicate whether to retrieve all the leaf nodes in the result node branch.
//   - getSiblings: flag to indicate whether to retrieve the previous and next node to the result nodes.
//   - getParent: flag to indicate whether to retrieve the parent object.
//   - getChildren: flag to indicate whether to retrieve the children objects.
//
// Returns:
//   - databaseResponse: the similarity search results
func SimilaritySearch(
	collectionName string,
	embeddedVector []float32,
	maxRetrievalCount int,
	outputFields []string,
	filters sharedtypes.DbFilters,
	minScore float64,
	getLeafNodes bool,
	getSiblings bool,
	getParent bool,
	getChildren bool) (databaseResponse []sharedtypes.DbResponse) {
	logCtx := &logging.ContextMap{}
	client, err := qdrant_utils.QdrantClient()
	if err != nil {
		logPanic(logCtx, "unable to create qdrant client: %q", err)
	}

	// perform the qdrant query
	limit := uint64(maxRetrievalCount)
	scoreThreshold := float32(minScore)
	query := qdrant.QueryPoints{
		CollectionName: collectionName,
		Query:          qdrant.NewQueryDense(embeddedVector),
		Limit:          &limit,
		ScoreThreshold: &scoreThreshold,
		Filter:         qdrant_utils.DbFiltersAsQdrant(filters),
		WithVectors:    qdrant.NewWithVectorsEnable(false),
		WithPayload:    qdrant.NewWithPayloadInclude(outputFields...),
	}
	scoredPoints, err := client.Query(context.TODO(), &query)
	if err != nil {
		logPanic(logCtx, "error in qdrant query: %q", err)
	}
	logging.Log.Debugf(logCtx, "Got %d points from qdrant query", len(scoredPoints))

	// convert to allie type
	databaseResponse = make([]sharedtypes.DbResponse, len(scoredPoints))
	for i, scoredPoint := range scoredPoints {

		dbResponse, err := qdrant_utils.TryIntoWithJson[map[string]*qdrant.Value, sharedtypes.DbResponse](scoredPoint.Payload)
		if err != nil {
			logPanic(logCtx, "error converting qdrant payload to dbResponse: %q", err)
		}
		databaseResponse[i] = dbResponse
	}

	// get related nodes if requested
	if getLeafNodes {
		logging.Log.Debugf(logCtx, "getting leaf nodes")
		err := qdrant_utils.RetrieveLeafNodes(logCtx, client, collectionName, outputFields, &databaseResponse)
		if err != nil {
			logPanic(logCtx, "error getting leaf nodes: %q", err)
		}
	}
	if getSiblings {
		logging.Log.Debugf(logCtx, "getting sibling nodes")
		err := qdrant_utils.RetrieveDirectSiblingNodes(logCtx, client, collectionName, outputFields, &databaseResponse)
		if err != nil {
			logPanic(logCtx, "error getting sibling nodes: %q", err)
		}
	}
	if getParent {
		logging.Log.Debugf(logCtx, "getting parent nodes")
		err := qdrant_utils.RetrieveParentNodes(logCtx, client, collectionName, outputFields, &databaseResponse)
		if err != nil {
			logPanic(logCtx, "error getting parent nodes: %q", err)
		}
	}
	if getChildren {
		logging.Log.Debugf(logCtx, "getting child nodes")
		err := qdrant_utils.RetrieveChildNodes(logCtx, client, collectionName, outputFields, &databaseResponse)
		if err != nil {
			logPanic(logCtx, "error getting child nodes: %q", err)
		}
	}
	return databaseResponse
}

// CreateKeywordsDbFilter creates a keywords filter for the KnowledgeDB.
//
// The function returns the keywords filter.
//
// Tags:
//   - @displayName: Keywords Filter
//
// Parameters:
//   - keywords: the keywords to be used for the filter
//   - needAll: flag to indicate whether all keywords are needed
//
// Returns:
//   - databaseFilter: the keywords filter
func CreateKeywordsDbFilter(keywords []string, needAll bool) (databaseFilter sharedtypes.DbArrayFilter) {
	var keywordsFilters sharedtypes.DbArrayFilter

	// -- Add the keywords filter if needed
	if len(keywords) > 0 {
		keywordsFilters = createDbArrayFilter(keywords, needAll)
	}

	return keywordsFilters
}

// CreateTagsDbFilter creates a tags filter for the KnowledgeDB.
//
// The function returns the tags filter.
//
// Tags:
//   - @displayName: Tags Filter
//
// Parameters:
//   - tags: the tags to be used for the filter
//   - needAll: flag to indicate whether all tags are needed
//
// Returns:
//   - databaseFilter: the tags filter
func CreateTagsDbFilter(tags []string, needAll bool) (databaseFilter sharedtypes.DbArrayFilter) {
	var tagsFilters sharedtypes.DbArrayFilter

	// -- Add the tags filter if needed
	if len(tags) > 0 {
		tagsFilters = createDbArrayFilter(tags, needAll)
	}

	return tagsFilters
}

// CreateMetadataDbFilter creates a metadata filter for the KnowledgeDB.
//
// The function returns the metadata filter.
//
// Tags:
//   - @displayName: Metadata Filter
//
// Parameters:
//   - fieldName: the name of the field
//   - fieldType: the type of the field
//   - filterData: the filter data
//   - needAll: flag to indicate whether all data is needed
//
// Returns:
//   - databaseFilter: the metadata filter
func CreateMetadataDbFilter(fieldName string, fieldType string, filterData []string, needAll bool) (databaseFilter sharedtypes.DbJsonFilter) {
	return createDbJsonFilter(fieldName, fieldType, filterData, needAll)
}

// CreateDbFilter creates a filter for the KnowledgeDB.
//
// The function returns the filter.
//
// Tags:
//   - @displayName: Create Filter
//
// Parameters:
//   - guid: the guid filter
//   - documentId: the document ID filter
//   - documentName: the document name filter
//   - level: the level filter
//   - tags: the tags filter
//   - keywords: the keywords filter
//   - metadata: the metadata filter
//
// Returns:
//   - databaseFilter: the filter
func CreateDbFilter(
	guid []string,
	documentId []string,
	documentName []string,
	level []string,
	tags sharedtypes.DbArrayFilter,
	keywords sharedtypes.DbArrayFilter,
	metadata []sharedtypes.DbJsonFilter) (databaseFilter sharedtypes.DbFilters) {
	var filters sharedtypes.DbFilters

	// -- Add the guid filter if needed
	if len(guid) > 0 {
		filters.GuidFilter = guid
	}

	// -- Add the document ID filter if needed
	if len(documentId) > 0 {
		filters.DocumentIdFilter = documentId
	}

	// -- Add the document name filter if needed
	if len(documentName) > 0 {
		filters.DocumentNameFilter = documentName
	}

	// -- Add the level filter if needed
	if len(level) > 0 {
		filters.LevelFilter = level
	}

	// -- Add the tags filter if needed
	if len(tags.FilterData) > 0 {
		filters.TagsFilter = tags
	}

	// -- Add the keywords filter if needed
	if len(keywords.FilterData) > 0 {
		filters.KeywordsFilter = keywords
	}

	// -- Add the metadata filter if needed
	if len(metadata) > 0 {
		filters.MetadataFilter = metadata
	}

	return filters
}

// AddDataRequest sends a request to the add_data endpoint.
//
// Tags:
//   - @displayName: Add Data
//
// Parameters:
//   - collectionName: name of the collection the request is sent to.
//   - data: the data to add.
func AddDataRequest(collectionName string, documentData []sharedtypes.DbData) {
	points := make([]*qdrant.PointStruct, len(documentData))
	for i, doc := range documentData {
		id := qdrant.NewIDUUID(doc.Guid.String())
		vector := qdrant.NewVectorsDense(doc.Embedding)
		payloadMap, err := qdrant_utils.TryIntoWithJson[sharedtypes.DbData, map[string]any](doc)
		if err != nil {
			logPanic(nil, "unable to transform document data to json: %q", err)
		}
		delete(payloadMap, "guid")
		delete(payloadMap, "embedding")
		points[i] = &qdrant.PointStruct{
			Id:      id,
			Vectors: vector,
			Payload: qdrant.NewValueMap(payloadMap),
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

// CreateCollectionRequest sends a request to the collection endpoint.
//
// Tags:
//   - @displayName: Create Collection
//
// Parameters:
//   - collectionName: the name of the collection to create.
//   - vectorSize: the length of the vector embeddings
//   - vectorDistance: the vector similarity distance algorithm to use for the vector index (cosine, dot, euclid, manhattan)
func CreateCollectionRequest(collectionName string, vectorSize uint64, vectorDistance string) {
	logCtx := &logging.ContextMap{}

	client, err := qdrant_utils.QdrantClient()
	if err != nil {
		logPanic(logCtx, "unable to create qdrant client: %q", err)
	}

	ctx := context.TODO()

	// check if collection already exists
	collectionExists, err := client.CollectionExists(ctx, collectionName)
	if err != nil {
		logPanic(logCtx, "unable to determine if collection already exists: %v", err)
	}
	if collectionExists {
		logging.Log.Debugf(logCtx, "collection %q already exists, skipping creation", collectionName)
		return
	}

	// create the collection
	err = client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: collectionName,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     vectorSize,
			Distance: qdrant_utils.VectorDistance(vectorDistance),
		}),
	})
	if err != nil {
		logPanic(logCtx, "failed to create collection: %q", err)
	}
	logging.Log.Debugf(logCtx, "Created collection: %s", collectionName)

	// now create the default indexes (these are the things that other knowledgedb functions filter/search on)
	// does ID need to be indexed?
	indexes := []struct {
		name      string
		fieldType qdrant.FieldType
	}{
		{"level", qdrant.FieldType_FieldTypeKeyword},
		{"keywords", qdrant.FieldType_FieldTypeKeyword},
		{"document_id", qdrant.FieldType_FieldTypeKeyword},
		{"tags", qdrant.FieldType_FieldTypeKeyword},
	}
	for _, index := range indexes {
		request := qdrant.CreateFieldIndexCollection{
			CollectionName: collectionName,
			FieldName:      index.name,
			FieldType:      &index.fieldType,
		}
		res, err := client.CreateFieldIndex(ctx, &request)
		if err != nil {
			logPanic(logCtx, "error creating payload index on %q: %v", index.name, err)
		}
		logging.Log.Debugf(logCtx, "created payload index on %q: %q", index.name, res.Status)
	}
}
