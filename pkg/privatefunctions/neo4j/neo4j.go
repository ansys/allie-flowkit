package neo4j

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/ansys/allie-flowkit/pkg/internalstates"
	"github.com/ansys/allie-flowkit/pkg/privatefunctions/codegeneration"
	"github.com/ansys/allie-sharedtypes/pkg/logging"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type neo4j_Context struct {
	driver *neo4j.DriverWithContext
}

// Initialize DB login object
var Neo4j_Driver neo4j_Context

// Initialize neo4j database connection.
//
// Parameters:
//   - uri: URI of the neo4j database.
//   - username: Username of the neo4j database.
//   - password: Password of the neo4j database.
//
// Returns:
//   - funcError: Error object.
func Initialize(uri string, username string, password string) (funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(internalstates.Ctx, "Panic Initialize: %v", r)
			funcError = r.(error)
			return
		}
	}()

	// Create DB login object
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		logging.Log.Errorf(internalstates.Ctx, "Error during neo4j.NewDriverWithContext %v", err)
		return err
	}
	Neo4j_Driver = neo4j_Context{driver: &driver}

	// Check if DB connection is successfull
	db_ctx := context.Background()
	session := driver.NewSession(db_ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(db_ctx)

	_, err = session.ExecuteWrite(db_ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		result, err := transaction.Run(db_ctx,
			"RETURN 1",
			nil,
		)
		if err != nil {
			logging.Log.Errorf(internalstates.Ctx, "Error during session.ExecuteWrite: %v", err)
			return nil, err
		}

		if result.Next(db_ctx) {
			return result.Record().Values, nil
		}

		logging.Log.Error(internalstates.Ctx, "nothing returned by query")
		return nil, errors.New("nothing returned by query")
	})
	if err != nil {
		return err
	}

	// Log successfull connection
	logging.Log.Infof(internalstates.Ctx, "Initialized neo4j database connection to %v", uri)

	return nil
}

// AddNodes adds nodes to neo4j database.
//
// Parameters:
//   - nodes: List of nodes to be added.
//
// Returns:
//   - funcError: Error object.
func (neo4j_context *neo4j_Context) AddNodes(nodes []codegeneration.CodeGenerationElement) (funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(internalstates.Ctx, "Panic AddNodes: %v", r)
			funcError = r.(error)
			return
		}
	}()

	// Open session
	db_ctx := context.Background()
	session := (*neo4j_context.driver).NewSession(db_ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(db_ctx)

	counter := 0

	// Add nodes
	_, err := session.ExecuteWrite(db_ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		for _, node := range nodes {
			// Convert the node object to a map
			nodeType := string(node.Type) // Label for the node
			nodeName := node.Name         // Name of the node
			nodeMap := make(map[string]any)
			nodeJSON, err := json.Marshal(node) // Convert struct to JSON
			if err != nil {
				logging.Log.Errorf(internalstates.Ctx, "Error serializing node to JSON: %v", err)
				return false, err
			}
			err = json.Unmarshal(nodeJSON, &nodeMap) // Convert JSON to map
			if err != nil {
				logging.Log.Errorf(internalstates.Ctx, "Error deserializing JSON to map: %v", err)
				return false, err
			}

			// Flatten the map to avoid nested objects
			// flattenedMap := make(map[string]any)
			// generic.FlattenMap(nodeMap, "", flattenedMap)

			// Ensure "Type" is excluded from the properties
			delete(nodeMap, "type")
			delete(nodeMap, "name")

			// Add example in string format
			delete(nodeMap, "example")
			if node.Example.Description != "" {
				nodeMap["example"] = "Description: " + node.Example.Description + "\nCode: " + node.Example.Code.Text
			}

			// Add parameters in list of strings format
			delete(nodeMap, "parameters")
			parameters := make([]string, 0)
			for _, parameter := range node.Parameters {
				parameters = append(parameters, parameter.Name+": "+parameter.Description)
			}
			nodeMap["parameters"] = parameters

			// Create node dynamically using the map
			_, err = transaction.Run(db_ctx,
				"MERGE (n:"+nodeType+" {Name: $name}) SET n += $properties",
				map[string]any{
					"name":       nodeName,
					"properties": nodeMap,
				},
			)
			if err != nil {
				logging.Log.Errorf(internalstates.Ctx, "Error during transaction.Run: %v", err)
				return false, err
			}

			counter++
		}
		return true, nil
	})
	if err != nil {
		log.Printf("Error during session.ExecuteWrite: %v", err)
		return err
	}

	log.Printf("Added %v documents to neo4j", counter)
	return nil
}

