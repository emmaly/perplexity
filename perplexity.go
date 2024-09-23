package perplexity

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// Client is a client for the Perplexity AI API.
type Client struct {
	token   string
	client  *http.Client
	baseURL string
}

// DefaultBaseURL is the default base URL for the Perplexity API.
const DefaultBaseURL = "https://api.perplexity.ai"

// ClientOptions represents options for configuring a new Perplexity API client.
type ClientOptions struct {
	// HTTPClient is an optional *http.Client to use for requests.
	HTTPClient *http.Client

	// BaseURL is the base URL for the Perplexity API.
	// If empty, `DefaultBaseURL` is used.
	BaseURL string
}

// MessageRole represents the role of the speaker in a message.
type MessageRole string

const (
	// MessageRoleSystem represents a system message.
	MessageRoleSystem MessageRole = "system"
	// MessageRoleUser represents a user message.
	MessageRoleUser MessageRole = "user"
	// MessageRoleAssistant represents an assistant message.
	MessageRoleAssistant MessageRole = "assistant"
)

// Message represents a message in the conversation.
type Message struct {
	// Role of the speaker in this turn of conversation.
	// Allowed values are `MessageRoleSystem`, `MessageRoleUser`, or `MessageRoleAssistant`.
	Role MessageRole `json:"role"`

	// Content is the contents of the message in this turn of conversation.
	Content string `json:"content"`
}

// RecencyFilter represents a recency filter for search results.
type RecencyFilter string

const (
	// RecencyFilterHour represents a recency filter for the last hour.
	RecencyFilterHour RecencyFilter = "hour"
	// RecencyFilterDay represents a recency filter for the last day.
	RecencyFilterDay RecencyFilter = "day"
	// RecencyFilterWeek represents a recency filter for the last week.
	RecencyFilterWeek RecencyFilter = "week"
	// RecencyFilterMonth represents a recency filter for the last month.
	RecencyFilterMonth RecencyFilter = "month"
)

// Model represents a model that can complete prompts.
type Model string

const (
	// ModelLlama31SonarSmall128kOnline represents the Llama 3.1 Sonar Small 128k Online model.
	ModelLlama31SonarSmall128kOnline Model = "llama-3.1-sonar-small-128k-online"
	// ModelLlama31SonarLarge128kOnline represents the Llama 3.1 Sonar Large 128k Online model.
	ModelLlama31SonarLarge128kOnline Model = "llama-3.1-sonar-large-128k-online"
	// ModelLlama31SonarHuge128kOnline represents the Llama 3.1 Sonar Huge 128k Online model.
	ModelLlama31SonarHuge128kOnline Model = "llama-3.1-sonar-huge-128k-online"

	// ModelLlama31SonarSmall128kChat represents the Llama 3.1 Sonar Small 128k Chat model.
	ModelLlama31SonarSmall128kChat Model = "llama-3.1-sonar-small-128k-chat"
	// ModelLlama31SonarLarge128kChat represents the Llama 3.1 Sonar Large 128k Chat model.
	ModelLlama31SonarLarge128kChat Model = "llama-3.1-sonar-large-128k-chat"

	// ModelLlama31Instruct8b represents the Llama 3.1 Instruct 8b model.
	ModelLlama31Instruct8b Model = "llama-3.1-8b-instruct"
	// ModelLlama31Instruct70b represents the Llama 3.1 Instruct 70b model.
	ModelLlama31Instruct70b Model = "llama-3.1-70b-instruct"
)

