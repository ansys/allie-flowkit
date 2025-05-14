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

package qdrant

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ansys/aali-sharedtypes/pkg/config"
	"github.com/ansys/aali-sharedtypes/pkg/logging"
	"github.com/ansys/aali-sharedtypes/pkg/sharedtypes"
	"github.com/google/uuid"
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
		"tags[]":        dbArrayFilter(dbFilters.TagsFilter),
		"keywords[]":    dbArrayFilter(dbFilters.KeywordsFilter),
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
//   - data: Data to retrieve the leaf nodes for.
//
// Returns:
//   - error: Error if any issue occurs while retrieving the leaves.
func RetrieveLeafNodes(ctx *logging.ContextMap, client *qdrant.Client, collectionName string, data *[]sharedtypes.DbResponse) (funcError error) {
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
			WithPayload: qdrant.NewWithPayloadEnable(true),
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
			dbresp, err := QdrantPayloadToType[sharedtypes.DbData](point.Payload)
			if err != nil {
				logging.Log.Errorf(ctx, "error converting qdrant payload: %q", err)
				return err
			}
			id, err := uuid.Parse(point.Id.GetUuid())
			if err != nil {
				logging.Log.Errorf(ctx, "point ID is not parseable as a UUID: %v", err)
				return err
			}
			dbresp.Guid = id
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
//   - data: Data to retrieve the parent nodes for.
//
// Returns:
//   - error: Error if any issue occurs while retrieving the parents.
func RetrieveParentNodes(ctx *logging.ContextMap, client *qdrant.Client, collectionName string, data *[]sharedtypes.DbResponse) (funcError error) {
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
		filter := qdrant.Filter{}
		if dbresp.ParentId != nil {
			filter.Must = append(filter.Must, qdrant.NewHasID(qdrant.NewIDUUID(dbresp.ParentId.String())))
		}
		queries[i] = &qdrant.QueryPoints{
			CollectionName: collectionName,
			Filter:         &filter,
			Limit:          &limit,
			WithVectors:    qdrant.NewWithVectorsEnable(false),
			WithPayload:    qdrant.NewWithPayloadEnable(true),
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
			parent, err := QdrantPayloadToType[sharedtypes.DbData](batchRes.Result[0].Payload)
			if err != nil {
				logging.Log.Errorf(ctx, "error converting qdrant payload: %q", err)
				return err
			}
			id, err := uuid.Parse(batchRes.Result[0].Id.GetUuid())
			if err != nil {
				return fmt.Errorf("point ID is not parseable as a UUID: %v", err)
			}
			parent.Guid = id
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
//   - data: Data to retrieve the children for.
//
// Returns:
//   - error: Error if any issue occurs while retrieving the children.
func RetrieveChildNodes(ctx *logging.ContextMap, client *qdrant.Client, collectionName string, data *[]sharedtypes.DbResponse) (funcError error) {
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
			WithPayload: qdrant.NewWithPayloadEnable(true),
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
			child, err := QdrantPayloadToType[sharedtypes.DbData](point.Payload)
			if err != nil {
				logging.Log.Errorf(ctx, "error converting qdrant payload: %q", err)
				return err
			}
			id, err := uuid.Parse(point.Id.GetUuid())
			if err != nil {
				logging.Log.Errorf(ctx, "point ID is not parseable as a UUID: %v", err)
				return err
			}
			child.Guid = id
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
//   - data: Data to retrieve the siblings for.
//
// Returns:
//   - error: Error if any issue occurs while retrieving the siblings.
func RetrieveDirectSiblingNodes(ctx *logging.ContextMap, client *qdrant.Client, collectionName string, data *[]sharedtypes.DbResponse) (funcError error) {
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
			WithPayload: qdrant.NewWithPayloadEnable(true),
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
			sibling, err := QdrantPayloadToType[sharedtypes.DbData](point.Payload)
			if err != nil {
				logging.Log.Errorf(ctx, "error converting qdrant payload: %q", err)
				return err
			}
			id, err := uuid.Parse(point.Id.GetUuid())
			if err != nil {
				logging.Log.Errorf(ctx, "point ID is not parseable as a UUID: %v", err)
				return err
			}
			sibling.Guid = id
			siblingsDbData[j] = sibling
		}
		(*data)[i].Siblings = siblingsDbData
	}
	return nil
}

func qdrantValToAny(val *qdrant.Value) any {
	switch val.Kind.(type) {
	case *qdrant.Value_NullValue:
		return nil
	case *qdrant.Value_DoubleValue:
		return val.GetDoubleValue()
	case *qdrant.Value_IntegerValue:
		return val.GetIntegerValue()
	case *qdrant.Value_StringValue:
		return val.GetStringValue()
	case *qdrant.Value_BoolValue:
		return val.GetBoolValue()
	case *qdrant.Value_StructValue:
		structmap := val.GetStructValue().GetFields()
		valmap := make(map[string]any, len(structmap))
		for k, v := range structmap {
			valmap[k] = qdrantValToAny(v)
		}
		return valmap
	case *qdrant.Value_ListValue:
		list := val.GetListValue().GetValues()
		vallist := make([]any, len(list))
		for i, v := range list {
			vallist[i] = qdrantValToAny(v)
		}
		return vallist
	default:
		panic(fmt.Sprintf("unknown qdrant value kind %q", val.Kind))
	}
}

func QdrantPayloadToMap(payload map[string]*qdrant.Value) map[string]any {
	m := make(map[string]any, len(payload))
	for k, v := range payload {
		m[k] = qdrantValToAny(v)
	}
	return m
}

func QdrantPayloadToType[T any](payload map[string]*qdrant.Value) (T, error) {
	var final T

	qdrantMap := QdrantPayloadToMap(payload)
	jsonBytes, err := json.Marshal(qdrantMap)
	if err != nil {
		return final, fmt.Errorf("unable to marshal qdrant payload to bytes: %v", err)
	}

	err = json.Unmarshal(jsonBytes, &final)
	if err != nil {
		return final, fmt.Errorf("unable to unmarshal bytes to type %T: %v", final, err)
	}
	return final, nil
}

func ToQdrantPayload[T any](t T) (map[string]*qdrant.Value, error) {

	jsonBytes, err := json.Marshal(t)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal qdrant payload to bytes: %v", err)
	}

	var jsonMap map[string]any
	err = json.Unmarshal(jsonBytes, &jsonMap)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal bytes to map: %v", err)
	}

	qdrantPayloadTypeConversion(jsonMap)

	return qdrant.TryValueMap(jsonMap)
}

func qdrantPayloadTypeConversion(payloadMap map[string]any) {
	for k, v := range payloadMap {
		switch v := v.(type) {
		case []string:
			anys := make([]any, len(v))
			for i, v := range v {
				anys[i] = v
			}
			payloadMap[k] = anys
		default:
			continue
		}
	}
}
