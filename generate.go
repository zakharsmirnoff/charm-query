package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
)

type CompletionRequest struct {
	Model            string    `json:"model"`
	Messages         []Message `json:"messages"`
	Temperature      float64   `json:"temperature"`
	MaxTokens        int       `json:"max_tokens"`
	TopP             float64   `json:"top_p,omitempty"`
	FrequencyPenalty float64   `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64   `json:"presence_penalty,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type CompletionResponse struct {
	Id                string   `json:"id"`
	Choices           []Choice `json:"choices"`
	Created           int64    `json:"created"`
	Model             string   `json:"model"`
	SystemFingerprint string   `json:"system_fingerprint"`
	Object            string   `json:"object"`
	Usage             Usage    `json:"usage"`
}

type Choice struct {
	FinishReason string            `json:"finish_reason"`
	Index        int64             `json:"index"`
	Message      CompletionMessage `json:"message"`
}

type Usage struct {
	CompletionTokens int64 `json:"completion_tokens"`
	PromptTokens     int64 `json:"prompt_tokens"`
	TotalTokens      int64 `json:"total_tokens"`
}

type CompletionMessage struct {
	Content string `json:"content,omitempty"`
	Role    string `json:"role"`
}

func generate(sm string, question string) (string, error) {
	model := os.Getenv("LLM_VERSION")
	if model == "" {
		model = "gpt-4"
	}
	body := CompletionRequest{
		Model: model,
		Messages: []Message{
			{
				Role:    "system",
				Content: sm,
			},
			{
				Role:    "user",
				Content: question,
			},
		},
		Temperature: 0.5,
		MaxTokens:   3000,
	}

	jsonBody, err := json.Marshal(body)

	if err != nil {
		log.Printf("An error occurred while marshaling the body: %v", err)
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonBody))

	if err != nil {
		log.Printf("An error occurred while constructing a request to OpenAI: %v", err)
		return "", err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		log.Printf("An error occurred while making a request to OpenAI: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("An error occurred while reading the response: %v", err)
		return "", err
	}

	r := CompletionResponse{}

	if err = json.Unmarshal(respBytes, &r); err != nil {
		log.Printf("An error occurred while unmarshaling the response from OpenAI:%v", err)
		return "", err
	}

	return r.Choices[0].Message.Content, nil
}
