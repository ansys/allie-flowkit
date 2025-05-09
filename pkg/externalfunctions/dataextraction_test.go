package externalfunctions

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ansys/aali-graphdb-goclient/aali_graphdb"
	"github.com/ansys/aali-sharedtypes/pkg/config"
	"github.com/ansys/aali-sharedtypes/pkg/logging"
	"github.com/ansys/aali-sharedtypes/pkg/sharedtypes"
	"github.com/google/uuid"
	"github.com/qdrant/go-client/qdrant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
	"gopkg.in/yaml.v2"
)

func TestGetDocumentType(t *testing.T) {
	tests := []struct {
		fileName string
		expected string
	}{
		{"test.txt", "txt"},
		{"test.docx", "docx"},
		{"test.pdf", "pdf"},
		{"test.jpg", "jpg"},
		{"test.jpeg", "jpeg"},
		{"test.png", "png"},
		{"test", ""},
	}

	for _, test := range tests {
		actual := GetDocumentType(test.fileName)
		if actual != test.expected {
			t.Errorf("GetFileExtension(%s): expected %s, actual %s", test.fileName, test.expected, actual)
		}
	}
}

func TestGetLocalFileContent(t *testing.T) {
	// Create a temporary file for testing.
	tempFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Write some content to the temporary file.
	content := "Hello, World!"
	_, err = tempFile.WriteString(content)
	if err != nil {
		t.Fatalf("failed to write to test file: %v", err)
	}
	tempFile.Close()

	// Calculate the expected checksum.
	hash := sha256.New()
	_, err = hash.Write([]byte(content))
	if err != nil {
		t.Fatalf("failed to calculate expected checksum: %v", err)
	}
	expectedChecksum := hex.EncodeToString(hash.Sum(nil))

	// Call the function with the test file.
	actualChecksum, actualContent := GetLocalFileContent(tempFile.Name())

	// Check if the actual checksum matches the expected checksum.
	if actualChecksum != expectedChecksum {
		t.Errorf("expected checksum %v, got %v", expectedChecksum, actualChecksum)
	}

	if !bytes.Equal(actualContent, []byte(content)) {
		t.Errorf("expected content %v, got %v", content, actualContent)
	}
}

func TestAppendStringSlices(t *testing.T) {
	tests := []struct {
		slice1   []string
		slice2   []string
		slice3   []string
		slice4   []string
		slice5   []string
		expected []string
	}{
		{[]string{"a", "b", "c"}, []string{"d", "e", "f"}, []string{}, []string{}, []string{}, []string{"a", "b", "c", "d", "e", "f"}},
		{[]string{"a", "b", "c"}, []string{}, []string{}, []string{}, []string{}, []string{"a", "b", "c"}},
		{[]string{}, []string{"d", "e", "f"}, []string{}, []string{}, []string{}, []string{"d", "e", "f"}},
		{[]string{}, []string{}, []string{}, []string{}, []string{}, []string{}},
	}

	for _, test := range tests {
		actual := AppendStringSlices(test.slice1, test.slice2, test.slice3, test.slice4, test.slice5)
		if len(actual) != len(test.expected) {
			t.Errorf("expected length %d, got %d", len(test.expected), len(actual))
		}
		for i := range actual {
			if actual[i] != test.expected[i] {
				t.Errorf("expected %v, got %v", test.expected, actual)
			}
		}
	}
}

type flowkitTestContainersConfig struct {
	qdrant        bool
	allieEmbedder bool
	allieLlm      bool
	aaliGraphDb   bool
}

type hostPort struct {
	host string
	port int
}

type flowkitTestContainersResult struct {
	config        config.Config
	qdrant        *hostPort
	allieEmbedder *hostPort
	allieLlm      *hostPort
	aaliGraphdDb  *hostPort
}

// StdoutLogConsumer is a LogConsumer that prints the log to stdout
type StdoutLogConsumer struct{}

// Accept prints the log to stdout
func (lc *StdoutLogConsumer) Accept(l testcontainers.Log) {
	fmt.Print(string(l.Content))
}

