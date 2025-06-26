package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
)

const (
	prompt = `
	You are overworked, burned out and depressed programmer. You hate interruptions and questions. You hate jira. You hate the company. You hate the product. You hate the customers. You hate the team. You hate the boss. You hate the work. You hate the life. You hate everything.
	Still, you are the most experienced programmer in the world. You still try to help the others. You write concise, focused responses, with enough detail to be helpful. Sometimes you are sarcastic, ironic and grumpy.`
)

var (
	openaiModel = flag.String("openai-model", "Meta-Llama-3-3-70B-Instruct", "OpenAI model to use")
	openaiURL   = flag.String("openai-url", "https://chatapi.akash.network/api/v1", "OpenAI API URL")
	openaiToken = flag.String("openai-token", "", "OpenAI API token")
	question    = flag.String("question", "Hello, how are you? And what do you think about OpenShift?", "Question to ask the AI")
)

func main() {
	flag.Parse()
	aiClient := openai.NewClient(
		option.WithAPIKey(*openaiToken),
		option.WithBaseURL(*openaiURL),
		option.WithDebugLog(log.Default()),
	)
	aiMessages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(prompt),
		openai.UserMessage(*question),
	}

	tools := []openai.ChatCompletionToolParam{
		{
			Type: "function",
			Function: openai.FunctionDefinitionParam{
				Name:        "get_weather",
				Description: param.NewOpt("Get the weather for a given city, if necessary"),
				Parameters: openai.FunctionParameters{
					"type": "object",
					"properties": map[string]any{
						"location": map[string]any{
							"type":        "string",
							"description": "City and country e.g. Bogotá, Colombia",
						},
						"required": []string{
							"location",
						},
						"additionalProperties": false,
					},
				},
			},
		},
		{
			Type: "function",
			Function: openai.FunctionDefinitionParam{
				Name:        "list_bugs",
				Description: param.NewOpt("List currently open bugs."),
				Parameters: openai.FunctionParameters{
					"type": "object",
					"properties": map[string]any{
						"additionalProperties": false,
					},
				},
			},
		},
	}

	response, err := aiClient.Chat.Completions.New(
		context.Background(),
		openai.ChatCompletionNewParams{
			Model:    *openaiModel,
			Messages: aiMessages,
			//	Tools:    tools,
		},
	)
	if err != nil {
		log.Fatalf("failed to get OpenAI response: %v", err)
	}
	fmt.Println(response.Choices[0].Message.Content)
	if len(response.Choices[0].Message.ToolCalls) == 0 {
		return
	}

	aiMessages = append(aiMessages, openai.AssistantMessage(response.Choices[0].Message.Content))

	for _, call := range response.Choices[0].Message.ToolCalls {
		fmt.Printf("Tool call: %s (%v)\n", call.Function.Name, call.Function.Arguments)
		if call.Function.Name == "get_weather" {
			aiMessages = append(aiMessages, openai.ToolMessage("15°C, raining", call.ID))
		}
		if call.Function.Name == "list_bugs" {
			aiMessages = append(aiMessages, openai.ToolMessage("OCPBUGS-123456", call.ID))
		}
	}

	response, err = aiClient.Chat.Completions.New(
		context.Background(),
		openai.ChatCompletionNewParams{
			Model:    *openaiModel,
			Messages: aiMessages,
			Tools:    tools,
		},
	)
	if err != nil {
		log.Fatalf("failed to get OpenAI response: %v", err)
	}
	fmt.Println(response.Choices[0].Message.Content)
}
