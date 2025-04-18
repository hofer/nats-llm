package proxy

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

	msg, err := n.client.Request("ollama.chat", jsonStr, 420*time.Second)
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
