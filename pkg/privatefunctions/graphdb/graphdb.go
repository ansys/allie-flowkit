package graphdb

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"slices"
	"strings"

	"github.com/ansys/aali-graphdb-goclient/aali_graphdb"
	"github.com/ansys/aali-sharedtypes/pkg/logging"
	"github.com/ansys/aali-sharedtypes/pkg/sharedtypes"
	"github.com/ansys/allie-flowkit/pkg/privatefunctions/codegeneration"
)

type graphDbContext struct {
	client *aali_graphdb.Client
	dbname string
}

// Initialize DB login object
var GraphDbDriver graphDbContext

// Initialize graph database connection.
//
// Parameters:
//   - uri: URI of the graph database.
//
// Returns:
//   - funcError: Error object.
func Initialize(uri string) (funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic Initialize: %v", r)
			funcError = r.(error)
			return
		}
	}()

	logCtx := &logging.ContextMap{}

	// make sure address is absolute (for now, assume everything is http)
	scheme, _, found := strings.Cut(uri, "://")
	var addr string
	if found {
		if strings.ToLower(scheme) == "http" {
			addr = uri
		} else {
			return fmt.Errorf("expected http address but got scheme %q", scheme)
		}
	} else {
		addr = fmt.Sprintf("http://%v", uri)
	}

	// create client
	client, err := aali_graphdb.DefaultClient(addr)
	if err != nil {
		logging.Log.Errorf(logCtx, "Error creating graphdb client: %v", err)
		return err
	}

	GraphDbDriver = graphDbContext{
		client: client,
		dbname: "aali", // TODO: for now this is hard-coded, but may want configurable in the future
	}

	// Check if DB connection is successfull
	_, err = client.GetHealth()
	if err != nil {
		logging.Log.Errorf(logCtx, "Error during healthcheck: %v", err)
		return err
	}

	// if the database does not exist, create it
	dbs, err := client.GetDatabases()
	if err != nil {
		logging.Log.Errorf(logCtx, "unable to get existing databases")
		return err
	}
	if slices.Contains(dbs, GraphDbDriver.dbname) {
		logging.Log.Debugf(logCtx, "database %q already exists in the graphdb", GraphDbDriver.dbname)
	} else {
		logging.Log.Debugf(logCtx, "database %q does not exist, creating", GraphDbDriver.dbname)
		err := client.CreateDatabase(GraphDbDriver.dbname)
		if err != nil {
			logging.Log.Errorf(logCtx, "unable to create database %q", GraphDbDriver.dbname)
			return err
		}
	}

	// Log successfull connection
	logging.Log.Debugf(logCtx, "Initialized graphdb database connection to %v", addr)

	return nil
}

func graphdbStringList(l []string) aali_graphdb.ListValue {
	values := make([]aali_graphdb.Value, len(l))
	for i, e := range l {
		values[i] = aali_graphdb.StringValue(e)
	}
	return aali_graphdb.ListValue{
		LogicalType: aali_graphdb.StringLogicalType{},
		Values:      values,
	}
}

////////////// Write functions //////////////

