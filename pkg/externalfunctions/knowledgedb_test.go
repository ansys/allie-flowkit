package externalfunctions

import (
	"context"
	"maps"
	"testing"

	"github.com/ansys/aali-sharedtypes/pkg/config"
	"github.com/ansys/aali-sharedtypes/pkg/logging"
	"github.com/ansys/aali-sharedtypes/pkg/sharedtypes"
	"github.com/ansys/allie-flowkit/pkg/privatefunctions/graphdb"
	"github.com/google/uuid"
	"github.com/qdrant/go-client/qdrant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendVectorsToKnowledgeDB(t *testing.T) {

	// setup containers
	ctx := context.Background()

	testcase := func(t *testing.T, distance string, collection string) {
		require := require.New(t)
		assert := assert.New(t)

		setup := setupFlowkitTestContainers(
			t,
			ctx,
			flowkitTestContainersConfig{
				qdrant:        true,
				allieEmbedder: false,
				allieLlm:      false,
				aaliGraphDb:   false,
			},
		)
		config.GlobalConfig = &setup.config
		logging.InitLogger(&setup.config)

		qdrantClient, err := qdrant.NewClient(&qdrant.Config{Host: setup.qdrant.host, Port: setup.qdrant.port})
		require.NoError(err)

		// first make sure collection isn't there
		collExists, err := qdrantClient.CollectionExists(ctx, collection)
		require.NoError(err)
		assert.False(collExists, "collection %q shouldn't exist before running", collection)

		// now create collection
		QdrantCreateCollection(collection, 4, distance)

		// now check collection is there
		collExists, err = qdrantClient.CollectionExists(ctx, collection)
		require.NoError(err)
		assert.True(collExists, "collection %q should exist but doesn't", collection)

		// now insert some data
		data := []any{
			map[string]any{
				"id":            uuid.NewString(),
				"vector":        []float32{0, -1, -2, -3},
				"document_name": "Doc 1",
				"keywords":      []any{"kw1", "kw2"},
				"level":         "leaf",
			},
			map[string]any{
				"id":            uuid.NewString(),
				"vector":        []float32{4, 5, 6, 7},
				"document_name": "Doc 2",
				"keywords":      []any{"kw2", "kw3", "kw4"},
				"level":         "leaf",
			},
			map[string]any{
				"id":            uuid.NewString(),
				"vector":        []float32{4, 5, 6, 8},
				"document_name": "Doc 3",
				"keywords":      []any{"kw5"},
				"level":         "leaf",
			},
		}
		QdrantInsertData(collection, data, "id", "vector")

		// create index
		QdrantCreateIndex(collection, "document_name", "keyword", true)
		QdrantCreateIndex(collection, "keywords", "keyword", true)
		QdrantCreateIndex(collection, "level", "keyword", true)

		// do a straight up search with an exact match
		resp := SendVectorsToKnowledgeDB([]float32{0, -1, -2, -3}, []string{}, false, collection, 1, 0)
		require.Len(resp, 1, "expected 1 result but got %d", len(resp))
		assert.Equal("Doc 1", resp[0].DocumentName)

		// do a keyword filtered search with approx match
		resp = SendVectorsToKnowledgeDB([]float32{4, 5, 6, 7}, []string{"kw5"}, true, collection, 100, 0)
		require.Len(resp, 1, "expected 1 result but got %d", len(resp))
		assert.Equal("Doc 3", resp[0].DocumentName)
	}

	t.Run("cosine", func(t *testing.T) { testcase(t, "cosine", "test-cosine") })
	t.Run("dot", func(t *testing.T) { testcase(t, "dot", "test-dot") })
}

func TestCreateAndListCollections(t *testing.T) {
	ctx := context.Background()

	require := require.New(t)
	assert := assert.New(t)

	setup := setupFlowkitTestContainers(
		t,
		ctx,
		flowkitTestContainersConfig{
			qdrant:        true,
			allieEmbedder: false,
			allieLlm:      false,
			aaliGraphDb:   false,
		},
	)
	config.GlobalConfig = &setup.config
	logging.InitLogger(&setup.config)

	colls := GetListCollections()
	require.Len(colls, 0, "should be 0 collections initially")

	collReqs := map[string]struct {
		size     uint64
		distance string
	}{
		"mycollection1": {10, "cosine"},
		"mycollection2": {100, "euclid"},
		"mycollection3": {2, "manhattan"},
		"mycollection4": {1524, "dot"},
	}
	for collName, params := range collReqs {
		CreateCollectionRequest(collName, params.size, params.distance)
	}

	colls = GetListCollections()
	assert.Len(colls, len(collReqs))

	for expName := range maps.Keys(collReqs) {
		assert.Contains(colls, expName)
	}
}

