package gemini

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gemini-cli-go/internal/auth"
	"gemini-cli-go/internal/constants"
	"gemini-cli-go/internal/models"
	"gemini-cli-go/internal/types"
	"gemini-cli-go/internal/utils"
)

// Client handles communication with Google's Gemini API through the Code Assist endpoint
type Client struct {
	authManager *auth.AuthManager
	config      *types.Environment
	projectID   string
	httpClient  *http.Client
}

// NewClient creates a new Gemini API client
func NewClient(config *types.Environment, authManager *auth.AuthManager) *Client {
	return &Client{
		authManager: authManager,
		config:      config,
		httpClient: &http.Client{
			Timeout: time.Duration(config.RequestTimeout) * time.Second,
		},
	}
}

// DiscoverProjectID discovers the Google Cloud project ID
func (c *Client) DiscoverProjectID() (string, error) {
	if c.config.GeminiProjectID != "" {
		c.projectID = c.config.GeminiProjectID
		return c.projectID, nil
	}

	if c.projectID != "" {
		return c.projectID, nil
	}

	// Try to discover project ID using the Code Assist API
	response, err := c.authManager.CallEndpoint("loadCodeAssist", map[string]interface{}{
		"cloudaicompanionProject": "default-project",
		"metadata": map[string]interface{}{
			"duetProject": "default-project",
		},
	})

	if err != nil {
		return "", fmt.Errorf("%s: %w", constants.ErrProjectDiscoveryFailed, err)
	}

	// Parse the response to extract project ID
	if responseMap, ok := response.(map[string]interface{}); ok {
		if projectID, exists := responseMap["cloudaicompanionProject"].(string); exists {
			c.projectID = projectID
			return projectID, nil
		}
	}

	return "", fmt.Errorf(constants.ErrProjectDiscoveryFailed)
}

// StreamContent streams content from Gemini API
func (c *Client) StreamContent(ctx context.Context, modelID string, systemPrompt string, messages []types.ChatMessage, options *StreamOptions) (<-chan types.StreamChunk, error) {
	if err := c.authManager.InitializeAuth(); err != nil {
		return nil, err
	}

	projectID, err := c.DiscoverProjectID()
	if err != nil {
		return nil, err
	}

	// Extract system prompt and convert messages
	systemPrompt, otherMessages := c.extractSystemPrompt(messages)
	contents, err := c.convertMessagesToGeminiFormat(otherMessages)
	if err != nil {
		return nil, err
	}

	// Add system prompt if provided
	if systemPrompt != "" {
		systemContent := types.GeminiFormattedMessage{
			Role: "user",
			Parts: []types.GeminiPart{
				{Text: systemPrompt},
			},
		}
		contents = append([]types.GeminiFormattedMessage{systemContent}, contents...)
	}

	// Create generation config
	generationConfig := c.createGenerationConfig(modelID, options)

	// Create stream request
	streamRequest := map[string]interface{}{
		"model":   modelID,
		"project": projectID,
		"request": map[string]interface{}{
			"contents":         contents,
			"generationConfig": generationConfig,
		},
	}

	// Create channel for streaming chunks
	chunkChan := make(chan types.StreamChunk, 100)

	go func() {
		defer close(chunkChan)

		// Handle thinking mode
		if options != nil && options.EnableFakeThinking && models.SupportsThinking(modelID) {
			if err := c.generateFakeThinking(ctx, chunkChan, messages, options.StreamThinkingAsContent); err != nil {
				chunkChan <- types.StreamChunk{
					Type: types.StreamChunkTypeText,
					Data: fmt.Sprintf("Error generating thinking: %v", err),
				}
			}
		}

		// Perform stream request
		if err := c.performStreamRequest(ctx, chunkChan, streamRequest, options); err != nil {
			chunkChan <- types.StreamChunk{
				Type: types.StreamChunkTypeText,
				Data: fmt.Sprintf("Error: %v", err),
			}
		}
	}()

	return chunkChan, nil
}

// GetCompletion gets a complete response from Gemini API (non-streaming)
func (c *Client) GetCompletion(ctx context.Context, modelID string, messages []types.ChatMessage, options *StreamOptions) (*CompletionResult, error) {
	// Note: system prompt is handled by StreamContent
	chunkChan, err := c.StreamContent(ctx, modelID, "", messages, options)
	if err != nil {
		return nil, err
	}

	var content strings.Builder
	var usage *types.UsageData

	for chunk := range chunkChan {
		switch chunk.Type {
		case types.StreamChunkTypeText:
			if text, ok := chunk.Data.(string); ok {
				content.WriteString(text)
			}
		case types.StreamChunkTypeUsage:
			if usageData, ok := chunk.Data.(types.UsageData); ok {
				usage = &usageData
			}
		}
	}

	return &CompletionResult{
		Content: content.String(),
		Usage:   usage,
	}, nil
}

