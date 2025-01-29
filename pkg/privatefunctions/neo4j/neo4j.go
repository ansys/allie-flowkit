package neo4j

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/ansys/allie-flowkit/pkg/privatefunctions/codegeneration"
	"github.com/ansys/allie-sharedtypes/pkg/logging"
	"github.com/ansys/allie-sharedtypes/pkg/sharedtypes"
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
			logging.Log.Errorf(&logging.ContextMap{}, "Panic Initialize: %v", r)
			funcError = r.(error)
			return
		}
	}()

	// Create DB login object
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error during neo4j.NewDriverWithContext %v", err)
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
			logging.Log.Errorf(&logging.ContextMap{}, "Error during session.ExecuteWrite: %v", err)
			return nil, err
		}

		if result.Next(db_ctx) {
			return result.Record().Values, nil
		}

		logging.Log.Error(&logging.ContextMap{}, "nothing returned by query")
		return nil, errors.New("nothing returned by query")
	})
	if err != nil {
		return err
	}

	// Log successfull connection
	logging.Log.Infof(&logging.ContextMap{}, "Initialized neo4j database connection to %v", uri)

	return nil
}

////////////// Write functions //////////////

// AddCodeGenerationElementNodes adds nodes to neo4j database.
//
// Parameters:
//   - nodes: List of nodes to be added.
//
// Returns:
//   - funcError: Error object.
func (neo4j_context *neo4j_Context) AddCodeGenerationElementNodes(nodes []sharedtypes.CodeGenerationElement) (funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic AddCodeGenerationElementNodes: %v", r)
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
				logging.Log.Errorf(&logging.ContextMap{}, "Error serializing node to JSON: %v", err)
				return false, err
			}
			err = json.Unmarshal(nodeJSON, &nodeMap) // Convert JSON to map
			if err != nil {
				logging.Log.Errorf(&logging.ContextMap{}, "Error deserializing JSON to map: %v", err)
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
				logging.Log.Errorf(&logging.ContextMap{}, "Error during transaction.Run: %v", err)
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

// AddCodeGenerationExampleNodes adds nodes to neo4j database.
//
// Parameters:
//   - nodes: List of nodes to be added.
//
// Returns:
//   - funcError: Error object.
func (neo4j_context *neo4j_Context) AddCodeGenerationExampleNodes(nodes []sharedtypes.CodeGenerationExample) (funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic AddCodeGenerationExampleNodes: %v", r)
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
				logging.Log.Errorf(&logging.ContextMap{}, "Error serializing node to JSON: %v", err)
				return false, err
			}
			err = json.Unmarshal(nodeJSON, &nodeMap) // Convert JSON to map
			if err != nil {
				logging.Log.Errorf(&logging.ContextMap{}, "Error deserializing JSON to map: %v", err)
				return false, err
			}

			delete(nodeMap, "name")
			delete(nodeMap, "chunks")
			delete(nodeMap, "guid")

			// Add dependency equivalences map as a json string
			delete(nodeMap, "dependency_equivalences")
			dependencyEquivalencesJSON, err := json.Marshal(node.DependencyEquivalences)
			if err != nil {
				logging.Log.Errorf(&logging.ContextMap{}, "Error serializing dependency equivalences to JSON: %v", err)
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
				logging.Log.Errorf(&logging.ContextMap{}, "Error during transaction.Run: %v", err)
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
func (neo4j_context *neo4j_Context) AddUserGuideSectionNodes(nodes []sharedtypes.CodeGenerationUserGuideSection) (funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic AddUserGuideSectionNodes: %v", r)
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
				logging.Log.Errorf(&logging.ContextMap{}, "Error serializing node to JSON: %v", err)
				return false, err
			}
			err = json.Unmarshal(nodeJSON, &nodeMap) // Convert JSON to map
			if err != nil {
				logging.Log.Errorf(&logging.ContextMap{}, "Error deserializing JSON to map: %v", err)
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
				logging.Log.Errorf(&logging.ContextMap{}, "Error during transaction.Run: %v", err)
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

// CreateCodeGenerationExampleRelationships creates relationships between nodes in neo4j database.
//
// Parameters:
//   - relationships: List of relationships to be created.
//
// Returns:
//   - funcError: Error object.
func (neo4j_context *neo4j_Context) CreateCodeGenerationExampleRelationships(nodes []sharedtypes.CodeGenerationExample) (funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic CreateCodeGenerationExampleRelationships: %v", r)
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

		logging.Log.Infof(&logging.ContextMap{}, "Creating relationships for batch %v-%v", i, end)

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
						logging.Log.Errorf(&logging.ContextMap{}, "Error during transaction.Run: %v", err)
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

	logging.Log.Infof(&logging.ContextMap{}, "Created relationships for %v nodes", len(nodes))
	return nil
}

// CreateCodeGenerationRelationships creates relationships between nodes in neo4j database.
//
// Parameters:
//   - relationships: List of relationships to be created.
//
// Returns:
//   - funcError: Error object.
func (neo4j_context *neo4j_Context) CreateCodeGenerationRelationships(nodes []sharedtypes.CodeGenerationElement) (funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic CreateCodeGenerationRelationships: %v", r)
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

		logging.Log.Infof(&logging.ContextMap{}, "Creating relationships for batch %v-%v", i, end)

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
						logging.Log.Errorf(&logging.ContextMap{}, "Error during transaction.Run: %v", err)
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
						logging.Log.Errorf(&logging.ContextMap{}, "Error during transaction.Run: %v", err)
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
func (neo4j_context *neo4j_Context) CreateUserGuideSectionRelationships(nodes []sharedtypes.CodeGenerationUserGuideSection) (funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic CreateUserGuideSectionRelationships: %v", r)
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

		logging.Log.Infof(&logging.ContextMap{}, "Creating relationships for batch %v-%v", i, end)

		// Create relationships between sections and their references
		_, err := session.ExecuteWrite(db_ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
			for _, node := range batch {
				// Create relationships between each of the dependencies and the adjacent dependency
				for _, referenceLink := range node.ReferencedLinks {
					if node.Name == referenceLink || node.DocumentName == referenceLink {
						continue
					}

					// Check if reference link references the link of another section and create relationship
					_, err := transaction.Run(db_ctx,
						"MATCH (a {Name: $a}) MATCH (b {link: $b}) MERGE (a)-[:REFERENCES]->(b)",
						map[string]any{
							"a": node.Name,
							"b": referenceLink,
						},
					)
					if err != nil {
						logging.Log.Errorf(&logging.ContextMap{}, "Error during transaction.Run: %v", err)
						return false, err
					}

					// Check if reference link references a document and create relationship
					_, err = transaction.Run(db_ctx,
						"MATCH (a {Name: $a}) MATCH (b:UserGuide {document_name: $b}) WITH a, b ORDER BY b.level ASC LIMIT 1 MERGE (a)-[:REFERENCES]->(b)",
						map[string]any{
							"a": node.Name,
							"b": referenceLink,
						},
					)
					if err != nil {
						logging.Log.Errorf(&logging.ContextMap{}, "Error during transaction.Run: %v", err)
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
						logging.Log.Errorf(&logging.ContextMap{}, "Error during transaction.Run: %v", err)
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
						logging.Log.Errorf(&logging.ContextMap{}, "Error during transaction.Run: %v", err)
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
						logging.Log.Errorf(&logging.ContextMap{}, "Error during transaction.Run: %v", err)
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
					logging.Log.Errorf(&logging.ContextMap{}, "Error during transaction.Run: %v", err)
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

////////////// Read functions //////////////

// GetExamplesFromCodeGenerationElement gets examples from a code generation element.
//
// Parameters:
//   - elementType: Type of the code generation element.
//   - elementName: Name of the code generation element.
//
// Returns:
//   - exampleNames: List of example names.
//   - funcError: Error object.
func (neo4j_context *neo4j_Context) GetExamplesFromCodeGenerationElement(elementType string, elementName string) (exampleNames []string, funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic GetExamplesFromCodeGenerationElement: %v", r)
			funcError = r.(error)
			return
		}
	}()

	// Open session
	db_ctx := context.Background()
	session := (*neo4j_context.driver).NewSession(db_ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(db_ctx)

	exampleNames = []string{}

	// Get examples
	_, err := session.ExecuteRead(db_ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		result, err := transaction.Run(db_ctx,
			"MATCH (a:"+elementType+" {Name: $name})<-[:USES]-(b:Example) RETURN b.Name",
			map[string]any{
				"name": elementName,
			},
		)
		if err != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Error during transaction.Run: %v", err)
			return nil, err
		}

		for result.Next(db_ctx) {
			exampleNames = append(exampleNames, result.Record().Values[0].(string))
		}

		return nil, nil
	})

	if err != nil {
		log.Printf("Error during session.ExecuteRead: %v", err)
		return
	}

	return exampleNames, nil
}

// GetCodeGenerationElementAndDependencies gets a code generation element and its dependencies.
//
// Parameters:
//   - elementName: Name of the code generation element.
//   - maxHops: Maximum number of hops to search for dependencies.
//
// Returns:
//   - elements: List of code generation elements.
//   - funcError: Error object.
func (neo4j_context *neo4j_Context) GetCodeGenerationElementAndDependencies(elementName string, maxHops int) (elements []sharedtypes.CodeGenerationElement, funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic GetCodeGenerationElementAndDependencies: %v", r)
			funcError = r.(error)
			return
		}
	}()

	// Open session
	db_ctx := context.Background()
	session := (*neo4j_context.driver).NewSession(db_ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(db_ctx)

	elements = []sharedtypes.CodeGenerationElement{}

	// Execute query
	_, err := session.ExecuteRead(db_ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := fmt.Sprintf(`
			MATCH (start {Name: $element_name})
			OPTIONAL MATCH paths = (start)-[:BELONGS_TO*0..%d]->(node)
			WITH DISTINCT node
			OPTIONAL MATCH (node)-[:BELONGS_TO]->(dep)
			RETURN node,
				collect(DISTINCT dep.Name) AS dependencies,
				labels(node) AS type
		`, maxHops)

		params := map[string]any{
			"element_name": elementName,
		}

		result, err := tx.Run(db_ctx, query, params)
		if err != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Error during transaction.Run: %v", err)
			return nil, err
		}

		// Parse results
		for result.Next(db_ctx) {
			record := result.Record()

			node := record.Values[0].(neo4j.Node) // Extract node
			dependenciesRaw, _ := record.Get("dependencies")
			nodeType, _ := record.Get("type")

			// Convert dependencies to []string
			var dependencies []string
			if dependenciesRaw != nil {
				for _, dep := range dependenciesRaw.([]interface{}) {
					dependencies = append(dependencies, dep.(string))
				}
			}

			// Build CodeGenerationElement
			element := sharedtypes.CodeGenerationElement{
				Name:           getStringProp(node.Props, "Name"),
				NamePseudocode: getStringProp(node.Props, "namePseudocode"),
				NameFormatted:  getStringProp(node.Props, "nameFormatted"),
				Summary:        getStringProp(node.Props, "summary"),
				Dependencies:   dependencies,
			}

			if len(nodeType.([]interface{})) > 0 {
				nodeTypeString := nodeType.([]interface{})[0].(string)
				var nodeTypeEnum sharedtypes.CodeGenerationType

				nodeTypeEnum, err = codegeneration.StringToCodeGenerationType(nodeTypeString)
				if err != nil {
					logging.Log.Errorf(&logging.ContextMap{}, "Error converting node type to CodeGenerationType: %v", err)
					return nil, err
				}

				element.Type = nodeTypeEnum
			}

			elements = append(elements, element)
		}

		if err := result.Err(); err != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Error during result iteration: %v", err)
			return nil, err
		}

		return nil, nil
	})

	if err != nil {
		log.Printf("Error during session.ExecuteRead: %v", err)
		funcError = err
		return
	}

	return elements, nil
}

// GetUserGuideMainChapters gets the main chapters of the user guide.
//
// Returns:
//   - sections: List of user guide sections.
//   - funcError: Error object.
func (neo4j_context *neo4j_Context) GetUserGuideMainChapters() (sections []sharedtypes.CodeGenerationUserGuideSection, funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic GetUserGuideMainChapters: %v", r)
			funcError = r.(error)
			return
		}
	}()

	// Open session
	db_ctx := context.Background()
	session := (*neo4j_context.driver).NewSession(db_ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(db_ctx)

	// Execute query
	_, err := session.ExecuteRead(db_ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (firstChild {level: 1.0})
			WITH firstChild
			OPTIONAL MATCH path=(firstChild)-[:NEXT_SIBLING*]->(child {level: 1.0})
			ORDER BY length(path) DESC
			WITH firstChild, [n IN nodes(path)] AS chapters
			RETURN chapters
		`

		result, err := transaction.Run(db_ctx, query, map[string]any{})
		if err != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Error during transaction.Run: %v", err)
			return nil, err
		}

		for result.Next(db_ctx) {
			if result.Record().Values[0] == nil {
				continue
			}
			chapters := result.Record().Values[0].([]any)
			for _, chapter := range chapters {
				node, ok := chapter.(neo4j.Node)
				if !ok {
					logging.Log.Warnf(&logging.ContextMap{}, "Unexpected type in chapter: %v", chapter)
					continue
				}
				section := sharedtypes.CodeGenerationUserGuideSection{
					Name:         node.Props["Name"].(string),
					Title:        node.Props["title"].(string),
					Level:        int(node.Props["level"].(float64)),
					Parent:       node.Props["parent"].(string),
					DocumentName: node.Props["document_name"].(string),
				}

				sections = append(sections, section)
			}
			break
		}

		if err = result.Err(); err != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Error during result iteration: %v", err)
			return nil, err
		}

		return nil, nil
	})

	if err != nil {
		log.Printf("Error during session.ExecuteRead: %v", err)
		return
	}

	return sections, nil
}

// GetUserGuideSectionChildren gets the children of a user guide section.
//
// Parameters:
//   - sectionName: Name of the user guide section.
//
// Returns:
//   - sectionChildren: List of user guide sections.
//   - funcError: Error object.
func (neo4j_context *neo4j_Context) GetUserGuideSectionChildren(sectionName string) (sectionChildren []sharedtypes.CodeGenerationUserGuideSection, funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic in GetUserGuideSectionChildren: %v", r)
			funcError = r.(error)
			return
		}
	}()

	// Open session
	db_ctx := context.Background()
	session := (*neo4j_context.driver).NewSession(db_ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(db_ctx)

	// Execute query
	_, err := session.ExecuteRead(db_ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (n:UserGuide {Name: $sectionName})-[:HAS_CHILD]->(section_child)
			RETURN section_child
		`
		params := map[string]any{
			"sectionName": sectionName,
		}

		result, err := transaction.Run(db_ctx, query, params)
		if err != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Error during transaction.Run: %v", err)
			return nil, err
		}

		for result.Next(db_ctx) {
			record := result.Record()
			node := record.Values[0].(neo4j.Node) // Assumes the returned value is a Neo4j Node
			attributes := map[string]interface{}{}

			// Extract properties of the node
			for key, value := range node.Props {
				attributes[key] = value
			}

			child := sharedtypes.CodeGenerationUserGuideSection{
				Name:         node.Props["Name"].(string),
				Title:        node.Props["title"].(string),
				Level:        int(node.Props["level"].(float64)),
				Parent:       node.Props["parent"].(string),
				DocumentName: node.Props["document_name"].(string),
			}

			sectionChildren = append(sectionChildren, child)
		}

		if err := result.Err(); err != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Error during result iteration: %v", err)
			return nil, err
		}

		return nil, nil
	})

	if err != nil {
		log.Printf("Error during session.ExecuteRead: %v", err)
		return
	}

	return sectionChildren, nil
}