// AddCodeGenerationElementNodes adds nodes to neo4j database.
//
// Parameters:
//   - nodes: List of nodes to be added.
//
// Returns:
//   - funcError: Error object.
func (graphdb_context *graphDbContext) AddCodeGenerationElementNodes(nodes []sharedtypes.CodeGenerationElement) (funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic AddCodeGenerationElementNodes: %v", r)
			funcError = r.(error)
			return
		}
	}()

	// Add nodes
	for _, node := range nodes {
		// Convert the node object to a map
		nodeType := string(node.Type) // Label for the node

		// Add parameters in list of strings format
		parameters := make([]string, 0)
		for _, parameter := range node.Parameters {
			parameterString := "Name: " + parameter.Name
			if parameter.Type != "" {
				parameterString += "\nType: " + parameter.Type
			}
			if parameter.Description != "" {
				parameterString += "\nDescription: " + parameter.Description
			}
			parameters = append(parameters, parameterString)
		}

		// update all fields except for: type, name
		queryParams := aali_graphdb.ParameterMap{
			"guid":                aali_graphdb.UUIDValue(node.Guid),
			"name_pseudocode":     aali_graphdb.StringValue(node.NamePseudocode),
			"name_formatted":      aali_graphdb.StringValue(node.NameFormatted),
			"description":         aali_graphdb.StringValue(node.Description),
			"name":                aali_graphdb.StringValue(node.Name),
			"dependencies":        graphdbStringList(node.Dependencies),
			"summary":             aali_graphdb.StringValue(node.Summary),
			"return_type":         aali_graphdb.StringValue(node.ReturnType),
			"return_element_list": graphdbStringList(node.ReturnElementList),
			"return_description":  aali_graphdb.StringValue(node.ReturnDescription),
			"remarks":             aali_graphdb.StringValue(node.Remarks),
			"enum_values":         graphdbStringList(node.EnumValues),
			"parameters":          graphdbStringList(parameters),
			"example":             aali_graphdb.StringValue("Description: " + node.Example.Description + "\nCode: " + node.Example.Code.Text),
		}

		// build up the query
		query := fmt.Sprintf(
			`
			MERGE (n:%v {Name: $name}
			SET
				n.guid = $guid,
				n.name_pseudocode = $name_pseudocode,
				n.name_formatted = $name_formatted,
				n.description = $description,
				n.dependencies = $dependencies,
				n.summary = $summary,
				n.return_type = $return_type,
				n.return_element_list = $return_element_list,
				n.return_description = $return_description,
				n.remarks = $remarks,
				n.enum_values = $enum_values,
				n.parameters = $parameters,
				n.example = $example
			`,
			nodeType,
		)

		// Create node dynamically using the map
		_, err := graphdb_context.client.CypherQueryWrite(
			graphdb_context.dbname,
			query,
			queryParams,
		)
		if err != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Error during cypher query: %v", err)
			return err
		}
	}

	log.Printf("Added %v documents to graphdb", len(nodes))
	return nil
}

// AddCodeGenerationExampleNodes adds nodes to graph database.
//
// Parameters:
//   - nodes: List of nodes to be added.
//
// Returns:
//   - funcError: Error object.
func (graphdb_context *graphDbContext) AddCodeGenerationExampleNodes(nodes []sharedtypes.CodeGenerationExample) (funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic AddCodeGenerationExampleNodes: %v", r)
			funcError = r.(error)
			return
		}
	}()

	// Add nodes
	for _, node := range nodes {
		// Add dependency equivalences map as a json string
		dependencyEquivalencesJSON, err := json.Marshal(node.DependencyEquivalences)
		if err != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Error serializing dependency equivalences to JSON: %v", err)
			return err
		}

		query := "MERGE (n:Example {Name: $name}) SET n.dependencies = $dependencies, n.dependency_equivalences = $dependency_equivalences"
		parameters := aali_graphdb.ParameterMap{
			"name":                    aali_graphdb.StringValue(node.Name),
			"dependencies":            graphdbStringList(node.Dependencies),
			"dependency_equivalences": aali_graphdb.StringValue(string(dependencyEquivalencesJSON)),
		}

		_, err = graphdb_context.client.CypherQueryWrite(
			graphdb_context.dbname,
			query,
			parameters,
		)
		if err != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Error during cypher query: %v", err)
			return err
		}

	}

	log.Printf("Added %v documents to neo4j", len(nodes))
	return nil
}

// AddUserGuideSectionNodes adds nodes to graph database.
//
// Parameters:
//   - nodes: List of nodes to be added.
//   - label: Label for the nodes.
//
// Returns:
//   - funcError: Error object.
func (graphdb_context *graphDbContext) AddUserGuideSectionNodes(nodes []sharedtypes.CodeGenerationUserGuideSection, label string) (funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic AddUserGuideSectionNodes: %v", r)
			funcError = r.(error)
			return
		}
	}()
	nodeType := "UserGuide"
	if label != "" {
		nodeType = label
	}

	// create the schema if not exists
	createTableQuery := fmt.Sprintf(
		`
			CREATE NODE TABLE IF NOT EXISTS %v(
				name STRING,
				title STRING,
				document_name STRING,
				parent STRING,
				level INT64,
				link STRING,
				referenced_links STRING[],

				PRIMARY KEY (name)
			)
		`,
		nodeType,
	)
	_, err := graphdb_context.client.CypherQueryWrite(graphdb_context.dbname, createTableQuery, nil)
	if err != nil {
		errMsg := fmt.Sprintf("error creating table %q: %q", nodeType, err)
		logging.Log.Errorf(&logging.ContextMap{}, errMsg)
		return errors.New(errMsg)
	}
	logging.Log.Debugf(&logging.ContextMap{}, "successfully created (if not exist) table %q", nodeType)

	// Add nodes
	query := fmt.Sprintf(
		`
		MERGE (n:%v {Name: $name})
		SET
			n.title = $title,
			n.document_name = $document_name,
			n.parent = $parent,
			n.level = $level,
			n.link = $link,
			n.referenced_links = $referenced_links
		`,
		nodeType,
	)
	for _, node := range nodes {
		// Convert the node object to parameters
		parameters := aali_graphdb.ParameterMap{
			"name":             aali_graphdb.StringValue(node.Name),
			"title":            aali_graphdb.StringValue(node.Title),
			"document_name":    aali_graphdb.StringValue(node.DocumentName),
			"parent":           aali_graphdb.StringValue(node.Parent),
			"level":            aali_graphdb.Int64Value(node.Level),
			"link":             aali_graphdb.StringValue(node.Link),
			"referenced_links": graphdbStringList(node.ReferencedLinks),
		}

		_, err := graphdb_context.client.CypherQueryWrite(
			graphdb_context.dbname,
			query,
			parameters,
		)
		if err != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Error during cypher query: %v", err)
			return err
		}

	}

	log.Printf("Added %v documents to graph db", len(nodes))
	return nil
}

