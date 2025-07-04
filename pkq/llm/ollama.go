package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/ollama/ollama/api"
	"time"
)

const (
	ollamaChatSubject  = "ollama.chat"
	ollamaEmbedSubject = "ollama.embed"
	ollamaShowSubject  = "ollama.show"
)

func NewNatsOllamaLLM(nc *nats.Conn, modelName string) *NatsOllamaLLM {
	return &NatsOllamaLLM{
		client:    nc,
		modelName: modelName,
	}
}

type NatsOllamaLLM struct {
	client    *nats.Conn
	modelName string
}

func (n *NatsOllamaLLM) Chat(ctx context.Context, req *api.ChatRequest) (api.ChatResponse, error) {
	req.Model = n.modelName
	var response api.ChatResponse
	err := natsRequest(ctx, n.client, ollamaChatSubject, req, &response)
	return response, err
}

func (n *NatsOllamaLLM) Embed(ctx context.Context, req *api.EmbedRequest) (api.EmbedResponse, error) {
	req.Model = n.modelName
	var response api.EmbedResponse
	err := natsRequest(ctx, n.client, ollamaEmbedSubject, req, &response)
	return response, err
}

func (n *NatsOllamaLLM) Show(ctx context.Context, req *api.ShowRequest) (api.ShowResponse, error) {
	req.Model = n.modelName
	var response api.ShowResponse
	err := natsRequest(ctx, n.client, ollamaShowSubject, req, &response)
	return response, err
}

type ApiResponse interface {
	*api.ShowResponse | *api.EmbedResponse | *api.ChatResponse
}

type ApiRequest interface {
	*api.ShowRequest | *api.EmbedRequest | *api.ChatRequest
}

func natsRequest[T ApiRequest, A ApiResponse](ctx context.Context, n *nats.Conn, subject string, req T, resp A) error {
	jsonStr, err := json.Marshal(req)
	if err != nil {
		return err
	}

	remainingDuration := time.Second * 30
	deadline, ok := ctx.Deadline()
	if ok {
		remainingDuration = time.Until(deadline)
	}

	msg, err := n.Request(subject, jsonStr, remainingDuration)
	if err != nil {
		return err
	}

	if msg.Data == nil || len(msg.Data) == 0 {
		return fmt.Errorf("Failed to create a response from a given request")
	}

	err = json.Unmarshal(msg.Data, resp)
	return err
}
