package proxy

import "testing"

func TestGetGeminiSchemaChat(t *testing.T) {
	_, err := GetGeminiSchemaChat()
	if err != nil {
		t.Errorf("GetGeminiSchemaChat() error = %v", err)
	}
}
