package mlclient

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"os"
)

type MLClient struct {
	openAiClient *openai.Client
}

func NewMLClient() *MLClient {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil
	}

	client := openai.NewClient(apiKey)
	return &MLClient{openAiClient: client}
}

func (c *MLClient) GenerateCoverLetter(ctx context.Context, jobDesc, resumeText string) (string, error) {
	messages := createCoverletterGeneratorPrompt(jobDesc, resumeText)
	response, err := c.openAiClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       openai.GPT3Dot5Turbo,
		Messages:    messages,
		MaxTokens:   2000,
		Temperature: 0.2,
		Stream:      false,
		Stop:        []string{"\n."},
	})
	if err != nil {
		return "", err
	}
	return response.Choices[0].Message.Content, nil
}

func createCoverletterGeneratorPrompt(jobDesc, resume string) []openai.ChatCompletionMessage {
	if jobDesc == "" {
		return []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a tech recruiter who has reviewed thousands of resume and coverletter",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: fmt.Sprintf("Can you create a cover letter for my resume?\nResume:\n%s", resume),
			},
		}
	}
	return []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "You are a tech recruiter who has reviewed thousands of resume and coverletter",
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: fmt.Sprintf("I want to apply for this job:\nJob desc:\n%s\n can you help me write a coverletter for this?", jobDesc),
		},
		{
			Role:    openai.ChatMessageRoleAssistant,
			Content: "Ok, can you share your resume with me?",
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: fmt.Sprintf("Yes, here is the text content of resume:\n %s", resume),
		},
	}
}
