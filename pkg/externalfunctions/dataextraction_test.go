package externalfunctions

import (
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

	// Check if the actual content matches the expected content.
	if actualContent != content {
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