// GetUserGuideTableOfContents gets the table of contents of the user guide.
//
// Parameters:
//   - maxLevel: Maximum depth of the table of contents.
//
// Returns:
//   - tableOfContents: List of user guide sections.
//   - funcError: Error object.
func (neo4j_context *neo4j_Context) GetUserGuideTableOfContents(maxLevel int) (tableOfContents []map[string]interface{}, funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic in GetUserGuideTableOfContents: %v", r)
			funcError = r.(error)
			return
		}
	}()

	db_ctx := context.Background()
	session := (*neo4j_context.driver).NewSession(db_ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(db_ctx)

	// Fetch table of contents
	mainChapters, err := Neo4j_Driver.GetUserGuideMainChapters()
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error fetching main chapters: %v", err)
		return nil, err
	}

	tableOfContentsList := []map[string]interface{}{}

	// Process each section
	for _, section := range mainChapters {
		sectionNode, err := neo4j_context.getUserGuideNodeRecursive(section.Name, maxLevel)
		if err != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Error fetching section node %v: %v", section.Name, err)
			return nil, err
		}
		if sectionNode != nil {
			tableOfContentsList = append(tableOfContentsList, map[string]interface{}{
				"title":    sectionNode["title"],
				"Name":     sectionNode["Name"],
				"children": sectionNode["children"],
			})
		}
	}

	return tableOfContentsList, nil
}

