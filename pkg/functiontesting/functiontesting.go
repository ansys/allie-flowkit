package functiontesting

import (
	"fmt"

	"github.com/ansys/allie-flowkit/pkg/externalfunctions"
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

	defaultFields := []externalfunctions.AnsysGPTDefaultFields{
		{QueryWord: "course", FieldName: "type_of_asset", FieldDefaultValue: "aic"},
		{QueryWord: "apdl", FieldName: "product", FieldDefaultValue: "mechanical apdl"},
		{QueryWord: "lsdyna", FieldName: "product", FieldDefaultValue: "ls-dyna"},
	}

	filedValues := map[string][]string{
		"physics":       {"structures", "fluids", "electronics", "structural mechanics", "discovery", "optics", "photonics", "python", "scade", "materials", "stem", "student", "fluid dynamics", "semiconductors"},
		"type_of_asset": {"aic", "km", "documentation", "youtube", "general_faq", "alh", "article", "white-paper", "brochure"},
		"product":       {"forte", "scade", "mechanical", "mechanical apdl", "fluent", "embedded software", "avxcelerate", "designxplorer", "designmodeler", "cloud direct", "maxwell", "stk", "ls-dyna", "lsdyna", "gateway", "granta", "rocky", "icepak", "siwave", "cfx", "meshing", " lumerical", "motion", "autodyn", "minerva", "redhawk-sc", "totem", "totem-sc", "powerartist", "raptorx", "velocerf", "exalto", "pathfinder", "pathfinder-sc", "diakopto", "pragonx", "primex", "on-chip electromagnetics", "redhawk-sc electrothermal", "redhawk-sc security", "voltage-timing and clock jitter", "medini", "ensight", "forte", "discovery", "hfss", "sherlock", "spaceclaim", "twin builder", "additive prep", "additive print", "composite cure sim", "composite preppost", "ncode designlife", "spaceclaim directmodeler", "cfx pre", "cfx solver", "cfx turbogrid", "icem cfd", "workbench platform"},
	}

	indexNames := []string{"granular-ansysgpt", "ansysgpt-documentation-2023r2", "scade-documentation-2023r2", "ansys-dot-com-marketing", "ibp-app-brief", "ansysgpt-alh", "ansysgpt-scbu", "lsdyna-documentation-r14"}
	// indexNames := []string{indexName}

	filter := externalfunctions.AnsysGPTExtractFieldsFromQuery(query, filedValues, defaultFields)
	output := externalfunctions.AnsysGPTACSSemanticHybridSearchs(query, embeddedQuery, indexNames, filter, 10)
	fmt.Println(output)
}
