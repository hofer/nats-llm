package main

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/hofer/nats-llm/pkq/llm"
	"github.com/nats-io/nats.go"
	"github.com/ollama/ollama/api"
	log "github.com/sirupsen/logrus"
)

func main() {
	nc, err := nats.Connect(os.Getenv("NATS_SERVER_URL"))
	if err != nil {
		log.Fatal(err)
	}

	// Example to chat with Ollama
	chatWithOllama(nc)

	// Example to chat with Gemini
	chatWithGemini(nc)

	// Example with tool calling in Ollama
	toolCallingWithOllama(nc)

	// Example with tool calling in Gemini
	toolCallingWithGemini(nc)
}

// Note: This example is deliberately verbose, so it is easy to understand:
func chatWithOllama(nc *nats.Conn) {
	ollamaLlm := llm.NewNatsOllamaLLM(nc, "gemma3:27b")

	ctx, _ := context.WithTimeout(context.Background(), time.Minute*5)
	firstMessage := api.Message{
		Role:    "user",
		Content: "How many people are in this image?",
		Images: []api.ImageData{
			readImageData(),
		},
	}
	ollamaRes, err := ollamaLlm.Chat(ctx, &api.ChatRequest{
		Messages: []api.Message{
			firstMessage,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Info(ollamaRes.Message.Content)
}

// Note: This example is deliberately verbose, so it is easy to understand:
func chatWithGemini(nc *nats.Conn) {
	geminiLlm := llm.NewNatsGeminiLLM(nc, "gemini-2.5-flash")

	ctx, _ := context.WithTimeout(context.Background(), time.Second*60)
	firstMessage := api.Message{
		Role:    "user",
		Content: "What is the man in the middle holding in his hands?",
		Images: []api.ImageData{
			readImageData(),
		},
	}
	geminiRes1, err1 := geminiLlm.Chat(ctx, &api.ChatRequest{
		Messages: []api.Message{
			firstMessage,
		},
	})
	if err1 != nil {
		log.Fatal(err1)
	}

	log.Infof("First response from LLM: %s", geminiRes1.Message.Content)
	secondMessage := api.Message{
		Role:    "user",
		Content: "Is the person on the left holding flowers as well?",
	}
	geminiRes2, err2 := geminiLlm.Chat(ctx, &api.ChatRequest{
		Messages: []api.Message{
			firstMessage,
			geminiRes1.Message,
			secondMessage,
		},
	})
	if err2 != nil {
		log.Fatal(err2)
	}

	log.Infof("Second response from LLM: %s", geminiRes2.Message.Content)
}

func readImageData() api.ImageData {
	filePath := "example.png"
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Error opening file %s: %v", filePath, err)
	}
	// 2. Ensure the file is closed when the function exits
	defer file.Close()

	// 3. Read all bytes from the file
	data, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("Error reading file %s: %v", filePath, err)
	}
	return data
}

// Note: This example is deliberately verbose, so it is easy to understand:
func toolCallingWithOllama(nc *nats.Conn) {
	ollamaLlm := llm.NewNatsOllamaLLM(nc, "mistral-small:24b")

	ctx, _ := context.WithTimeout(context.Background(), time.Second*60)
	firstMessage := api.Message{
		Role:    "user",
		Content: "What is the current temperature in New York City??",
	}
	ollamaRes1, err1 := ollamaLlm.Chat(ctx, &api.ChatRequest{
		Messages: []api.Message{
			firstMessage,
		},
		Tools: getTools(),
	})
	if err1 != nil {
		log.Fatal(err1)
	}
	log.Infof("Is this the expected tool call? %v", ollamaRes1.Message.ToolCalls)

	toolResult := `{"status": "success", "data": "21 degrees celsius.", "name": "get_temperature"}` // Example tool output (often JSON)
	secondMessage := api.Message{
		Role:    "tool",
		Content: toolResult,
	}
	ollamaRes2, err2 := ollamaLlm.Chat(ctx, &api.ChatRequest{
		Messages: []api.Message{
			firstMessage,
			ollamaRes1.Message,
			secondMessage,
		},
	})
	if err2 != nil {
		log.Fatal(err2)
	}

	log.Infof("Second response from LLM: %s", ollamaRes2.Message.Content)
}

// Note: This example is deliberately verbose, so it is easy to understand:
func toolCallingWithGemini(nc *nats.Conn) {
	geminiLlm := llm.NewNatsGeminiLLM(nc, "gemini-2.5-flash-preview-04-17")

	ctx, _ := context.WithTimeout(context.Background(), time.Second*60)
	firstMessage := api.Message{
		Role:    "user",
		Content: "What time is it?",
	}
	geminiRes1, err1 := geminiLlm.Chat(ctx, &api.ChatRequest{
		Messages: []api.Message{
			firstMessage,
		},
		Tools: getTools(),
	})
	if err1 != nil {
		log.Fatal(err1)
	}
	log.Infof("Is this the expected tool call? %s", geminiRes1.Message.ToolCalls)

	toolResult := `{"status": "success", "data": "21 degrees celsius.", "name": "get_temperature"}` // Example tool output (often JSON)
	secondMessage := api.Message{
		Role:    "tool",
		Content: toolResult,
	}
	geminiRes2, err2 := geminiLlm.Chat(ctx, &api.ChatRequest{
		Messages: []api.Message{
			firstMessage,
			geminiRes1.Message,
			secondMessage,
		},
	})
	if err2 != nil {
		log.Fatal(err2)
	}

	log.Infof("Second response from LLM: %s", geminiRes2.Message.Content)
}

func getTools() []api.Tool {
	return []api.Tool{
		{
			Type: "tool",
			Function: api.ToolFunction{
				Name:        "get_temperature",
				Description: "Returns the current temperature for a given city name",
				Parameters: api.ToolFunctionParameters{
					Type: "object",
					Properties: map[string]api.ToolProperty{
						"city": {
							Type: api.PropertyType{
								"string",
							},
							Description: "The name of the city",
						},
					},
				},
			},
		},
		{
			Type: "tool",
			Function: api.ToolFunction{
				Name:        "get_time",
				Description: "Returns the current time",
			},
		},
	}
}