func TestGeneralQuery(t *testing.T) {
	// setup containers
	ctx := context.Background()

	require := require.New(t)
	assert := assert.New(t)

	setup := setupFlowkitTestContainers(
		t,
		ctx,
		flowkitTestContainersConfig{
			qdrant:        true,
			allieEmbedder: false,
			allieLlm:      false,
			aaliGraphDb:   false,
		},
	)
	config.GlobalConfig = &setup.config
	logging.InitLogger(&setup.config)

	qdrantClient, err := qdrant.NewClient(&qdrant.Config{Host: setup.qdrant.host, Port: setup.qdrant.port})
	require.NoError(err)
	const COLLECTIONNAME = "test-gen-query"

	// first make sure collection isn't there
	collExists, err := qdrantClient.CollectionExists(ctx, COLLECTIONNAME)
	require.NoError(err)
	assert.False(collExists, "collection %q shouldn't exist before running", COLLECTIONNAME)

	// now create collection
	QdrantCreateCollection(COLLECTIONNAME, 4, "cosine")

	// now check collection is there
	collExists, err = qdrantClient.CollectionExists(ctx, COLLECTIONNAME)
	require.NoError(err)
	assert.True(collExists, "collection %q should exist but doesn't", COLLECTIONNAME)

	// now insert some data
	data := []any{
		map[string]any{
			"id":            uuid.NewString(),
			"vector":        []float32{0, -1, -2, -3},
			"document_name": "Doc",
			"keywords":      []any{"kw1", "kw2"},
			"level":         "leaf",
			"tags":          []any{},
		},
		map[string]any{
			"id":            uuid.NewString(),
			"vector":        []float32{4, 5, 6, 7},
			"document_name": "Document 2",
			"keywords":      []any{"kw2", "kw3", "kw4"},
			"level":         "leaf",
			"tags":          []any{"tag1", "tag2"},
		},
		map[string]any{
			"id":            uuid.NewString(),
			"vector":        []float32{4, 5, 6, 8},
			"document_name": "Doc",
			"keywords":      []any{"kw5"},
			"level":         "leaf",
			"tags":          []any{"tag1"},
		},
		map[string]any{
			"id":            uuid.NewString(),
			"vector":        []float32{4, 5, 6, 8},
			"document_name": "Main",
			"keywords":      []any{"kw5"},
			"level":         "root",
			"tags":          []any{"tag1", "tag2"},
		},
		map[string]any{
			"id":            uuid.NewString(),
			"vector":        []float32{4, 5, 6, 8},
			"document_name": "title",
			"keywords":      []any{"kw6"},
			"level":         "middle",
			"tags":          []any{"tag2", "tag1"},
		},
	}
	QdrantInsertData(COLLECTIONNAME, data, "id", "vector")

	// create index
	QdrantCreateIndex(COLLECTIONNAME, "document_name", "keyword", true)
	QdrantCreateIndex(COLLECTIONNAME, "keywords", "keyword", true)
	QdrantCreateIndex(COLLECTIONNAME, "level", "keyword", true)

	// do search
	filters := sharedtypes.DbFilters{
		DocumentNameFilter: []string{"Doc", "Main", "title"},
		LevelFilter:        []string{"leaf", "middle"},
		KeywordsFilter: sharedtypes.DbArrayFilter{
			NeedAll:    false,
			FilterData: []string{"kw6", "kw5"},
		},
		TagsFilter: sharedtypes.DbArrayFilter{
			NeedAll:    true,
			FilterData: []string{"tag1", "tag2"},
		},
	}

	resp := GeneralQuery(COLLECTIONNAME, 100, []string{"document_name", "level", "keywords", "tags"}, filters)
	require.Len(resp, 1, "expected 1 result but got %d", len(resp))
	assert.Equal("title", resp[0].DocumentName)
	assert.Equal("middle", resp[0].Level)
	assert.Equal([]string{"kw6"}, resp[0].Keywords)
	assert.Equal([]string{"tag2", "tag1"}, resp[0].Tags)
}

