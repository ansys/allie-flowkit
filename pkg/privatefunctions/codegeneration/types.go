package codegeneration

import (
	"encoding/xml"
)

// Structs representing the XML structure
type XMLObjectDefinitionDocument struct {
	XMLName  xml.Name            `xml:"doc"`
	Assembly XMLAssembly         `xml:"assembly"`
	Members  []XMLAssemblyMember `xml:"members>member"`
}

type XMLAssembly struct {
	Name string `xml:"name"`
}

type XMLAssemblyMember struct {
	Name       string           `xml:"name,attr"`
	Summary    string           `xml:"summary"`
	ReturnType string           `xml:"returnType"`
	Params     []XMLMemberParam `xml:"param"`             // Handles multiple <param> elements
	Example    XMLMemberExample `xml:"example,omitempty"` // Optional <example> element
	Remarks    string           `xml:"remarks,omitempty"`
	EnumValues string           `xml:"enumValues,omitempty"`
}

type XMLMemberParam struct {
	Name        string `xml:"name" json:"name"`             // Attribute for <param>
	Type        string `xml:"type,omitempty" json:"type"`   // Attribute for <param>
	Description string `xml:",chardata" json:"description"` // Text content of <param>
}

type XMLMemberExample struct {
	Description string               `xml:",chardata" json:"description"` // Text content of <example>
	Code        XMLMemberExampleCode `xml:"code,omitempty" json:"code"`   // Optional <code> element
}

type XMLMemberExampleCode struct {
	Type string `xml:"type,attr" json:"type"` // Attribute for <code>
	Text string `xml:",chardata" json:"text"` // Text content of <code>
}

type CodeGenerationElement struct {
	Guid string             `json:"guid"`
	Type CodeGenerationType `json:"type"`

	NamePseudocode string `json:"name_pseudocode"` // Function name without dependencies
	NameFormatted  string `json:"name_formatted"`  // Name of the function with spaces and without parameters
	Description    string `json:"description"`

	Name              string   `json:"name"`
	Dependencies      []string `json:"dependencies"`
	Summary           string   `json:"summary"`
	ReturnType        string   `json:"return"`
	ReturnElementList []string `json:"return_element_list"`
	Remarks           string   `json:"remarks"`

	// Only for type "function" or "method"
	Parameters []XMLMemberParam `json:"parameters"`
	Example    XMLMemberExample `json:"example"`

	// Only for type "enum"
	EnumValues []string `json:"enum_values"`
}

// Enum values for CodeGenerationType
type CodeGenerationType string

const (
	Function  CodeGenerationType = "Function"
	Method    CodeGenerationType = "Method"
	Class     CodeGenerationType = "Class"
	Parameter CodeGenerationType = "Parameter"
	Enum      CodeGenerationType = "Enum"
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
	NameFormatted  string           `json:"name_formatted"`
	Description    string           `json:"description"`
	ParentClass    string           `json:"parent_class"`
}

type VectorDatabaseExample struct {
	Guid                   string            `json:"guid"`
	DenseVector            []float32         `json:"dense_vector"`
	SparseVector           map[uint]float32  `json:"sparse_vector"`
	DocumentName           string            `json:"document_name"`
	Dependencies           []string          `json:"dependencies"`
	DependencyEquivalences map[string]string `json:"dependency_equivalences"`
	Text                   string            `json:"text"`
	PreviousChunk          string            `json:"previous_chunk"`
	NextChunk              string            `json:"next_chunk"`
}

type GraphDatabaseElement struct {
	Guid           string           `json:"guid"`
	Type           string           `json:"type"`
	NamePseudocode string           `json:"name_pseudocode"`
	Description    string           `json:"description"`
	Summary        string           `json:"summary"`
	Examples       string           `json:"examples"`
	Parameters     []XMLMemberParam `json:"parameters"`
	Dependencies   []string         `json:"dependencies"`
	ReturnType     string           `json:"returnType"`
	Remarks        string           `json:"remarks"`
}

type CodeGenerationExample struct {
	Guid                   string            `json:"guid"`
	Name                   string            `json:"name"`
	Dependencies           []string          `json:"dependencies"`
	DependencyEquivalences map[string]string `json:"dependency_equivalences"`
	Chunks                 []string          `json:"chunks"`
}

type CodeGenerationUserGuideSection struct {
	Name            string   `json:"name"`
	Title           string   `json:"title"`
	IsFirstChild    bool     `json:"is_first_child"`
	NextSibling     string   `json:"next_sibling"`
	NextParent      string   `json:"next_parent"`
	DocumentName    string   `json:"document_name"`
	Parent          string   `json:"parent"`
	Content         string   `json:"content"`
	Level           int      `json:"level"`
	Link            string   `json:"link"`
	ReferencedLinks []string `json:"referenced_links"`
	Chunks          []string `json:"chunks"`
}

type VectorDatabaseUserGuideSection struct {
	Guid              string           `json:"guid"`
	SectionName       string           `json:"section_name"`
	DocumentName      string           `json:"document_name"`
	ParentSectionName string           `json:"parent_section_name"`
	Text              string           `json:"text"`
	Level             int              `json:"level"`
	PreviousChunk     string           `json:"previous_chunk"`
	NextChunk         string           `json:"next_chunk"`
	DenseVector       []float32        `json:"dense_vector"`
	SparseVector      map[uint]float32 `json:"sparse_vector"`
}
