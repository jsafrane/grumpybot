package openai

import (
	"context"
	"errors"
	"grumpybot/slack"
	"log"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// AIClient is a wrapper around the openai client that allows to send messages to the AI.
// It also allows to set a personality for the AI.
type AIClient struct {
	aiClient    *openai.Client
	logger      *log.Logger
	model       string
	personality string
}

func NewAIClient(logger *log.Logger, personality, url, token, model string) *AIClient {
	aiClient := openai.NewClient(
		option.WithAPIKey(token),
		option.WithBaseURL(url),
		option.WithDebugLog(logger),
	)

	return &AIClient{
		aiClient:    &aiClient,
		logger:      logger,
		model:       model,
		personality: personality,
	}
}

func (c *AIClient) AskAI(ctx context.Context, messages []slack.Message) (string, error) {
	aiMessages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(c.personality),
	}

	// Squash the messages by the same sender into a single string.
	// models.corp expect messages in a conversation to alternate between user and assistant.
	accumulatedSender := slack.MessageSender("")
	accumulatedText := ""
	for _, msg := range messages {
		if msg.Sender == accumulatedSender {
			accumulatedText = accumulatedText + "\n" + msg.Text
			continue
		}
		// New sender detected, add the accumulated text to the conversation.
		switch accumulatedSender {
		case slack.MessageSenderUser:
			aiMessages = append(aiMessages, openai.UserMessage(accumulatedText))
		case slack.MessageSenderBot:
			aiMessages = append(aiMessages, openai.AssistantMessage(accumulatedText))
		default:
			// ignore possible accumulatedSender==""
		}
		accumulatedText = msg.Text
		accumulatedSender = msg.Sender
	}
	// add the last messages to the conversation
	switch accumulatedSender {
	case slack.MessageSenderUser:
		aiMessages = append(aiMessages, openai.UserMessage(accumulatedText))
	case slack.MessageSenderBot:
		aiMessages = append(aiMessages, openai.AssistantMessage(accumulatedText))
	default:
		// ignore possible accumulatedSender==""
	}

	response, err := c.aiClient.Chat.Completions.New(
		ctx,
		openai.ChatCompletionNewParams{
			Model:    c.model,
			Messages: aiMessages,
		},
	)
	if err != nil {
		return "", err
	}

	if len(response.Choices) == 0 {
		return "", errors.New("no choices returned from OpenAI")
	}

	return response.Choices[0].Message.Content, nil
}
