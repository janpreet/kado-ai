package ai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
)

type AIClient struct {
	apiKey     string
	model      string
	clientType string
	iacPath    string
}

func NewAIClient(iacPath string, configPath string) (*AIClient, error) {
	config, err := loadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	apiKey, apiKeyExists := config["AI_API_KEY"]
	model, modelExists := config["AI_MODEL"]
	clientType, clientTypeExists := config["AI_CLIENT"]

	if !apiKeyExists || !modelExists || !clientTypeExists {
		return nil, fmt.Errorf("AI_API_KEY, AI_MODEL, or AI_CLIENT is not set in config")
	}

	return &AIClient{
		apiKey:     apiKey,
		model:      model,
		clientType: clientType,
		iacPath:    iacPath,
	}, nil
}

func loadConfig(configPath string) (map[string]string, error) {
	if configPath == "" {
		usr, err := user.Current()
		if err != nil {
			return nil, err
		}
		configPath = filepath.Join(usr.HomeDir, ".kdconfig")
	}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			config[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return config, nil
}

func (c *AIClient) RunAI() (string, error) {
	terraformAndRegoCode := c.scanDirectory(filepath.Join(c.iacPath, "terraform"), []string{".tf", ".rego"})
	ansibleAndRegoCode := c.scanDirectory(filepath.Join(c.iacPath, "ansible"), []string{".yml", ".yaml", ".rego"})

	terraformPlanPath := filepath.Join(c.iacPath, "terraform", "plan.json")
	terraformPlan, err := c.extractFileContent(terraformPlanPath)
	if err != nil {
		terraformPlan = "Terraform plan not found"
	} else {
		terraformPlan = c.sanitizeContent(terraformPlan)
	}

	input := fmt.Sprintf(`Please provide comprehensive infrastructure recommendations based on the following:

Terraform Code and OPA Rego Policies:
%s

Ansible Code and OPA Rego Policies:
%s

Terraform Plan:
%s

Consider all aspects including infrastructure provisioning, configuration management, security policies, and best practices.`, 
		c.sanitizeContent(terraformAndRegoCode),
		c.sanitizeContent(ansibleAndRegoCode),
		terraformPlan)

	if err := c.saveAIInput(input); err != nil {
		return "", fmt.Errorf("failed to save AI input: %v", err)
	}

	fmt.Printf("AI input has been saved to %s\n", filepath.Join(c.iacPath, "ai_input.txt"))
	fmt.Print("Do you want to proceed with sending this data to the AI for analysis? (yes/no): ")
	var response string
	fmt.Scanln(&response)
	if strings.ToLower(response) != "yes" {
		return "", fmt.Errorf("operation cancelled by user")
	}

	recommendations, err := c.getRecommendations(input)
	if err != nil {
		return "", fmt.Errorf("failed to get recommendations: %v", err)
	}

	var aiResponse map[string]interface{}
	err = json.Unmarshal([]byte(recommendations), &aiResponse)
	if err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	content, ok := aiResponse["content"].([]interface{})
	if !ok || len(content) == 0 {
		return "", fmt.Errorf("no content found in the response")
	}

	textContent, ok := content[0].(map[string]interface{})["text"].(string)
	if !ok {
		return "", fmt.Errorf("unable to extract text content from the response")
	}

	return textContent, nil
}

func (c *AIClient) getRecommendations(input string) (string, error) {
	var url string
	var requestBody []byte
	var err error

	switch c.clientType {
	case "chatgpt":
		url = "https://api.openai.com/v1/chat/completions"
		requestBody, err = json.Marshal(map[string]interface{}{
			"model":    c.model,
			"messages": []map[string]string{{"role": "user", "content": input}},
		})
	case "anthropic_messages":
		url = "https://api.anthropic.com/v1/messages"
		requestBody, err = json.Marshal(map[string]interface{}{
			"model": c.model,
			"max_tokens": 1024,
			"messages": []map[string]string{
				{"role": "user", "content": input},
			},
		})
	default:
		return "", fmt.Errorf("unsupported AI client: %s", c.clientType)
	}

	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.clientType == "chatgpt" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	} else if c.clientType == "anthropic_messages" {
		req.Header.Set("x-api-key", c.apiKey)
		req.Header.Set("anthropic-version", "2023-06-01")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (c *AIClient) extractFileContent(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c *AIClient) scanDirectory(dir string, extensions []string) string {
	var content strings.Builder
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			for _, ext := range extensions {
				if strings.HasSuffix(info.Name(), ext) {
					fileContent, err := c.extractFileContent(path)
					if err == nil {
						content.WriteString(fmt.Sprintf("File: %s\n%s\n\n", path, fileContent))
					}
					break
				}
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Sprintf("Error scanning directory %s: %v", dir, err)
	}
	return content.String()
}

func (c *AIClient) sanitizeContent(content string) string {
	sensitivePatterns := []string{
		`(?i)(aws_access_key|aws_secret_key|password|token|secret|api_key)(\s*[=:]\s*)['"]?[^\s'",]+['"]?`,
		`(?i)(private_key)(\s*[=:]\s*)['"]?-----BEGIN[^'",]*-----END[^'",]*['"]?`,
		`(?i)(connection_string)(\s*[=:]\s*)['"]?[^\s'",]+['"]?`,
		`(?i)(bearer\s+)['"]?[^\s'",]+['"]?`,
		`(?i)("?\w*password"?\s*[:=]?\s*\{?\s*"?value"?\s*[:=]?\s*)['"]?[^\s'",}]+['"]?`,
		`(?i)("?\w*user"?\s*[:=]?\s*\{?\s*"?value"?\s*[:=]?\s*)['"]?[^\s'",}]+['"]?`,
		`(?i)("?\w*(password|secret|key|token)"?\s*[:=]?\s*["'])[^"']+["']`,
		`(?i)("?\w*(password|secret|key|token)"?\s*[:=]?\s*\{?\s*"?value"?\s*[:=]?\s*)['"]?[^\s'",}]+['"]?`,
		`\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b`,
		`\b(?:(?:[0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|(?:[0-9a-fA-F]{1,4}:){1,7}:|(?:[0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|(?:[0-9a-fA-F]{1,4}:){1,5}(?::[0-9a-fA-F]{1,4}){1,2}|(?:[0-9a-fA-F]{1,4}:){1,4}(?::[0-9a-fA-F]{1,4}){1,3}|(?:[0-9a-fA-F]{1,4}:){1,3}(?::[0-9a-fA-F]{1,4}){1,4}|(?:[0-9a-fA-F]{1,4}:){1,2}(?::[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:(?:(?::[0-9a-fA-F]{1,4}){1,6})|:(?:(?::[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(?::[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(?:ffff(?::0{1,4}){0,1}:){0,1}(?:(?:25[0-5]|(?:2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(?:25[0-5]|(?:2[0-4]|1{0,1}[0-9]){0,1}[0-9])|(?:[0-9a-fA-F]{1,4}:){1,4}:(?:(?:25[0-5]|(?:2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(?:25[0-5]|(?:2[0-4]|1{0,1}[0-9]){0,1}[0-9]))\b`,
	}

	for _, pattern := range sensitivePatterns {
		re := regexp.MustCompile(pattern)
		content = re.ReplaceAllString(content, "[REDACTED]")
	}

	urlPattern := `(https?://)([\w.-]+)(\/?\S*)`
	re := regexp.MustCompile(urlPattern)
	content = re.ReplaceAllString(content, "${1}[REDACTED]${3}")

	return content
}

func (c *AIClient) saveAIInput(input string) error {
	inputFilePath := filepath.Join(c.iacPath, "ai_input.txt")
	err := os.WriteFile(inputFilePath, []byte(input), 0644)
	if err != nil {
		return fmt.Errorf("failed to save AI input to file: %v", err)
	}
	return nil
}