// ChatCompletionRequest represents a request for a chat completion.
type ChatCompletionRequest struct {
	// Model is the name of the model that will complete your prompt.
	// Refer to Supported Models to find all the models offered.
	// *This field is required.*
	Model Model `json:"model"`

	// Messages is a list of messages comprising the conversation so far.
	// *This field is required.*
	//
	// ***Warning:** "We do not raise any exceptions if your chat inputs contain messages with special
	// tokens. If avoiding prompt injections is a concern for your use case, it is recommended that
	// you check for special tokens prior to calling the API. For more details, read
	// (Meta’s recommendations for Llama)[https://github.com/meta-llama/llama/blob/008385a/UPDATES.md#token-sanitization-update]."*
	Messages []Message `json:"messages"`

	// MaxTokens is the maximum number of completion tokens returned by the API.
	// If left unspecified, the model will generate tokens until it reaches a stop token
	// or the end of its context window.
	MaxTokens int `json:"max_tokens,omitempty"`

	// Temperature controls the amount of randomness in the response,
	// valued between 0 (inclusive) and 2 (exclusive).
	// Higher values result in more random outputs, while lower values are more deterministic.
	// Default: `0.2`
	Temperature float64 `json:"temperature,omitempty"`

	// TopP is the nucleus sampling threshold, valued between 0 and 1 (inclusive).
	// For each token, the model considers the results of the tokens with TopP probability mass.
	// It's recommended to adjust either TopK or TopP, but not both.
	// Default: `0.9`
	TopP float64 `json:"top_p,omitempty"`

	// ReturnCitations determines whether the response should include citations.
	// Default: `false`
	// *This feature is only available via closed beta access.*
	ReturnCitations bool `json:"return_citations"`

	// SearchDomainFilter limits the citations used by the online model to URLs from the specified domains.
	// Currently limited to only 3 domains for allowlisting and blocklisting.
	// For blocklisting, add a "-" to the beginning of the domain string.
	SearchDomainFilter []string `json:"search_domain_filter"`

	// ReturnImages determines whether the response should include images.
	// Default: `false`
	// *This feature is only available via closed beta access.*
	ReturnImages bool `json:"return_images"`

	// ReturnRelatedQuestions determines whether the response should include related questions.
	// Default: `false`
	// *This feature is only available via closed beta access.*
	ReturnRelatedQuestions bool `json:"return_related_questions"`

	// SearchRecencyFilter restricts search results to within the specified time interval.
	// Does not apply to images.
	// Valid values are `RecencyFilterMonth`, `RecencyFilterWeek`, `RecencyFilterDay`, `RecencyFilterHour`.
	SearchRecencyFilter RecencyFilter `json:"search_recency_filter,omitempty"`

	// TopK is the number of tokens to keep for highest top-k filtering,
	// specified as an integer between 0 and 2048 (inclusive).
	// If set to 0, top-k filtering is disabled.
	// It's recommended to adjust either TopK or TopP, but not both.
	// Default: `0`
	TopK int `json:"top_k,omitempty"`

	// Stream is a callback function that is called when new tokens are generated.
	// If provided, the response will be streamed incrementally as the model generates new tokens.
	// Default: `nil`
	Stream OnUpdateHandler `json:"-"` // this will be set as a bool in MarshalJSON for the API

	// PresencePenalty is a value between -2.0 and 2.0.
	// Positive values penalize new tokens based on whether they appear in the text so far,
	// increasing the model's likelihood to discuss new topics.
	// Incompatible with FrequencyPenalty.
	// Default: `0.0`
	PresencePenalty float64 `json:"presence_penalty,omitempty"`

	// FrequencyPenalty is a multiplicative penalty greater than 0.
	// Values greater than 1.0 penalize new tokens based on their existing frequency in the text so far,
	// decreasing the model's likelihood to repeat the same line verbatim.
	// A value of 1.0 means no penalty.
	// Incompatible with PresencePenalty.
	// Default: `1.0`
	FrequencyPenalty float64 `json:"frequency_penalty,omitempty"`
}

// OnUpdateHandler is a callback function that is called when new tokens are generated.
type OnUpdateHandler func(delta ChatCompletionResponse)

// ChatCompletionResponse represents a response from the chat completion API.
type ChatCompletionResponse struct {
	// ID is an ID generated uniquely for each response.
	ID string `json:"id"`

	// Model is the model used to generate the response.
	Model string `json:"model"`

	// Object is the object type, which always equals `chat.completion`.
	Object string `json:"object"`

	// Created is the Unix timestamp (in seconds) of when the completion was created.
	Created int64 `json:"created"`

	// Choices is the list of completion choices the model generated for the input prompt.
	Choices []Choice `json:"choices"`

	// Usage contains usage statistics for the completion request.
	Usage Usage `json:"usage"`
}

// Choice represents a single completion choice generated by the model.
type Choice struct {
	// Index is the index of this completion in the list.
	Index int `json:"index"`

	// FinishReason is the reason the model stopped generating tokens.
	// Possible values include `FinishReasonStop` if the model hit a natural stopping point,
	// or `FinishReasonLength` if the maximum number of tokens specified in the request was reached.
	FinishReason FinishReason `json:"finish_reason"`

	// Message is the message generated by the model.
	Message Message `json:"message"`

	// Delta is the incrementally streamed next tokens.
	// Only meaningful when Stream is `true`.
	Delta Message `json:"delta"`
}

// Usage contains usage statistics for the completion request.
type Usage struct {
	// PromptTokens is the number of tokens provided in the request prompt.
	PromptTokens int `json:"prompt_tokens"`

	// CompletionTokens is the number of tokens generated in the response output.
	CompletionTokens int `json:"completion_tokens"`

	// TotalTokens is the total number of tokens used in the chat completion (prompt + completion).
	TotalTokens int `json:"total_tokens"`
}

// FinishReason represents the reason the model stopped generating tokens.
type FinishReason string