// AddExampleNodes adds nodes to neo4j database.
//
// Parameters:
//   - nodes: List of nodes to be added.
//
// Returns:
//   - funcError: Error object.
func (neo4j_context *neo4j_Context) AddExampleNodes(nodes []codegeneration.CodeGenerationExample) (funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(internalstates.Ctx, "Panic AddNodes: %v", r)
			funcError = r.(error)
			return
		}
	}()

	// Open session
	db_ctx := context.Background()
	session := (*neo4j_context.driver).NewSession(db_ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(db_ctx)

	counter := 0

	// Add nodes
	_, err := session.ExecuteWrite(db_ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		for _, node := range nodes {
			// Convert the node object to a map
			nodeType := "Example"
			nodeName := node.Name
			nodeMap := make(map[string]any)
			nodeJSON, err := json.Marshal(node) // Convert struct to JSON
			if err != nil {
				logging.Log.Errorf(internalstates.Ctx, "Error serializing node to JSON: %v", err)
				return false, err
			}
			err = json.Unmarshal(nodeJSON, &nodeMap) // Convert JSON to map
			if err != nil {
				logging.Log.Errorf(internalstates.Ctx, "Error deserializing JSON to map: %v", err)
				return false, err
			}

			delete(nodeMap, "name")
			delete(nodeMap, "chunks")
			delete(nodeMap, "guid")

			// Add dependency equivalences map as a json string
			delete(nodeMap, "dependency_equivalences")
			dependencyEquivalencesJSON, err := json.Marshal(node.DependencyEquivalences)
			if err != nil {
				logging.Log.Errorf(internalstates.Ctx, "Error serializing dependency equivalences to JSON: %v", err)
				return false, err
			}
			nodeMap["dependency_equivalences"] = string(dependencyEquivalencesJSON)

			// Create node dynamically using the map
			_, err = transaction.Run(db_ctx,
				"MERGE (n:"+nodeType+" {Name: $name}) SET n += $properties",
				map[string]any{
					"name":       nodeName,
					"properties": nodeMap,
				},
			)
			if err != nil {
				logging.Log.Errorf(internalstates.Ctx, "Error during transaction.Run: %v", err)
				return false, err
			}

			counter++
		}
		return true, nil
	})
	if err != nil {
		log.Printf("Error during session.ExecuteWrite: %v", err)
		return err
	}

	log.Printf("Added %v documents to neo4j", counter)
	return nil
}

// AddUserGuideSectionNodes adds nodes to neo4j database.
//
// Parameters:
//   - nodes: List of nodes to be added.
//
// Returns:
//   - funcError: Error object.
func (neo4j_context *neo4j_Context) AddUserGuideSectionNodes(nodes []codegeneration.CodeGenerationUserGuideSection) (funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(internalstates.Ctx, "Panic AddNodes: %v", r)
			funcError = r.(error)
			return
		}
	}()

	// Open session
	db_ctx := context.Background()
	session := (*neo4j_context.driver).NewSession(db_ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(db_ctx)

	counter := 0

	// Add nodes
	_, err := session.ExecuteWrite(db_ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		for _, node := range nodes {
			// Convert the node object to a map
			nodeType := "UserGuide"
			nodeName := node.Name
			nodeMap := make(map[string]any)
			nodeJSON, err := json.Marshal(node) // Convert struct to JSON
			if err != nil {
				logging.Log.Errorf(internalstates.Ctx, "Error serializing node to JSON: %v", err)
				return false, err
			}
			err = json.Unmarshal(nodeJSON, &nodeMap) // Convert JSON to map
			if err != nil {
				logging.Log.Errorf(internalstates.Ctx, "Error deserializing JSON to map: %v", err)
				return false, err
			}

			delete(nodeMap, "name")
			delete(nodeMap, "content")
			delete(nodeMap, "chunks")
			delete(nodeMap, "is_first_child")
			delete(nodeMap, "next_sibling")
			delete(nodeMap, "next_parent")

			// Create node dynamically using the map
			_, err = transaction.Run(db_ctx,
				"MERGE (n:"+nodeType+" {Name: $name}) SET n += $properties",
				map[string]any{
					"name":       nodeName,
					"properties": nodeMap,
				},
			)
			if err != nil {
				logging.Log.Errorf(internalstates.Ctx, "Error during transaction.Run: %v", err)
				return false, err
			}

			counter++
		}
		return true, nil
	})
	if err != nil {
		log.Printf("Error during session.ExecuteWrite: %v", err)
		return err
	}

	log.Printf("Added %v documents to neo4j", counter)
	return nil
}

