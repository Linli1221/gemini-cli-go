package models

import (
	"fmt"
	"gemini-cli-go/internal/constants"
	"gemini-cli-go/internal/types"
)

// GeminiCliModels contains configuration for all supported Gemini models
var GeminiCliModels = map[string]types.ModelInfo{
	"gemini-2.5-pro": {
		MaxTokens:           65536,
		ContextWindow:       1048576,
		SupportsImages:      true,
		SupportsPromptCache: false,
		InputPrice:          0,
		OutputPrice:         0,
		Description:         "Google's Gemini 2.5 Pro model via OAuth (free tier)",
		Thinking:            true,
	},
	"gemini-2.5-flash": {
		MaxTokens:           65536,
		ContextWindow:       1048576,
		SupportsImages:      true,
		SupportsPromptCache: false,
		InputPrice:          0,
		OutputPrice:         0,
		Description:         "Google's Gemini 2.5 Flash model via OAuth (free tier)",
		Thinking:            true,
	},
	"gemini-2.0-flash-001": {
		MaxTokens:           65536,
		ContextWindow:       1048576,
		SupportsImages:      true,
		SupportsPromptCache: false,
		InputPrice:          0,
		OutputPrice:         0,
		Description:         "Google's Gemini 2.0 Flash model via OAuth (free tier)",
		Thinking:            false,
	},
	"gemini-2.0-flash-lite-preview-02-05": {
		MaxTokens:           65536,
		ContextWindow:       1048576,
		SupportsImages:      true,
		SupportsPromptCache: false,
		InputPrice:          0,
		OutputPrice:         0,
		Description:         "Google's Gemini 2.0 Flash Lite Preview model via OAuth (free tier)",
		Thinking:            false,
	},
	"gemini-2.0-pro-exp-02-05": {
		MaxTokens:           65536,
		ContextWindow:       1048576,
		SupportsImages:      true,
		SupportsPromptCache: false,
		InputPrice:          0,
		OutputPrice:         0,
		Description:         "Google's Gemini 2.0 Pro Experimental model via OAuth (free tier)",
		Thinking:            false,
	},
	"gemini-1.5-pro": {
		MaxTokens:           65536,
		ContextWindow:       1048576,
		SupportsImages:      true,
		SupportsPromptCache: false,
		InputPrice:          0,
		OutputPrice:         0,
		Description:         "Google's Gemini 1.5 Pro model via OAuth (free tier)",
		Thinking:            false,
	},
	"gemini-1.5-flash": {
		MaxTokens:           65536,
		ContextWindow:       1048576,
		SupportsImages:      true,
		SupportsPromptCache: false,
		InputPrice:          0,
		OutputPrice:         0,
		Description:         "Google's Gemini 1.5 Flash model via OAuth (free tier)",
		Thinking:            false,
	},
}

// GetModelInfo returns model information for the given model ID
func GetModelInfo(modelID string) (*types.ModelInfo, bool) {
	info, exists := GeminiCliModels[modelID]
	if !exists {
		return nil, false
	}
	return &info, true
}

// GetAllModelIDs returns a slice of all available model IDs
func GetAllModelIDs() []string {
	modelIDs := make([]string, 0, len(GeminiCliModels))
	for modelID := range GeminiCliModels {
		modelIDs = append(modelIDs, modelID)
	}
	return modelIDs
}

// IsValidModel checks if the given model ID is valid
func IsValidModel(modelID string) bool {
	_, exists := GeminiCliModels[modelID]
	return exists
}

// GetDefaultModel returns the default model ID
func GetDefaultModel() string {
	return constants.DefaultModel
}

// SupportsImages checks if the model supports image inputs
func SupportsImages(modelID string) bool {
	info, exists := GeminiCliModels[modelID]
	if !exists {
		return false
	}
	return info.SupportsImages
}

// SupportsThinking checks if the model supports thinking/reasoning
func SupportsThinking(modelID string) bool {
	info, exists := GeminiCliModels[modelID]
	if !exists {
		return false
	}
	return info.Thinking
}

// GetMaxTokens returns the maximum number of tokens for the model
func GetMaxTokens(modelID string) int {
	info, exists := GeminiCliModels[modelID]
	if !exists {
		return constants.MaxTokensDefault
	}
	return info.MaxTokens
}

// GetContextWindow returns the context window size for the model
func GetContextWindow(modelID string) int {
	info, exists := GeminiCliModels[modelID]
	if !exists {
		return constants.ContextWindowDefault
	}
	return info.ContextWindow
}

// GetModelDescription returns the description for the model
func GetModelDescription(modelID string) string {
	info, exists := GeminiCliModels[modelID]
	if !exists {
		return "Unknown model"
	}
	return info.Description
}

// GetVisionCapableModels returns a slice of model IDs that support vision
func GetVisionCapableModels() []string {
	var visionModels []string
	for modelID, info := range GeminiCliModels {
		if info.SupportsImages {
			visionModels = append(visionModels, modelID)
		}
	}
	return visionModels
}

// GetThinkingCapableModels returns a slice of model IDs that support thinking
func GetThinkingCapableModels() []string {
	var thinkingModels []string
	for modelID, info := range GeminiCliModels {
		if info.Thinking {
			thinkingModels = append(thinkingModels, modelID)
		}
	}
	return thinkingModels
}

// ValidateModelForImages validates if a model supports image inputs
func ValidateModelForImages(modelID string) error {
	if !SupportsImages(modelID) {
		return fmt.Errorf("model '%s' does not support image inputs. Please use a vision-capable model", modelID)
	}
	return nil
}

// ValidateModelForThinking validates if a model supports thinking
func ValidateModelForThinking(modelID string) error {
	if !SupportsThinking(modelID) {
		return fmt.Errorf("model '%s' does not support thinking. Please use a thinking-capable model", modelID)
	}
	return nil
}