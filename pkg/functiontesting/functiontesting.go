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

package functiontesting

import (
	"fmt"

	"github.com/ansys/aali-flowkit/pkg/externalfunctions"
)

// TestAnsysGPTACSSearchIndex tests the AnsysGPTACSSemanticHybridSearch function.
// The function takes an index name and a query string as input.
// The function performs vector embedding on the query string and extracts fields from the query.
// The function then performs a semantic hybrid search on the index using the embedded query and extracted fields.
// The function prints the output of the search.
//
// Parameters:
//   - indexName: the name of the index to search
//   - query: the query string to search for
func TestAnsysGPTACSSearchIndex(indexName string, query string) {
	embeddedQuery := externalfunctions.PerformVectorEmbeddingRequest(query)

	// defaultFields := []sharedtypes.AnsysGPTDefaultFields{
	// 	{QueryWord: "course", FieldName: "type_of_asset", FieldDefaultValue: "aic"},
	// 	{QueryWord: "apdl", FieldName: "product", FieldDefaultValue: "mechanical apdl"},
	// 	{QueryWord: "lsdyna", FieldName: "product", FieldDefaultValue: "ls-dyna"},
	// }

	// filedValues := map[string][]string{
	// 	"physics":       {"structures", "fluids", "electronics", "structural mechanics", "discovery", "optics", "photonics", "python", "scade", "materials", "stem", "student", "fluid dynamics", "semiconductors"},
	// 	"type_of_asset": {"aic", "km", "documentation", "youtube", "general_faq", "alh", "article", "white-paper", "brochure"},
	// 	"product":       {"additive prep", "additive print", "autodyn", "avxcelerate", "cfx", "cfx pre", "cfx solver", "cfx turbogrid", "clock jitter flow", "cloud direct", "composite cure sim", "composite preppost", "designmodeler", "designxplorer", "diakopto", "discovery", "embedded software", "ensight", "exalto", "fluent", "forte", "gateway", "granta", "hfss", "icem cfd", "icepak", "ls-dyna", "lsdyna", "lumerical", "maxwell", "mechanical", "mechanical apdl", "medini", "meshing", "minerva", "motion", "ncode designlife", "pathfinder", "pathfinder-sc", "powerartist", "pragonx", "primex", "raptorh", "raptorx", "redhawk-sc", "redhawk-sc electrothermal", "redhawk-sc security", "rocky", "scade", "sherlock", "siwave", "spaceclaim", "spaceclaim directmodeler", "stk", "totem", "totem-sc", "twin builder", "velocerf", "voltage-timing", "workbench platform"},
	// }

	// indexNames := []string{"granular-ansysgpt", "ansysgpt-documentation-2023r2", "scade-documentation-2023r2", "ansys-dot-com-marketing", "ibp-app-brief", "ansysgpt-alh", "ansysgpt-scbu", "lsdyna-documentation-r14"}
	indexNames := []string{indexName}

	// ACS endpoint, API key, and API version
	acsEndpoint := ""
	acsApiKey := ""
	acsApiVersion := ""
	physics := []string{}

	// Extract fields from the query
	// filter := externalfunctions.AnsysGPTExtractFieldsFromQuery(query, filedValues, defaultFields)
	// output := externalfunctions.AnsysGPTACSSemanticHybridSearchs(acsEndpoint, acsApiKey, acsApiVersion, query, embeddedQuery, indexNames, filter, 10)
	output := externalfunctions.AisAcsSemanticHybridSearchs(acsEndpoint, acsApiKey, acsApiVersion, query, embeddedQuery, indexNames, physics, 10)
	fmt.Println(len(output))
}