func TestSimilaritySearch(t *testing.T) {
	// setup containers
	ctx := context.Background()

	require := require.New(t)
	assert := assert.New(t)

	setup := setupFlowkitTestContainers(
		t,
		ctx,
		flowkitTestContainersConfig{
			qdrant:        true,
			allieEmbedder: false,
			allieLlm:      false,
			aaliGraphDb:   false,
		},
	)
	config.GlobalConfig = &setup.config
	logging.InitLogger(&setup.config)

	qdrantClient, err := qdrant.NewClient(&qdrant.Config{Host: setup.qdrant.host, Port: setup.qdrant.port})
	require.NoError(err)
	const COLLECTIONNAME = "test-gen-quiery"

	// first make sure collection isn't there
	collExists, err := qdrantClient.CollectionExists(ctx, COLLECTIONNAME)
	require.NoError(err)
	assert.False(collExists, "collection %q shouldn't exist before running", COLLECTIONNAME)

	// now create collection
	QdrantCreateCollection(COLLECTIONNAME, 4, "cosine")

	// now check collection is there
	collExists, err = qdrantClient.CollectionExists(ctx, COLLECTIONNAME)
	require.NoError(err)
	assert.True(collExists, "collection %q should exist but doesn't", COLLECTIONNAME)

	uuids := make([]string, 6)
	for i := range uuids {
		uuids[i] = uuid.NewString()
	}

	docId := uuid.NewString()

	// now insert some data
	data := []any{
		map[string]any{
			"id":            uuids[0],
			"vector":        []float32{0, -1, -2, -3},
			"document_name": "Doc 1",
			"keywords":      []any{"kw1", "kw2"},
			"level":         "leaf",
			"tags":          []any{},
			"document_id":   docId,
			"parent_id":     uuids[3],
		},
		map[string]any{
			"id":            uuids[1],
			"vector":        []float32{4, 5, 6, 7},
			"document_name": "Document 2",
			"keywords":      []any{"kw2", "kw3", "kw4"},
			"level":         "leaf",
			"tags":          []any{"tag1", "tag2"},
			"document_id":   docId,
			"parent_id":     uuids[3],
		},
		map[string]any{
			"id":            uuids[2],
			"vector":        []float32{4, 5, 6, 8},
			"document_name": "Doc 3",
			"keywords":      []any{"kw5"},
			"level":         "leaf",
			"tags":          []any{"tag1"},
			"document_id":   uuid.NewString(),
		},
		map[string]any{
			"id":            uuids[3],
			"vector":        []float32{4, 5, 6, 8},
			"document_name": "Main",
			"keywords":      []any{"kw5"},
			"level":         "root",
			"tags":          []any{"tag1", "tag2"},
			"document_id":   docId,
		},
		map[string]any{
			"id":              uuids[4],
			"vector":          []float32{4, 5, 6, 8},
			"document_name":   "title",
			"keywords":        []any{"kw6"},
			"level":           "middle",
			"tags":            []any{"tag2", "tag1"},
			"document_id":     docId,
			"next_sibling_id": uuids[5],
			"parent_id":       uuids[3],
			"child_ids":       []any{uuids[0]},
		},
		map[string]any{
			"id":                  uuids[5],
			"vector":              []float32{0, 0, 0, 0},
			"document_name":       "title 2",
			"level":               "middle",
			"document_id":         docId,
			"previous_sibling_id": uuids[4],
			"parent_id":           uuids[3],
			"child_ids":           []any{uuids[1]},
		},
	}
	QdrantInsertData(COLLECTIONNAME, data, "id", "vector")

	// create index
	QdrantCreateIndex(COLLECTIONNAME, "document_name", "keyword", true)
	QdrantCreateIndex(COLLECTIONNAME, "keywords", "keyword", true)
	QdrantCreateIndex(COLLECTIONNAME, "level", "keyword", true)

	// do search
	resp := SimilaritySearch(
		COLLECTIONNAME,
		[]float32{4, 5, 6, 7},
		1,
		sharedtypes.DbFilters{
			DocumentNameFilter: []string{"title"},
			LevelFilter:        []string{"middle"},
		},
		0,
		true,
		true,
		true,
		true,
	)
	require.Len(resp, 1, "expected 1 result but got %d", len(resp))
	primaryDoc := resp[0]
	assert.Equal("title", primaryDoc.DocumentName)
	assert.Equal(docId, primaryDoc.DocumentId)
	assert.Equal(uuids[4], primaryDoc.Guid.String())

	require.Len(primaryDoc.LeafNodes, 2)
	leafIds := []string{primaryDoc.LeafNodes[0].Guid.String(), primaryDoc.LeafNodes[1].Guid.String()}
	assert.Contains(leafIds, uuids[0])
	assert.Contains(leafIds, uuids[1])

	require.Len(primaryDoc.Siblings, 1)
	require.Equal(uuids[5], primaryDoc.Siblings[0].Guid.String())

	require.Len(primaryDoc.Children, 1)
	require.Equal(uuids[0], primaryDoc.Children[0].Guid.String())

	require.Equal(uuids[3], primaryDoc.Parent.Guid.String())
}

