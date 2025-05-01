package proxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/generative-ai-go/genai"
	"github.com/ollama/ollama/api"
	"net/http"
	"strings"
	"time"
)

func GetGeminiSchemaChat() (string, error) {
	return marshalSchema(&api.ChatRequest{}, &api.ChatResponse{})
}

func createHistoryContent(reqData api.ChatRequest) []*genai.Content {
	if len(reqData.Messages) == 1 {
		return []*genai.Content{}
	}

	result := []*genai.Content{}
	messages := reqData.Messages[:len(reqData.Messages)-1]
	for _, message := range messages {
		role := message.Role
		if role == "assistant" {
			role = "model"
		}
		result = append(result, &genai.Content{
			Role:  role,
			Parts: createContentParts(message),
		})
	}
	return result
}

func createUserContentParts(reqData api.ChatRequest) ([]genai.Part, error) {
	// we assume that the last message is a user inMessage:
	if len(reqData.Messages) == 0 {
		return nil, errors.New("no message content found in the request")
	}

	userMessage := reqData.Messages[len(reqData.Messages)-1]
	if strings.ToLower(userMessage.Role) != "user" && strings.ToLower(userMessage.Role) != "tool" {
		return nil, errors.New(fmt.Sprintf("message role must be 'user' or 'tool' but was '%s'", userMessage.Role))
	}

	return createContentParts(userMessage), nil
}

func createContentParts(message api.Message) []genai.Part {
	parts := []genai.Part{}
	if len(message.Content) > 0 && message.Role != "tool" {
		parts = append(parts, genai.Text(message.Content))
	}

	if message.Role == "tool" {
		toolResult := jsonToMap(message.Content)
		parts = append(parts, genai.FunctionResponse{
			Name:     toolResult["name"].(string),
			Response: toolResult,
		})
	}

	for _, toolCall := range message.ToolCalls {
		parts = append(parts, genai.FunctionCall{
			Name: toolCall.Function.Name,
			Args: toolCall.Function.Arguments,
		})
	}

	for _, imageData := range message.Images {
		mimeType := http.DetectContentType(imageData)
		parts = append(parts, genai.ImageData(strings.Split(mimeType, "/")[1], imageData))
	}
	return parts
}

func createGeminiToolSchema(reqData api.ChatRequest) []*genai.Tool {
	result := make([]*genai.Tool, 0)
	for _, tool := range reqData.Tools {
		props := map[string]*genai.Schema{}
		for name, prop := range tool.Function.Parameters.Properties {
			props[name] = &genai.Schema{
				Type:        mapOllamaType(prop.Type),
				Description: prop.Description,
			}
		}

		parametersSchema := &genai.Schema{
			Type:       genai.TypeObject,
			Properties: props,
		}
		geminiToolTool := &genai.Tool{
			FunctionDeclarations: []*genai.FunctionDeclaration{{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				Parameters:  parametersSchema,
			}},
		}
		result = append(result, geminiToolTool)
	}
	return result
}

func createOllamaResponse(resp *genai.GenerateContentResponse) (api.ChatResponse, error) {
	if len(resp.Candidates) > 1 {
		return api.ChatResponse{}, errors.New("too many candidates. expecting only one candidate")
	}

	responseText := ""
	toolCalls := []api.ToolCall{}

	candidate := resp.Candidates[0]
	if candidate.Content != nil {
		for _, part := range candidate.Content.Parts {
			switch part.(type) {
			case genai.Text:
				responseText += fmt.Sprint(part)
			case genai.FunctionCall:
				fc := part.(genai.FunctionCall)
				toolCalls = append(toolCalls, api.ToolCall{
					Function: api.ToolCallFunction{
						Name:      fc.Name,
						Arguments: fc.Args,
					},
				})
			default:
				// Not handled...
			}
		}
	}

	return api.ChatResponse{
		CreatedAt: time.Now(),
		Message: api.Message{
			Content:   responseText,
			Role:      "assistant",
			ToolCalls: toolCalls,
		},
		DoneReason: candidate.FinishReason.String(),
		Done:       candidate.FinishReason == genai.FinishReasonStop,
	}, nil
}

func mapOllamaType(propertyType api.PropertyType) genai.Type {
	var typesForNames = map[string]genai.Type{
		"string":  genai.TypeString,
		"double":  genai.TypeNumber,
		"float":   genai.TypeNumber,
		"integer": genai.TypeInteger,
		"bool":    genai.TypeBoolean,
		"boolean": genai.TypeBoolean,
		"array":   genai.TypeArray,
		"object":  genai.TypeObject,
	}

	result, ok := typesForNames[propertyType[0]]
	if !ok {
		return genai.TypeUnspecified
	}
	return result
}

func jsonToMap(jsonStr string) map[string]interface{} {
	result := make(map[string]interface{})
	json.Unmarshal([]byte(jsonStr), &result)
	return result
}