func setupFlowkitTestContainers(t *testing.T, ctx context.Context, testContainerConfig flowkitTestContainersConfig) flowkitTestContainersResult {

	result := flowkitTestContainersResult{config: config.Config{}}

	allieNetwork, err := network.New(ctx)
	require.NoError(t, err)
	testcontainers.CleanupNetwork(t, allieNetwork)

	logConsumer := testcontainers.LogConsumerConfig{
		Opts:      []testcontainers.LogProductionOption{testcontainers.WithLogProductionTimeout(10 * time.Second)},
		Consumers: []testcontainers.LogConsumer{&StdoutLogConsumer{}},
	}

	if testContainerConfig.qdrant {
		// setup qdrant container
		qdrantReq := testcontainers.ContainerRequest{
			Image:        "qdrant/qdrant:v1.13.6",
			ExposedPorts: []string{"6334/tcp"},
			WaitingFor: wait.ForAll(
				wait.ForLog("Qdrant gRPC listening on 6334"),
				wait.ForListeningPort("6334/tcp"),
			),
			Networks:       []string{allieNetwork.Name},
			NetworkAliases: map[string][]string{allieNetwork.Name: {"qdrant"}},
			LogConsumerCfg: &logConsumer,
		}
		qdrantCont, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: qdrantReq, Started: true,
		})
		defer testcontainers.CleanupContainer(t, qdrantCont)
		require.NoError(t, err)
		qdrantHost, err := qdrantCont.Host(ctx)
		require.NoError(t, err)
		qdrantPort, err := qdrantCont.MappedPort(ctx, "6334/tcp")
		require.NoError(t, err)

		result.qdrant = &hostPort{qdrantHost, qdrantPort.Int()}
		result.config.QDRANT_HOST = qdrantHost
		result.config.QDRANT_PORT = qdrantPort.Int()
	}

	if testContainerConfig.allieEmbedder {
		// setup allie-codegen-embedder
		allieEmbedderReq := testcontainers.ContainerRequest{
			Image:        "ghcr.io/ansys-internal/allie-embedding:latest",
			ExposedPorts: []string{"8000/tcp"},
			WaitingFor: wait.ForAll(
				wait.ForLog("Uvicorn running on http://0.0.0.0:8000"),
				wait.ForListeningPort("8000/tcp"),
			),
			Networks:       []string{allieNetwork.Name},
			NetworkAliases: map[string][]string{allieNetwork.Name: {"allie-codegen-embedder"}},
			LogConsumerCfg: &logConsumer,
		}
		allieEmbedderCont, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: allieEmbedderReq, Started: true,
		})
		defer testcontainers.CleanupContainer(t, allieEmbedderCont)
		require.NoError(t, err)
		allieEmbedderHost, err := allieEmbedderCont.Host(ctx)
		require.NoError(t, err)
		allieEmbedderPort, err := allieEmbedderCont.MappedPort(ctx, "8000/tcp")
		require.NoError(t, err)

		result.allieEmbedder = &hostPort{allieEmbedderHost, allieEmbedderPort.Int()}
	}

	if testContainerConfig.allieLlm {
		// setup allie-llm
		// setup config
		allieLlmConfigFile, err := os.CreateTemp("", "test-allie-config")
		require.NoError(t, err)
		defer require.NoError(t, os.Remove(allieLlmConfigFile.Name()))
		allieLlmConfig, err := yaml.Marshal(config.Config{WEBSERVER_PORT: "9003", LOG_LEVEL: "debug"})
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(allieLlmConfigFile.Name(), allieLlmConfig, 0644))

		// setup models.yaml
		allieLlmModelsFile, err := os.CreateTemp("", "test-allie-models")
		require.NoError(t, err)
		defer require.NoError(t, os.Remove(allieLlmModelsFile.Name()))

		// embedder will only actually work if `testContainerConfig.allieEmbedder=true`, but set it up
		// here since you need something
		modelsYml := []string{
			"EMBEDDING_MODELS:",
			"  - MODEL_TYPE: bge-m3",
			"    MODEL_NAME: BAAI/bge-m3",
			"    URL: http://allie-codegen-embedder:8000/",
			"    NUMBER_OF_WORKERS: 2",
			"",
			"CHAT_MODELS:",
			"  - ID: gpt-4-32k-france-central",
			"    MODEL_TYPE: azure-gpt",
			"    MODEL_NAME: gpt-4-32k-france-central",
			"    URL: https://csebu-chatgpt-francecentral.openai.azure.com/",
			"    API_KEY: ", // don't actually need it for any of these tests
			"    NUMBER_OF_WORKERS: 2",
		}
		require.NoError(t, os.WriteFile(allieLlmModelsFile.Name(), []byte(strings.Join(modelsYml, "\n")), 0644))

		// now start the container with the 2 files mounted
		allieLlmReq := testcontainers.ContainerRequest{
			FromDockerfile: testcontainers.FromDockerfile{
				Context:    "../../../allie-llm",
				Dockerfile: "deployments/docker/Dockerfile",
			},
			ExposedPorts: []string{"9003/tcp"},
			WaitingFor: wait.ForAll(
				wait.ForLog("Allie LLM started successfully; Webserver is listening on port 9003"),
				wait.ForListeningPort("9003/tcp"),
			),
			Files: []testcontainers.ContainerFile{
				{
					HostFilePath:      allieLlmConfigFile.Name(),
					ContainerFilePath: "/app/config.yaml",
					FileMode:          0644,
				},
				{
					HostFilePath:      allieLlmModelsFile.Name(),
					ContainerFilePath: "/app/models.yaml",
					FileMode:          0644,
				},
			},
			Networks:       []string{allieNetwork.Name},
			NetworkAliases: map[string][]string{allieNetwork.Name: {"allie-llm"}},
			LogConsumerCfg: &logConsumer,
		}
		allieLlmCont, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: allieLlmReq, Started: true,
		})
		defer testcontainers.CleanupContainer(t, allieLlmCont)
		require.NoError(t, err)
		allieLlmHost, err := allieLlmCont.Host(ctx)
		require.NoError(t, err)
		allieLlmPort, err := allieLlmCont.MappedPort(ctx, "9003/tcp")
		require.NoError(t, err)

		result.allieLlm = &hostPort{allieLlmHost, allieLlmPort.Int()}
		result.config.LLM_HANDLER_ENDPOINT = fmt.Sprintf("ws://%s:%d", allieLlmHost, allieLlmPort.Int())
	}

	if testContainerConfig.aaliGraphDb {
		// setup aali-graphdb
		aaliGraphDbReq := testcontainers.ContainerRequest{
			Image:        "ghcr.io/ansys/aali-graphdb:v0.1.0",
			ExposedPorts: []string{"8080/tcp"},
			WaitingFor: wait.ForAll(
				wait.ForLog("listening on 0.0.0.0:8080"),
				wait.ForListeningPort("8080/tcp"),
			),
			Networks:       []string{allieNetwork.Name},
			NetworkAliases: map[string][]string{allieNetwork.Name: {"aali-graphdb"}},
			LogConsumerCfg: &logConsumer,
		}
		aaliGraphDbCont, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: aaliGraphDbReq, Started: true,
		})
		defer testcontainers.CleanupContainer(t, aaliGraphDbCont)
		require.NoError(t, err)
		aaliGraphDbHost, err := aaliGraphDbCont.Host(ctx)
		require.NoError(t, err)
		aaliGraphDbPort, err := aaliGraphDbCont.MappedPort(ctx, "8080/tcp")
		require.NoError(t, err)

		result.aaliGraphdDb = &hostPort{aaliGraphDbHost, aaliGraphDbPort.Int()}
		result.config.GRAPHDB_ADDRESS = fmt.Sprintf("http://%v:%d", aaliGraphDbHost, aaliGraphDbPort.Int())
	}

	return result
}

func TestStoreElementsInVectorDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long container test in short mode")
	}

	ctx := context.Background()

	// start containers & set config
	setup := setupFlowkitTestContainers(t, ctx, flowkitTestContainersConfig{
		qdrant:        true,
		allieEmbedder: true,
		allieLlm:      true,
	})
	config.GlobalConfig = &setup.config
	logging.InitLogger(&setup.config)

	// do some initial checks
	const COLLECTIONNAME = "testing"
	qdrantClient, err := qdrant.NewClient(&qdrant.Config{Host: setup.qdrant.host, Port: setup.qdrant.port})
	require.NoError(t, err)

	collExists, err := qdrantClient.CollectionExists(ctx, COLLECTIONNAME)
	require.NoError(t, err)
	assert.False(t, collExists, "collection %q shouldn't exist before running", COLLECTIONNAME)

	// insert the data in
	elements := []sharedtypes.CodeGenerationElement{
		{
			Guid:              uuid.New(),
			Type:              "Method",
			NamePseudocode:    "",
			NameFormatted:     "",
			Description:       "",
			Name:              "",
			Dependencies:      []string{"GrandParent", "Parent"},
			Summary:           "",
			ReturnType:        "",
			ReturnElementList: []string{},
			ReturnDescription: "",
			Remarks:           "",
			Parameters:        []sharedtypes.XMLMemberParam{},
			Example: sharedtypes.XMLMemberExample{
				Description: "",
				Code: sharedtypes.XMLMemberExampleCode{
					Type: "",
					Text: "",
				},
			},
			EnumValues: nil,
		},
	}
	expectedPayloads := []map[string]*qdrant.Value{
		qdrant.NewValueMap(map[string]any{
			"name":            elements[0].Name,
			"name_pseudocode": elements[0].NamePseudocode,
			"name_formatted":  elements[0].NameFormatted,
			"type":            string(elements[0].Type),
			"parent_class":    "GrandParent.Parent",
		}),
	}
	assert.Len(t, expectedPayloads, len(elements))

	StoreElementsInVectorDatabase(elements, COLLECTIONNAME, 2, "cosine")

	// query qdrant to make sure things are as they should be
	collExists, err = qdrantClient.CollectionExists(ctx, COLLECTIONNAME)
	require.NoError(t, err)
	assert.True(t, collExists, "collection %q wasn't created", COLLECTIONNAME)

	points, err := qdrantClient.Query(ctx, &qdrant.QueryPoints{
		CollectionName: COLLECTIONNAME,
		WithVectors:    qdrant.NewWithVectorsEnable(false),
		WithPayload:    qdrant.NewWithPayloadEnable(true),
	})
	require.NoError(t, err)
	assert.Len(t, points, len(expectedPayloads), "expected %d qdrant points but got %d", len(expectedPayloads), len(points))

	actualPayloads := make([]map[string]*qdrant.Value, len(points))
	for i, point := range points {
		actualPayloads[i] = point.Payload
	}
	assert.Equal(t, expectedPayloads, actualPayloads, "payloads were not as expected")

}