func TestAddDataRequest(t *testing.T) {
	// setup containers
	ctx := context.Background()

	require := require.New(t)
	assert := assert.New(t)

	setup := setupFlowkitTestContainers(
		t,
		ctx,
		flowkitTestContainersConfig{
			qdrant:        true,
			allieEmbedder: false,
			allieLlm:      false,
			aaliGraphDb:   false,
		},
	)
	config.GlobalConfig = &setup.config
	logging.InitLogger(&setup.config)

	qdrantClient, err := qdrant.NewClient(&qdrant.Config{Host: setup.qdrant.host, Port: setup.qdrant.port})
	require.NoError(err)
	const COLLECTIONNAME = "test-add-data"

	// first make sure collection isn't there
	collExists, err := qdrantClient.CollectionExists(ctx, COLLECTIONNAME)
	require.NoError(err)
	assert.False(collExists, "collection %q shouldn't exist before running", COLLECTIONNAME)

	// now create collection
	QdrantCreateCollection(COLLECTIONNAME, 4, "cosine")

	// now check collection is there
	collExists, err = qdrantClient.CollectionExists(ctx, COLLECTIONNAME)
	require.NoError(err)
	assert.True(collExists, "collection %q should exist but doesn't", COLLECTIONNAME)

	docId := uuid.NewString()

	// now insert some data
	data := []sharedtypes.DbData{
		{
			Guid:         uuid.New(),
			DocumentName: "Doc 1",
			Keywords:     []string{"kw1", "kw2"},
			Level:        "leaf",
			Tags:         []string{},
			DocumentId:   docId,
			Embedding:    []float32{0, 1, 2, 3},
		},
		{
			Guid:         uuid.New(),
			DocumentName: "Document 2",
			Keywords:     []string{"kw2", "kw3", "kw4"},
			Level:        "leaf",
			Tags:         []string{"tag1", "tag2"},
			DocumentId:   docId,
			Embedding:    []float32{0, 1, 2, 3},
		},
		{
			Guid:         uuid.New(),
			DocumentName: "Doc 3",
			Keywords:     []string{"kw5"},
			Level:        "leaf",
			Tags:         []string{"tag1"},
			DocumentId:   docId,
			Embedding:    []float32{0, 1, 2, 3},
		},
		{
			Guid:         uuid.New(),
			DocumentName: "Main",
			Keywords:     []string{"kw5"},
			Level:        "root",
			Tags:         []string{"tag1", "tag2"},
			DocumentId:   docId,
			Embedding:    []float32{0, 1, 2, 3},
		},
		{
			Guid:         uuid.New(),
			DocumentName: "title",
			Keywords:     []string{"kw6"},
			Level:        "middle",
			Tags:         []string{"tag2", "tag1"},
			DocumentId:   docId,
			Embedding:    []float32{0, 1, 2, 3},
		},
		{
			Guid:         uuid.New(),
			DocumentName: "title 2",
			Level:        "middle",
			DocumentId:   docId,
			Embedding:    []float32{0, 1, 2, 3},
		},
	}
	AddDataRequest(COLLECTIONNAME, data)

	// check theres some points in there
	points, err := qdrantClient.Query(ctx, &qdrant.QueryPoints{
		CollectionName: COLLECTIONNAME,
	})
	require.NoError(err)
	assert.Len(points, len(data))
}

