package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/emmaly/perplexity"
)

func main() {
	// Set a timeout for the API call
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	// Retrieve the API token from the environment variable
	token := os.Getenv("PERPLEXITY_API_TOKEN")
	if token == "" {
		log.Fatal("Please set the PERPLEXITY_API_TOKEN environment variable.")
	}

	// Create a new Perplexity client
	client := perplexity.NewClient(token, nil)

	// Prepare the chat completion request
	req := perplexity.ChatCompletionRequest{
		Model: perplexity.ModelLlama31SonarSmall128kOnline,
		Messages: []perplexity.Message{
			{
				Role:    perplexity.MessageRoleSystem,
				Content: "Be precise and concise. Be witty and engaging, with a touch of humor.",
				// Note: The system prompt is processed by the model but not by the online search subsystem.
			},
			{
				Role: perplexity.MessageRoleUser,
				Content: `This user likes fantasy and sci-fi with strong female leads. Use online resources to find books that match the user's interests.

				The following is the user's query:
				Recommend a book about things I'm interested in.`,
			},
		},
		SearchRecencyFilter: perplexity.RecencyFilterMonth,
		Stream: func(delta perplexity.ChatCompletionResponse) {
			// Handle incremental updates by printing the assistant's response as it streams
			for _, choice := range delta.Choices {
				fmt.Print(choice.Delta.Content)
			}
		},
	}

	// Send the chat completion request
	response, err := client.ChatCompletion(ctx, req)
	if err != nil {
		log.Fatalf("Error calling ChatCompletion: %v", err)
	}

	// If streaming is enabled, the response will be nil, and content is handled in the stream callback
	if response == nil {
		fmt.Println("\nStreaming completed.")
		return
	}

	// Handle non-streaming response
	if len(response.Choices) > 0 {
		fmt.Printf("Assistant's reply:\n%s\n", response.Choices[0].Message.Content)
	} else {
		fmt.Println("No choices found in the response.")
	}
}
