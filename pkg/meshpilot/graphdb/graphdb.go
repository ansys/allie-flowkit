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

package graphdb

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ansys/aali-sharedtypes/pkg/aali_graphdb"
	"github.com/ansys/aali-sharedtypes/pkg/logging"
)

type graphDbContext struct {
	client *aali_graphdb.Client
	dbname string
}

// Initialize DB login object
var GraphDbDriver graphDbContext

// EstablishConnection graph database connection.
//
// Parameters:
//   - uri: URI of the graph database.
//   - db_name: Name of the database to connect to.
//
// Returns:
//   - funcError: Error object.
func EstablishConnection(uri string, db_name string) (funcError error) {
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
		dbname: db_name,
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
		logging.Log.Debugf(logCtx, "database %q exists in the graphdb", GraphDbDriver.dbname)
	} else {
		return fmt.Errorf("database %q does not exists in the graphdb", GraphDbDriver.dbname)
	}

	return nil
}

func (graphdb_context *graphDbContext) GetProperties(description string, query string) (properties []string, funcError error) {

	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic get properties from graphdb: %v", r)
			funcError = r.(error)
			return
		}
	}()

	params := aali_graphdb.ParameterMap{
		"description": aali_graphdb.StringValue(description),
	}

	result, err := aali_graphdb.CypherQueryReadGeneric[map[string][]string](graphdb_context.client, graphdb_context.dbname, query, params)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error during cypher query: %v", err)
		return nil, err
	}

	if len(result) == 0 {
		logging.Log.Debugf(&logging.ContextMap{}, "No properties found for description: %s", description)
		return []string{}, nil
	}

	// Extract properties from the result
	if propertiesList, ok := result[0]["properties"]; ok {
		properties = propertiesList
	} else {
		logging.Log.Debugf(&logging.ContextMap{}, "No properties found in the result for description: %s", description)
		return []string{}, nil
	}

	return properties, nil
}

func (graphdb_context *graphDbContext) GetSummaries(description string, query string) (summaries string, funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic get properties from graphdb: %v", r)
			funcError = r.(error)
			return
		}
	}()

	params := aali_graphdb.ParameterMap{
		"description": aali_graphdb.StringValue(description),
	}

	result, err := aali_graphdb.CypherQueryReadGeneric[map[string]interface{}](GraphDbDriver.client, GraphDbDriver.dbname, query, params)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error during cypher query: %v", err)
		return "", err
	}

	if len(result) == 0 {
		logging.Log.Debugf(&logging.ContextMap{}, "No summaries found for description: %s", description)
		return "", nil
	}

	// Extract summaries from the result
	if stateNode, ok := result[0]["stateNode"]; ok {
		summaries = stateNode.(map[string]interface{})["Summary"].(string)
	} else {
		logging.Log.Debugf(&logging.ContextMap{}, "No summaries found in the result for description: %s", description)
		return "", nil
	}

	return summaries, nil
}

func (graphdb_context *graphDbContext) GetActions(description string, query string) (actions []map[string]string, funcError error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic get properties from graphdb: %v", r)
			funcError = r.(error)
			return
		}
	}()

	params := aali_graphdb.ParameterMap{
		"description": aali_graphdb.StringValue(description),
	}

	// Execute the query and retrieve the result
	result, err := aali_graphdb.CypherQueryReadGeneric[map[string]interface{}](graphdb_context.client, graphdb_context.dbname, query, params)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error during cypher query: %v", err)
		return nil, err
	}

	if len(result) == 0 {
		logging.Log.Debugf(&logging.ContextMap{}, "No actions found for description: %s", description)
		return []map[string]string{}, nil
	}

	// Process the result
	for _, item := range result {
		actionNodeRaw, ok := item["actionNode"]
		if !ok {
			logging.Log.Debugf(&logging.ContextMap{}, "No actionNode found in the result for description: %s", description)
			continue
		}

		// Ensure actionNode is a map
		actionNode, ok := actionNodeRaw.(map[string]interface{})
		if !ok {
			logging.Log.Errorf(&logging.ContextMap{}, "Unexpected type for actionNode: %T", actionNodeRaw)
			continue
		}

		if "ACTION" != actionNode["node_type"] {
			logging.Log.Debugf(&logging.ContextMap{}, "Node label ACTION does not match action node type %s for description: %s", actionNode["node_type"], description)
			continue
		}

		action := make(map[string]string)
		for key, value := range actionNode {
			if slices.Contains([]string{"_id", "_label", "node_type", "sha1_id", "begin_conn_id", "end_conn_id", "_ID", "_LABEL"}, key) {
				continue
			}
			if key == "ActApi" && value == "" {
				logging.Log.Infof(&logging.ContextMap{}, "ActApi is None or empty for action node: %v", actionNode)
				continue
			}
			if strValue, ok := value.(string); ok && strValue != "" {
				action[key] = strValue
			}
		}

		if len(action) > 0 {
			actions = append(actions, action)
		}
	}

	if len(actions) == 0 {
		logging.Log.Debugf(&logging.ContextMap{}, "No actions found in the result for description: %s", description)
		return []map[string]string{}, nil
	}

	logging.Log.Debugf(&logging.ContextMap{}, "Actions found for description: %s, actions: %v", description, actions)
	return actions, nil
}

func (graphdb_context *graphDbContext) GetSolutions(fmFailureCode, primeMeshFailureCode, query string) (solutions []string, funcError error) {

	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic get properties from graphdb: %v", r)
			funcError = r.(error)
			return
		}
	}()

	logging.Log.Debugf(&logging.ContextMap{}, "failure codes: %s, %s", fmFailureCode, primeMeshFailureCode)

	params := aali_graphdb.ParameterMap{
		"fm_failure_code":    aali_graphdb.StringValue(fmFailureCode),
		"prime_failure_code": aali_graphdb.StringValue(primeMeshFailureCode),
	}

	logging.Log.Debugf(&logging.ContextMap{}, "Executing query with parameters: %v", params)

	result, err := aali_graphdb.CypherQueryReadGeneric[map[string]interface{}](graphdb_context.client, graphdb_context.dbname, query, params)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error during cypher query: %v", err)
		return []string{}, err
	}

	if len(result) == 0 {
		logging.Log.Debugf(&logging.ContextMap{}, "No solutions found for failure codes: %s, %s", fmFailureCode, primeMeshFailureCode)
		return []string{}, nil
	}

	// Extract summaries from the result
	for _, item := range result {
		if stateNode, ok := item["stateNode"]; ok {
			solution := stateNode.(map[string]interface{})["Description"].(string)
			if solution != "" {
				solutions = append(solutions, solution)
			} else {
				logging.Log.Debugf(&logging.ContextMap{}, "One of the solution description found in the result is empty for failure codes: %s, %s", fmFailureCode, primeMeshFailureCode)
			}
		} else {
			logging.Log.Debugf(&logging.ContextMap{}, "No stateNode found in the result for failure codes: %s, %s", fmFailureCode, primeMeshFailureCode)
		}
	}

	if len(solutions) == 0 {
		logging.Log.Debugf(&logging.ContextMap{}, "No solutions found in the result for failure codes: %s, %s", fmFailureCode, primeMeshFailureCode)
		return []string{}, nil
	}

	logging.Log.Debugf(&logging.ContextMap{}, "Solutions found for failure codes: %s, %s, solutions: %v", fmFailureCode, primeMeshFailureCode, solutions)

	return solutions, nil
}