// CreateCodeGenerationExampleRelationships creates relationships between nodes in graph database.
//
// Parameters:
//   - relationships: List of relationships to be created.
//
// Returns:
//   - funcError: Error object.
func (graphdb_context *graphDbContext) CreateCodeGenerationExampleRelationships(nodes []sharedtypes.CodeGenerationExample) (funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic CreateCodeGenerationExampleRelationships: %v", r)
			funcError = r.(error)
			return
		}
	}()

	// Create relationships
	for _, node := range nodes {
		for _, dependency := range node.Dependencies {
			query := "MATCH (a {Name: $a}) MATCH (b {Name: $b}) MERGE (a)-[:USES]->(b)"
			parameters := aali_graphdb.ParameterMap{

				"a": aali_graphdb.StringValue(node.Name),
				"b": aali_graphdb.StringValue(dependency),
			}
			_, err := graphdb_context.client.CypherQueryWrite(graphdb_context.dbname, query, parameters)
			if err != nil {
				logging.Log.Errorf(&logging.ContextMap{}, "Error during cypher query: %v", err)
				return err
			}

		}
	}

	logging.Log.Debugf(&logging.ContextMap{}, "Created relationships for %v nodes", len(nodes))
	return nil
}

// CreateCodeGenerationRelationships creates relationships between nodes in graph database.
//
// Parameters:
//   - relationships: List of relationships to be created.
//
// Returns:
//   - funcError: Error object.
func (graphdb_context *graphDbContext) CreateCodeGenerationRelationships(nodes []sharedtypes.CodeGenerationElement) (funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic CreateCodeGenerationRelationships: %v", r)
			funcError = r.(error)
			return
		}
	}()

	alreadyAddedRelationships := make(map[string]bool)

	// Create relationships
	for _, node := range nodes {
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

			_, err := graphdb_context.client.CypherQueryWrite(
				graphdb_context.dbname,
				"MERGE (a {Name: $a}) MERGE (b {Name: $b}) MERGE (a)-[:BELONGS_TO]->(b)",
				aali_graphdb.ParameterMap{
					"a": aali_graphdb.StringValue(dependency),
					"b": aali_graphdb.StringValue(dependencyList[i+1]),
				},
			)
			if err != nil {
				logging.Log.Errorf(&logging.ContextMap{}, "Error during cypher query: %v", err)
				return err
			}
		}

		// Create relationships between the node and its return values
		for _, returnElement := range node.ReturnElementList {
			_, err := graphdb_context.client.CypherQueryWrite(
				graphdb_context.dbname,
				"MATCH (a {Name: $a}) MATCH (b {Name: $b}) MERGE (a)-[:RETURNS]->(b)",
				aali_graphdb.ParameterMap{
					"a": aali_graphdb.StringValue(node.Name),
					"b": aali_graphdb.StringValue(returnElement),
				},
			)
			if err != nil {
				logging.Log.Errorf(&logging.ContextMap{}, "Error during cypher query: %v", err)
				return err
			}
		}

		// Create relationships between the node and its parameters
		parameterList := []string{}
		for _, parameter := range node.Parameters {
			patameterTypes, err := codegeneration.CreateReturnList(parameter.Type)
			if err != nil {
				logging.Log.Errorf(&logging.ContextMap{}, "Error during codegeneration.CreateReturnList: %v", err)
				return err
			}
			parameterList = append(parameterList, patameterTypes...)
		}

		for _, parameter := range parameterList {
			_, err := graphdb_context.client.CypherQueryWrite(
				graphdb_context.dbname,
				"MATCH (a {Name: $a}) MATCH (b {Name: $b}) MERGE (a)-[:USES_PARAMETER]->(b)",
				aali_graphdb.ParameterMap{
					"a": aali_graphdb.StringValue(node.Name),
					"b": aali_graphdb.StringValue(parameter),
				},
			)
			if err != nil {
				logging.Log.Errorf(&logging.ContextMap{}, "Error during cypher query: %v", err)
				return err
			}
		}
	}

	log.Printf("Created %v relationships in graph db", len(nodes))
	return nil
}