// convertMessagesToGeminiFormat converts OpenAI messages to Gemini format
func (c *Client) convertMessagesToGeminiFormat(messages []types.ChatMessage) ([]types.GeminiFormattedMessage, error) {
	var contents []types.GeminiFormattedMessage

	for _, msg := range messages {
		geminiMsg, err := c.convertMessageToGeminiFormat(msg)
		if err != nil {
			return nil, err
		}
		contents = append(contents, *geminiMsg)
	}

	return contents, nil
}

// convertMessageToGeminiFormat converts a single message to Gemini format
func (c *Client) convertMessageToGeminiFormat(msg types.ChatMessage) (*types.GeminiFormattedMessage, error) {
	role := msg.Role
	if role == "assistant" {
		role = "model"
	}

	// Handle string content
	if text, ok := msg.Content.(string); ok {
		return &types.GeminiFormattedMessage{
			Role: role,
			Parts: []types.GeminiPart{
				{Text: text},
			},
		}, nil
	}

	// Handle array content (multimodal)
	if contentArray, ok := msg.Content.([]interface{}); ok {
		var parts []types.GeminiPart

		for _, content := range contentArray {
			if contentMap, ok := content.(map[string]interface{}); ok {
				part, err := c.convertContentToGeminiPart(contentMap)
				if err != nil {
					return nil, err
				}
				if part != nil {
					parts = append(parts, *part)
				}
			}
		}

		return &types.GeminiFormattedMessage{
			Role:  role,
			Parts: parts,
		}, nil
	}

	// Fallback: convert to string
	return &types.GeminiFormattedMessage{
		Role: role,
		Parts: []types.GeminiPart{
			{Text: fmt.Sprintf("%v", msg.Content)},
		},
	}, nil
}

// convertContentToGeminiPart converts content to Gemini part
func (c *Client) convertContentToGeminiPart(content map[string]interface{}) (*types.GeminiPart, error) {
	contentType, ok := content["type"].(string)
	if !ok {
		return nil, fmt.Errorf("content type is required")
	}

	switch contentType {
	case "text":
		if text, ok := content["text"].(string); ok {
			return &types.GeminiPart{Text: text}, nil
		}

	case "image_url":
		if imageURL, ok := content["image_url"].(map[string]interface{}); ok {
			if url, ok := imageURL["url"].(string); ok {
				return c.convertImageURLToGeminiPart(url)
			}
		}
	}

	return nil, nil
}

// convertImageURLToGeminiPart converts image URL to Gemini part
func (c *Client) convertImageURLToGeminiPart(imageURL string) (*types.GeminiPart, error) {
	validation := utils.ValidateImageURL(imageURL)
	if !validation.IsValid {
		return nil, fmt.Errorf("invalid image: %s", validation.Error)
	}

	if strings.HasPrefix(imageURL, "data:") {
		// Handle base64 encoded images
		mimeType, base64Data, err := utils.ExtractBase64Data(imageURL)
		if err != nil {
			return nil, err
		}

		return &types.GeminiPart{
			InlineData: &struct {
				MimeType string `json:"mimeType"`
				Data     string `json:"data"`
			}{
				MimeType: mimeType,
				Data:     base64Data,
			},
		}, nil
	}

	// Handle URL images
	return &types.GeminiPart{
		FileData: &struct {
			MimeType string `json:"mimeType"`
			FileURI  string `json:"fileUri"`
		}{
			MimeType: validation.MimeType,
			FileURI:  imageURL,
		},
	}, nil
}

