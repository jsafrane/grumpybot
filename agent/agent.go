package agent

import (
	"context"
	"grumpybot/openai"
	"grumpybot/slack"
	"log"
)

type Agent struct {
	aiClient *openai.AIClient
	slackBot *slack.Bot
	logger   *log.Logger
}

func NewAgent(aiClient *openai.AIClient, slackBot *slack.Bot, logger *log.Logger) *Agent {
	agent := &Agent{
		aiClient: aiClient,
		logger:   logger,
		slackBot: slackBot,
	}
	return agent
}

func (a *Agent) handleMessage(ctx context.Context, channel, threadTS string, thread []slack.Message) {
	go func() {
		// send the messages to the AI client
		response, err := a.aiClient.AskAI(ctx, thread)
		if err != nil {
			a.logger.Printf("Error sending message to the AI client: %v", err)
			return
		}

		a.slackBot.SendMessage(ctx, channel, threadTS, response)
	}()
}

func (a *Agent) Run(ctx context.Context) {
	a.slackBot.SetMessageListener(a.handleMessage)
	a.slackBot.Run(ctx)
}
