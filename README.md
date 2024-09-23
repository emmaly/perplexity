# Perplexity Go Client

[![Go Reference](https://pkg.go.dev/badge/github.com/emmaly/perplexity.svg)](https://pkg.go.dev/github.com/emmaly/perplexity)

A Go client library for interacting with the [Perplexity AI API](https://docs.perplexity.ai/api-reference/chat-completions).

## Installation

To install the package, run:

```bash
go get github.com/emmaly/perplexity
```

## Features

- **Chat Completions**: Generate responses using various models provided by Perplexity AI.
- **Streaming Responses**: Support for handling streaming responses with callback functions.
- **Customization**: Configure request parameters like temperature, max tokens, top_p, and more.
- **Error Handling**: Comprehensive error handling for API responses and network issues.
- **Fully Typed**: Typed API requests and responses for a better development experience.
- **No Dependencies**: Pure Go implementation with no dependencies outside of the standard library.

## Usage

Here's a basic example of how to use the client:

```go
package main

import (
    "context"
    "fmt"

    "github.com/emmaly/perplexity"
)

func main() {
    // Replace with your actual API token
    token := "your_api_token_here"

    // Create a new client
    client := perplexity.NewClient(token, nil)

    // Prepare the chat completion request
    req := perplexity.ChatCompletionRequest{
        Model: perplexity.ModelLlama31SonarLarge128kChat,
        Messages: []perplexity.Message{
            {
                Role:    perplexity.MessageRoleUser,
                Content: "Hello, how are you?",
            },
        },
        MaxTokens:   150,
        Temperature: 0.7,
    }

    // Send the chat completion request
    resp, err := client.ChatCompletion(context.Background(), req)
    if err != nil {
        panic(err)
    }

    // Print the assistant's reply
    fmt.Println(resp.Choices[0].Message.Content)
}
```

### Streaming Responses

To handle streaming responses, provide a callback function to the `Stream` field:

```go
req := perplexity.ChatCompletionRequest{
    Model: perplexity.ModelLlama31SonarLarge128kChat,
    Messages: []perplexity.Message{
        {
            Role:    perplexity.MessageRoleUser,
            Content: "Tell me a joke.",
        },
    },
    Stream: func(delta perplexity.ChatCompletionResponse) {
        // Handle incremental updates
        fmt.Print(delta.Choices[0].Delta.Content)
    },
}

// The response will be nil when streaming
_, err := client.ChatCompletion(context.Background(), req)
if err != nil {
    panic(err)
}
```

## Available Models

The client supports several models:

- `ModelLlama31SonarSmall128kOnline`
- `ModelLlama31SonarLarge128kOnline`
- `ModelLlama31SonarHuge128kOnline`
- `ModelLlama31SonarSmall128kChat`
- `ModelLlama31SonarLarge128kChat`
- `ModelLlama31Instruct8b`
- `ModelLlama31Instruct70b`

## Request Parameters

- `Model`: **Required.** The name of the model to use.
- `Messages`: **Required.** Conversation history as a list of messages.
- `MaxTokens`: Maximum number of tokens to generate.
- `Temperature`: Controls randomness in the output (`0` to `2`).
- `TopP`: Nucleus sampling threshold (`0` to `1`).
- `ReturnCitations`: Include citations in the response (requires beta access).
- `ReturnImages`: Include images in the response (requires beta access).
- `ReturnRelatedQuestions`: Include related questions (requires beta access).
- `SearchRecencyFilter`: Limit search results to recent content (`hour`, `day`, `week`, `month`).
- `TopK`: Number of highest probability tokens to keep for top-k filtering.
- `PresencePenalty`: Penalize new tokens based on their presence in the text so far (`-2.0` to `2.0`).
- `FrequencyPenalty`: Penalize new tokens based on their frequency in the text so far (>= `0`).

## Error Handling

Errors returned by the API are wrapped in Go error types. Check for errors when making API calls:

```go
resp, err := client.ChatCompletion(context.Background(), req)
if err != nil {
    // Handle error
    fmt.Println("Error:", err)
    return
}
```

## Documentation

For more detailed information, refer to the [GoDoc documentation](https://pkg.go.dev/github.com/emmaly/perplexity).

## Contributing

Contributions are welcome! Please open an issue or submit a pull request on GitHub.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

**Note:** Replace `"your_api_token_here"` with your actual Perplexity AI API token. Ensure you handle sensitive information like API tokens securely and avoid hardcoding them in your source code.
