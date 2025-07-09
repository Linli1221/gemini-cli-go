package stream

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gemini-cli-go/internal/constants"
	"gemini-cli-go/internal/types"

	"github.com/google/uuid"
)

// Transformer converts Gemini stream chunks to OpenAI format
type Transformer struct {
	model         string
	completionID  string
	chunkIndex    int
	created       int64
}

// NewTransformer creates a new stream transformer
func NewTransformer(model string) *Transformer {
	return &Transformer{
		model:        model,
		completionID: constants.ChatCompletionIDPrefix + strings.ReplaceAll(uuid.New().String(), "-", ""),
		chunkIndex:   0,
		created:      time.Now().Unix(),
	}
}

// Transform converts a Gemini stream chunk to OpenAI format
func (t *Transformer) Transform(chunk types.StreamChunk) (string, error) {
	t.chunkIndex++

	switch chunk.Type {
	case types.StreamChunkTypeText:
		return t.transformTextChunk(chunk)
	case types.StreamChunkTypeReasoning:
		return t.transformReasoningChunk(chunk)
	case types.StreamChunkTypeThinkingContent:
		return t.transformThinkingContentChunk(chunk)
	case types.StreamChunkTypeRealThinking:
		return t.transformRealThinkingChunk(chunk)
	case types.StreamChunkTypeUsage:
		return t.transformUsageChunk(chunk)
	default:
		return "", fmt.Errorf("unknown chunk type: %s", chunk.Type)
	}
}

// transformTextChunk transforms a text chunk to OpenAI format
func (t *Transformer) transformTextChunk(chunk types.StreamChunk) (string, error) {
	text, ok := chunk.Data.(string)
	if !ok {
		return "", fmt.Errorf("invalid text chunk data type")
	}

	response := types.ChatCompletionResponse{
		ID:      t.completionID,
		Object:  constants.OpenAIChatCompletionObject,
		Created: t.created,
		Model:   t.model,
		Choices: []types.ChatCompletionChoice{
			{
				Index: 0,
				Delta: &types.ChatCompletionDelta{
					Content: text,
				},
				FinishReason: nil,
			},
		},
	}

	return t.formatSSEChunk(response)
}

// transformReasoningChunk transforms a reasoning chunk to OpenAI format
func (t *Transformer) transformReasoningChunk(chunk types.StreamChunk) (string, error) {
	reasoningData, ok := chunk.Data.(types.ReasoningData)
	if !ok {
		return "", fmt.Errorf("invalid reasoning chunk data type")
	}

	response := types.ChatCompletionResponse{
		ID:      t.completionID,
		Object:  constants.OpenAIChatCompletionObject,
		Created: t.created,
		Model:   t.model,
		Choices: []types.ChatCompletionChoice{
			{
				Index: 0,
				Delta: &types.ChatCompletionDelta{
					Reasoning: reasoningData.Reasoning,
				},
				FinishReason: nil,
			},
		},
	}

	return t.formatSSEChunk(response)
}

// transformThinkingContentChunk transforms thinking content to OpenAI format
func (t *Transformer) transformThinkingContentChunk(chunk types.StreamChunk) (string, error) {
	text, ok := chunk.Data.(string)
	if !ok {
		return "", fmt.Errorf("invalid thinking content chunk data type")
	}

	response := types.ChatCompletionResponse{
		ID:      t.completionID,
		Object:  constants.OpenAIChatCompletionObject,
		Created: t.created,
		Model:   t.model,
		Choices: []types.ChatCompletionChoice{
			{
				Index: 0,
				Delta: &types.ChatCompletionDelta{
					Content: text,
				},
				FinishReason: nil,
			},
		},
	}

	return t.formatSSEChunk(response)
}

// transformRealThinkingChunk transforms real thinking to OpenAI format
func (t *Transformer) transformRealThinkingChunk(chunk types.StreamChunk) (string, error) {
	text, ok := chunk.Data.(string)
	if !ok {
		return "", fmt.Errorf("invalid real thinking chunk data type")
	}

	response := types.ChatCompletionResponse{
		ID:      t.completionID,
		Object:  constants.OpenAIChatCompletionObject,
		Created: t.created,
		Model:   t.model,
		Choices: []types.ChatCompletionChoice{
			{
				Index: 0,
				Delta: &types.ChatCompletionDelta{
					Reasoning: text,
				},
				FinishReason: nil,
			},
		},
	}

	return t.formatSSEChunk(response)
}

