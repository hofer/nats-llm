package proxy

//import (
//	"context"
//	"fmt"
//	"github.com/google/generative-ai-go/genai"
//	"google.golang.org/api/option"
//	"os"
//	"testing"
//)
//
//func TestGeminiShow(t *testing.T) {
//	ctx := context.Background()
//	client, _ := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
//	defer client.Close()
//
//	model := client.GenerativeModel("gemini-2.5-pro-exp-03-25")
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
