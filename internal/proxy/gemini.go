package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/huh/spinner"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
	"github.com/ollama/ollama/api"
	log "github.com/sirupsen/logrus"
	"google.golang.org/genai"
	"runtime"
)

func StartNatsGeminiProxy(natsUrl string, apiKey string) error {
	nc, err := nats.Connect(natsUrl)
	if err != nil {
		return err
	}

	natsGeminiProxy := NewNatsGeminiProxy(apiKey)
	err = natsGeminiProxy.Start(nc)
	if err != nil {
		return err
	}

	runtime.Goexit()
	return nil
}

type NatsGeminiProxy struct {
	apiKey string
	client *genai.Client
}

func NewNatsGeminiProxy(apiKey string) *NatsGeminiProxy {
	return &NatsGeminiProxy{
		apiKey: apiKey,
	}
}

func (n *NatsGeminiProxy) Start(nc *nats.Conn) error {

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: n.apiKey,
	})

	if err != nil {
		return err
	}
	n.client = client
	//defer client.Close()

	srv, err := micro.AddService(nc, micro.Config{
		Name:        "NatsGemini",
		Version:     "0.0.1",
		Description: "Nats microservice acting as a proxy for Gemini.",
	})
	if err != nil {
		return err
	}
	//defer srv.Stop()

	root := srv.AddGroup("gemini")

	// Chat
	chatSchema, err := GetGeminiSchemaChat()
	if err != nil {
		return err
	}
	err = root.AddEndpoint("chat", micro.HandlerFunc(n.chatHandler), micro.WithEndpointMetadata(map[string]string{
		"schema": chatSchema,
	}))
	if err != nil {
		return err
	}

	// Show
	showSchema, err := GetGeminiSchemaShow()
	if err != nil {
		return err
	}
	err = root.AddEndpoint("show", micro.HandlerFunc(n.showHandler), micro.WithEndpointMetadata(map[string]string{
		"schema": showSchema,
	}))

	return err
}

func (n *NatsGeminiProxy) chatHandler(req micro.Request) {
	var reqData api.ChatRequest
	err := json.Unmarshal(req.Data(), &reqData)
	if err != nil {
		req.Error("400", err.Error(), nil)
		return
	}

	// Create the chat session with the Gemini model:
	history := createHistoryContent(reqData)
	chat, err := n.client.Chats.Create(context.Background(), reqData.Model, &genai.GenerateContentConfig{
		Tools:             createGeminiToolSchema(reqData),
		SystemInstruction: createGeminiSystemPrompt(reqData),
	}, history)
	if err != nil {
		req.Error("500", err.Error(), nil)
		return
	}

	// User content can either be a user input or a tool response:
	userContentParts, contentErr := createUserContentParts(reqData)
	if contentErr != nil {
		log.Errorf("session.SendMessage: %v", contentErr)
		req.Error("500", contentErr.Error(), nil)
		return
	}

	var res *genai.GenerateContentResponse
	sp := spinner.New()
	action := func() {
		res, err = chat.Send(context.Background(), userContentParts...)
	}

	sp.Title(fmt.Sprintf("Generate content with model '%s'...", reqData.Model)).Action(action).Run()
	if err != nil {
		log.Errorf("session.SendMessage: %v", err)
		req.Error("500", err.Error(), nil)
		return
	}

	ollamaResp, err := createOllamaChatResponse(res)
	if err != nil {
		log.Errorf("cannot create a response: %v", err)
		req.Error("400", err.Error(), nil)
		return
	}

	responseData, err := json.Marshal(ollamaResp)
	if err != nil {
		log.Errorf("cannot create a response: %v", err)
		req.Error("400", err.Error(), nil)
		return
	}

	log.Debug(string(responseData))
	err = req.Respond(responseData)
}

func (n *NatsGeminiProxy) showHandler(req micro.Request) {
	var reqData api.ShowRequest
	err := json.Unmarshal(req.Data(), &reqData)
	if err != nil {
		req.Error("400", err.Error(), nil)
		return
	}

	// Get the generative model requested by the user:
	model, err := n.client.Models.Get(context.Background(), reqData.Model, &genai.GetModelConfig{})
	if err != nil {
		log.Error(err)
		req.Error("500", err.Error(), nil)
		return
	}

	ollamaResp, err := createOllamaShowResponse(model)
	if err != nil {
		log.Errorf("cannot create a response: %v", err)
		req.Error("400", err.Error(), nil)
		return
	}

	responseData, err := json.Marshal(ollamaResp)
	if err != nil {
		log.Errorf("cannot create a response: %v", err)
		req.Error("400", err.Error(), nil)
		return
	}

	log.Debug(string(responseData))
	err = req.Respond(responseData)
}
