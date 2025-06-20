package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/huh/spinner"
	"github.com/google/generative-ai-go/genai"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
	"github.com/ollama/ollama/api"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"
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
	client, err := genai.NewClient(ctx, option.WithAPIKey(n.apiKey))
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

	// Get the generative model requested by the user:
	model := n.client.GenerativeModel(reqData.Model)

	// Set system prompt if available in the message history:
	systemPrompt := createGeminiSystemPrompt(reqData)
	if systemPrompt != nil {
		model.SystemInstruction = systemPrompt
	}

	// Before initiating a conversation, we tell the model which tools it has
	// at its disposal.
	model.Tools = createGeminiToolSchema(reqData)

	// For using tools, the chat mode is useful because it provides the required
	// chat context/history.
	session := model.StartChat()
	session.History = createHistoryContent(reqData)

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
		res, err = session.SendMessage(context.Background(), userContentParts...)
		//res, err = model.GenerateContent(context.Background(), userContentParts...)
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
	model := n.client.GenerativeModel(reqData.Model)

	info, err := model.Info(context.Background())
	if err != nil {
		log.Error(err)
		req.Error("500", err.Error(), nil)
		return
	}

	//// Set system prompt if available in the message history:
	//systemPrompt := createGeminiSystemPrompt(reqData)
	//if systemPrompt != nil {
	//	model.SystemInstruction = systemPrompt
	//}
	//
	//// Before initiating a conversation, we tell the model which tools it has
	//// at its disposal.
	//model.Tools = createGeminiToolSchema(reqData)
	//
	//// For using tools, the chat mode is useful because it provides the required
	//// chat context/history.
	//session := model.StartChat()
	//session.History = createHistoryContent(reqData)
	//
	//// User content can either be a user input or a tool response:
	//userContentParts, contentErr := createUserContentParts(reqData)
	//if contentErr != nil {
	//	log.Errorf("session.SendMessage: %v", contentErr)
	//	req.Error("500", contentErr.Error(), nil)
	//	return
	//}
	//
	//var res *genai.GenerateContentResponse
	//sp := spinner.New()
	//action := func() {
	//	res, err = session.SendMessage(ctx, userContentParts...)
	//	//res, err = model.GenerateContent(ctx, userContentParts...)
	//}
	//
	//sp.Title(fmt.Sprintf("Generate content with model '%s'...", reqData.Model)).Action(action).Run()
	//if err != nil {
	//	log.Errorf("session.SendMessage: %v", err)
	//	req.Error("500", err.Error(), nil)
	//	return
	//}

	ollamaResp, err := createOllamaShowResponse(info)
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