// CreateUserGuideSectionRelationships creates relationships between nodes in graph database.
//
// Parameters:
//   - nodes: List of relationships to be created.
//   - label: Label for the nodes (UserGuide by default).
//
// Returns:
//   - funcError: Error object.
func (graphdb_context *graphDbContext) CreateUserGuideSectionRelationships(nodes []sharedtypes.CodeGenerationUserGuideSection, label string) (funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic CreateUserGuideSectionRelationships: %v", r)
			funcError = r.(error)
			return
		}
	}()

	// Set label (node type)
	nodeType := "UserGuide"
	if label != "" {
		nodeType = label
	}

	// create references table
	createReferences := fmt.Sprintf("CREATE REL TABLE IF NOT EXISTS REFERENCES(FROM %s TO %s)", nodeType, nodeType)
	_, err := graphdb_context.client.CypherQueryWrite(graphdb_context.dbname, createReferences, nil)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "error during cypher query creating REFERENCES table: %v", err)
		return err
	}

	// Create relationships between sections and their references
	for _, node := range nodes {
		// Create relationships between each of the dependencies and the adjacent dependency
		for _, referenceLink := range node.ReferencedLinks {
			if node.Name == referenceLink || node.DocumentName == referenceLink {
				continue
			}

			// Check if reference link references the link of another section and create relationship
			query := fmt.Sprintf(
				"MATCH (a:%s {Name: $a}) MATCH (b:%s {link: $b}) MERGE (a)-[:REFERENCES]->(b)",
				nodeType,
				nodeType,
			)
			_, err := graphdb_context.client.CypherQueryWrite(
				graphdb_context.dbname,
				query,
				aali_graphdb.ParameterMap{
					"a": aali_graphdb.StringValue(node.Name),
					"b": aali_graphdb.StringValue(referenceLink),
				},
			)
			if err != nil {
				logging.Log.Errorf(&logging.ContextMap{}, "Error during cypher query: %v", err)
				return err
			}

			// Check if reference link references a document and create relationship
			query = fmt.Sprintf(
				"MATCH (a:%s {Name: $a}) MATCH (b:%s {document_name: $b}) WITH a, b ORDER BY b.level ASC LIMIT 1 MERGE (a)-[:REFERENCES]->(b)",
				nodeType,
				nodeType,
			)
			_, err = graphdb_context.client.CypherQueryWrite(
				graphdb_context.dbname,
				query,
				aali_graphdb.ParameterMap{
					"a": aali_graphdb.StringValue(node.Name),
					"b": aali_graphdb.StringValue(referenceLink),
				},
			)
			if err != nil {
				logging.Log.Errorf(&logging.ContextMap{}, "Error during cypher query: %v", err)
				return err
			}
		}

		// Create relationship between each section and the next section
		if node.NextSibling != "" {
			// create next sibling table
			createNextSibling := fmt.Sprintf("CREATE REL TABLE IF NOT EXISTS NEXT_SIBLING(FROM %s TO %s)", nodeType, nodeType)
			_, err := graphdb_context.client.CypherQueryWrite(graphdb_context.dbname, createNextSibling, nil)
			if err != nil {
				logging.Log.Errorf(&logging.ContextMap{}, "error during cypher query creating NEXT_SIBLING table: %v", err)
				return err
			}

			query := fmt.Sprintf(
				"MATCH (a:%s {Name: $a}) MATCH (b:%s {Name: $b}) MERGE (a)-[:NEXT_SIBLING]->(b)",
				nodeType,
				nodeType,
			)
			_, err = graphdb_context.client.CypherQueryWrite(
				graphdb_context.dbname,
				query,
				aali_graphdb.ParameterMap{
					"a": aali_graphdb.StringValue(node.Name),
					"b": aali_graphdb.StringValue(node.NextSibling),
				},
			)
			if err != nil {
				logging.Log.Errorf(&logging.ContextMap{}, "Error during cypher query: %v", err)
				return err
			}
		}

		// Create relationship between last child and following section parent
		if node.NextParent != "" {
			// create next parent table
			createNextParent := fmt.Sprintf("CREATE REL TABLE IF NOT EXISTS NEXT_PARENT(FROM %s TO %s)", nodeType, nodeType)
			_, err := graphdb_context.client.CypherQueryWrite(graphdb_context.dbname, createNextParent, nil)
			if err != nil {
				logging.Log.Errorf(&logging.ContextMap{}, "error during cypher query creating NEXT_PARENT table: %v", err)
				return err
			}

			query := fmt.Sprintf(
				"MATCH (a:%s {Name: $a}) MATCH (b:%s {Name: $b}) MERGE (a)-[:NEXT_PARENT]->(b)",
				nodeType,
				nodeType,
			)

			_, err = graphdb_context.client.CypherQueryWrite(
				graphdb_context.dbname,
				query,
				aali_graphdb.ParameterMap{
					"a": aali_graphdb.StringValue(node.Name),
					"b": aali_graphdb.StringValue(node.NextParent),
				},
			)
			if err != nil {
				logging.Log.Errorf(&logging.ContextMap{}, "Error during cypher query: %v", err)
				return err
			}
		}

		// Create relationship between each first child and its parent
		if node.IsFirstChild {
			// create has first child table
			createHasFirstChild := fmt.Sprintf("CREATE REL TABLE IF NOT EXISTS HAS_FIRST_CHILD(FROM %s TO %s)", nodeType, nodeType)
			_, err := graphdb_context.client.CypherQueryWrite(graphdb_context.dbname, createHasFirstChild, nil)
			if err != nil {
				logging.Log.Errorf(&logging.ContextMap{}, "error during cypher query creating HAS_FIRST_CHILD table: %v", err)
				return err
			}

			query := fmt.Sprintf(
				"MATCH (a:%s {Name: $a}) MATCH (b:%s {Name: $b}) MERGE (b)-[:HAS_FIRST_CHILD]->(a)",
				nodeType,
				nodeType,
			)

			_, err = graphdb_context.client.CypherQueryWrite(
				graphdb_context.dbname,
				query,
				aali_graphdb.ParameterMap{
					"a": aali_graphdb.StringValue(node.Name),
					"b": aali_graphdb.StringValue(node.Parent),
				},
			)
			if err != nil {
				logging.Log.Errorf(&logging.ContextMap{}, "Error during cypher query: %v", err)
				return err
			}
		}

		// create has_child table if it doesn't exist
		createHasChild := fmt.Sprintf("CREATE REL TABLE IF NOT EXISTS HAS_CHILD(FROM %s TO %s)", nodeType, nodeType)
		_, err := graphdb_context.client.CypherQueryWrite(graphdb_context.dbname, createHasChild, nil)
		if err != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "error during cypher query creating HAS_CHILD table: %v", err)
			return err
		}

		// Create relationships between each of the dependencies and the adjacent dependency
		query := fmt.Sprintf(
			"MATCH (a:%s {Name: $a}) MATCH (b:%s {Name: $b}) MERGE (a)<-[:HAS_CHILD]-(b)",
			nodeType,
			nodeType,
		)

		_, err = graphdb_context.client.CypherQueryWrite(
			graphdb_context.dbname,
			query,
			aali_graphdb.ParameterMap{
				"a": aali_graphdb.StringValue(node.Name),
				"b": aali_graphdb.StringValue(node.Parent),
			},
		)
		if err != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Error during cypher query: %v", err)
			return err
		}
	}

	log.Printf("Created %v relationships in graphdb", len(nodes))
	return nil
}