// createGenerationConfig creates generation configuration for the request
func (c *Client) createGenerationConfig(modelID string, options *StreamOptions) map[string]interface{} {
	config := map[string]interface{}{
		"temperature": constants.DefaultTemperature,
	}

	if options != nil {
		if options.Temperature != nil {
			config["temperature"] = *options.Temperature
		}

		if options.MaxTokens != nil {
			config["maxOutputTokens"] = *options.MaxTokens
		}

		if options.TopP != nil {
			config["topP"] = *options.TopP
		}

		// Handle thinking configuration - use correct format for Gemini API
		if models.SupportsThinking(modelID) {
			thinkingBudget := constants.DefaultThinkingBudget
			if options.ThinkingBudget != nil {
				thinkingBudget = *options.ThinkingBudget
			}
			
			// Validate thinking budget (can't be 0 for thinking models)
			if thinkingBudget == 0 {
				thinkingBudget = constants.DefaultThinkingBudget
			}

			if options.EnableRealThinking {
				// Enable thinking with proper structure
				config["thinkingConfig"] = map[string]interface{}{
					"thinkingBudget":  thinkingBudget,
					"includeThoughts": true,
				}
			} else {
				// Disable thinking visibility but still provide valid budget
				config["thinkingConfig"] = map[string]interface{}{
					"thinkingBudget":  thinkingBudget,
					"includeThoughts": false,
				}
			}
		}
		// Don't add thinkingConfig for non-thinking models
	}

	return config
}

// performStreamRequest performs the actual stream request
func (c *Client) performStreamRequest(ctx context.Context, chunkChan chan<- types.StreamChunk, streamRequest map[string]interface{}, options *StreamOptions) error {
	return c.performStreamRequestWithRetry(ctx, chunkChan, streamRequest, options, false)
}

// performStreamRequestWithRetry performs stream request with retry logic
func (c *Client) performStreamRequestWithRetry(ctx context.Context, chunkChan chan<- types.StreamChunk, streamRequest map[string]interface{}, options *StreamOptions, isRetry bool) error {
	bodyBytes, err := json.Marshal(streamRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal stream request: %w", err)
	}

	url := fmt.Sprintf("%s/%s:streamGenerateContent?alt=sse", constants.CodeAssistEndpoint, constants.CodeAssistAPIVersion)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", constants.ContentTypeJSON)
	req.Header.Set("Authorization", constants.BearerPrefix+c.authManager.GetAccessToken())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized && !isRetry {
		c.authManager.ClearTokenCache()
		if err := c.authManager.InitializeAuth(); err != nil {
			return err
		}
		return c.performStreamRequestWithRetry(ctx, chunkChan, streamRequest, options, true)
	}

	if resp.StatusCode != http.StatusOK {
		// Read the error response body for debugging
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("stream request failed with status %d (could not read error body: %v)", resp.StatusCode, err)
		}
		return fmt.Errorf("stream request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return c.parseSSEStream(ctx, chunkChan, resp.Body, options)
}

// parseSSEStream parses the Server-Sent Events stream
func (c *Client) parseSSEStream(ctx context.Context, chunkChan chan<- types.StreamChunk, body io.ReadCloser, options *StreamOptions) error {
	scanner := bufio.NewScanner(body)
	var buffer strings.Builder
	hasClosedThinking := false
	hasStartedThinking := false

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Text()
		
		if line == "" {
			if buffer.Len() > 0 {
				if err := c.processSSEData(chunkChan, buffer.String(), options, &hasStartedThinking, &hasClosedThinking); err != nil {
					return err
				}
				buffer.Reset()
			}
			continue
		}

		if strings.HasPrefix(line, "data: ") {
			buffer.WriteString(line[6:])
		}
	}

	return scanner.Err()
}

// processSSEData processes SSE data and sends chunks
func (c *Client) processSSEData(chunkChan chan<- types.StreamChunk, data string, options *StreamOptions, hasStartedThinking *bool, hasClosedThinking *bool) error {
	if data == "[DONE]" {
		return nil
	}

	var geminiResp types.GeminiResponse
	if err := json.Unmarshal([]byte(data), &geminiResp); err != nil {
		return fmt.Errorf("failed to parse SSE data: %w", err)
	}

	if geminiResp.Response == nil {
		return nil
	}

	// Process candidates
	for _, candidate := range geminiResp.Response.Candidates {
		if candidate.Content != nil {
			for _, part := range candidate.Content.Parts {
				if err := c.processPart(chunkChan, part, options, hasStartedThinking, hasClosedThinking); err != nil {
					return err
				}
			}
		}
	}

	// Process usage metadata
	if geminiResp.Response.UsageMetadata != nil {
		usage := types.UsageData{
			InputTokens:  geminiResp.Response.UsageMetadata.PromptTokenCount,
			OutputTokens: geminiResp.Response.UsageMetadata.CandidatesTokenCount,
		}
		chunkChan <- types.StreamChunk{
			Type: types.StreamChunkTypeUsage,
			Data: usage,
		}
	}

	return nil
}

