package codegeneration

import (
	"encoding/xml"
)

// Structs representing the XML structure
type MechanicalObjectDefinitionDocument struct {
	XMLName  xml.Name                   `xml:"doc"`
	Assembly MechanicalAssembly         `xml:"assembly"`
	Members  []MechanicalAssemblyMember `xml:"members>member"`
}

type MechanicalAssembly struct {
	Name string `xml:"name"`
}

type MechanicalAssemblyMember struct {
	Name       string                  `xml:"name,attr"`
	Summary    string                  `xml:"summary"`
	ReturnType string                  `xml:"returnType"`
	Params     []MechanicalMemberParam `xml:"param"`             // Handles multiple <param> elements
	Example    MechanicalMemberExample `xml:"example,omitempty"` // Optional <example> element
	Remarks    string                  `xml:"remarks,omitempty"`
}

type MechanicalMemberParam struct {
	Name        string `xml:"name,attr" json:"name"`        // Attribute for <param>
	Description string `xml:",chardata" json:"description"` // Text content of <param>
}

type MechanicalMemberExample struct {
	Description string                      `xml:",chardata" json:"description"` // Text content of <example>
	Code        MechanicalMemberExampleCode `xml:"code,omitempty" json:"code"`   // Optional <code> element
}

type MechanicalMemberExampleCode struct {
	Type string `xml:"type,attr" json:"type"` // Attribute for <code>
	Text string `xml:",chardata" json:"text"` // Text content of <code>
}

type CodeGenerationElement struct {
	Guid string             `json:"guid"`
	Type CodeGenerationType `json:"type"`

	NamePseudocode string `json:"name_pseudocode"`
	Description    string `json:"description"`

	Name         string   `json:"name"`
	Dependencies []string `json:"dependencies"`
	Summary      string   `json:"summary"`
	ReturnType   string   `json:"return"`
	Remarks      string   `json:"remarks"`

	// Only for type "function" or "method"
	Parameters []MechanicalMemberParam `json:"parameters"`
	Example    MechanicalMemberExample `json:"example"`
}

// Enum values for CodeGenerationType
type CodeGenerationType string

const (
	Function  CodeGenerationType = "Function"
	Method    CodeGenerationType = "Method"
	Class     CodeGenerationType = "Class"
	Parameter CodeGenerationType = "Parameter"
)

type CodeGenerationPseudocodeResponse struct {
	Signature   string `json:"signature"`
	Description string `json:"description"`
}

type VectorDatabaseElement struct {
	Guid           string           `json:"guid"`
	DenseVector    []float32        `json:"dense_vector"`
	SparseVector   map[uint]float32 `json:"sparse_vector"`
	Type           string           `json:"type"`
	Name           string           `json:"name"`
	NamePseudocode string           `json:"name_pseudocode"`
	Description    string           `json:"description"`
}

type GraphDatabaseElement struct {
	Guid           string                  `json:"guid"`
	Type           string                  `json:"type"`
	NamePseudocode string                  `json:"name_pseudocode"`
	Description    string                  `json:"description"`
	Summary        string                  `json:"summary"`
	Examples       string                  `json:"examples"`
	Parameters     []MechanicalMemberParam `json:"parameters"`
	Dependencies   []string                `json:"dependencies"`
	ReturnType     string                  `json:"returnType"`
	Remarks        string                  `json:"remarks"`
}
