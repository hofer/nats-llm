package llm

import (
	"context"
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/ollama/ollama/api"
	"time"
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
	jsonStr, err := json.Marshal(req)
	if err != nil {
		return api.ChatResponse{}, err
	}

	remainingDuration := time.Second * 30
	deadline, ok := ctx.Deadline()
	if ok {
		remainingDuration = time.Until(deadline)
	}

	msg, err := n.client.Request("gemini.chat", jsonStr, remainingDuration)
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

func (n *NatsGeminiLLM) Show(ctx context.Context, req *api.ShowRequest) (api.ShowResponse, error) {
	jsonStr, err := json.Marshal(req)
	if err != nil {
		return api.ShowResponse{}, err
	}

	remainingDuration := time.Second * 30
	deadline, ok := ctx.Deadline()
	if ok {
		remainingDuration = time.Until(deadline)
	}

	msg, err := n.client.Request("gemini.show", jsonStr, remainingDuration)
	if err != nil {
		return api.ShowResponse{}, err
	}

	var showResponse api.ShowResponse
	err = json.Unmarshal(msg.Data, &showResponse)
	if err != nil {
		return api.ShowResponse{}, err
	}

	return showResponse, nil
}