// transformUsageChunk transforms usage data to OpenAI format
func (t *Transformer) transformUsageChunk(chunk types.StreamChunk) (string, error) {
	usageData, ok := chunk.Data.(types.UsageData)
	if !ok {
		return "", fmt.Errorf("invalid usage chunk data type")
	}

	response := types.ChatCompletionResponse{
		ID:      t.completionID,
		Object:  constants.OpenAIChatCompletionObject,
		Created: t.created,
		Model:   t.model,
		Choices: []types.ChatCompletionChoice{
			{
				Index:        0,
				Delta:        &types.ChatCompletionDelta{},
				FinishReason: stringPtr("stop"),
			},
		},
		Usage: &types.ChatCompletionUsage{
			PromptTokens:     usageData.InputTokens,
			CompletionTokens: usageData.OutputTokens,
			TotalTokens:      usageData.InputTokens + usageData.OutputTokens,
		},
	}

	return t.formatSSEChunk(response)
}

// CreateFinalChunk creates the final [DONE] chunk
func (t *Transformer) CreateFinalChunk() string {
	return constants.SSEDoneMessage + constants.SSENewLine + constants.SSENewLine
}

// CreateRoleChunk creates a role chunk (first chunk with role)
func (t *Transformer) CreateRoleChunk() (string, error) {
	response := types.ChatCompletionResponse{
		ID:      t.completionID,
		Object:  constants.OpenAIChatCompletionObject,
		Created: t.created,
		Model:   t.model,
		Choices: []types.ChatCompletionChoice{
			{
				Index: 0,
				Delta: &types.ChatCompletionDelta{
					Role: "assistant",
				},
				FinishReason: nil,
			},
		},
	}

	return t.formatSSEChunk(response)
}

// formatSSEChunk formats a response as SSE chunk
func (t *Transformer) formatSSEChunk(response types.ChatCompletionResponse) (string, error) {
	jsonData, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	return constants.SSEDataPrefix + string(jsonData) + constants.SSENewLine + constants.SSENewLine, nil
}

// StreamWriter handles writing SSE data to a response writer
type StreamWriter struct {
	writer    ResponseWriter
	transformer *Transformer
	hasStarted bool
}

// ResponseWriter interface for writing streaming responses
type ResponseWriter interface {
	Write(data []byte) (int, error)
	Flush()
	Header() map[string]string
	SetHeader(key, value string)
	WriteHeader(statusCode int)
}

// NewStreamWriter creates a new stream writer
func NewStreamWriter(writer ResponseWriter, model string) *StreamWriter {
	return &StreamWriter{
		writer:      writer,
		transformer: NewTransformer(model),
		hasStarted:  false,
	}
}

// WriteHeaders sets the appropriate headers for SSE
func (sw *StreamWriter) WriteHeaders() {
	sw.writer.SetHeader("Content-Type", constants.ContentTypeSSE)
	sw.writer.SetHeader("Cache-Control", constants.CacheControlNoCache)
	sw.writer.SetHeader("Connection", constants.ConnectionKeepAlive)
	sw.writer.SetHeader("Access-Control-Allow-Origin", constants.CORSAllowOrigin)
	sw.writer.SetHeader("Access-Control-Allow-Methods", constants.CORSAllowMethods)
	sw.writer.SetHeader("Access-Control-Allow-Headers", constants.CORSAllowHeaders)
	sw.writer.WriteHeader(constants.StatusOK)
}

