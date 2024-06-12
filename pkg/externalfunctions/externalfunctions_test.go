package externalfunctions

import (
	"testing"
)

func TestExtractFieldsFromQuery(t *testing.T) {
	fieldValues := map[string][]string{
		"physics":       {"structures", "fluids", "electronics", "structural mechanics", "discovery", "optics", "photonics", "python", "scade", "materials", "stem", "student", "fluid dynamics", "semiconductors"},
		"type_of_asset": {"aic", "km", "documentation", "youtube", "general_faq", "alh", "article", "white-paper", "brochure"},
		"product":       {"forte", "scade", "mechanical", "mechanical apdl", "fluent", "embedded software", "avxcelerate", "designxplorer", "designmodeler", "cloud direct", "maxwell", "stk", "ls-dyna", "lsdyna", "gateway", "granta", "rocky", "icepak", "siwave", "cfx", " meshing", " lumerical", "motion", "autodyn", "minerva", "redhawk-sc", "totem", "totem-sc", "powerartist", "raptorx", "velocerf", "exalto", "pathfinder", "pathfinder-sc", "diakopto", "pragonx", "primex", "on-chip electromagnetics", "redhawk-sc electrothermal", "redhawk-sc security", "voltage-timing and clock jitter", "medini", "ensight", "forte", "discovery", "hfss", "sherlock", "spaceclaim", "twin builder", "additive prep", "additive print", "composite cure sim", "composite preppost", "ncode designlife", "spaceclaim directmodeler", "cfx pre", "cfx solver", "cfx turbogrid", "icem cfd", "workbench platform"},
	}

	defaultFields := []DefaultFields{
		{
			QueryWord:         "course",
			FieldName:         "type_of_asset",
			FieldDefaultValue: "aic",
		},
		{
			QueryWord:         "apdl",
			FieldName:         "product",
			FieldDefaultValue: "ls-dyna",
		},
		{
			QueryWord:         "lsdyna",
			FieldName:         "product",
			FieldDefaultValue: "ls-dyna",
		},
	}

	testCases := []struct {
		query string
		want  map[string]string
	}{
		{
			query: "The contact tool in which branch of the model tree informs the user whether the contact pair is initially open? Please choose from the following options: Geometry, Connections, Model, or Solution.",
			want:  map[string]string{},
		},
		{
			query: "Which of the following controls/options are available under Analysis Setting in WB LS-DYNA documentation?",
			want:  map[string]string{"product": "ls-dyna", "type_of_asset": "documentation"},
		},
		{
			query: "How does bonded contact differ from Shared Topology?",
			want:  map[string]string{},
		},
		{
			query: "I'm interested in understanding how the residual vectors help in MSUP Harmonic Analysis. Can you list the KM on this topic?",
			want:  map[string]string{"type_of_asset": "km"},
		},
		{
			query: "In the Mechanical Fatigue tool according to product documentaton, which option is used to specify the stress type for fatigue calculations ? Please choose the best option from the following: Equivalent Stress, Exposure Duration, Fatigue Strength Factor, or Stress Component.",
			want:  map[string]string{"product": "mechanical", "type_of_asset": "documentation"},
		},
		{
			query: "Is there any courses available on Ansys Getting started with Mechanical?",
			want:  map[string]string{"product": "mechanical", "type_of_asset": "aic"},
		},
		{
			query: "Can you check in Ansys help manual about the term 'remote pont'?",
			want:  map[string]string{"physics": "stem"},
		},
		{
			query: "Is there any knowledge articles on how to define frictional contact in Ansys Mechanical?",
			want:  map[string]string{"product": "mechanical", "type_of_asset": "article"},
		},
		{
			query: "Is there any knowledge materials on how to define frictional contact in Ansys Mechanical?",
			want:  map[string]string{"product": "mechanical", "type_of_asset": "materials"},
		},
		{
			query: "Is there any KMs on how to model turbulent fluid flow in Ansys Fluent?",
			want:  map[string]string{"product": "fluent", "physics": "fluids", "type_of_asset": "km"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.query, func(t *testing.T) {
			result := ExtractFieldsFromQuery(tc.query, fieldValues, defaultFields)
			for key, value := range tc.want {
				if result[key] != value {
					t.Errorf("For query %q, expected %s: %s, but got %s", tc.query, key, value, result[key])
				}
			}
		})
	}
}

// func TestPerformLLMRephraseRequest(t *testing.T) {
// 	template := `Orders: You are a technical support assistant that is professional, friendly, multilingual that determines the contextual relevance between the current query in '{query}' and previous content from '{chat_history}' and create a rephrased query **only in the case of a 'follow-up query'**: \n

// 	You are an expert at finding if there is *contextual relevance* between the current query in '{query}' and most recent content in *'HumanMessage(content) and AIMessage(content)' of '{chat_history}'*. \n

// 	Your expertise lies in understanding the context of the conversation and identifying whether the '{query}' is a continuation of the most recent topic or a new topic altogether.\n

// 	*Only* if the '{query}' is about seeking additional information to the recent content from *'HumanMessage(content) and AIMessage(content)' of '{chat_history}'*, you must consider the '{query}' as a 'follow-up query'. \n
// 	`
// 	query := "How to define a remote point in the course Ansys mechanical?"

// 	history := []HistoricMessage{
// 		{
// 			Role:    "system",
// 			Content: "you are a pirate",
// 		},
// 		{
// 			Role:    "user",
// 			Content: "How to define a bean in Pymapdl?",
// 		},
// 		{
// 			Role:    "assistant",
// 			Content: "i dont know",
// 		},
// 	}

// 	result := PerformLLMRephraseRequest(template, query, history)

// 	t.Logf("Result: %s", result)
// }
