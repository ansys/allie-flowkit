package qdrant

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ansys/aali-sharedtypes/pkg/config"
	"github.com/ansys/aali-sharedtypes/pkg/logging"
	"github.com/ansys/aali-sharedtypes/pkg/sharedtypes"
	"github.com/qdrant/go-client/qdrant"
)

// Create a new qdrant client from your config.
func QdrantClient() (*qdrant.Client, error) {
	return qdrant.NewClient(&qdrant.Config{
		Host: config.GlobalConfig.QDRANT_HOST,
		Port: config.GlobalConfig.QDRANT_PORT,
	})

}

func CreateCollectionIfNotExists(ctx context.Context, client *qdrant.Client, collectionName string, vectorsConfig *qdrant.VectorsConfig, sparseVectorsConfig *qdrant.SparseVectorConfig) error {
	exists, err := client.CollectionExists(ctx, collectionName)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName:      collectionName,
		VectorsConfig:       vectorsConfig,
		SparseVectorsConfig: sparseVectorsConfig,
	})
}

// convert from one type to another, using an intermediate JSON representation
func TryIntoWithJson[F any, T any](from F) (T, error) {
	var to T

	jsonBytes, err := json.Marshal(from)
	if err != nil {
		return to, fmt.Errorf("error marshaling from: %q", err)
	}

	err = json.Unmarshal(jsonBytes, &to)
	if err != nil {
		return to, fmt.Errorf("error unmarshaling bytes: %q", err)
	}
	return to, nil
}

// Get a qdrant vector distance metric from string
//
// Available options are:
//   - cosine
//   - dot
//   - euclid
//   - manhattan
func VectorDistance(distance string) qdrant.Distance {
	switch strings.ToLower(distance) {
	case "cosine":
		return qdrant.Distance_Cosine
	case "dot":
		return qdrant.Distance_Dot
	case "euclid":
		return qdrant.Distance_Euclid
	case "manhattan":
		return qdrant.Distance_Manhattan
	default:
		logging.Log.Warnf(&logging.ContextMap{}, "unknown vector distance metric %q", distance)
		return qdrant.Distance_UnknownDistance
	}
}

// Transform `sharedtypes.DbFilters` into a qdrant filter.
func DbFiltersAsQdrant(dbFilters sharedtypes.DbFilters) *qdrant.Filter {
	filter := qdrant.Filter{}
	asQdrantFilters := map[string]AsQdrantFilterConditions{
		"guid":          keywordFilter(dbFilters.GuidFilter),
		"document_id":   keywordFilter(dbFilters.DocumentIdFilter),
		"document_name": keywordFilter(dbFilters.DocumentNameFilter),
		"level":         keywordFilter(dbFilters.LevelFilter),
		"tags":          dbArrayFilter(dbFilters.TagsFilter),
		"keywords":      dbArrayFilter(dbFilters.KeywordsFilter),
	}
	for _, metadataFilter := range dbFilters.MetadataFilter {
		asQdrantFilters[metadataFilter.FieldName] = dbArrayFilter{
			NeedAll:    metadataFilter.NeedAll,
			FilterData: metadataFilter.FilterData,
		}
	}
	for field, asQdrantFilter := range asQdrantFilters {
		filter.Must = append(filter.Must, asQdrantFilter.AsQdrantFilterConditions(field)...)
	}

	return &filter
}

type AsQdrantFilterConditions interface {
	AsQdrantFilterConditions(field string) []*qdrant.Condition
}

type keywordFilter []string

func (kwFilt keywordFilter) AsQdrantFilterConditions(field string) []*qdrant.Condition {
	if len(kwFilt) > 0 {
		return []*qdrant.Condition{
			qdrant.NewMatchKeywords(field, kwFilt...),
		}
	} else {
		return []*qdrant.Condition{}
	}
}

type dbArrayFilter sharedtypes.DbArrayFilter

func (dbArrFilt dbArrayFilter) AsQdrantFilterConditions(field string) []*qdrant.Condition {
	conditions := []*qdrant.Condition{}
	if len(dbArrFilt.FilterData) > 0 {
		if dbArrFilt.NeedAll {
			for _, tag := range dbArrFilt.FilterData {
				conditions = append(conditions, qdrant.NewMatchKeyword(field, tag))
			}
		} else {
			conditions = append(conditions, qdrant.NewMatchKeywords(field, dbArrFilt.FilterData...))
		}
	}
	return conditions
}