// WriteChunk writes a chunk to the stream
func (sw *StreamWriter) WriteChunk(chunk types.StreamChunk) error {
	// Write role chunk first if this is the start
	if !sw.hasStarted && chunk.Type == types.StreamChunkTypeText {
		roleChunk, err := sw.transformer.CreateRoleChunk()
		if err != nil {
			return err
		}
		if _, err := sw.writer.Write([]byte(roleChunk)); err != nil {
			return err
		}
		sw.writer.Flush()
		sw.hasStarted = true
	}

	// Transform and write the chunk
	sseChunk, err := sw.transformer.Transform(chunk)
	if err != nil {
		return err
	}

	if _, err := sw.writer.Write([]byte(sseChunk)); err != nil {
		return err
	}
	sw.writer.Flush()

	return nil
}

// WriteFinalChunk writes the final [DONE] chunk
func (sw *StreamWriter) WriteFinalChunk() error {
	finalChunk := sw.transformer.CreateFinalChunk()
	if _, err := sw.writer.Write([]byte(finalChunk)); err != nil {
		return err
	}
	sw.writer.Flush()
	return nil
}

// ChunkBuffer buffers chunks for processing
type ChunkBuffer struct {
	chunks []types.StreamChunk
	limit  int
}

// NewChunkBuffer creates a new chunk buffer
func NewChunkBuffer(limit int) *ChunkBuffer {
	return &ChunkBuffer{
		chunks: make([]types.StreamChunk, 0, limit),
		limit:  limit,
	}
}

// Add adds a chunk to the buffer
func (cb *ChunkBuffer) Add(chunk types.StreamChunk) {
	if len(cb.chunks) >= cb.limit {
		// Remove oldest chunk if at limit
		cb.chunks = cb.chunks[1:]
	}
	cb.chunks = append(cb.chunks, chunk)
}

// GetAll returns all chunks in the buffer
func (cb *ChunkBuffer) GetAll() []types.StreamChunk {
	return cb.chunks
}

// Clear clears the buffer
func (cb *ChunkBuffer) Clear() {
	cb.chunks = cb.chunks[:0]
}

// Size returns the number of chunks in the buffer
func (cb *ChunkBuffer) Size() int {
	return len(cb.chunks)
}

// Helper functions

// stringPtr returns a pointer to the given string
func stringPtr(s string) *string {
	return &s
}

// CombineTextChunks combines consecutive text chunks for efficiency
func CombineTextChunks(chunks []types.StreamChunk) []types.StreamChunk {
	if len(chunks) == 0 {
		return chunks
	}

	var combined []types.StreamChunk
	var textBuffer strings.Builder
	var hasText bool

	for _, chunk := range chunks {
		if chunk.Type == types.StreamChunkTypeText {
			if text, ok := chunk.Data.(string); ok {
				textBuffer.WriteString(text)
				hasText = true
			}
		} else {
			// Flush any accumulated text
			if hasText {
				combined = append(combined, types.StreamChunk{
					Type: types.StreamChunkTypeText,
					Data: textBuffer.String(),
				})
				textBuffer.Reset()
				hasText = false
			}
			// Add the non-text chunk
			combined = append(combined, chunk)
		}
	}

	// Flush any remaining text
	if hasText {
		combined = append(combined, types.StreamChunk{
			Type: types.StreamChunkTypeText,
			Data: textBuffer.String(),
		})
	}

	return combined
}

// ValidateChunk validates a stream chunk
func ValidateChunk(chunk types.StreamChunk) error {
	switch chunk.Type {
	case types.StreamChunkTypeText, types.StreamChunkTypeThinkingContent:
		if _, ok := chunk.Data.(string); !ok {
			return fmt.Errorf("text chunk must contain string data")
		}
	case types.StreamChunkTypeReasoning:
		if _, ok := chunk.Data.(types.ReasoningData); !ok {
			return fmt.Errorf("reasoning chunk must contain ReasoningData")
		}
	case types.StreamChunkTypeUsage:
		if _, ok := chunk.Data.(types.UsageData); !ok {
			return fmt.Errorf("usage chunk must contain UsageData")
		}
	case types.StreamChunkTypeRealThinking:
		if _, ok := chunk.Data.(string); !ok {
			return fmt.Errorf("real thinking chunk must contain string data")
		}
	default:
		return fmt.Errorf("unknown chunk type: %s", chunk.Type)
	}
	return nil
}