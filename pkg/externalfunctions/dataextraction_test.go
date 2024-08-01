package externalfunctions

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"testing"
)

func TestDataExtractionGetDocumentType(t *testing.T) {
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
		actual := DataExtractionGetDocumentType(test.fileName)
		if actual != test.expected {
			t.Errorf("GetFileExtension(%s): expected %s, actual %s", test.fileName, test.expected, actual)
		}
	}
}

func TestDataExtractionGetLocalFileContent(t *testing.T) {
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
	actualChecksum, actualContent := DataExtractionGetLocalFileContent(tempFile.Name())

	// Check if the actual checksum matches the expected checksum.
	if actualChecksum != expectedChecksum {
		t.Errorf("expected checksum %v, got %v", expectedChecksum, actualChecksum)
	}

	// Check if the actual content matches the expected content.
	if actualContent != content {
		t.Errorf("expected content %v, got %v", content, actualContent)
	}
}
