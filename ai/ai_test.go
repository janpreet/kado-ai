package ai

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewAIClient(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "kado-ai-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test .kdconfig file
	kdconfigPath := filepath.Join(tmpDir, ".kdconfig")
	kdconfigContent := `
AI_API_KEY=test-api-key
AI_MODEL=test-model
AI_CLIENT=test-client
`
	err = os.WriteFile(kdconfigPath, []byte(kdconfigContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write test .kdconfig: %v", err)
	}

	// Create a new AIClient with the test config path
	client, err := NewAIClient("/path/to/iac", kdconfigPath)
	if err != nil {
		t.Fatalf("NewAIClient failed: %v", err)
	}

	// Check if the client was created with the correct values
	if client.apiKey != "test-api-key" {
		t.Errorf("Expected API key 'test-api-key', got '%s'", client.apiKey)
	}
	if client.model != "test-model" {
		t.Errorf("Expected model 'test-model', got '%s'", client.model)
	}
	if client.clientType != "test-client" {
		t.Errorf("Expected client type 'test-client', got '%s'", client.clientType)
	}
	if client.iacPath != "/path/to/iac" {
		t.Errorf("Expected IAC path '/path/to/iac', got '%s'", client.iacPath)
	}
}

func TestSanitizeContent(t *testing.T) {
	client := &AIClient{}
	testCases := []struct {
		input    string
		expected string
	}{
		{"password = 'secret123'", "[REDACTED]"},
		{"aws_access_key = 'AKIAIOSFODNN7EXAMPLE'", "[REDACTED]"},
		{"https://example.com/path", "https://[REDACTED]/path"},
		{"127.0.0.1", "[REDACTED]"},
		{"2001:0db8:85a3:0000:0000:8a2e:0370:7334", "[REDACTED]"},
	}

	for _, tc := range testCases {
		result := client.sanitizeContent(tc.input)
		if result != tc.expected {
			t.Errorf("For input '%s', expected '%s', but got '%s'", tc.input, tc.expected, result)
		}
	}
}