const (
	FinishReasonStop   FinishReason = "stop"
	FinishReasonLength FinishReason = "length"
)

type apiError struct {
	Error string `json:"error"`
}

// NewClient creates a new Perplexity API client with the given token.
// Optionally, you can pass a custom *http.Client to override default settings.
// If httpClient is nil, a default client with reasonable timeouts is used.
func NewClient(token string, options *ClientOptions) *Client {
	var httpClient *http.Client
	if options != nil {
		httpClient = options.HTTPClient
	}

	if httpClient == nil {
		// Set up a default Transport with reasonable timeouts.
		transport := &http.Transport{
			// Dialer with connection timeout
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second, // Timeout for establishing a connection
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   10 * time.Second,  // Timeout for TLS handshake
			ResponseHeaderTimeout: 300 * time.Second, // Timeout for reading response headers
			ExpectContinueTimeout: 1 * time.Second,   // Timeout for expect continue
		}

		httpClient = &http.Client{
			Transport: transport,
			// Do not set Timeout here; rely on context for per-request timeouts.
		}
	}

	baseURL := DefaultBaseURL
	if options != nil && options.BaseURL != "" {
		baseURL = options.BaseURL
	}

	return &Client{
		token:   token,
		client:  httpClient,
		baseURL: baseURL,
	}
}

// MarshalJSON marshals a ChatCompletionRequest into JSON.
func (req *ChatCompletionRequest) MarshalJSON() ([]byte, error) {
	type Alias ChatCompletionRequest
	return json.Marshal(&struct {
		*Alias
		Stream bool `json:"stream,omitempty"`
	}{
		Alias:  (*Alias)(req),
		Stream: req.Stream != nil,
	})
}

// ChatCompletion sends a chat completion request to the Perplexity AI API.
func (c *Client) ChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	url := c.baseURL + "/chat/completions"

	// Validate the request
	if req.Model == "" {
		return nil, errors.New("model is required")
	}
	if len(req.Messages) == 0 {
		return nil, errors.New("at least one message is required")
	}
	if req.Messages[len(req.Messages)-1].Role != MessageRoleUser {
		return nil, errors.New("the last message must be from the user")
	}
	if req.PresencePenalty != 0.0 && req.FrequencyPenalty != 0.0 {
		return nil, errors.New("PresencePenalty and FrequencyPenalty are incompatible; only one should be set")
	}

	// Marshal the payload to JSON
	jsonData, err := json.Marshal(&req) // if this is not a pointer, it will not use the custom MarshalJSON
	if err != nil {
		return nil, err
	}

	// Create the HTTP request with the provided context
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}

	// Set headers
	httpReq.Header.Set("Authorization", "Bearer "+c.token)
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute the HTTP request
	res, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// Check for HTTP errors
	if res.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(res.Body)
		var e apiError
		if err := json.Unmarshal(bodyBytes, &e); err == nil && e.Error != "" {
			return nil, fmt.Errorf("API error: %s", e.Error)
		}
		return nil, fmt.Errorf("unexpected status code: %s", res.Status)
	}

	// Check if the response is a Server-Sent Events stream
	contentType := res.Header.Get("Content-Type")
	if contentType == "text/event-stream" || strings.HasPrefix(contentType, "text/event-stream;") {
		if req.Stream == nil {
			return nil, errors.New("streaming response received but no stream handler provided")
		}
		return nil, c.handleStreamingResponse(res, req.Stream)
	}

	// Read and unmarshal the response body
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var response ChatCompletionResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// handleStreamingResponse handles streaming responses from the Perplexity AI API.
// It reads the Server-Sent Events (SSE) from the response and constructs the final ChatCompletionResponse.
func (c *Client) handleStreamingResponse(res *http.Response, onUpdate func(delta ChatCompletionResponse)) error {
	defer res.Body.Close()

	// onUpdate must be set
	if onUpdate == nil {
		return errors.New("onUpdate handler is required for streaming responses")
	}

	// Create a scanner to read the response line by line
	scanner := bufio.NewScanner(res.Body)

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if line == "" {
			continue
		}

		// Handle the end of the stream
		if line == "data: [DONE]" {
			break
		}

		// Check if the line starts with "data: "
		if len(line) >= 6 && line[:6] == "data: " {
			// Extract the JSON part
			jsonData := line[6:]

			// Parse the JSON data into an response structure
			var response ChatCompletionResponse
			err := json.Unmarshal([]byte(jsonData), &response)
			if err != nil {
				return fmt.Errorf("failed to unmarshal streaming event: %w", err)
			}

			// Call the onUpdate handler
			onUpdate(response)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading streaming response: %w", err)
	}

	return nil
}