// processPart processes a Gemini part and sends appropriate chunks
func (c *Client) processPart(chunkChan chan<- types.StreamChunk, part types.GeminiPart, options *StreamOptions, hasStartedThinking *bool, hasClosedThinking *bool) error {
	// Handle thinking content
	if part.Thought && part.Text != "" {
		if options != nil && options.StreamThinkingAsContent {
			if !*hasStartedThinking {
				chunkChan <- types.StreamChunk{
					Type: types.StreamChunkTypeThinkingContent,
					Data: constants.ThinkingOpenTag,
				}
				*hasStartedThinking = true
			}
			chunkChan <- types.StreamChunk{
				Type: types.StreamChunkTypeThinkingContent,
				Data: part.Text,
			}
		} else {
			chunkChan <- types.StreamChunk{
				Type: types.StreamChunkTypeRealThinking,
				Data: part.Text,
			}
		}
		return nil
	}

	// Handle regular text content
	if part.Text != "" && !part.Thought {
		// Close thinking tag if needed
		if *hasStartedThinking && !*hasClosedThinking {
			chunkChan <- types.StreamChunk{
				Type: types.StreamChunkTypeThinkingContent,
				Data: constants.ThinkingCloseTag,
			}
			*hasClosedThinking = true
		}

		chunkChan <- types.StreamChunk{
			Type: types.StreamChunkTypeText,
			Data: part.Text,
		}
	}

	return nil
}

// generateFakeThinking generates fake thinking output
func (c *Client) generateFakeThinking(ctx context.Context, chunkChan chan<- types.StreamChunk, messages []types.ChatMessage, streamAsContent bool) error {
	// Get the last user message
	var lastUserMessage string
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			if text, ok := messages[i].Content.(string); ok {
				lastUserMessage = text
				break
			}
		}
	}

	// Generate preview
	requestPreview := lastUserMessage
	if len(requestPreview) > 100 {
		requestPreview = requestPreview[:100] + "..."
	}

	if streamAsContent {
		// Stream as content with thinking tags
		chunkChan <- types.StreamChunk{
			Type: types.StreamChunkTypeThinkingContent,
			Data: constants.ThinkingOpenTag,
		}

		time.Sleep(constants.ReasoningChunkDelay)

		for _, template := range constants.ReasoningMessages {
			message := strings.ReplaceAll(template, "{requestPreview}", requestPreview)
			
			// Split into chunks
			chunks := c.splitIntoChunks(message, constants.ThinkingContentChunkSize)
			for _, chunk := range chunks {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				chunkChan <- types.StreamChunk{
					Type: types.StreamChunkTypeThinkingContent,
					Data: chunk,
				}
				time.Sleep(constants.StreamingChunkDelay)
			}
		}
	} else {
		// Stream as reasoning field
		for _, template := range constants.ReasoningMessages {
			message := strings.ReplaceAll(template, "{requestPreview}", requestPreview)

			// Split into smaller chunks for more realistic streaming
			chunks := c.splitIntoChunks(message, constants.ThinkingContentChunkSize)
			for _, chunk := range chunks {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				chunkChan <- types.StreamChunk{
					Type: types.StreamChunkTypeReasoning,
					Data: types.ReasoningData{Reasoning: chunk},
				}
				time.Sleep(constants.StreamingChunkDelay)
			}
		}
	}

	return nil
}

// splitIntoChunks splits text into chunks of specified size
func (c *Client) splitIntoChunks(text string, chunkSize int) []string {
	if len(text) <= chunkSize {
		return []string{text}
	}

	var chunks []string
	for i := 0; i < len(text); i += chunkSize {
		end := i + chunkSize
		if end > len(text) {
			end = len(text)
		}
		chunks = append(chunks, text[i:end])
	}

	return chunks
}

// extractSystemPrompt extracts system prompt from messages
func (c *Client) extractSystemPrompt(messages []types.ChatMessage) (string, []types.ChatMessage) {
	var systemPrompt string
	var otherMessages []types.ChatMessage

	for _, msg := range messages {
		if msg.Role == "system" {
			if text, ok := msg.Content.(string); ok {
				systemPrompt = text
			} else if contentArray, ok := msg.Content.([]interface{}); ok {
				// Extract text content from array
				var textParts []string
				for _, content := range contentArray {
					if contentMap, ok := content.(map[string]interface{}); ok {
						if contentType, ok := contentMap["type"].(string); ok && contentType == "text" {
							if text, ok := contentMap["text"].(string); ok {
								textParts = append(textParts, text)
							}
						}
					}
				}
				systemPrompt = strings.Join(textParts, " ")
			}
		} else {
			otherMessages = append(otherMessages, msg)
		}
	}

	return systemPrompt, otherMessages
}