// WriteCypherQuery executes a cypher query with write access.
//
// Parameters:
//   - query: Cypher query to execute.
//
// Returns:
//   - results: array of map[string]any, keys are determined by the specific cypher query that was passed in
//   - err: error, if any
func (graphdb_context *graphDbContext) WriteCypherQuery(query string) ([]map[string]any, error) {
	return graphdb_context.client.CypherQueryWrite(graphdb_context.dbname, query, nil)
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
func (graphdb_context *graphDbContext) GetExamplesFromCodeGenerationElement(elementType string, elementName string) (exampleNames []string, funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic GetExamplesFromCodeGenerationElement: %v", r)
			funcError = r.(error)
			return
		}
	}()

	type exampleName struct {
		Name string `json:"b.Name"`
	}

	// Get examples
	query := fmt.Sprintf("MATCH (a:%v {Name: $name})<-[:USES]-(b:Example) RETURN b.Name", elementType)
	examples, err := aali_graphdb.CypherQueryReadGeneric[exampleName](
		graphdb_context.client,
		graphdb_context.dbname,
		query,
		aali_graphdb.ParameterMap{
			"name": aali_graphdb.StringValue(elementName),
		},
	)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error during cypher query: %v", err)
		return nil, err
	}

	exampleNames = make([]string, len(examples))
	for i, example := range examples {
		exampleNames[i] = example.Name
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
func (graphdb_context *graphDbContext) GetCodeGenerationElementAndDependencies(elementName string, maxHops int) (elements []sharedtypes.CodeGenerationElement, funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic GetCodeGenerationElementAndDependencies: %v", r)
			funcError = r.(error)
			return
		}
	}()

	// Execute query
	query := fmt.Sprintf(`
			MATCH (start {Name: $element_name})
			OPTIONAL MATCH paths = (start)-[:BELONGS_TO*0..%d]->(node)
			WITH DISTINCT node
			OPTIONAL MATCH (node)-[:BELONGS_TO]->(dep)
			RETURN
				node.Name                  AS name,
				node.namePseudocode        AS name_pseudocode,
				node.nameFormatted         AS name_formatted,
				node.summary               AS summary,
				collect(DISTINCT dep.Name) AS dependencies,
				labels(node)               AS type
		`, maxHops)

	params := aali_graphdb.ParameterMap{
		"element_name": aali_graphdb.StringValue(elementName),
	}
	elements, err := aali_graphdb.CypherQueryReadGeneric[sharedtypes.CodeGenerationElement](graphdb_context.client, graphdb_context.dbname, query, params)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error during cypher query: %v", err)
		return nil, err
	}

	return elements, nil
}

