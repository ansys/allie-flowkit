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

package externalfunctions

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"testing"
)

func TestGetDocumentType(t *testing.T) {
	tests := []struct {
		fileName string
		expected string
	}{
		{"test.txt", "txt"},
		{"test.docx", "docx"},
		{"test.pdf", "pdf"},
		{"test.jpg", "jpg"},
		{"test.jpeg", "jpeg"},
		{"test.png", "png"},
		{"test", ""},
	}

	for _, test := range tests {
		actual := GetDocumentType(test.fileName)
		if actual != test.expected {
			t.Errorf("GetFileExtension(%s): expected %s, actual %s", test.fileName, test.expected, actual)
		}
	}
}

func TestGetLocalFileContent(t *testing.T) {
	// Create a temporary file for testing.
	tempFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Write some content to the temporary file.
	content := "Hello, World!"
	_, err = tempFile.WriteString(content)
	if err != nil {
		t.Fatalf("failed to write to test file: %v", err)
	}
	tempFile.Close()

	// Calculate the expected checksum.
	hash := sha256.New()
	_, err = hash.Write([]byte(content))
	if err != nil {
		t.Fatalf("failed to calculate expected checksum: %v", err)
	}
	expectedChecksum := hex.EncodeToString(hash.Sum(nil))

	// Call the function with the test file.
	actualChecksum, actualContent := GetLocalFileContent(tempFile.Name())

	// Check if the actual checksum matches the expected checksum.
	if actualChecksum != expectedChecksum {
		t.Errorf("expected checksum %v, got %v", expectedChecksum, actualChecksum)
	}

	if !bytes.Equal(actualContent, []byte(content)) {
		t.Errorf("expected content %v, got %v", content, actualContent)
	}
}

func TestAppendStringSlices(t *testing.T) {
	tests := []struct {
		slice1   []string
		slice2   []string
		slice3   []string
		slice4   []string
		slice5   []string
		expected []string
	}{
		{[]string{"a", "b", "c"}, []string{"d", "e", "f"}, []string{}, []string{}, []string{}, []string{"a", "b", "c", "d", "e", "f"}},
		{[]string{"a", "b", "c"}, []string{}, []string{}, []string{}, []string{}, []string{"a", "b", "c"}},
		{[]string{}, []string{"d", "e", "f"}, []string{}, []string{}, []string{}, []string{"d", "e", "f"}},
		{[]string{}, []string{}, []string{}, []string{}, []string{}, []string{}},
	}

	for _, test := range tests {
		actual := AppendStringSlices(test.slice1, test.slice2, test.slice3, test.slice4, test.slice5)
		if len(actual) != len(test.expected) {
			t.Errorf("expected length %d, got %d", len(test.expected), len(actual))
		}
		for i := range actual {
			if actual[i] != test.expected[i] {
				t.Errorf("expected %v, got %v", test.expected, actual)
			}
		}
	}
}
