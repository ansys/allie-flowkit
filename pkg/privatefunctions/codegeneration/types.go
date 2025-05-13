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

package codegeneration

import (
	"encoding/xml"

	"github.com/ansys/aali-sharedtypes/pkg/sharedtypes"
	"github.com/google/uuid"
)

// Structs representing the XML structure
type XMLObjectDefinitionDocument struct {
	XMLName  xml.Name         `xml:"doc"`
	Assembly XMLAssembly      `xml:"assembly"`
	Members  []AssemblyMember `xml:"members>member"`
}

type XMLAssembly struct {
	Name string `xml:"name"`
}

type AssemblyMember struct {
	Name       string                       `xml:"name,attr" json:"name"`
	Summary    string                       `xml:"summary" json:"summary"`
	ReturnType string                       `xml:"returnType" json:"return_type"`
	Returns    string                       `xml:"returns,omitempty" json:"returns"`
	Params     []sharedtypes.XMLMemberParam `xml:"param" json:"parameters"`          // Handles multiple <param> elements
	Example    sharedtypes.XMLMemberExample `xml:"example,omitempty" json:"example"` // Optional <example> element
	Remarks    string                       `xml:"remarks,omitempty" json:"remarks"` // Optional <remarks> element
	EnumValues string                       `xml:"enumValues,omitempty" json:"enum_values"`
}

type CodeGenerationPseudocodeResponse struct {
	Signature   string `json:"signature"`
	Description string `json:"description"`
}

type VectorDatabaseElement struct {
	Guid           uuid.UUID        `json:"guid"`
	DenseVector    []float32        `json:"dense_vector"`
	SparseVector   map[uint]float32 `json:"sparse_vector"`
	Type           string           `json:"type"`
	Name           string           `json:"name"`
	NamePseudocode string           `json:"name_pseudocode"`
	NameFormatted  string           `json:"name_formatted"`
	ParentClass    string           `json:"parent_class"`
}

type VectorDatabaseExample struct {
	Guid                   uuid.UUID         `json:"guid"`
	DenseVector            []float32         `json:"dense_vector"`
	SparseVector           map[uint]float32  `json:"sparse_vector"`
	DocumentName           string            `json:"document_name"`
	Dependencies           []string          `json:"dependencies"`
	DependencyEquivalences map[string]string `json:"dependency_equivalences"`
	Text                   string            `json:"text"`
	PreviousChunk          *uuid.UUID        `json:"previous_chunk"`
	NextChunk              *uuid.UUID        `json:"next_chunk"`
}

type GraphDatabaseElement struct {
	Guid           uuid.UUID                    `json:"guid"`
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
	Guid              uuid.UUID `json:"guid"`
	SectionName       string    `json:"section_name"`
	DocumentName      string    `json:"document_name"`
	Title             string
	ParentSectionName string           `json:"parent_section_name"`
	Text              string           `json:"text"`
	Level             int              `json:"level"`
	PreviousChunk     *uuid.UUID       `json:"previous_chunk"`
	NextChunk         *uuid.UUID       `json:"next_chunk"`
	DenseVector       []float32        `json:"dense_vector"`
	SparseVector      map[uint]float32 `json:"sparse_vector"`
}