// GetUserGuideMainChapters gets the main chapters of the user guide.
//
// Returns:
//   - sections: List of user guide sections.
//   - funcError: Error object.
func (graphdb_context *graphDbContext) GetUserGuideMainChapters() (sections []sharedtypes.CodeGenerationUserGuideSection, funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic GetUserGuideMainChapters: %v", r)
			funcError = r.(error)
			return
		}
	}()

	// Execute query
	query := `
			MATCH (firstChild {level: 1.0})
			WITH firstChild
			OPTIONAL MATCH path=(firstChild)-[:NEXT_SIBLING*]->(child {level: 1.0})
			ORDER BY length(path) DESC
			WITH firstChild, [n IN nodes(path)] AS chapters
			RETURN chapters
		`
	result, err := aali_graphdb.CypherQueryReadGeneric[[]sharedtypes.CodeGenerationUserGuideSection](graphdb_context.client, graphdb_context.dbname, query, nil)

	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error during cypher query: %v", err)
		return nil, err
	}

	switch len(result) {
	case 0:
		logging.Log.Warnf(&logging.ContextMap{}, "did not find any first child")
		return []sharedtypes.CodeGenerationUserGuideSection{}, nil
	case 1:
		sections = result[0]
	default:
		errMsg := fmt.Sprintf("got more than 1 first child: %d", len(result))
		logging.Log.Error(&logging.ContextMap{}, errMsg)
		return nil, errors.New(errMsg)
	}

	logging.Log.Debugf(&logging.ContextMap{}, "found %d chapters", len(sections))
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
func (graphdb_context *graphDbContext) GetUserGuideSectionChildren(sectionName string) (sectionChildren []sharedtypes.CodeGenerationUserGuideSection, funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic in GetUserGuideSectionChildren: %v", r)
			funcError = r.(error)
			return
		}
	}()

	// Execute query
	query := `
			MATCH (n:UserGuide {Name: $sectionName})-[:HAS_CHILD]->(section_child)
			RETURN
				section_child.Name          AS name,
				section_child.title         AS title,
				section_child.level         AS level,
				section_child.parent        AS parent,
				section_child.document_name AS document_name
		`
	params := aali_graphdb.ParameterMap{
		"sectionName": aali_graphdb.StringValue(sectionName),
	}
	sectionChildren, err := aali_graphdb.CypherQueryReadGeneric[sharedtypes.CodeGenerationUserGuideSection](graphdb_context.client, graphdb_context.dbname, query, params)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error during cypher query: %v", err)
		return nil, err
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
func (graphdb_context *graphDbContext) GetUserGuideTableOfContents(maxLevel int) (tableOfContents []map[string]interface{}, funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic in GetUserGuideTableOfContents: %v", r)
			funcError = r.(error)
			return
		}
	}()

	// Fetch table of contents
	mainChapters, err := GraphDbDriver.GetUserGuideMainChapters()
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error fetching main chapters: %v", err)
		return nil, err
	}

	tableOfContentsList := []map[string]interface{}{}

	// Process each section
	for _, section := range mainChapters {
		sectionNode, err := graphdb_context.getUserGuideNodeRecursive(section.Name, maxLevel)
		if err != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Error fetching section node %v: %v", section.Name, err)
			return nil, err
		}
		if sectionNode != nil {
			tableOfContentsList = append(tableOfContentsList, map[string]interface{}{
				"title":    sectionNode.Title,
				"Name":     sectionNode.Name,
				"children": sectionNode.Children,
			})
		}
	}

	return tableOfContentsList, nil
}

