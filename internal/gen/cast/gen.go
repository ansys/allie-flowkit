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

package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"unicode"
)

func capitalizeWord(word string) string {
	if word == "" {
		return ""
	}
	runes := []rune(word)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func funcName(from string, to string) string {
	fromName := capitalizeWord(from)
	toName := capitalizeWord(to)
	return fmt.Sprintf("Cast%vTo%v", fromName, toName)
}

func displayName(from string, to string) string {
	fromName := capitalizeWord(from)
	toName := capitalizeWord(to)
	return fmt.Sprintf("Cast %v to %v", fromName, toName)
}

type CastFuncs struct {
	CastAssertFromInterfaces []CastAssertFromInterface
	CastAssertToInterfaces   []CastAssertToInterface
}

type CastAssertFromInterface struct {
	Name        string
	DisplayName string
	FromType    string
	ToType      string
}

func BasicCastAssertFromInterface(from string, to string) CastAssertFromInterface {
	return CastAssertFromInterface{funcName(from, to), displayName(from, to), from, to}
}

type CastAssertToInterface struct {
	Name        string
	DisplayName string
	FromType    string
	ToType      string
}

func BasicCastAssertToInterface(from string, to string) CastAssertToInterface {
	return CastAssertToInterface{funcName(from, to), displayName(from, to), from, to}
}

func main() {
	castFuncs := CastFuncs{
		[]CastAssertFromInterface{},
		[]CastAssertToInterface{
			{"CastArrayMapStringAnyToAny", "Cast []map[string]any to any", "[]map[string]any", "any"},
			{"CastAnyToInterface", "Cast any to interface{}", "any", "interface {}"},
			{"CastInterfaceToAny", "Cast interface{} to any", "interface {}", "any"},
		},
	}

	primitives := []string{
		"string", "bool", "int8", "int16", "int32", "int64", "int", "uint8", "uint16", "uint32", "uint64", "uint",
		"float32", "float64", "complex64", "complex128", "byte", "rune",
	}
	for _, primitive := range primitives {
		castFuncs.CastAssertFromInterfaces = append(castFuncs.CastAssertFromInterfaces, BasicCastAssertFromInterface("any", primitive))
		castFuncs.CastAssertToInterfaces = append(castFuncs.CastAssertToInterfaces, BasicCastAssertToInterface(primitive, "any"))
	}

	_, thisFile, _, _ := runtime.Caller(0)
	genDir := filepath.Dir(thisFile)
	tmplFile := filepath.Join(genDir, "cast.gotmpl")
	outFile := filepath.Join(genDir, "../../../pkg/externalfunctions/cast.go")

	tmpl := template.Must(
		template.New("").Funcs(template.FuncMap{
			"toLower": strings.ToLower,
			"toUpper": strings.ToUpper,
		}).ParseFiles(tmplFile))

	// execute template w/ data
	var buf bytes.Buffer
	err := tmpl.ExecuteTemplate(&buf, "cast.gotmpl", castFuncs)
	if err != nil {
		panic(fmt.Sprintf("unable to execute template: %v", err))
	}

	// format the generated code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		panic(fmt.Sprintf("unable to format generated code: %v", err))
	}

	// write to file
	err = os.WriteFile(outFile, formatted, 0644)
	if err != nil {
		panic(fmt.Sprintf("unable to write generated code to file: %v", err))
	}
}
