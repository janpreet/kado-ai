[![Continuous Integration](https://github.com/janpreet/kado-ai/actions/workflows/ci.yaml/badge.svg)](https://github.com/janpreet/kado-ai/actions/workflows/ci.yaml)[![Publish to pkg.go.dev](https://github.com/janpreet/kado-ai/actions/workflows/publish.yaml/badge.svg)](https://github.com/janpreet/kado-ai/actions/workflows/publish.yaml)
# Kado AI

Kado AI is an AI-assisted infrastructure recommendation engine designed to analyze Terraform configurations, Ansible playbooks, OPA Rego policies, and Terraform plans. It leverages AI services to generate comprehensive infrastructure recommendations based on the provided code and configurations.

This package is an offshoot and integral part of the larger Kado project, which aims to streamline and enhance infrastructure operations through intelligent analysis and automation.

## Table of Contents

1. [Installation](#installation)
2. [Configuration](#configuration)
3. [Usage](#usage)
4. [Security Considerations](#security-considerations)
5. [Development](#development)
6. [Testing](#testing)
7. [Continuous Integration and Deployment](#continuous-integration-and-deployment)
8. [Contributing](#contributing)
9. [License](#license)
10. [Support](#support)
11. [Relationship to Kado](#relationship-to-kado)

## Installation

To use Kado AI in your Go project, run:

```bash
go get github.com/janpreet/kado-ai@latest
```

Ensure you're using Go 1.16 or later.

## Configuration

Kado AI requires a `.kdconfig` file in the user's home directory. This file should contain the following configuration:

```
AI_API_KEY=your_api_key_here
AI_MODEL=your_model_here
AI_CLIENT=your_client_type_here
```

- `AI_API_KEY`: Your API key for the AI service (ChatGPT or Anthropic).
- `AI_MODEL`: The AI model to use (e.g., "gpt-4" for ChatGPT or "claude-3-sonnet-20240229" for Anthropic).
- `AI_CLIENT`: The AI client type ("chatgpt" or "anthropic_messages").

To set up the configuration:

1. Create the `.kdconfig` file in your home directory:
   ```bash
   touch ~/.kdconfig
   ```

2. Add your configuration to the file:
   ```bash
   echo "AI_API_KEY=your_api_key_here" >> ~/.kdconfig
   echo "AI_MODEL=your_model_here" >> ~/.kdconfig
   echo "AI_CLIENT=your_client_type_here" >> ~/.kdconfig
   ```

3. Set appropriate permissions to protect your API key:
   ```bash
   chmod 600 ~/.kdconfig
   ```

## Usage

Here's a basic example of how to use Kado AI in your Go code:

```go
package main

import (
    "fmt"
    "log"

    kadoai "github.com/janpreet/kado-ai/ai"
)

func main() {
    client, err := kadoai.NewAIClient("/path/to/your/iac/code", "")
    if err != nil {
        log.Fatalf("Error creating AI client: %v", err)
    }

    recommendations, err := client.RunAI()
    if err != nil {
        if err.Error() == "operation cancelled by user" {
            fmt.Println("AI analysis cancelled.")
            return
        }
        log.Fatalf("Error running AI: %v", err)
    }

    fmt.Println("Infrastructure Recommendations:")
    fmt.Println(recommendations)
}
```

When you run this code:

1. It will analyze your Infrastructure as Code files.
2. It will save the sanitized input to a file named `ai_input.txt` in your specified IaC directory.
3. You will be prompted to review the input and confirm if you want to proceed with sending the data to the AI.
4. If you confirm, it will send the data to the AI service and return the recommendations.
5. If you cancel, the operation will stop without sending any data to the AI service.

This approach allows you to review the sanitized data before it's sent to the AI, providing an additional layer of security and control.

## Security Considerations

1. **API Key Protection**: Store your API key securely in the `.kdconfig` file and ensure it has restricted permissions (600).

2. **Data Sanitization**: Kado AI sanitizes sensitive information before sending it to the AI service. This includes:
   - Passwords
   - API keys
   - Tokens
   - Private keys
   - IP addresses
   - URLs (domain parts are redacted)

3. **Local Storage**: The sanitized input is saved locally in `ai_input.txt` within your IaC directory. Ensure this file is protected and cleaned up after use.

4. **No Persistent Storage**: The AI service does not store your data, but the interaction is part of the API call. Ensure compliance with your data handling policies.

5. **HTTPS**: All communications with AI services use HTTPS.

6. **Version Control**: Do not commit the `.kdconfig` file or any files containing sensitive information to version control.

7. **User Confirmation**: Before sending any data to the AI service, the user is prompted to review the sanitized input and must explicitly confirm to proceed. This allows for a final check to ensure no sensitive information is being sent unintentionally.

## Development

To set up the development environment:

1. Clone the repository:
   ```bash
   git clone https://github.com/janpreet/kado-ai.git
   cd kado-ai
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Make your changes to the code.

4. Run tests:
   ```bash
   make test
   ```

5. Build the package:
   ```bash
   make build
   ```

## Testing

Run tests using:

```bash
make test
```

This will run all unit tests in the package. Ensure all tests pass before submitting a pull request.

## Continuous Integration and Deployment

This project uses GitHub Actions for CI/CD. The workflows are defined in `.github/workflows/`:

- `ci.yml`: Runs tests and build on every push and pull request to the main branch.
- `publish.yml`: Runs tests, builds the package, and publishes to pkg.go.dev when a new release is created.

To create a new release:

1. Update the version in your code if necessary.
2. Commit and push your changes.
3. Create a new tag:
   ```bash
   make tag VERSION=v0.1.0
   ```
4. Go to the GitHub repository and create a new release based on this tag.

## Contributing

1. Fork the repository.
2. Create a new branch for your feature or bug fix.
3. Make your changes and write tests for them.
4. Run `make test` to ensure all tests pass.
5. Submit a pull request with a clear description of your changes.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Support

For support, please open an issue in the GitHub repository. Provide as much context as possible, including:

- Your Go version
- Your operating system
- The version of Kado AI you're using
- A detailed description of the issue
- Steps to reproduce the issue
- Any relevant code snippets or error messages

Please do not share any sensitive information (like API keys) in your support requests.

## Relationship to Kado

Kado AI is a component of the larger Kado project, which is designed to provide comprehensive infrastructure management and optimization tools. While Kado AI focuses on AI-assisted recommendations for infrastructure configurations, it integrates seamlessly with other Kado components to offer a full-featured solution for modern infrastructure challenges.

For more information about the overall Kado project and how Kado AI fits into the broader ecosystem, please visit the main Kado project repository.