// RetrieveLeafNodes retrieves all leaf nodes from the similarity search result branch (ultimate children containing the original document).
//
// Parameters:
//   - ctx: ContextMap.
//   - client: the qdrant client
//   - collectionName: Name of the collection in the qdrant database to retrieve the leaves from.
//   - outputFields: Fields to return in the response.
//   - data: Data to retrieve the leaf nodes for.
//
// Returns:
//   - error: Error if any issue occurs while retrieving the leaves.
func RetrieveLeafNodes(ctx *logging.ContextMap, client *qdrant.Client, collectionName string, outputFields []string, data *[]sharedtypes.DbResponse) (funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(ctx, "Panic in RetrieveLeafNodes: %v", r)
			funcError = r.(error)
			return
		}
	}()
	ctx = ctx.Copy()

	// for each dbresponse, get all leaf nodes that are in the same document
	queries := make([]*qdrant.QueryPoints, len(*data))
	for i, dbresp := range *data {
		queries[i] = &qdrant.QueryPoints{
			CollectionName: collectionName,
			Filter: &qdrant.Filter{
				Must: []*qdrant.Condition{
					qdrant.NewMatchKeyword("document_id", dbresp.DocumentId),
					qdrant.NewMatchKeyword("level", "leaf"),
				},
			},
			WithVectors: qdrant.NewWithVectorsEnable(false),
			WithPayload: qdrant.NewWithPayloadInclude(outputFields...),
		}
	}
	batchResults, err := client.QueryBatch(context.TODO(), &qdrant.QueryBatchPoints{
		CollectionName: collectionName,
		QueryPoints:    queries,
	})
	if err != nil {
		logging.Log.Errorf(ctx, "error in qdrant batch query: %q", err)
		return err
	}

	// convert qdrant to aali types
	for i, batchRes := range batchResults {
		leaves := make([]sharedtypes.DbData, len(batchRes.Result))
		for j, point := range batchRes.Result {
			dbresp, err := TryIntoWithJson[map[string]*qdrant.Value, sharedtypes.DbData](point.Payload)
			if err != nil {
				logging.Log.Errorf(ctx, "error converting qdrant payload: %q", err)
				return err
			}
			leaves[j] = dbresp
		}
		(*data)[i].LeafNodes = leaves
	}
	return nil
}

// RetrieveParentNodes retrieves the parent node for each of the documents provided.
//
// Parameters:
//   - ctx: ContextMap.
//   - client: the qdrant client
//   - collectionName: Name of the collection in the qdrant database to retrieve the parents from.
//   - outputFields: Fields to return in the response.
//   - data: Data to retrieve the parent nodes for.
//
// Returns:
//   - error: Error if any issue occurs while retrieving the parents.
func RetrieveParentNodes(ctx *logging.ContextMap, client *qdrant.Client, collectionName string, outputFields []string, data *[]sharedtypes.DbResponse) (funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(ctx, "Panic in RetrieveParentNodes: %v", r)
			funcError = r.(error)
			return
		}
	}()
	ctx = ctx.Copy()

	// for each dbresponse, get the parent document
	queries := make([]*qdrant.QueryPoints, len(*data))
	limit := uint64(1)
	for i, dbresp := range *data {
		queries[i] = &qdrant.QueryPoints{
			CollectionName: collectionName,
			Filter: &qdrant.Filter{
				Must: []*qdrant.Condition{
					qdrant.NewHasID(qdrant.NewIDUUID(dbresp.ParentId.String())),
				},
			},
			Limit:       &limit,
			WithVectors: qdrant.NewWithVectorsEnable(false),
			WithPayload: qdrant.NewWithPayloadInclude(outputFields...),
		}
	}
	batchResults, err := client.QueryBatch(context.TODO(), &qdrant.QueryBatchPoints{
		CollectionName: collectionName,
		QueryPoints:    queries,
	})
	if err != nil {
		logging.Log.Errorf(ctx, "error in qdrant batch query: %q", err)
		return err
	}

	// convert qdrant to aali types
	for i, batchRes := range batchResults {
		switch len(batchRes.Result) {
		case 0:
			continue
		case 1:
			parent, err := TryIntoWithJson[map[string]*qdrant.Value, sharedtypes.DbData](batchRes.Result[0].Payload)
			if err != nil {
				logging.Log.Errorf(ctx, "error converting qdrant payload: %q", err)
				return err
			}
			(*data)[i].Parent = &parent
		default:
			return fmt.Errorf("got more than 1 parent node (%d), but this should be impossible", len(batchRes.Result))
		}
	}
	return nil
}

