package codegeneration

import (
	"encoding/xml"

	"github.com/ansys/allie-sharedtypes/pkg/sharedtypes"
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
	Name       string                       `xml:"name,attr"`
	Summary    string                       `xml:"summary"`
	ReturnType string                       `xml:"returnType" json:"return_type"`
	Returns    string                       `xml:"returns,omitempty"`
	Params     []sharedtypes.XMLMemberParam `xml:"param" json:"parameters"` // Handles multiple <param> elements
	Example    sharedtypes.XMLMemberExample `xml:"example,omitempty"`       // Optional <example> element
	Remarks    string                       `xml:"remarks,omitempty"`
	EnumValues string                       `xml:"enumValues,omitempty"`
}

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
	Guid           string                       `json:"guid"`
	Type           string                       `json:"type"`
	NamePseudocode string                       `json:"name_pseudocode"`
	Description    string                       `json:"description"`
	Summary        string                       `json:"summary"`
	Examples       string                       `json:"examples"`
	Parameters     []sharedtypes.XMLMemberParam `json:"parameters"`
	Dependencies   []string                     `json:"dependencies"`
	ReturnType     string                       `json:"returnType"`
	Remarks        string                       `json:"remarks"`
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
