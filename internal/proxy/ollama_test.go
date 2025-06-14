package proxy

//
//func TestOllamaClientShow(t *testing.T) {
//	client, _ := api.ClientFromEnvironment()
//	showResult, _ := client.Show(context.Background(), &api.ShowRequest{
//		//Model: "magistral:24b",
//		Model: "gemma3:12b",
//		//Model: "snowflake-arctic-embed2:latest",
//	})
//
//	family := showResult.Details.Family
//	fmt.Println(family)
//
//	parameterSize := showResult.Details.ParameterSize
//	fmt.Println(parameterSize)
//
//	contextLength := showResult.ModelInfo[fmt.Sprintf("%s.context_length", family)]
//	fmt.Println(contextLength)
//
//	embeddingLength := showResult.ModelInfo[fmt.Sprintf("%s.embedding_length", family)]
//	fmt.Println(embeddingLength)
//
//	tokenizerModel := showResult.ModelInfo["tokenizer.ggml.model"]
//	fmt.Println(tokenizerModel)
//
//	for _, capability := range showResult.Capabilities {
//		fmt.Println(capability)
//	}
//}