// GetUserGuideNodeRecursive fetches a user guide node and its children recursively.
//
// Parameters:
//   - nodeName: Name of the user guide node.
//   - maxLevel: Maximum depth of the table of contents.
//
// Returns:
//   - nodeProps: Properties of the user guide node.
//   - funcError: Error object.
func (neo4j_context *neo4j_Context) getUserGuideNodeRecursive(nodeName string, maxLevel int) (map[string]interface{}, error) {
	db_ctx := context.Background()
	session := (*neo4j_context.driver).NewSession(db_ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(db_ctx)

	query := `
		MATCH (node {Name: $node_name})
		OPTIONAL MATCH (node)-[:HAS_FIRST_CHILD]->(firstChild)
		OPTIONAL MATCH path=(firstChild)-[:NEXT_SIBLING*]->(child)
		WITH node, firstChild, path
		ORDER BY length(path) DESC
		LIMIT 1
		WITH node, firstChild, [n IN nodes(path) | n.Name] AS orderedChildren
		RETURN node, node.name_pseudocode AS name_pseudocode, orderedChildren
	`
	params := map[string]any{
		"node_name": nodeName,
	}

	result, err := session.Run(context.Background(), query, params)
	if err != nil {
		return nil, fmt.Errorf("error during transaction.Run: %v", err)
	}

	record, err := result.Single(context.Background())
	if err != nil {
		return nil, nil // No record found, return nil without error
	}

	// Parse node properties
	node := record.Values[0].(neo4j.Node)
	nodeProps := map[string]interface{}{
		"title":    node.Props["title"],
		"Name":     node.Props["Name"],
		"children": []map[string]interface{}{},
	}

	// Recursively fetch children
	orderedChildrenNames, _ := record.Get("orderedChildren")
	if orderedChildrenNames != nil && maxLevel > 1 {
		for _, childName := range orderedChildrenNames.([]interface{}) {
			childNode, err := neo4j_context.getUserGuideNodeRecursive(childName.(string), maxLevel-1)
			if err != nil {
				return nil, fmt.Errorf("error fetching child node %v: %v", childName, err)
			}
			if childNode != nil {
				nodeProps["children"] = append(nodeProps["children"].([]map[string]interface{}), childNode)
			}
		}
	}

	return nodeProps, nil
}

// GetUserGuideSectionReferences gets the references of a user guide section.
//
// Parameters:
//   - sectionName: Name of the user guide section.
//
// Returns:
//   - referencedSections: List of referenced sections.
//   - funcError: Error object.
func (neo4j_context *neo4j_Context) GetUserGuideSectionReferences(sectionName string) (referencedSections string, funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic in GetUserGuideSectionReferences: %v", r)
			funcError = r.(error)
			return
		}
	}()

	// Open session
	db_ctx := context.Background()
	session := (*neo4j_context.driver).NewSession(db_ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(db_ctx)

	// Execute query
	_, err := session.ExecuteRead(db_ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (n:UserGuide {Name: $sectionName})-[:REFERENCES]->(reference)
			RETURN reference.Name
		`
		params := map[string]any{
			"sectionName": sectionName,
		}

		result, err := transaction.Run(db_ctx, query, params)
		if err != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Error during transaction.Run: %v", err)
			return nil, err
		}

		referencedSections = ""
		for result.Next(db_ctx) {
			record := result.Record()
			referencedSections += record.Values[0].(string) + ", "
		}

		if err := result.Err(); err != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Error during result iteration: %v", err)
			return nil, err
		}

		return nil, nil
	})
	if err != nil {
		log.Printf("Error during session.ExecuteRead: %v", err)
		return
	}

	return referencedSections, nil
}

// //////////// Helper functions //////////////

// getStringProp gets a string property from a map.
//
// Parameters:
//   - props: Map of properties.
//   - key: Key of the property.
//
// Returns:
//   - strVal: Value of the property.
func getStringProp(props map[string]interface{}, key string) string {
	if val, ok := props[key]; ok {
		if strVal, isString := val.(string); isString {
			return strVal
		}
	}
	return ""
}