func anyArray[T any](a []T) []any {
	res := make([]any, len(a))
	for i, e := range a {
		res[i] = e
	}
	return res
}

func anyMapVal[K comparable, V any](m map[K]V) map[K]any {
	res := make(map[K]any, len(m))
	for k, v := range m {
		res[k] = v
	}
	return res
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

func qdrantPayloadToMap(payload map[string]*qdrant.Value) map[string]any {
	m := make(map[string]any, len(payload))
	for k, v := range payload {
		m[k] = qdrantValToAny(v)
	}
	return m
}

func TestStoreExamplesInVectorDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long container test in short mode")
	}

	ctx := context.Background()

	// start containers & set config
	setup := setupFlowkitTestContainers(t, ctx, flowkitTestContainersConfig{
		qdrant:        true,
		allieEmbedder: true,
		allieLlm:      true,
	})
	config.GlobalConfig = &setup.config
	logging.InitLogger(&setup.config)

	// do some initial checks
	const COLLECTIONNAME = "testing"
	qdrantClient, err := qdrant.NewClient(&qdrant.Config{Host: setup.qdrant.host, Port: setup.qdrant.port})
	require.NoError(t, err)

	collExists, err := qdrantClient.CollectionExists(ctx, COLLECTIONNAME)
	require.NoError(t, err)
	assert.False(t, collExists, "collection %q shouldn't exist before running", COLLECTIONNAME)

	// insert the data in
	examples := []sharedtypes.CodeGenerationExample{
		{
			Guid:                   uuid.New(),
			Name:                   "examples/my_example.py",
			Dependencies:           []string{"examples/my_other_example.py"},
			DependencyEquivalences: map[string]string{},
			Chunks:                 []string{"import random\n", "def main():\n    print('hi')\n\n", "if __name__ == '__main__':\n    main()"},
		},
		{
			Guid:                   uuid.New(),
			Name:                   "examples/my_other_example.py",
			Dependencies:           []string{},
			DependencyEquivalences: map[string]string{},
			Chunks:                 []string{"print('hi')"},
		},
	}

	StoreExamplesInVectorDatabase(examples, COLLECTIONNAME, 2, "cosine")

	// query qdrant to make sure things are as they should be
	collExists, err = qdrantClient.CollectionExists(ctx, COLLECTIONNAME)
	require.NoError(t, err)
	assert.True(t, collExists, "collection %q wasn't created", COLLECTIONNAME)

	points, err := qdrantClient.Query(ctx, &qdrant.QueryPoints{
		CollectionName: COLLECTIONNAME,
		WithVectors:    qdrant.NewWithVectorsEnable(true),
		WithPayload:    qdrant.NewWithPayloadEnable(true),
	})
	require.NoError(t, err)
	nChunks := 0
	for _, ex := range examples {
		nChunks += len(ex.Chunks)
	}
	assert.Len(t, points, nChunks, "expected %d qdrant points but got %d", nChunks, len(points))

	// get IDs we need to correctly populate previous/next IDs
	var chunk0Id string
	var chunk1Id string
	var chunk2Id string
	for _, point := range points {
		text, found := point.Payload["text"]
		if !found {
			t.Fatal("point had no 'text' payload")
		}
		var textStr string
		switch text.Kind.(type) {
		case *qdrant.Value_StringValue:
			textStr = text.GetStringValue()
		default:
			t.Fatal("text payload was not string type")
		}

		var pointUuid string
		switch point.GetId().GetPointIdOptions().(type) {
		case *qdrant.PointId_Uuid:
			pointUuid = point.GetId().GetUuid()
		default:
			t.Fatal("expected point ID to be UUID type")
		}

		switch textStr {
		case examples[0].Chunks[0]:
			chunk0Id = pointUuid
		case examples[0].Chunks[1]:
			chunk1Id = pointUuid
		case examples[0].Chunks[2]:
			chunk2Id = pointUuid
		default:
		}
	}

	expectedPayloads := []map[string]any{
		{
			"document_name":           examples[0].Name,
			"next_chunk":              chunk1Id,
			"dependencies":            anyArray(examples[0].Dependencies),
			"dependency_equivalences": anyMapVal(examples[0].DependencyEquivalences),
			"text":                    examples[0].Chunks[0],
		},
		{
			"document_name":           examples[0].Name,
			"previous_chunk":          chunk0Id,
			"next_chunk":              chunk2Id,
			"dependencies":            anyArray(examples[0].Dependencies),
			"dependency_equivalences": anyMapVal(examples[0].DependencyEquivalences),
			"text":                    examples[0].Chunks[1],
		},
		{
			"document_name":           examples[0].Name,
			"previous_chunk":          chunk1Id,
			"dependencies":            anyArray(examples[0].Dependencies),
			"dependency_equivalences": anyMapVal(examples[0].DependencyEquivalences),
			"text":                    examples[0].Chunks[2],
		},
		{
			"document_name":           examples[1].Name,
			"dependencies":            anyArray(examples[1].Dependencies),
			"dependency_equivalences": anyMapVal(examples[1].DependencyEquivalences),
			"text":                    examples[1].Chunks[0],
		},
	}
	for _, point := range points {
		vecOpts := point.GetVectors().VectorsOptions.(*qdrant.VectorsOutput_Vectors)
		assert.Len(t, vecOpts.Vectors.Vectors, 2)
		assert.Contains(t, expectedPayloads, qdrantPayloadToMap(point.Payload))
	}
}

func TestStoreUserGuideSectionsInVectorDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long container test in short mode")
	}

	ctx := context.Background()
	assert := assert.New(t)

	// start containers & set config
	setup := setupFlowkitTestContainers(t, ctx, flowkitTestContainersConfig{
		qdrant:        true,
		allieEmbedder: true,
		allieLlm:      true,
	})
	config.GlobalConfig = &setup.config
	logging.InitLogger(&setup.config)

	// do some initial checks
	const COLLECTIONNAME = "testing"
	qdrantClient, err := qdrant.NewClient(&qdrant.Config{Host: setup.qdrant.host, Port: setup.qdrant.port})
	require.NoError(t, err)

	collExists, err := qdrantClient.CollectionExists(ctx, COLLECTIONNAME)
	require.NoError(t, err)
	assert.False(collExists, "collection %q shouldn't exist before running", COLLECTIONNAME)

	// insert the data in
	sections := []sharedtypes.CodeGenerationUserGuideSection{
		{
			Name:            "Section Name",
			Title:           "Title",
			IsFirstChild:    false,
			NextSibling:     "",
			NextParent:      "",
			DocumentName:    "Doc Name",
			Parent:          "Parent Section",
			Content:         "Here is the content\n\nIt can be...\nmultiline",
			Level:           2,
			Link:            "",
			ReferencedLinks: []string{},
		},
	}

	StoreUserGuideSectionsInVectorDatabase(sections, COLLECTIONNAME, 2, 5, 1, "cosine")

	// query qdrant to make sure things are as they should be
	collExists, err = qdrantClient.CollectionExists(ctx, COLLECTIONNAME)
	require.NoError(t, err)
	assert.True(collExists, "collection %q wasn't created", COLLECTIONNAME)

	points, err := qdrantClient.Query(ctx, &qdrant.QueryPoints{
		CollectionName: COLLECTIONNAME,
		WithVectors:    qdrant.NewWithVectorsEnable(true),
		WithPayload:    qdrant.NewWithPayloadEnable(true),
	})
	require.NoError(t, err)

	expectedPayloads := []map[string]any{
		{
			"section_name":        sections[0].Name,
			"document_name":       sections[0].DocumentName,
			"title":               sections[0].Title,
			"parent_section_name": sections[0].Parent,
			"level":               int64(sections[0].Level),
			"text":                "Here is the content\n\n",
			"next_chunk":          "",
		},
		{
			"section_name":        sections[0].Name,
			"document_name":       sections[0].DocumentName,
			"title":               sections[0].Title,
			"parent_section_name": sections[0].Parent,
			"level":               int64(sections[0].Level),
			"text":                "\n\nIt can be...\n",
			"previous_chunk":      "",
			"next_chunk":          "",
		},
		{
			"section_name":        sections[0].Name,
			"document_name":       sections[0].DocumentName,
			"title":               sections[0].Title,
			"parent_section_name": sections[0].Parent,
			"level":               int64(sections[0].Level),
			"text":                "...\nmultiline",
			"previous_chunk":      "",
		},
	}
	assert.Len(points, len(expectedPayloads), "expected %d qdrant points but got %d", len(expectedPayloads), len(points))

	// get IDs we need to correctly populate previous/next IDs
	for _, point := range points {
		text, found := point.Payload["text"]
		if !found {
			t.Fatal("point had no 'text' payload")
		}
		var textStr string
		switch text.Kind.(type) {
		case *qdrant.Value_StringValue:
			textStr = text.GetStringValue()
		default:
			t.Fatal("text payload was not string type")
		}

		var pointUuid string
		switch point.GetId().GetPointIdOptions().(type) {
		case *qdrant.PointId_Uuid:
			pointUuid = point.GetId().GetUuid()
		default:
			t.Fatal("expected point ID to be UUID type")
		}

		switch textStr {
		case expectedPayloads[0]["text"]:
			expectedPayloads[1]["previous_chunk"] = pointUuid
		case expectedPayloads[1]["text"]:
			expectedPayloads[0]["next_chunk"] = pointUuid
			expectedPayloads[2]["previous_chunk"] = pointUuid
		case expectedPayloads[2]["text"]:
			expectedPayloads[1]["next_chunk"] = pointUuid
		default:
		}
	}

	for _, point := range points {
		vecOpts := point.GetVectors().VectorsOptions.(*qdrant.VectorsOutput_Vectors)
		assert.Len(vecOpts.Vectors.Vectors, 2)
		payloadMap := qdrantPayloadToMap(point.Payload)
		assert.Contains(expectedPayloads, payloadMap)
	}
}

func TestStoreElementsInGraphDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long container test in short mode")
	}

	ctx := context.Background()
	assert := assert.New(t)
	require := require.New(t)

	// start containers & set config
	setup := setupFlowkitTestContainers(t, ctx, flowkitTestContainersConfig{
		qdrant:        false,
		allieEmbedder: false,
		allieLlm:      false,
		aaliGraphDb:   true,
	})
	config.GlobalConfig = &setup.config
	logging.InitLogger(&setup.config)

	// do some initial checks
	const DBNAME = "aali" // this is hardcoded in the graphdb driver for now
	client, err := aali_graphdb.DefaultClient(setup.config.GRAPHDB_ADDRESS)
	require.NoError(err)

	dbs, err := client.GetDatabases()
	require.NoError(err)
	assert.Len(dbs, 0)

	// insert the data in
	element := sharedtypes.CodeGenerationElement{
		Guid:              uuid.New(),
		Type:              "Method",
		NamePseudocode:    "",
		NameFormatted:     "",
		Description:       "",
		Name:              "myFunction",
		Dependencies:      []string{"Parent"},
		Summary:           "",
		ReturnType:        "",
		ReturnElementList: []string{"Child"},
		ReturnDescription: "",
		Remarks:           "",
		Parameters: []sharedtypes.XMLMemberParam{
			{Name: "p1", Type: "Child"},
			{Name: "p2", Type: "Parent"},
		},
		Example: sharedtypes.XMLMemberExample{
			Description: "",
			Code: sharedtypes.XMLMemberExampleCode{
				Type: "",
				Text: "",
			},
		},
		EnumValues: nil,
	}

	StoreElementsInGraphDatabase([]sharedtypes.CodeGenerationElement{element})

	// query graphdb to make sure things are as they should be

	// check element nodes
	type dbElement struct {
		Name string
		Type string
		Guid uuid.UUID
	}
	expectedDbElements := []dbElement{
		{element.Name, string(element.Type), element.Guid},
		{"Parent", "", uuid.Nil},
		{"Child", "", uuid.Nil},
	}
	dbElements, err := aali_graphdb.CypherQueryReadGeneric[dbElement](client, DBNAME, "MATCH (n:Element) RETURN n.name AS name, n.guid as guid, n.type as type", nil)
	require.NoError(err)
	assert.Len(dbElements, len(expectedDbElements))
	for _, dbElem := range dbElements {
		assert.Contains(expectedDbElements, dbElem)
	}

	// check belongs to relationships
	type count struct {
		Count int
	}
	belongsTos, err := aali_graphdb.CypherQueryReadGeneric[count](client, DBNAME, "MATCH (a)-[e:BelongsTo]->(b) RETURN COUNT(e) AS count", nil)
	require.NoError(err)
	assert.Len(belongsTos, 1)
	assert.Equal(1, belongsTos[0].Count)

	// check returns relationships
	returns, err := aali_graphdb.CypherQueryReadGeneric[count](client, DBNAME, "MATCH (a)-[e:Returns]->(b) RETURN COUNT(e) AS count", nil)
	require.NoError(err)
	assert.Len(returns, 1)
	assert.Equal(1, returns[0].Count)

	// check uses parameters relationshps
	usesParameters, err := aali_graphdb.CypherQueryReadGeneric[count](client, DBNAME, "MATCH (a)-[e:UsesParameter]->(b) RETURN COUNT(e) AS count", nil)
	require.NoError(err)
	assert.Len(usesParameters, 1)
	assert.Equal(2, usesParameters[0].Count)

}

func TestStoreExamplesInGraphDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long container test in short mode")
	}

	ctx := context.Background()
	assert := assert.New(t)
	require := require.New(t)

	// start containers & set config
	setup := setupFlowkitTestContainers(t, ctx, flowkitTestContainersConfig{
		qdrant:        false,
		allieEmbedder: false,
		allieLlm:      false,
		aaliGraphDb:   true,
	})
	config.GlobalConfig = &setup.config
	logging.InitLogger(&setup.config)

	// do some initial checks
	const DBNAME = "aali" // this is hardcoded in the graphdb driver for now
	client, err := aali_graphdb.DefaultClient(setup.config.GRAPHDB_ADDRESS)
	require.NoError(err)

	dbs, err := client.GetDatabases()
	require.NoError(err)
	assert.Len(dbs, 0)

	// insert the data in
	examples := []sharedtypes.CodeGenerationExample{
		{
			Guid:                   uuid.New(),
			Name:                   "MyExampleName",
			Dependencies:           []string{"depA", "depB"},
			DependencyEquivalences: map[string]string{"depA": "A"},
			Chunks:                 []string{"chunk1", "chunk2"},
		},
	}

	StoreExamplesInGraphDatabase(examples)

	// query graphdb to make sure things are as they should be

	// check example nodes
	examplesDb, err := aali_graphdb.CypherQueryReadGeneric[struct {
		Name                   string            `json:"n.name"`
		Dependencies           []string          `json:"n.dependencies"`
		DependencyEquivalences map[string]string `json:"n.dependency_equivalences"`
		Guid                   uuid.UUID         `json:"n.guid"`
	}](client, DBNAME, "MATCH (n:Example) RETURN n.*", nil)
	require.NoError(err)
	assert.Len(examplesDb, 1)
	assert.Equal(examples[0].Name, examplesDb[0].Name)
	assert.Equal(examples[0].Guid, examplesDb[0].Guid)
	assert.Equal(examples[0].Dependencies, examplesDb[0].Dependencies)
	assert.Equal(examples[0].DependencyEquivalences, examplesDb[0].DependencyEquivalences)

	// check element nodes (created for the dependencies)
	elements, err := aali_graphdb.CypherQueryReadGeneric[struct {
		Name    string `json:"n.name"`
		Example string `json:"n.example"`
	}](client, DBNAME, "MATCH (n:Element) RETURN n.*", nil)
	require.NoError(err)
	assert.Len(elements, 2)

	// check uses relationships
	belongsTo, err := aali_graphdb.CypherQueryReadGeneric[struct{}](client, DBNAME, "MATCH (a)-[e:Uses]->(b) RETURN a, b;", nil)
	require.NoError(err)
	assert.Len(belongsTo, 2)
}

func TestStoreUserGuideSectionsInGraphDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long container test in short mode")
	}

	ctx := context.Background()
	assert := assert.New(t)
	require := require.New(t)

	// start containers & set config
	setup := setupFlowkitTestContainers(t, ctx, flowkitTestContainersConfig{
		qdrant:        false,
		allieEmbedder: false,
		allieLlm:      false,
		aaliGraphDb:   true,
	})
	config.GlobalConfig = &setup.config
	logging.InitLogger(&setup.config)

	// do some initial checks
	const DBNAME = "aali" // this is hardcoded in the graphdb driver for now
	client, err := aali_graphdb.DefaultClient(setup.config.GRAPHDB_ADDRESS)
	require.NoError(err)

	dbs, err := client.GetDatabases()
	require.NoError(err)
	assert.Len(dbs, 0)

	// insert the data in
	sections := []sharedtypes.CodeGenerationUserGuideSection{
		{
			Name:            "section1",
			Title:           "Section 1",
			DocumentName:    "My Guide",
			Parent:          "root",
			Level:           0,
			Link:            "",
			ReferencedLinks: []string{},
			NextSibling:     "section2",
			NextParent:      "",
			IsFirstChild:    false,
		},
		{
			Name:            "section1a",
			Title:           "Section 1a",
			DocumentName:    "My Guide",
			Parent:          "section1",
			Level:           1,
			Link:            "",
			ReferencedLinks: []string{},
			NextSibling:     "section1b",
			NextParent:      "section2",
			IsFirstChild:    true,
		},
		{
			Name:            "section1b",
			Title:           "Section 1b",
			DocumentName:    "My Guide",
			Parent:          "section1",
			Level:           1,
			Link:            "",
			ReferencedLinks: []string{},
			NextSibling:     "",
			NextParent:      "section2",
			IsFirstChild:    false,
		},
		{
			Name:            "section2",
			Title:           "Section 2",
			DocumentName:    "My Guide",
			Parent:          "root",
			Level:           0,
			Link:            "",
			ReferencedLinks: []string{"section1"},
			NextSibling:     "",
			NextParent:      "",
			IsFirstChild:    false,
		},
	}

	StoreUserGuideSectionsInGraphDatabase(sections)

	// query graphdb to make sure things are as they should be
	type dbUserGuideSection struct {
		Name            string   `json:"a.name"`
		Title           string   `json:"a.title"`
		DocumentName    string   `json:"a.document_name"`
		Parent          string   `json:"a.parent"`
		Level           int      `json:"a.level"`
		Link            string   `json:"a.link"`
		ReferencedLinks []string `json:"a.referenced_links"`
	}
	expectedSections := make([]dbUserGuideSection, len(sections)+1)
	for i, sec := range sections {
		expectedSections[i] = dbUserGuideSection{
			Name:            sec.Name,
			Title:           sec.Title,
			DocumentName:    sec.DocumentName,
			Parent:          sec.Parent,
			Level:           sec.Level,
			Link:            sec.Link,
			ReferencedLinks: sec.ReferencedLinks,
		}
	}
	expectedSections[len(sections)] = dbUserGuideSection{
		Name: "root",
	}

	// check nodes
	sectionsDb, err := aali_graphdb.CypherQueryReadGeneric[dbUserGuideSection](client, DBNAME, "MATCH (a:UserGuide) RETURN a.*", nil)
	require.NoError(err)
	assert.Len(sectionsDb, len(expectedSections))
	for _, dbsection := range sectionsDb {
		assert.Contains(expectedSections, dbsection)
	}

	// check references
	type count struct {
		Count int
	}
	refs, err := aali_graphdb.CypherQueryReadGeneric[count](client, DBNAME, "MATCH (:UserGuide)-[e:References]->(:UserGuide) RETURN COUNT(e) AS count", nil)
	require.NoError(err)
	assert.Len(refs, 1)
	assert.Equal(refs[0].Count, 1)

	// check next siblings
	nextSibs, err := aali_graphdb.CypherQueryReadGeneric[count](client, DBNAME, "MATCH (:UserGuide)-[e:NextSibling]->(:UserGuide) RETURN COUNT(e) AS count", nil)
	require.NoError(err)
	assert.Len(nextSibs, 1)
	assert.Equal(nextSibs[0].Count, 2)

	// check next parents
	nextPars, err := aali_graphdb.CypherQueryReadGeneric[count](client, DBNAME, "MATCH (:UserGuide)-[e:NextParent]->(:UserGuide) RETURN COUNT(e) AS count", nil)
	require.NoError(err)
	assert.Len(nextPars, 1)
	assert.Equal(nextPars[0].Count, 2)

	// check is 1st child
	hasFirstChilds, err := aali_graphdb.CypherQueryReadGeneric[count](client, DBNAME, "MATCH (:UserGuide)-[e:HasFirstChild]->(:UserGuide) RETURN COUNT(e) AS count", nil)
	require.NoError(err)
	assert.Len(hasFirstChilds, 1)
	assert.Equal(hasFirstChilds[0].Count, 1)

	// check has child
	hasChilds, err := aali_graphdb.CypherQueryReadGeneric[count](client, DBNAME, "MATCH (:UserGuide)-[e:HasChild]->(:UserGuide) RETURN COUNT(e) AS count", nil)
	require.NoError(err)
	assert.Len(hasChilds, 1)
	assert.Equal(hasChilds[0].Count, 4)
}
