package llm

import (
	"context"
	"github.com/nats-io/nats.go"
	"github.com/ollama/ollama/api"
)

func NewNatsGeminiLLM(nc *nats.Conn) *NatsGeminiLLM {
	return &NatsGeminiLLM{
		client: nc,
	}
}

type NatsGeminiLLM struct {
	client *nats.Conn
}

func (n *NatsGeminiLLM) Chat(ctx context.Context, req *api.ChatRequest) (api.ChatResponse, error) {
	var response api.ChatResponse
	err := natsRequest(ctx, n.client, "ollama.chat", req, &response)
	return response, err
}

func (n *NatsGeminiLLM) Embed(ctx context.Context, req *api.EmbedRequest) (api.EmbedResponse, error) {
	var response api.EmbedResponse
	err := natsRequest(ctx, n.client, "ollama.embed", req, &response)
	return response, err
}

func (n *NatsGeminiLLM) Show(ctx context.Context, req *api.ShowRequest) (api.ShowResponse, error) {
	var response api.ShowResponse
	err := natsRequest(ctx, n.client, "ollama.show", req, &response)
	return response, err
}
