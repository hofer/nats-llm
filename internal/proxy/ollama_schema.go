package proxy

import (
	"encoding/json"

	"github.com/ollama/ollama/api"
	"github.com/swaggest/assertjson"
	"github.com/swaggest/jsonschema-go"
)

type NatsLlmProxySchema struct {
	Request  string `json:"request"`
	Response string `json:"response"`
}

func schema(request any, response any) (*NatsLlmProxySchema, error) {
	reflector := jsonschema.Reflector{}
	reqSchema, err := reflector.Reflect(request)
	if err != nil {
		return nil, err
	}
	reqSchemaStr, _ := assertjson.MarshalIndentCompact(reqSchema, "", " ", 80)

	resSchema, err := reflector.Reflect(response)
	if err != nil {
		return nil, err
	}
	resSchemaStr, _ := assertjson.MarshalIndentCompact(resSchema, "", " ", 80)

	return &NatsLlmProxySchema{
		Request:  string(reqSchemaStr),
		Response: string(resSchemaStr),
	}, nil
}

func marshalSchema(request any, response any) (string, error) {
	serviceSchema, _ := schema(request, response)
	schemaData, err := json.Marshal(serviceSchema)
	if err != nil {
		return "", err
	}
	return string(schemaData), nil
}

func GetSchemaGenerate() (string, error) {
	return marshalSchema(&api.GenerateRequest{}, &api.GenerateResponse{})
}

func GetSchemaEmbed() (string, error) {
	return marshalSchema(&api.EmbedRequest{}, &api.EmbedResponse{})
}

func GetSchemaEmbedding() (string, error) {
	return marshalSchema(&api.EmbeddingRequest{}, &api.EmbeddingResponse{})
}

func GetSchemaChat() (string, error) {
	return marshalSchema(&api.ChatRequest{}, &api.ChatResponse{})
}

func GetSchemaShow() (string, error) {
	return marshalSchema(&api.ShowRequest{}, &api.ShowResponse{})
}