// CreateExampleRelationships creates relationships between nodes in neo4j database.
//
// Parameters:
//   - relationships: List of relationships to be created.
//
// Returns:
//   - funcError: Error object.
func (neo4j_context *neo4j_Context) CreateExampleRelationships(nodes []codegeneration.CodeGenerationExample) (funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(internalstates.Ctx, "Panic CreateRelationships: %v", r)
			funcError = r.(error)
			return
		}
	}()

	// Open session
	db_ctx := context.Background()
	session := (*neo4j_context.driver).NewSession(db_ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(db_ctx)

	// Create relationships in batches
	maxBatchSize := 1000
	for i := 0; i < len(nodes); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(nodes) {
			end = len(nodes)
		}

		// Create batch of nodes
		batch := nodes[i:end]

		logging.Log.Infof(internalstates.Ctx, "Creating relationships for batch %v-%v", i, end)

		_, err := session.ExecuteWrite(db_ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
			for _, node := range batch {
				// Create relationships between each of the dependencies and the adjacent dependency
				for _, dependency := range node.Dependencies {
					_, err := transaction.Run(db_ctx,
						"MATCH (a {Name: $a}) MATCH (b {Name: $b}) MERGE (a)-[:USES]->(b)",
						map[string]any{
							"a": node.Name,
							"b": dependency,
						},
					)
					if err != nil {
						logging.Log.Errorf(internalstates.Ctx, "Error during transaction.Run: %v", err)
						return false, err
					}
				}
			}
			return true, nil
		})
		if err != nil {
			log.Printf("Error during session.ExecuteWrite: %v", err)
			return
		}
	}

	logging.Log.Infof(internalstates.Ctx, "Created relationships for %v nodes", len(nodes))
	return nil
}

// CreateRelationships creates relationships between nodes in neo4j database.
//
// Parameters:
//   - relationships: List of relationships to be created.
//
// Returns:
//   - funcError: Error object.
func (neo4j_context *neo4j_Context) CreateRelationships(nodes []codegeneration.CodeGenerationElement) (funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(internalstates.Ctx, "Panic CreateRelationships: %v", r)
			funcError = r.(error)
			return
		}
	}()

	// Open session
	db_ctx := context.Background()
	session := (*neo4j_context.driver).NewSession(db_ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(db_ctx)

	counter := 0

	alreadyAddedRelationships := make(map[string]bool)

	// Create relationships in batches
	maxBatchSize := 1000
	for i := 0; i < len(nodes); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(nodes) {
			end = len(nodes)
		}

		// Create batch of nodes
		batch := nodes[i:end]

		logging.Log.Infof(internalstates.Ctx, "Creating relationships for batch %v-%v", i, end)

		_, err := session.ExecuteWrite(db_ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
			for _, node := range batch {
				// Reorder dependencies and create the complete dependency name
				dependencyList := make([]string, 0)
				// Add the node itself to the list of dependencies
				dependencyList = append(dependencyList, node.Name)
				for i := len(node.Dependencies) - 1; i >= 0; i-- {
					// Create the complete name of the dependency
					dependencyCompleteName := ""
					if i > 0 {
						for _, dependencyName := range node.Dependencies[:i] {
							dependencyCompleteName += dependencyName + "."
						}
					}
					dependencyCompleteName += node.Dependencies[i]

					dependencyList = append(dependencyList, dependencyCompleteName)
				}

				// Create relationships between each of the dependencies and the adjacent dependency
				for i, dependency := range dependencyList[:len(dependencyList)-1] {
					// Check if the relationship has already been added
					if alreadyAddedRelationships[dependency+"-"+dependencyList[i+1]] {
						continue
					}
					alreadyAddedRelationships[dependency+"-"+dependencyList[i+1]] = true

					_, err := transaction.Run(db_ctx,
						"MERGE (a {Name: $a}) MERGE (b {Name: $b}) MERGE (a)-[:BELONGS_TO]->(b)",
						map[string]any{
							"a": dependency,
							"b": dependencyList[i+1],
						},
					)
					if err != nil {
						logging.Log.Errorf(internalstates.Ctx, "Error during transaction.Run: %v", err)
						return false, err
					}
				}

				// Create relationships between the node and its return values
				for _, returnElement := range node.ReturnElementList {
					_, err := transaction.Run(db_ctx,
						"MATCH (a {Name: $a}) MATCH (b {Name: $b}) MERGE (a)-[:RETURNS]->(b)",
						map[string]any{
							"a": node.Name,
							"b": returnElement,
						},
					)
					if err != nil {
						logging.Log.Errorf(internalstates.Ctx, "Error during transaction.Run: %v", err)
						return false, err
					}
				}
			}
			return true, nil
		})
		if err != nil {
			log.Printf("Error during session.ExecuteWrite: %v", err)
			return
		}
	}

	log.Printf("Created %v relationships in neo4j", counter)
	return nil
}

