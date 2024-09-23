// Package perplexity provides a client for interacting with the Perplexity AI API.
//
// Example usage:
//
//	package main
//
//	import (
//	    "context"
//	    "fmt"
//	    "log"
//	    "time"
//
//	    "github.com/emmaly/perplexity"
//	)
//
//	func main() {
//	    // Replace with your actual API token
//	    token := "<your_api_token>"
//
//	    client := perplexity.NewClient(token, nil) // Pass nil to use default settings
//
//	    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	    defer cancel()
//
//	    req := perplexity.ChatCompletionRequest{
//	        Model: "llama-3.1-sonar-small-128k-online",
//	        Messages: []perplexity.Message{
//	            {
//	                Role:    perplexity.MessageRoleSystem,
//	                Content: "Be precise and concise.",
//	            },
//	            {
//	                Role:    perplexity.MessageRoleUser,
//	                Content: "How many stars are there in our galaxy?",
//	            },
//	        },
//	        MaxTokens:              100,
//	        Temperature:            0.2,
//	        TopP:                   0.9,
//	        ReturnCitations:        true,
//	        SearchDomainFilter:     []string{"perplexity.ai"},
//	        ReturnImages:           false,
//	        ReturnRelatedQuestions: false,
//	        SearchRecencyFilter:    perplexity.RecencyFilterMonth,
//	        TopK:                   0,
//	        Stream:                 false,
//	        PresencePenalty:        0.0,
//	        FrequencyPenalty:       1.0,
//	    }
//
//	    response, err := client.ChatCompletion(ctx, req)
//	    if err != nil {
//	        log.Fatalf("Error calling ChatCompletion: %v", err)
//	    }
//
//	    if len(response.Choices) > 0 {
//	        fmt.Printf("Assistant's reply: %s\n", response.Choices[0].Message.Content)
//	    } else {
//	        fmt.Println("No choices found in the response.")
//	    }
//	}
package perplexity
