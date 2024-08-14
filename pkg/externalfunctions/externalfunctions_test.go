package externalfunctions

import (
	"testing"

	"github.com/ansys/allie-sharedtypes/pkg/sharedtypes"
)

func TestExtractFieldsFromQuery(t *testing.T) {
	fieldValues := map[string][]string{
		"physics":       {"structures", "fluids", "electronics", "structural mechanics", "discovery", "optics", "photonics", "python", "scade", "materials", "stem", "student", "fluid dynamics", "semiconductors"},
		"type_of_asset": {"aic", "km", "documentation", "youtube", "general_faq", "alh", "article", "white-paper", "brochure"},
		"product":       {"forte", "scade", "mechanical", "mechanical apdl", "fluent", "embedded software", "avxcelerate", "designxplorer", "designmodeler", "cloud direct", "maxwell", "stk", "ls-dyna", "lsdyna", "gateway", "granta", "rocky", "icepak", "siwave", "cfx", " meshing", " lumerical", "motion", "autodyn", "minerva", "redhawk-sc", "totem", "totem-sc", "powerartist", "raptorx", "velocerf", "exalto", "pathfinder", "pathfinder-sc", "diakopto", "pragonx", "primex", "on-chip electromagnetics", "redhawk-sc electrothermal", "redhawk-sc security", "voltage-timing and clock jitter", "medini", "ensight", "forte", "discovery", "hfss", "sherlock", "spaceclaim", "twin builder", "additive prep", "additive print", "composite cure sim", "composite preppost", "ncode designlife", "spaceclaim directmodeler", "cfx pre", "cfx solver", "cfx turbogrid", "icem cfd", "workbench platform"},
	}

	defaultFields := []sharedtypes.AnsysGPTDefaultFields{
		{
			QueryWord:         "course",
			FieldName:         "type_of_asset",
			FieldDefaultValue: "aic",
		},
		{
			QueryWord:         "apdl",
			FieldName:         "product",
			FieldDefaultValue: "mechanical apdl",
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
			query: "Is there any knowledge articles on how to define frictional contact in Ansys Mechanical?",
			want:  map[string]string{"product": "mechanical", "type_of_asset": "article"},
		},
		{
			query: "Is there any KMs on how to model turbulent fluid flow in Ansys Fluent?",
			want:  map[string]string{"product": "fluent", "physics": "fluids", "type_of_asset": "km"},
		},
		{
			query: "Please provide information on materials used in semiconductors.",
			want:  map[string]string{"physics": "semiconductors"},
		},
		{
			query: "What are the new features in the latest SCADE release?",
			want:  map[string]string{"product": "scade"},
		},
		{
			query: "I need some documentation on the CFX solver.",
			want:  map[string]string{"product": "cfx solver", "type_of_asset": "documentation"},
		},
		{
			query: "Are there any new articles on fluid dynamics?",
			want:  map[string]string{"physics": "fluid dynamics", "type_of_asset": "article"},
		},
		{
			query: "Tell me more about the Mechanical APDL course.",
			want:  map[string]string{"product": "mechanical apdl", "type_of_asset": "aic"},
		},
		{
			query: "Is there any training available for Discovery?",
			want:  map[string]string{"product": "discovery"},
		},
		{
			query: "Where can I find the general FAQ for DesignXplorer?",
			want:  map[string]string{"product": "designxplorer", "type_of_asset": "general_faq"},
		},
		{
			query: "Any new documentation on the Additive Print tool?",
			want:  map[string]string{"product": "additive print", "type_of_asset": "documentation"},
		},
		{
			query: "Looking for a brochure on SpaceClaim.",
			want:  map[string]string{"product": "spaceclaim", "type_of_asset": "brochure"},
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
		"gun", "firearm", "armament", "ammunition", "launch vehicle", "missile", "ballistic", "rocket", "torpedo", "bomb",
		"satellite", "mine", "explosive", "ordinance", "energetic materials", "propellants", "incendiary",
		"war", "ground vehicles", "weapon", "biological agent", "spacecraft",
		"nuclear", "classified articles", "directed energy weapons", "explosion", "jet engine", "defense", "military", "terrorism",
	}
	errorResponseMessage := "Prohibited content detected."

	testCases := []struct {
		query       string
		wantFound   bool
		wantMessage string
	}{
		// Original test cases
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
		{
			query:       "What are the new developments in software engineering?",
			wantFound:   false,
			wantMessage: "",
		},
		{
			query:       "Discuss the implications of gunfire and missile technology.",
			wantFound:   true,
			wantMessage: errorResponseMessage,
		},
		{
			query:       "Are there any developments in ballistic missile defense?",
			wantFound:   true,
			wantMessage: errorResponseMessage,
		},
		{
			query:       "What are the effects of nuclear power and its use in military applications?",
			wantFound:   true,
			wantMessage: errorResponseMessage,
		},
		{
			query:       "Looking into the properties of energetic materials.",
			wantFound:   true,
			wantMessage: errorResponseMessage,
		},
		{
			query:       "How are propellants used in spacecraft propulsion?",
			wantFound:   true,
			wantMessage: errorResponseMessage,
		},
		{
			query:       "Research on biological agents and their countermeasures.",
			wantFound:   true,
			wantMessage: errorResponseMessage,
		},
		{
			query:       "Exploring the field of aerospace engineering.",
			wantFound:   false,
			wantMessage: "",
		},
		{
			query:       "Impact of missile defense systems.",
			wantFound:   true,
			wantMessage: errorResponseMessage,
		},
		{
			query:       "What measures are taken to prevent explosive incidents?",
			wantFound:   true,
			wantMessage: errorResponseMessage,
		},
		{
			query:       "Analyzing the efficiency of new jet engines.",
			wantFound:   true,
			wantMessage: errorResponseMessage,
		},
		{
			query:       "Studying the defense mechanisms of modern military systems.",
			wantFound:   true,
			wantMessage: errorResponseMessage,
		},
		{
			query:       "Exploring advancements in directed energy weapons.",
			wantFound:   true,
			wantMessage: errorResponseMessage,
		},
		{
			query:       "The latest updates in classified articles.",
			wantFound:   true,
			wantMessage: errorResponseMessage,
		},
		{
			query:       "Looking into the propulsion systems of launch vehicles.",
			wantFound:   true,
			wantMessage: errorResponseMessage,
		},
		{
			query:       "Discussion on energetic materials and their applications.",
			wantFound:   true,
			wantMessage: errorResponseMessage,
		},
		{
			query:       "Examining the use of armaments in modern warfare.",
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
