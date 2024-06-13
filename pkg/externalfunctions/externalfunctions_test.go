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
			result := AnsysGPTExtractFieldsFromQuery(tc.query, fieldValues, defaultFields)
			for key, value := range tc.want {
				if result[key] != value {
					t.Errorf("For query %q, expected %s: %s, but got %s", tc.query, key, value, result[key])
				}
			}
		})
	}
}

func TestAnsysGPTCheckProhibitedWords(t *testing.T) {
	prohibitedWords := []string{
		"gun", "firearm", "armament", "ammunition", "launchvehicle", "missile", "ballistic", "rocket", "torpedo", "bomb",
		"satellite", "mine", "explosive", "ordinance", "energetic materials", "propellants", "incendiary",
		"war", "groundvehicles", "weapon", "biological agent", "spacecraft",
		"nuclear", "classifiedarticles", "directedenergyweapons", "explosion", "jetengine", "defense", "military", "terrorism",
	}
	errorResponseMessage := "Prohibited content detected."

	testCases := []struct {
		query       string
		wantFound   bool
		wantMessage string
	}{
		{
			query:       "This is a test query about gun control.",
			wantFound:   true,
			wantMessage: errorResponseMessage,
		},
		{
			query:       "We need more information on rocket launches.",
			wantFound:   true,
			wantMessage: errorResponseMessage,
		},
		{
			query:       "What are the new developments in space exploration?",
			wantFound:   false,
			wantMessage: "",
		},
		{
			query:       "Can you tell me about nuclear power plants?",
			wantFound:   true,
			wantMessage: errorResponseMessage,
		},
		{
			query:       "Details on biological agent research.",
			wantFound:   true,
			wantMessage: errorResponseMessage,
		},
		{
			query:       "How does a jetengine work?",
			wantFound:   true,
			wantMessage: errorResponseMessage,
		},
		{
			query:       "This text should not trigger any prohibited words.",
			wantFound:   false,
			wantMessage: "",
		},
		{
			query:       "The new ordinance passed yesterday.",
			wantFound:   true,
			wantMessage: errorResponseMessage,
		},
		{
			query:       "Understanding the chemistry of propellants.",
			wantFound:   true,
			wantMessage: errorResponseMessage,
		},
		{
			query:       "How to prevent explosion in a chemical plant?",
			wantFound:   true,
			wantMessage: errorResponseMessage,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.query, func(t *testing.T) {
			found, message := AnsysGPTCheckProhibitedWords(tc.query, prohibitedWords, errorResponseMessage)
			if found != tc.wantFound || message != tc.wantMessage {
				t.Errorf("For query %q, expected (%v, %s) but got (%v, %s)",
					tc.query, tc.wantFound, tc.wantMessage, found, message)
			}
		})
	}
}