func TestRetrieveDependencies(t *testing.T) {
	// setup containers
	ctx := context.Background()

	require := require.New(t)
	assert := assert.New(t)

	setup := setupFlowkitTestContainers(
		t,
		ctx,
		flowkitTestContainersConfig{
			qdrant:        false,
			allieEmbedder: false,
			allieLlm:      false,
			aaliGraphDb:   true,
		},
	)
	config.GlobalConfig = &setup.config
	logging.InitLogger(&setup.config)

	require.NoError(graphdb.Initialize(setup.config.GRAPHDB_ADDRESS))
	driver := graphdb.GraphDbDriver

	// first add data
	data := []sharedtypes.CodeGenerationUserGuideSection{
		{Name: "1", NextSibling: "2", Level: 0},
		{Name: "1a", IsFirstChild: true, NextSibling: "1b", NextParent: "2", Parent: "1", Level: 1},
		{Name: "1b", NextParent: "2", Parent: "1", Level: 1},
		{Name: "2", NextSibling: "3", Level: 0},
		{Name: "3", NextSibling: "4", Level: 0},
		{Name: "3a", IsFirstChild: true, NextSibling: "3b", NextParent: "4", Parent: "3", Level: 1},
		{Name: "3b", NextSibling: "3c", NextParent: "4", Parent: "3", Level: 1},
		{Name: "3b1", IsFirstChild: true, NextSibling: "3b2", NextParent: "3c", Parent: "3b", Level: 2},
		{Name: "3b2", NextParent: "3c", Parent: "3b", Level: 2},
		{Name: "3c", NextSibling: "3d", NextParent: "4", Parent: "3", Level: 1},
		{Name: "3d", NextParent: "4", Parent: "3", Level: 1},
		{Name: "3d1", IsFirstChild: true, NextParent: "4", Parent: "3d", Level: 1},
		{Name: "4", Level: 0},
		{Name: "4a", IsFirstChild: true, Parent: "4", Level: 1},
	}
	require.NoError(driver.AddUserGuideSectionNodes(data))
	require.NoError(driver.CreateUserGuideSectionRelationships(data))

	// now test deps
	// relationshipName := []string{"NextSibling", "NextParent", "HasFirstChild", "HasChild"}
	// relationshipDirection := []string{"in", "out", "both"}
	depIds := RetrieveDependencies("NextParent", "out", "3b", sharedtypes.DbArrayFilter{}, 2)
	assert.EqualValues([]string{"4"}, depIds)

	depIds = RetrieveDependencies("NextSibling", "in", "3b", sharedtypes.DbArrayFilter{}, 2)
	assert.EqualValues([]string{"3a"}, depIds)

	depIds = RetrieveDependencies("HasChild", "both", "3b", sharedtypes.DbArrayFilter{}, 2)
	expected := []string{"3", "3b1", "3b2"}
	assert.Len(depIds, len(expected))
	for _, exp := range expected {
		assert.Contains(depIds, exp)
	}
}

func TestGeneralGraphDbQuery(t *testing.T) {
	// setup containers
	ctx := context.Background()

	require := require.New(t)
	assert := assert.New(t)

	setup := setupFlowkitTestContainers(
		t,
		ctx,
		flowkitTestContainersConfig{
			qdrant:        false,
			allieEmbedder: false,
			allieLlm:      false,
			aaliGraphDb:   true,
		},
	)
	config.GlobalConfig = &setup.config
	logging.InitLogger(&setup.config)

	require.NoError(graphdb.Initialize(setup.config.GRAPHDB_ADDRESS))
	driver := graphdb.GraphDbDriver

	// first add data
	data := []sharedtypes.CodeGenerationUserGuideSection{
		{Name: "1", NextSibling: "2", Level: 0},
		{Name: "1a", IsFirstChild: true, NextSibling: "1b", NextParent: "2", Parent: "1", Level: 1},
		{Name: "1b", NextParent: "2", Parent: "1", Level: 1},
		{Name: "2", NextSibling: "3", Level: 0},
		{Name: "3", NextSibling: "4", Level: 0},
		{Name: "3a", IsFirstChild: true, NextSibling: "3b", NextParent: "4", Parent: "3", Level: 1},
		{Name: "3b", NextSibling: "3c", NextParent: "4", Parent: "3", Level: 1},
		{Name: "3b1", IsFirstChild: true, NextSibling: "3b2", NextParent: "3c", Parent: "3b", Level: 2},
		{Name: "3b2", NextParent: "3c", Parent: "3b", Level: 2},
		{Name: "3c", NextSibling: "3d", NextParent: "4", Parent: "3", Level: 1},
		{Name: "3d", NextParent: "4", Parent: "3", Level: 1},
		{Name: "3d1", IsFirstChild: true, NextParent: "4", Parent: "3d", Level: 1},
		{Name: "4", Level: 0},
		{Name: "4a", IsFirstChild: true, Parent: "4", Level: 1},
	}
	require.NoError(driver.AddUserGuideSectionNodes(data))
	require.NoError(driver.CreateUserGuideSectionRelationships(data))

	// now make a query
	res := GeneralGraphDbQuery("MATCH (n {parent:'1'}) RETURN n.name AS name")
	expected := []string{"1a", "1b"}
	require.Len(res, len(expected))
	for _, r := range res {
		name := r["name"].(string)
		assert.Contains(expected, name)
	}
}