type userGuideNodeProps struct {
	Title    string
	Name     string
	Children []userGuideNodeProps
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
func (graphdb_context *graphDbContext) getUserGuideNodeRecursive(nodeName string, maxLevel int) (*userGuideNodeProps, error) {
	type userGuideNodePropsRaw struct {
		Title    string   `json:"title"`
		Name     string   `json:"name"`
		Children []string `json:"children"`
	}

	query := `
		MATCH (node {Name: $node_name})
		OPTIONAL MATCH (node)-[:HAS_FIRST_CHILD]->(firstChild)
		OPTIONAL MATCH path=(firstChild)-[:NEXT_SIBLING*]->(child)
		WITH node, firstChild, path
		ORDER BY length(path) DESC
		LIMIT 1
		WITH node, firstChild, [n IN nodes(path) | n.Name] AS orderedChildren
		RETURN
			node.title      AS title,
			node.Name       AS name,
			orderedChildren AS children
	`
	params := aali_graphdb.ParameterMap{
		"node_name": aali_graphdb.StringValue(nodeName),
	}

	result, err := aali_graphdb.CypherQueryReadGeneric[userGuideNodePropsRaw](graphdb_context.client, graphdb_context.dbname, query, params)
	if err != nil {
		return nil, fmt.Errorf("error during cypher query: %v", err)
	}

	if len(result) != 1 {
		return nil, nil // No record found, return nil without error
	}
	record := result[0]

	// Recursively fetch children
	orderedChildrenNames := record.Children
	children := []userGuideNodeProps{}
	if orderedChildrenNames != nil && maxLevel > 1 {
		for _, childName := range orderedChildrenNames {
			childNode, err := graphdb_context.getUserGuideNodeRecursive(childName, maxLevel-1)
			if err != nil {
				return nil, fmt.Errorf("error fetching child node %v: %v", childName, err)
			}
			if childNode != nil {
				children = append(children, *childNode)
			}
		}
	}

	nodeProps := userGuideNodeProps{
		Title:    record.Title,
		Name:     record.Name,
		Children: children,
	}
	return &nodeProps, nil
}

// GetUserGuideSectionReferences gets the references of a user guide section.
//
// Parameters:
//   - sectionName: Name of the user guide section.
//
// Returns:
//   - referencedSections: List of referenced sections.
//   - funcError: Error object.
func (graphdb_context *graphDbContext) GetUserGuideSectionReferences(sectionName string) (referencedSections string, funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic in GetUserGuideSectionReferences: %v", r)
			funcError = r.(error)
			return
		}
	}()

	// Execute query
	type reference struct {
		Reference string `json:"reference.Name"`
	}
	query := `
		MATCH (n:UserGuide {Name: $sectionName})-[:REFERENCES]->(reference)
		RETURN reference.Name
		`
	params := aali_graphdb.ParameterMap{
		"sectionName": aali_graphdb.StringValue(sectionName),
	}
	result, err := aali_graphdb.CypherQueryReadGeneric[reference](graphdb_context.client, graphdb_context.dbname, query, params)

	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error during cypher query: %v", err)
		return "", err
	}

	referencedSections = ""
	for _, reference := range result {
		referencedSections += reference.Reference + ", "
	}

	return referencedSections, nil
}

