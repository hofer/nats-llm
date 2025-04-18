package llm

import (
	"context"
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/ollama/ollama/api"
	"time"
)

func NewNatsOllamaLLM(nc *nats.Conn) *NatsOllamaLLM {
	return &NatsOllamaLLM{
		client: nc,
	}
}

type NatsOllamaLLM struct {
	client *nats.Conn
}

func (n *NatsOllamaLLM) Chat(ctx context.Context, req *api.ChatRequest) (api.ChatResponse, error) {
	jsonStr, err := json.Marshal(req)
	if err != nil {
		return api.ChatResponse{}, err
	}

	deadline, _ := ctx.Deadline()
	remainingDuration := time.Until(deadline)

	msg, err := n.client.Request("ollama.chat", jsonStr, remainingDuration)
	if err != nil {
		return api.ChatResponse{}, err
	}

	var chatResponse api.ChatResponse
	err = json.Unmarshal(msg.Data, &chatResponse)
	if err != nil {
		return api.ChatResponse{}, err
	}

	return chatResponse, nil
}

func (n *NatsOllamaLLM) Embed(ctx context.Context, req *api.EmbedRequest) (api.EmbedResponse, error) {

	jsonStr, err := json.Marshal(req)
	if err != nil {
		return api.EmbedResponse{}, err
	}

	deadline, _ := ctx.Deadline()
	remainingDuration := time.Until(deadline)

	msg, err := n.client.Request("ollama.embed", jsonStr, remainingDuration)
	if err != nil {
		return api.EmbedResponse{}, err
	}

	var embedResponse api.EmbedResponse
	err = json.Unmarshal(msg.Data, &embedResponse)
	if err != nil {
		return api.EmbedResponse{}, err
	}

	return embedResponse, nil
}
