package graphdb

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"
	"text/template"

	"github.com/ansys/aali-flowkit/pkg/privatefunctions/codegeneration"
	"github.com/ansys/aali-graphdb-goclient/aali_graphdb"
	"github.com/ansys/aali-sharedtypes/pkg/logging"
	"github.com/ansys/aali-sharedtypes/pkg/sharedtypes"
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

	// initialize the schema
	err = GraphDbDriver.CreateSchema()
	if err != nil {
		logging.Log.Errorf(logCtx, "unable to create graph DB schema: %q", err)
		return err
	}

	// Log successfull connection
	logging.Log.Debugf(logCtx, "Initialized graphdb database connection to %v", addr)

	return nil
}

func (graphdb_context *graphDbContext) CreateSchema() error {
	stmts := []string{
		`CREATE NODE TABLE IF NOT EXISTS Element(
			type STRING,
			guid UUID,
			name_pseudocode STRING,
			name_formatted STRING,
			description STRING,
			name STRING,
			dependencies STRING[],
			summary STRING,
			return_type STRING,
			return_element_list STRING[],
			return_description STRING,
			remarks STRING,
			enum_values STRING[],
			parameters STRING[],
			example STRING,

			PRIMARY KEY (name)
		)`,
		`CREATE NODE TABLE IF NOT EXISTS Example(
			name STRING,
			dependencies STRING[],
			dependency_equivalences MAP(STRING, STRING),
			guid UUID,

			PRIMARY KEY (guid)
		)`,
		`CREATE NODE TABLE IF NOT EXISTS UserGuide(
			name STRING,
			title STRING,
			document_name STRING,
			parent STRING,
			level INT64,
			link STRING,
			referenced_links STRING[],

			PRIMARY KEY (name)
		)`,
		"CREATE REL TABLE IF NOT EXISTS Uses(FROM Example TO Element)",
		"CREATE REL TABLE IF NOT EXISTS BelongsTo(FROM Element TO Element)",
		"CREATE REL TABLE IF NOT EXISTS Returns(FROM Element TO Element)",
		"CREATE REL TABLE IF NOT EXISTS UsesParameter(FROM Element TO Element)",
		"CREATE REL TABLE IF NOT EXISTS References(FROM UserGuide TO UserGuide)",
		"CREATE REL TABLE IF NOT EXISTS NextSibling(FROM UserGuide TO UserGuide)",
		"CREATE REL TABLE IF NOT EXISTS NextParent(FROM UserGuide TO UserGuide)",
		"CREATE REL TABLE IF NOT EXISTS HasFirstChild(FROM UserGuide TO UserGuide)",
		"CREATE REL TABLE IF NOT EXISTS HasChild(FROM UserGuide TO UserGuide)"}

	for _, stmt := range stmts {
		_, err := graphdb_context.client.CypherQueryWrite(graphdb_context.dbname, stmt, nil)
		if err != nil {
			return err
		}
	}
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

// AddCodeGenerationElementNodes adds nodes to graphdb database.
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
			"type":                aali_graphdb.StringValue(node.Type),
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
		query := `
			MERGE (n:Element {type: $type, name: $name})
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
			`

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
		dependencyEquivalencesKuzu := map[aali_graphdb.Value]aali_graphdb.Value{}
		for k, v := range node.DependencyEquivalences {
			dependencyEquivalencesKuzu[aali_graphdb.StringValue(k)] = aali_graphdb.StringValue(v)
		}

		query := "MERGE (n:Example {guid: $guid}) SET n.name = $name, n.dependencies = $dependencies, n.dependency_equivalences = $dependency_equivalences"
		parameters := aali_graphdb.ParameterMap{
			"name":         aali_graphdb.StringValue(node.Name),
			"dependencies": graphdbStringList(node.Dependencies),
			"dependency_equivalences": aali_graphdb.MapValue{
				KeyType:   aali_graphdb.StringLogicalType{},
				ValueType: aali_graphdb.StringLogicalType{},
				Pairs:     dependencyEquivalencesKuzu,
			},
			"guid": aali_graphdb.UUIDValue(node.Guid),
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

		for _, dep := range node.Dependencies {
			query := `
				MERGE (n:Element {name: $name})
				SET n.example = $example
			`
			params := aali_graphdb.ParameterMap{
				"name":    aali_graphdb.StringValue(dep),
				"example": aali_graphdb.StringValue(node.Name),
			}
			_, err := graphdb_context.client.CypherQueryWrite(graphdb_context.dbname, query, params)
			if err != nil {
				logging.Log.Errorf(&logging.ContextMap{}, "Error during cypher query: %v", err)
				return err
			}
		}
	}

	log.Printf("Added %v documents to graphdb", len(nodes))
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
func (graphdb_context *graphDbContext) AddUserGuideSectionNodes(nodes []sharedtypes.CodeGenerationUserGuideSection) (funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic AddUserGuideSectionNodes: %v", r)
			funcError = r.(error)
			return
		}
	}()

	// Add nodes
	query := `
		MERGE (n:UserGuide {name: $name})
		SET
			n.title = $title,
			n.document_name = $document_name,
			n.parent = $parent,
			n.level = $level,
			n.link = $link,
			n.referenced_links = $referenced_links
		`
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
			query := `
				MATCH (example:Example), (element:Element)
				WHERE example.name = $example AND element.name = $element
				MERGE (example)-[:Uses]->(element)
			`
			// query := "MERGE (:Example {name: $example})-[:Uses]->(:Element {name: $element})"
			parameters := aali_graphdb.ParameterMap{
				"example": aali_graphdb.StringValue(node.Name),
				"element": aali_graphdb.StringValue(dependency),
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
				"MERGE (a:Element {name: $a}) MERGE (b:Element {name: $b}) MERGE (a)-[:BelongsTo]->(b)",
				aali_graphdb.ParameterMap{
					"a": aali_graphdb.StringValue(dependency),
					"b": aali_graphdb.StringValue(dependencyList[i+1]),
				},
			)
			if err != nil {
				errMsg := fmt.Sprintf("Error during cypher query adding BelongsTo relationships: %v", err)
				logging.Log.Errorf(&logging.ContextMap{}, errMsg)
				return errors.New(errMsg)
			}
		}

		// Create relationships between the node and its return values
		for _, returnElement := range node.ReturnElementList {
			_, err := graphdb_context.client.CypherQueryWrite(
				graphdb_context.dbname,
				"MERGE (a:Element {name: $a}) MERGE (b:Element {name: $b}) MERGE (a)-[:Returns]->(b)",
				aali_graphdb.ParameterMap{
					"a": aali_graphdb.StringValue(node.Name),
					"b": aali_graphdb.StringValue(returnElement),
				},
			)
			if err != nil {
				errMsg := fmt.Sprintf("Error during cypher query adding Returns relationships: %v", err)
				logging.Log.Errorf(&logging.ContextMap{}, errMsg)
				return errors.New(errMsg)
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
				"MERGE (a:Element {name: $a}) MERGE (b:Element {name: $b}) MERGE (a)-[:UsesParameter]->(b)",
				aali_graphdb.ParameterMap{
					"a": aali_graphdb.StringValue(node.Name),
					"b": aali_graphdb.StringValue(parameter),
				},
			)
			if err != nil {
				errMsg := fmt.Sprintf("Error during cypher query adding UsesParameter relationships: %v", err)
				logging.Log.Errorf(&logging.ContextMap{}, errMsg)
				return errors.New(errMsg)
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
func (graphdb_context *graphDbContext) CreateUserGuideSectionRelationships(nodes []sharedtypes.CodeGenerationUserGuideSection) (funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic CreateUserGuideSectionRelationships: %v", r)
			funcError = r.(error)
			return
		}
	}()

	// Create relationships between sections and their references
	for _, node := range nodes {
		// Create relationships between each of the dependencies and the adjacent dependency
		for _, referenceLink := range node.ReferencedLinks {
			if node.Name == referenceLink || node.DocumentName == referenceLink {
				continue
			}

			// Check if reference link references the link of another section and create relationship
			query := "MERGE (a:UserGuide {name: $section}) MERGE (b:UserGuide {name: $link}) MERGE (a)-[:References]->(b)"
			_, err := graphdb_context.client.CypherQueryWrite(
				graphdb_context.dbname,
				query,
				aali_graphdb.ParameterMap{
					"section": aali_graphdb.StringValue(node.Name),
					"link":    aali_graphdb.StringValue(referenceLink),
				},
			)
			if err != nil {
				errMsg := fmt.Sprintf("Error during cypher query adding References relationships: %v", err)
				logging.Log.Errorf(&logging.ContextMap{}, errMsg)
				return errors.New(errMsg)
			}
		}

		// Create relationship between each section and the next section
		if node.NextSibling != "" {
			query := "MERGE (a:UserGuide {name: $a}) MERGE (b:UserGuide {name: $b}) MERGE (a)-[:NextSibling]->(b)"
			_, err := graphdb_context.client.CypherQueryWrite(
				graphdb_context.dbname,
				query,
				aali_graphdb.ParameterMap{
					"a": aali_graphdb.StringValue(node.Name),
					"b": aali_graphdb.StringValue(node.NextSibling),
				},
			)
			if err != nil {
				errMsg := fmt.Sprintf("Error during cypher query adding NextSibling relationships: %v", err)
				logging.Log.Errorf(&logging.ContextMap{}, errMsg)
				return errors.New(errMsg)
			}
		}

		// Create relationship between last child and following section parent
		if node.NextParent != "" {
			query := "MERGE (a:UserGuide {name: $a}) MERGE (b:UserGuide {name: $b}) MERGE (a)-[:NextParent]->(b)"

			_, err := graphdb_context.client.CypherQueryWrite(
				graphdb_context.dbname,
				query,
				aali_graphdb.ParameterMap{
					"a": aali_graphdb.StringValue(node.Name),
					"b": aali_graphdb.StringValue(node.NextParent),
				},
			)
			if err != nil {
				errMsg := fmt.Sprintf("Error during cypher query adding NextParent relationships: %v", err)
				logging.Log.Errorf(&logging.ContextMap{}, errMsg)
				return errors.New(errMsg)
			}
		}

		// Create relationship between each first child and its parent
		if node.IsFirstChild {
			// create has first child table
			query := "MERGE (a:UserGuide {name: $a}) MERGE (b:UserGuide {name: $b}) MERGE (b)-[:HasFirstChild]->(a)"

			_, err := graphdb_context.client.CypherQueryWrite(
				graphdb_context.dbname,
				query,
				aali_graphdb.ParameterMap{
					"a": aali_graphdb.StringValue(node.Name),
					"b": aali_graphdb.StringValue(node.Parent),
				},
			)
			if err != nil {
				errMsg := fmt.Sprintf("Error during cypher query adding HasFirstChild relationships: %v", err)
				logging.Log.Errorf(&logging.ContextMap{}, errMsg)
				return errors.New(errMsg)
			}
		}

		// Create relationships between each of the dependencies and the adjacent dependency
		query := "MERGE (a:UserGuide {name: $a}) MERGE (b:UserGuide {name: $b}) MERGE (a)<-[:HasChild]-(b)"

		_, err := graphdb_context.client.CypherQueryWrite(
			graphdb_context.dbname,
			query,
			aali_graphdb.ParameterMap{
				"a": aali_graphdb.StringValue(node.Name),
				"b": aali_graphdb.StringValue(node.Parent),
			},
		)
		if err != nil {
			errMsg := fmt.Sprintf("Error during cypher query adding HasChild relationships: %v", err)
			logging.Log.Errorf(&logging.ContextMap{}, errMsg)
			return errors.New(errMsg)
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
	query := fmt.Sprintf("MATCH (a:%v {Name: $name})<-[:Uses]-(b:Example) RETURN b.Name", elementType)
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
			OPTIONAL MATCH paths = (start)-[:BelongsTo*0..%d]->(node)
			WITH DISTINCT node
			OPTIONAL MATCH (node)-[:BelongsTo]->(dep)
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
			OPTIONAL MATCH path=(firstChild)-[:NextSibling*]->(child {level: 1.0})
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
			MATCH (n:UserGuide {Name: $sectionName})-[:HasChild]->(section_child)
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
		OPTIONAL MATCH (node)-[:HasFirstChild]->(firstChild)
		OPTIONAL MATCH path=(firstChild)-[:NextSibling*]->(child)
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
		MATCH (n:UserGuide {Name: $sectionName})-[:References]->(reference)
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

// RetrieveDependencies retrieves the dependencies for the specified document from graph database.
//
// Parameters:
//   - ctx: ContextMap object.
//   - relationshipName: Name of the relationship.
//   - relationshipDirection: Direction of the relationship.
//   - sourceDocumentId: Id of the document.
//   - nodeTypesFilter: Node types filter.
//   - maxHops: Maximum number of hops.
//
// Returns:
//   - dependenciesIds: List of dependencies ids.
//   - funcError: Error object.
func (graphdb_context *graphDbContext) RetrieveDependencies(ctx *logging.ContextMap, relationshipName string, relationshipDirection string, sourceDocumentName string, nodeTypesFilter sharedtypes.DbArrayFilter, tagsFilter []string, maxHops int) (dependencyNames []string, funcError error) {
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
	whereClause := " "
	if len(tagsFilter) > 0 {
		quotedSlice := make([]string, len(tagsFilter))
		for i, s := range tagsFilter {
			quotedSlice[i] = fmt.Sprintf("'%s'", s)
		}
		whereClause = fmt.Sprintf(" WHERE n.documentType IN [%v] ", strings.Join(quotedSlice, ", "))
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
			MATCH (r {name: $sourceDocumentId}){{.firstArrow}}[:{{.relationshipName}}*1..{{.maxHops}}]{{.secondArrow}}(m{{.nodeTypeExpression}} ){{.whereClause}}
			RETURN m.name AS name
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
		"whereClause":        whereClause,
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
		"sourceDocumentId": aali_graphdb.StringValue(sourceDocumentName),
	}
	result, err := aali_graphdb.CypherQueryReadGeneric[struct {
		Name string `json:"name"`
	}](graphdb_context.client, graphdb_context.dbname, query, params)
	if err != nil {
		logging.Log.Errorf(ctx, "error during cypher query: %v", err)
		return nil, err
	}

	dependencyNames = make([]string, len(result))
	for i, dep := range result {
		dependencyNames[i] = dep.Name
	}

	return dependencyNames, nil
}
