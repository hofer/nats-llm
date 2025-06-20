package llm

import (
	"context"
	"github.com/nats-io/nats.go"
	"github.com/ollama/ollama/api"
)

const (
	geminiChatSubject  = "gemini.chat"
	geminiEmbedSubject = "gemini.embed"
	geminiShowSubject  = "gemini.show"
)

func NewNatsGeminiLLM(nc *nats.Conn, modelName string) *NatsGeminiLLM {
	return &NatsGeminiLLM{
		client:    nc,
		modelName: modelName,
	}
}

type NatsGeminiLLM struct {
	client    *nats.Conn
	modelName string
}

func (n *NatsGeminiLLM) Chat(ctx context.Context, req *api.ChatRequest) (api.ChatResponse, error) {
	req.Model = n.modelName
	var response api.ChatResponse
	err := natsRequest(ctx, n.client, geminiChatSubject, req, &response)
	return response, err
}

func (n *NatsGeminiLLM) Embed(ctx context.Context, req *api.EmbedRequest) (api.EmbedResponse, error) {
	req.Model = n.modelName
	var response api.EmbedResponse
	err := natsRequest(ctx, n.client, geminiEmbedSubject, req, &response)
	return response, err
}

func (n *NatsGeminiLLM) Show(ctx context.Context, req *api.ShowRequest) (api.ShowResponse, error) {
	req.Model = n.modelName
	var response api.ShowResponse
	err := natsRequest(ctx, n.client, geminiShowSubject, req, &response)
	return response, err
}
