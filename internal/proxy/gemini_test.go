package proxy

import (
	"encoding/json"
	"github.com/nats-io/nats.go/micro"
	"github.com/ollama/ollama/api"
)

//func TestStartNatsGeminiProxy(t *testing.T) {
//	proxy := NewNatsGeminiProxy(os.Getenv("GEMINI_API_KEY"))
//
//	req := &DummyRequest{}
//	proxy.chatHandler(req)
//}

//
//func TestCreateUserContentParts(8 *testing.T) {
//	create
//}

type DummyRequest struct {
}

// Respond sends the response for the request.
// Additional headers can be passed using [WithHeaders] option.
func (r *DummyRequest) Respond([]byte, ...micro.RespondOpt) error {
	return nil
}

// RespondJSON marshals the given response value and responds to the request.
// Additional headers can be passed using [WithHeaders] option.
func (r *DummyRequest) RespondJSON(any, ...micro.RespondOpt) error {
	return nil
}

// Error prepares and publishes error response from a handler.
// A response error should be set containing an error code and description.
// Optionally, data can be set as response payload.
func (r *DummyRequest) Error(code, description string, data []byte, opts ...micro.RespondOpt) error {
	return nil
}

// Data returns request data.
func (r *DummyRequest) Data() []byte {
	reqData := api.ChatRequest{
		Model: "gemini-2.5-flash",
		Messages: []api.Message{
			{Content: "What is your name?", Role: "user"},
		},
	}
	data, _ := json.Marshal(reqData)
	return data
}

// Headers returns request headers.
func (r *DummyRequest) Headers() micro.Headers {
	return nil
}

// Subject returns underlying NATS message subject.
func (r *DummyRequest) Subject() string {
	return ""
}

// Reply returns underlying NATS message reply subject.
func (r *DummyRequest) Reply() string {
	return ""
}

//func TestGeminiShow(t *testing.T) {
//	ctx := context.Background()
//	client, _ := genai.NewClient(ctx, &genai.ClientConfig{
//		APIKey: os.Getenv("GEMINI_API_KEY"),
//	})
//	defer client.Close()
//
//	model := client.Models.Get()GenerativeModel("gemini-2.5-flash")
//	//model.CountTokens()
//
//	modelInfo, err := model.Info(context.Background())
//	if err != nil {
//		t.Error(err)
//	}
//	for _, genMethod := range modelInfo.SupportedGenerationMethods {
//		fmt.Println(genMethod)
//	}
//
//	contextLength := modelInfo.InputTokenLimit
//	fmt.Println(contextLength)
//
//	family := modelInfo.BaseModelID
//	fmt.Println(family)
//}
