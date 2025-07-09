package gemini

import (
	"strings"
	"gemini-cli-go/internal/types"
)

// StreamOptions represents options for streaming requests
type StreamOptions struct {
	// Temperature controls randomness in the response
	Temperature *float64 `json:"temperature,omitempty"`
	
	// MaxTokens limits the number of tokens in the response
	MaxTokens *int `json:"max_tokens,omitempty"`
	
	// TopP controls nucleus sampling
	TopP *float64 `json:"top_p,omitempty"`
	
	// ThinkingBudget controls the token budget for thinking
	ThinkingBudget *int `json:"thinking_budget,omitempty"`
	
	// EnableRealThinking enables real thinking from Gemini
	EnableRealThinking bool `json:"enable_real_thinking"`
	
	// EnableFakeThinking enables fake thinking generation
	EnableFakeThinking bool `json:"enable_fake_thinking"`
	
	// StreamThinkingAsContent streams thinking as content with tags
	StreamThinkingAsContent bool `json:"stream_thinking_as_content"`
	
	// IncludeReasoning includes reasoning in the response
	IncludeReasoning bool `json:"include_reasoning"`
}

// CompletionResult represents the result of a completion request
type CompletionResult struct {
	Content string           `json:"content"`
	Usage   *types.UsageData `json:"usage,omitempty"`
}

// StreamingContext holds context for streaming operations
type StreamingContext struct {
	HasStartedThinking bool
	HasClosedThinking  bool
	NeedsThinkingClose bool
}

// GenerationConfig represents Gemini generation configuration
type GenerationConfig struct {
	Temperature     float64 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
	TopP            float64 `json:"topP,omitempty"`
	ThinkingBudget  int     `json:"thinkingBudget,omitempty"`
	StopSequences   []string `json:"stopSequences,omitempty"`
}

// StreamRequest represents a streaming request to Gemini
type StreamRequest struct {
	Model   string                           `json:"model"`
	Project string                           `json:"project"`
	Request StreamRequestContent             `json:"request"`
}

// StreamRequestContent represents the content of a stream request
type StreamRequestContent struct {
	Contents         []types.GeminiFormattedMessage `json:"contents"`
	GenerationConfig GenerationConfig               `json:"generationConfig"`
}

// ErrorInfo represents error information
type ErrorInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// ProcessingState tracks the state of stream processing
type ProcessingState struct {
	HasStartedThinking bool
	HasClosedThinking  bool
	LastChunkType      types.StreamChunkType
	ContentBuffer      []string
}

// Reset resets the processing state
func (ps *ProcessingState) Reset() {
	ps.HasStartedThinking = false
	ps.HasClosedThinking = false
	ps.LastChunkType = ""
	ps.ContentBuffer = ps.ContentBuffer[:0]
}

// ShouldCloseThinking determines if thinking should be closed
func (ps *ProcessingState) ShouldCloseThinking() bool {
	return ps.HasStartedThinking && !ps.HasClosedThinking
}

// MarkThinkingStarted marks thinking as started
func (ps *ProcessingState) MarkThinkingStarted() {
	ps.HasStartedThinking = true
}

// MarkThinkingClosed marks thinking as closed
func (ps *ProcessingState) MarkThinkingClosed() {
	ps.HasClosedThinking = true
}

// ChunkProcessor handles processing of different chunk types
type ChunkProcessor struct {
	state   *ProcessingState
	options *StreamOptions
}

// NewChunkProcessor creates a new chunk processor
func NewChunkProcessor(options *StreamOptions) *ChunkProcessor {
	return &ChunkProcessor{
		state:   &ProcessingState{},
		options: options,
	}
}

// ProcessPart processes a Gemini part and returns appropriate chunks
func (cp *ChunkProcessor) ProcessPart(part types.GeminiPart) []types.StreamChunk {
	var chunks []types.StreamChunk

	// Handle thinking content
	if part.Thought && part.Text != "" {
		if cp.options != nil && cp.options.StreamThinkingAsContent {
			// Stream as content with thinking tags
			if !cp.state.HasStartedThinking {
				chunks = append(chunks, types.StreamChunk{
					Type: types.StreamChunkTypeThinkingContent,
					Data: "<thinking>\n",
				})
				cp.state.MarkThinkingStarted()
			}
			chunks = append(chunks, types.StreamChunk{
				Type: types.StreamChunkTypeThinkingContent,
				Data: part.Text,
			})
		} else {
			// Stream as separate reasoning field
			chunks = append(chunks, types.StreamChunk{
				Type: types.StreamChunkTypeRealThinking,
				Data: part.Text,
			})
		}
		return chunks
	}

	// Handle regular text content
	if part.Text != "" && !part.Thought {
		// Close thinking tag if needed
		if cp.state.ShouldCloseThinking() {
			chunks = append(chunks, types.StreamChunk{
				Type: types.StreamChunkTypeThinkingContent,
				Data: "\n</thinking>\n\n",
			})
			cp.state.MarkThinkingClosed()
		}

		chunks = append(chunks, types.StreamChunk{
			Type: types.StreamChunkTypeText,
			Data: part.Text,
		})
	}

	return chunks
}

// ReasoningGenerator generates reasoning content
type ReasoningGenerator struct {
	templates []string
}

// NewReasoningGenerator creates a new reasoning generator
func NewReasoningGenerator() *ReasoningGenerator {
	return &ReasoningGenerator{
		templates: []string{
			"🔍 **Analyzing the request: \"{requestPreview}\"**\n\n",
			"🤔 Let me think about this step by step... ",
			"💭 I need to consider the context and provide a comprehensive response. ",
			"🎯 Based on my understanding, I should address the key points while being accurate and helpful. ",
			"✨ Let me formulate a clear and structured answer.\n\n",
		},
	}
}

// Generate generates reasoning content for a given request
func (rg *ReasoningGenerator) Generate(requestPreview string) []string {
	var reasoning []string
	for _, template := range rg.templates {
		reasoning = append(reasoning, 
			strings.ReplaceAll(template, "{requestPreview}", requestPreview))
	}
	return reasoning
}

// TextChunker splits text into chunks for streaming
type TextChunker struct {
	chunkSize int
}

// NewTextChunker creates a new text chunker
func NewTextChunker(chunkSize int) *TextChunker {
	return &TextChunker{chunkSize: chunkSize}
}

// Split splits text into chunks with word boundary awareness
func (tc *TextChunker) Split(text string) []string {
	if len(text) <= tc.chunkSize {
		return []string{text}
	}

	var chunks []string
	remaining := text

	for len(remaining) > 0 {
		if len(remaining) <= tc.chunkSize {
			chunks = append(chunks, remaining)
			break
		}

		// Find a good break point
		chunkEnd := tc.chunkSize
		searchSpace := remaining[:chunkEnd+10] // Look ahead a bit
		
		// Look for good break characters
		goodBreaks := []string{" ", "\n", ".", ",", "!", "?", ";", ":"}
		for _, breakChar := range goodBreaks {
			if idx := strings.LastIndex(searchSpace, breakChar); idx > tc.chunkSize*7/10 {
				chunkEnd = idx + 1
				break
			}
		}

		chunks = append(chunks, remaining[:chunkEnd])
		remaining = remaining[chunkEnd:]
	}

	return chunks
}