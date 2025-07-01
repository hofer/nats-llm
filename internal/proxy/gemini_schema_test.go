package proxy

import (
	"github.com/ollama/ollama/api"
	"github.com/stretchr/testify/assert"
	"google.golang.org/genai"

	"testing"
)

func TestCreateContentParts(t *testing.T) {
	tt := []struct {
		testName     string
		inMessage    api.Message
		expectedPart []*genai.Part
	}{
		{
			testName: "user message",
			inMessage: api.Message{
				Content: "Hello World",
				Role:    "user",
			},
			expectedPart: []*genai.Part{
				genai.NewPartFromText("Hello World"),
			},
		},
		{
			testName: "user message with image",
			inMessage: api.Message{
				Content: "Hello World",
				Role:    "user",
				Images: []api.ImageData{
					[]byte{71, 111},
				},
			},
			expectedPart: []*genai.Part{
				genai.NewPartFromText("Hello World"),
				genai.NewPartFromBytes([]byte{71, 111}, "plain; charset=utf-8"),
			},
		},
		{
			testName: "user message with tools",
			inMessage: api.Message{
				Content: "Hello World",
				Role:    "user",
				ToolCalls: []api.ToolCall{
					{Function: api.ToolCallFunction{Name: "hello_world"}},
				},
			},
			expectedPart: []*genai.Part{
				genai.NewPartFromText("Hello World"),
				genai.NewPartFromFunctionCall("hello_world", nil),
			},
		},
		{
			testName: "tool response",
			inMessage: api.Message{
				Role:    "tool",
				Content: "{\"data\": \"whatever result from function call\", \"name\": \"hello_world\"}",
			},
			expectedPart: []*genai.Part{
				genai.NewPartFromFunctionResponse("hello_world", map[string]any{
					"data": "whatever result from function call",
					"name": "hello_world",
				}),
			},
		},
	}

	for _, td := range tt {
		t.Run(td.testName, func(t *testing.T) {
			//act
			parts := createContentParts(td.inMessage)

			//assert
			assert.Equal(t, td.expectedPart, parts)
		})
	}
}