// RetrieveDependencies retrieves the dependencies for the specified document from neo4j database.
//
// Parameters:
//   - ctx: ContextMap object.
//   - collectionName: Name of the collection.
//   - relationshipName: Name of the relationship.
//   - relationshipDirection: Direction of the relationship.
//   - sourceDocumentId: Id of the document.
//   - nodeTypesFilter: Node types filter.
//   - maxHops: Maximum number of hops.
//
// Returns:
//   - dependenciesIds: List of dependencies ids.
//   - funcError: Error object.
func (graphdb_context *graphDbContext) RetrieveDependencies(ctx *logging.ContextMap, collectionName string, relationshipName string, relationshipDirection string, sourceDocumentId string, nodeTypesFilter sharedtypes.DbArrayFilter, tagsFilter []string, maxHops int) (dependenciesIds []string, funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(ctx, "Panic retrieveDependencies: %v", r)
			funcError = r.(error)
			return
		}
	}()

	// Create node type expression
	nodeTypeExpression := ""
	if len(nodeTypesFilter.FilterData) > 0 {
		if nodeTypesFilter.NeedAll {
			nodeTypeExpression = ":" + strings.Join(nodeTypesFilter.FilterData, ":")
		} else {
			nodeTypeExpression = ":" + strings.Join(nodeTypesFilter.FilterData, "|")
		}
	}

	// Create tags filter
	neo4jWhereClause := " "
	if len(tagsFilter) > 0 {
		quotedSlice := make([]string, len(tagsFilter))
		for i, s := range tagsFilter {
			quotedSlice[i] = fmt.Sprintf("'%s'", s)
		}
		neo4jWhereClause = fmt.Sprintf(" WHERE n.documentType IN [%v] ", strings.Join(quotedSlice, ", "))
	}

	// Set the relationship direction
	var firstArrow, secondArrow string
	switch relationshipDirection {
	case "out":
		firstArrow = "-"
		secondArrow = "->"
	case "in":
		firstArrow = "<-"
		secondArrow = "-"
	case "both":
		firstArrow = "-"
		secondArrow = "-"
	}

	if maxHops > 0 {
		maxHops = maxHops - 1
	}

	// Retrieve dependencies
	queryTemplate, err := template.New("query").Parse(
		`
			MATCH (r {documentId: $sourceDocumentId, collectionName: $collectionName}){{.firstArrow}}[:{{.relationshipName}}]{{.secondArrow}}(n{{.nodeTypeExpression}}{collectionName: $collectionName}){{.firstArrow}}[:{{.relationshipName}}*0..{{.maxHops}}]{{.secondArrow}}(m{{.nodeTypeExpression}} {collectionName: $collectionName}){{.neo4jWhereClause}}
			RETURN m.documentId AS documentId
		`,
	)
	if err != nil {
		logging.Log.Errorf(ctx, "error parsing query template: %v", err)
		return nil, err
	}
	var queryBuf bytes.Buffer
	queryVars := map[string]any{
		"firstArrow":         firstArrow,
		"secondArrow":        secondArrow,
		"nodeTypeExpression": nodeTypeExpression,
		"neo4jWhereClause":   neo4jWhereClause,
		"relationshipName":   relationshipName,
		"maxHops":            maxHops,
	}
	err = queryTemplate.Execute(&queryBuf, queryVars)
	if err != nil {
		logging.Log.Errorf(ctx, "error executing query template: %v", err)
		return nil, err
	}
	query := queryBuf.String()
	params := aali_graphdb.ParameterMap{
		"sourceDocumentId": aali_graphdb.StringValue(sourceDocumentId),
		"collectionName":   aali_graphdb.StringValue(collectionName),
	}
	// result, err := graphdb_context.client.CypherQueryRead(graphdb_context.dbname, query, params)
	result, err := aali_graphdb.CypherQueryReadGeneric[struct {
		DocumentId string `json:"documentId"`
	}](graphdb_context.client, graphdb_context.dbname, query, params)
	if err != nil {
		logging.Log.Errorf(ctx, "error during cypher query: %v", err)
		return nil, err
	}

	dependenciesIds = make([]string, len(result))
	for i, dep := range result {
		dependenciesIds[i] = dep.DocumentId
	}

	return dependenciesIds, nil
}
