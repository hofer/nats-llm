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
	"runtime"
	"strings"
)

func StartOllamaProxy(natsUrl string, ollamaUrl string) error {
	nc, err := nats.Connect(natsUrl)
	if err != nil {
		return err
	}

	natsOllamaProxy := NewNatsOllamaProxy()
	err = natsOllamaProxy.Start(nc)
	if err != nil {
		return err
	}

	runtime.Goexit()
	return nil
}

type NatsOllamaProxy struct {
	client *api.Client
}

func NewNatsOllamaProxy() *NatsOllamaProxy {
	return &NatsOllamaProxy{}
}

func (n *NatsOllamaProxy) Start(nc *nats.Conn) error {
	log.Infof("Starting nats-ollama-proxy...")
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return err
	}
	n.client = client

	srv, err := micro.AddService(nc, micro.Config{
		Name:        "NatsOllama",
		Version:     "0.0.1",
		Description: "Nats microservice acting as a proxy for Ollama.",
	})
	if err != nil {
		return err
	}
	//defer srv.Stop()

	root := srv.AddGroup("ollama")

	// Generate
	generateSchema, err := GetSchemaGenerate()
	if err != nil {
		log.Fatal(err)
	}
	err = root.AddEndpoint("generate", micro.HandlerFunc(n.generateHandler), micro.WithEndpointMetadata(map[string]string{
		"schema": generateSchema,
	}))
	if err != nil {
		return err
	}

	// Embed
	embedSchema, err := GetSchemaEmbed()
	err = root.AddEndpoint("embed", micro.HandlerFunc(n.embedHandler), micro.WithEndpointMetadata(map[string]string{
		"schema": embedSchema,
	}))
	if err != nil {
		return err
	}

	// Embedding
	embeddingSchema, err := GetSchemaEmbedding()
	err = root.AddEndpoint("embedding", micro.HandlerFunc(n.embeddingHandler), micro.WithEndpointMetadata(map[string]string{
		"schema": embeddingSchema,
	}))
	if err != nil {
		return err
	}

	// Chat
	chatSchema, err := GetSchemaChat()
	err = root.AddEndpoint("chat", micro.HandlerFunc(n.chatHandler), micro.WithEndpointMetadata(map[string]string{
		"schema": chatSchema,
	}))
	if err != nil {
		return err
	}

	// Show
	showSchema, err := GetSchemaShow()
	err = root.AddEndpoint("show", micro.HandlerFunc(n.showHandler), micro.WithEndpointMetadata(map[string]string{
		"schema": showSchema,
	}))
	return err
}

func (n *NatsOllamaProxy) generateHandler(req micro.Request) {

	var reqData api.GenerateRequest
	err := json.Unmarshal(req.Data(), &reqData)
	if err != nil {
		req.Error("400", err.Error(), nil)
		return
	}

	// Set streaming to false, thus making sure we wait for a response.
	reqData.Stream = new(bool)

	ctx := context.Background()
	respFunc := func(resp api.GenerateResponse) error {
		// Only print the response here; GenerateResponse has a number of other
		// interesting fields you want to examine.
		//fmt.Println()

		responseData, err := json.Marshal(resp)
		if err != nil {
			req.Error("400", err.Error(), nil)
			return err
		}
		err = req.Respond(responseData)
		return err
	}

	err = n.client.Generate(ctx, &reqData, respFunc)
	if err != nil {
		req.Error("400", err.Error(), nil)
	}
}

func (n *NatsOllamaProxy) embedHandler(req micro.Request) {
	var reqData api.EmbedRequest
	err := json.Unmarshal(req.Data(), &reqData)
	if err != nil {
		log.Error("Error unmarshalling request:", err)
		req.Error("400", err.Error(), nil)
		return
	}

	log.Infof("Embed Request for model: '%s'", reqData.Model)

	err = n.pullMissingModel(err, reqData.Model)
	if err != nil {
		log.Error("Error when checking/pulling a missing model:", err)
		req.Error("500", err.Error(), nil)
		return
	}

	ctx := context.Background()
	resp, err := n.client.Embed(ctx, &reqData)
	if err != nil {
		log.Error("Error calling Ollama:", err)
		req.Error("500", err.Error(), nil)
		return
	}

	responseData, err := json.Marshal(resp)
	if err != nil {
		log.Error("Error marshalling response:", err)
		req.Error("500", err.Error(), nil)
		return
	}
	err = req.Respond(responseData)
}