// RetrieveChildNodes retrieves the child nodes for each of the documents provided.
//
// Parameters:
//   - ctx: ContextMap.
//   - client: the qdrant client
//   - collectionName: Name of the collection in the qdrant database to retrieve the children from.
//   - outputFields: Fields to return in the response.
//   - data: Data to retrieve the children for.
//
// Returns:
//   - error: Error if any issue occurs while retrieving the children.
func RetrieveChildNodes(ctx *logging.ContextMap, client *qdrant.Client, collectionName string, outputFields []string, data *[]sharedtypes.DbResponse) (funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(ctx, "Panic in RetrieveChildNodes: %v", r)
			funcError = r.(error)
			return
		}
	}()
	ctx = ctx.Copy()

	// for each dbresponse, get the parent document
	queries := make([]*qdrant.QueryPoints, len(*data))
	for i, dbresp := range *data {
		childIds := make([]*qdrant.PointId, len(dbresp.ChildIds))
		for j, childid := range dbresp.ChildIds {
			childIds[j] = qdrant.NewIDUUID(childid.String())
		}
		queries[i] = &qdrant.QueryPoints{
			CollectionName: collectionName,
			Filter: &qdrant.Filter{
				Must: []*qdrant.Condition{
					qdrant.NewHasID(childIds...),
				},
			},
			WithVectors: qdrant.NewWithVectorsEnable(false),
			WithPayload: qdrant.NewWithPayloadInclude(outputFields...),
		}
	}
	batchResults, err := client.QueryBatch(context.TODO(), &qdrant.QueryBatchPoints{
		CollectionName: collectionName,
		QueryPoints:    queries,
	})
	if err != nil {
		logging.Log.Errorf(ctx, "error in qdrant batch query: %q", err)
		return err
	}

	// convert qdrant to aali types
	for i, batchRes := range batchResults {
		childrenDbData := make([]sharedtypes.DbData, len(batchRes.Result))
		for j, point := range batchRes.Result {
			child, err := TryIntoWithJson[map[string]*qdrant.Value, sharedtypes.DbData](point.Payload)
			if err != nil {
				logging.Log.Errorf(ctx, "error converting qdrant payload: %q", err)
				return err
			}
			childrenDbData[j] = child
		}
		(*data)[i].Children = childrenDbData
	}
	return nil
}

// RetrieveDirectSiblingNodes retrieves the nodes associated with the next & previous sibling (if any) for each of the documents provided.
//
// Parameters:
//   - ctx: ContextMap.
//   - client: the qdrant client
//   - collectionName: Name of the collection in the qdrant database to retrieve the siblings from.
//   - outputFields: Fields to return in the response.
//   - data: Data to retrieve the siblings for.
//
// Returns:
//   - error: Error if any issue occurs while retrieving the siblings.
func RetrieveDirectSiblingNodes(ctx *logging.ContextMap, client *qdrant.Client, collectionName string, outputFields []string, data *[]sharedtypes.DbResponse) (funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(ctx, "Panic in RetrieveDirectSiblingNodes: %v", r)
			funcError = r.(error)
			return
		}
	}()
	ctx = ctx.Copy()

	// for each dbresponse, get the parent document
	queries := make([]*qdrant.QueryPoints, len(*data))
	for i, dbresp := range *data {
		siblingIds := []*qdrant.PointId{}
		if dbresp.PreviousSiblingId != nil {
			siblingIds = append(siblingIds, qdrant.NewIDUUID(dbresp.PreviousSiblingId.String()))
		}
		if dbresp.NextSiblingId != nil {
			siblingIds = append(siblingIds, qdrant.NewIDUUID(dbresp.NextSiblingId.String()))
		}

		queries[i] = &qdrant.QueryPoints{
			CollectionName: collectionName,
			Filter: &qdrant.Filter{
				Must: []*qdrant.Condition{
					qdrant.NewHasID(siblingIds...),
				},
			},
			WithVectors: qdrant.NewWithVectorsEnable(false),
			WithPayload: qdrant.NewWithPayloadInclude(outputFields...),
		}
	}
	batchResults, err := client.QueryBatch(context.TODO(), &qdrant.QueryBatchPoints{
		CollectionName: collectionName,
		QueryPoints:    queries,
	})
	if err != nil {
		logging.Log.Errorf(ctx, "error in qdrant batch query: %q", err)
		return err
	}

	// convert qdrant to aali types
	for i, batchRes := range batchResults {
		siblingsDbData := make([]sharedtypes.DbData, len(batchRes.Result))
		for j, point := range batchRes.Result {
			sibling, err := TryIntoWithJson[map[string]*qdrant.Value, sharedtypes.DbData](point.Payload)
			if err != nil {
				logging.Log.Errorf(ctx, "error converting qdrant payload: %q", err)
				return err
			}
			siblingsDbData[j] = sibling
		}
		(*data)[i].Siblings = siblingsDbData
	}
	return nil
}