// CreateUserGuideSectionRelationships creates relationships between nodes in neo4j database.
//
// Parameters:
//   - nodes: List of relationships to be created.
//
// Returns:
//   - funcError: Error object.
func (neo4j_context *neo4j_Context) CreateUserGuideSectionRelationships(nodes []codegeneration.CodeGenerationUserGuideSection) (funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(internalstates.Ctx, "Panic CreateRelationships: %v", r)
			funcError = r.(error)
			return
		}
	}()

	// Open session
	db_ctx := context.Background()
	session := (*neo4j_context.driver).NewSession(db_ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(db_ctx)

	counter := 0

	// Create relationships in batches
	maxBatchSize := 1000
	for i := 0; i < len(nodes); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(nodes) {
			end = len(nodes)
		}

		// Create batch of nodes
		batch := nodes[i:end]

		logging.Log.Infof(internalstates.Ctx, "Creating relationships for batch %v-%v", i, end)

		// Create relationships between sections and their references
		_, err := session.ExecuteWrite(db_ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
			for _, node := range batch {
				// Create relationships between each of the dependencies and the adjacent dependency
				for _, referenceLink := range node.ReferencedLinks {
					if node.Name == referenceLink {
						continue
					}
					_, err := transaction.Run(db_ctx,
						"MATCH (a {Name: $a}) MATCH (b {link: $b}) MERGE (a)-[:REFERENCES]->(b)",
						map[string]any{
							"a": node.Name,
							"b": referenceLink,
						},
					)
					if err != nil {
						logging.Log.Errorf(internalstates.Ctx, "Error during transaction.Run: %v", err)
						return false, err
					}
				}

				// Create relationship between each section and the next section
				if node.NextSibling != "" {
					_, err := transaction.Run(db_ctx,
						"MATCH (a {Name: $a}) MATCH (b {Name: $b}) MERGE (a)-[:NEXT_SIBLING]->(b)",
						map[string]any{
							"a": node.Name,
							"b": node.NextSibling,
						},
					)
					if err != nil {
						logging.Log.Errorf(internalstates.Ctx, "Error during transaction.Run: %v", err)
						return false, err
					}
				}

				// Create relationship between last child and following section parent
				if node.NextParent != "" {
					_, err := transaction.Run(db_ctx,
						"MATCH (a {Name: $a}) MATCH (b {Name: $b}) MERGE (a)-[:NEXT_PARENT]->(b)",
						map[string]any{
							"a": node.Name,
							"b": node.NextParent,
						},
					)
					if err != nil {
						logging.Log.Errorf(internalstates.Ctx, "Error during transaction.Run: %v", err)
						return false, err
					}
				}

				// Create relationship between each first child and its parent
				if node.IsFirstChild {
					_, err := transaction.Run(db_ctx,
						"MATCH (a {Name: $a}) MATCH (b {Name: $b}) MERGE (b)-[:HAS_FIRST_CHILD]->(a)",
						map[string]any{
							"a": node.Name,
							"b": node.Parent,
						},
					)
					if err != nil {
						logging.Log.Errorf(internalstates.Ctx, "Error during transaction.Run: %v", err)
						return false, err
					}
				}

				// Create relationships between each of the dependencies and the adjacent dependency
				_, err := transaction.Run(db_ctx,
					"MATCH (a {Name: $a}) MATCH (b {Name: $b}) MERGE (a)<-[:HAS_CHILD]-(b)",
					map[string]any{
						"a": node.Name,
						"b": node.Parent,
					},
				)
				if err != nil {
					logging.Log.Errorf(internalstates.Ctx, "Error during transaction.Run: %v", err)
					return false, err
				}
			}
			return true, nil
		})
		if err != nil {
			log.Printf("Error during session.ExecuteWrite: %v", err)
			return
		}
	}

	log.Printf("Created %v relationships in neo4j", counter)
	return nil
}