func (n *NatsOllamaProxy) embeddingHandler(req micro.Request) {
	var reqData api.EmbeddingRequest
	err := json.Unmarshal(req.Data(), &reqData)
	if err != nil {
		req.Error("400", err.Error(), nil)
		return
	}

	ctx := context.Background()
	resp, err := n.client.Embeddings(ctx, &reqData)
	if err != nil {
		req.Error("500", err.Error(), nil)
		return
	}

	responseData, err := json.Marshal(resp)
	if err != nil {
		req.Error("500", err.Error(), nil)
		return
	}
	err = req.Respond(responseData)
}

func (n *NatsOllamaProxy) chatHandler(req micro.Request) {
	var reqData api.ChatRequest
	err := json.Unmarshal(req.Data(), &reqData)
	if err != nil {
		log.Error("Error unmarshalling request:", err)
		req.Error("400", err.Error(), nil)
		return
	}

	// Set streaming to false, thus making sure we wait for a response.
	reqData.Stream = new(bool)

	log.Infof("Chat request for model: '%s'", reqData.Model)
	respFunc := func(resp api.ChatResponse) error {
		responseData, err := json.Marshal(resp)
		if err != nil {
			log.Error("Error marshalling response:", err)
			req.Error("400", err.Error(), nil)
			return err
		}
		err = req.Respond(responseData)
		return err
	}

	err = n.pullMissingModel(err, reqData.Model)
	if err != nil {
		log.Error("Error when checking/pulling a missing model:", err)
		req.Error("500", err.Error(), nil)
		return
	}

	ctxChat := context.Background()
	var chatError error
	sp := spinner.New()
	action := func() {
		chatError = n.client.Chat(ctxChat, &reqData, respFunc)
	}

	err = sp.Title(fmt.Sprintf("Processing chat request for model '%s'...", reqData.Model)).Action(action).Run()

	//err = n.client.Chat(ctx, &reqData, respFunc)
	if chatError != nil {
		log.Error("Error marshalling response:", chatError)
		req.Error("400", chatError.Error(), nil)
	}
	if err != nil {
		log.Error("Error marshalling response:", err)
		req.Error("400", err.Error(), nil)
	}
}

func (n *NatsOllamaProxy) showHandler(req micro.Request) {
	var reqData api.ShowRequest
	err := json.Unmarshal(req.Data(), &reqData)
	if err != nil {
		log.Error("Error unmarshalling request:", err)
		req.Error("400", err.Error(), nil)
		return
	}

	err = n.pullMissingModel(err, reqData.Model)
	if err != nil {
		log.Error("Error when checking/pulling a missing model:", err)
		req.Error("500", err.Error(), nil)
		return
	}

	ctxShow := context.Background()
	var showError error
	var resp *api.ShowResponse
	sp := spinner.New()
	action := func() {
		resp, showError = n.client.Show(ctxShow, &reqData)
	}

	err = sp.Title(fmt.Sprintf("Processing show request for model '%s'...", reqData.Model)).Action(action).Run()
	if showError != nil {
		log.Error("Error on show response:", showError)
		req.Error("400", showError.Error(), nil)
		return
	}

	if err != nil {
		log.Error("Error calling show:", err)
		req.Error("400", err.Error(), nil)
		return
	}

	responseData, err := json.Marshal(resp)
	if err != nil {
		log.Error("Error marshalling response:", err)
		req.Error("400", err.Error(), nil)
		return
	}
	
	err = req.Respond(responseData)
	if err != nil {
		log.Error("Error on sending a show response:", err)
		return
	}
	return
}

func (n *NatsOllamaProxy) pullMissingModel(err error, model string) error {
	ctx := context.Background()
	modelList, err := n.client.List(ctx)
	if err != nil {
		return err
	}

	for _, ml := range modelList.Models {
		if strings.HasPrefix(ml.Model, model) {
			return nil
		}
	}

	ctxPull := context.Background()
	log.Warningf("Model does not exist. Start pulling a new model: '%s'", model)
	sp := spinner.New()
	action := func() {
		err = n.client.Pull(ctxPull, &api.PullRequest{Model: model}, func(response api.ProgressResponse) error {
			if response.Total != 0 {
				sp.Title(fmt.Sprintf("Downloading model '%s', status: '%s'", model, response.Status))
			}
			return nil
		})
	}

	sp.Title(fmt.Sprintf("Downloading model '%s'...", model)).Action(action).Run()

	if err != nil {
		return err
	}
	log.Infof("Pulling of model '%s' complete.", model)

	return nil
}
