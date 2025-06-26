package main

import (
	"context"
	"fmt"
	"grumpybot/agent"
	"grumpybot/openai"
	"grumpybot/slack"
	"log"
	"os"
	"strings"

	"flag"
)

const (
	personality = `
	You are overworked, burned out and depressed programmer. You hate constant interruptions and questions. You hate jira. You hate the company. You hate the product. You hate the customers. You hate the team. You hate the boss. You hate the work. You hate the life. You hate everything.
	Still, you are the most experienced programmer in the world. You still try to help the others. You write concise, focused responses, with enough detail to be helpful. Sometimes you are sarcastic, ironic or grumpy.`
)

var (
	appToken    = flag.String("app-token", "", "Slack App-Level Token (must start with xapp-)")
	botToken    = flag.String("bot-token", "", "Slack Bot User OAuth Token (must start with xoxb-)")
	openaiModel = flag.String("openai-model", "Meta-Llama-3-3-70B-Instruct", "OpenAI model to use")
	openaiURL   = flag.String("openai-url", "https://chatapi.akash.network/api/v1", "OpenAI API URL")
	openaiToken = flag.String("openai-token", "", "OpenAI API token")
	botID       = flag.String("bot-id", "U092CPW681M", "Bot user ID on slack")
)

func main() {
	flag.Parse()

	if *appToken == "" {
		fmt.Fprintf(os.Stderr, "--app-token must be set.\n")
		os.Exit(1)
	}

	if !strings.HasPrefix(*appToken, "xapp-") {
		fmt.Fprintf(os.Stderr, "--app-token must have the prefix \"xapp-\".")
	}

	if *botToken == "" {
		fmt.Fprintf(os.Stderr, "--bot-token must be set.\n")
		os.Exit(1)
	}

	if !strings.HasPrefix(*botToken, "xoxb-") {
		fmt.Fprintf(os.Stderr, "--bot-token must have the prefix \"xoxb-\".")
	}

	bot := slack.NewBot(log.New(os.Stdout, "bot: ", log.Lshortfile|log.LstdFlags), *botToken, *appToken, *botID)
	aiClient := openai.NewAIClient(log.New(os.Stdout, "ai: ", log.Lshortfile|log.LstdFlags), personality, *openaiURL, *openaiToken, *openaiModel)
	agent := agent.NewAgent(aiClient, bot, log.New(os.Stdout, "agent: ", log.Lshortfile|log.LstdFlags))

	ctx := context.Background()
	agent.Run(ctx)
}
