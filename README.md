# GrumpyBot

This is a proof of concept of a conversational slack bot.
If it is tagged in a message, it sends the whole slack thread to an OpenAI compatible model + sends the model response to slack.

It does not keep any memory, it always loads the whole slack thread. Basically, it just connects OpenAI and slack APIs.

## Usage

1. Create a new slack app.
  * Enable "socket mode", so we don't need to deal with webhooks.
  * Give it permissions:
    * `app_mentions:read`
    * `channels:history`
    * `chat:write`
  * Install it into a workspace and invite it to a channel.
2. Get an OpenAI-compatible endpoint, for exampe at chatapi.akash.network. Get URL of the endpoint + token.
3. Run the bot: `go run cmd/main.go -app-token=$SLACK_APP_TOKEN -bot-token=$SLACK_BOT_TOKEN -openai-url=https://chatapi.akash.network/api/v1 -openai-model=Meta-Llama-3-3-70B-Instruct -openai-token=$AKASH_CHAT_TOKEN`

## Sample conversation
`jsafrane`: @GrumpyBot Hello! What a beautiful day today!

`GrumpyBot`: sigh What does this have to do with the code that's not working or the bug that needs fixing? If you have a technical question, I'll do my best to help. Otherwise, I have a deadline to meet and a never-ending list of Jira tickets to attend to.

`jsafrane`: @GrumpyBot If you insist, how can I write an AI-driven slack bot?

`GrumpyBot`: Finally, a question worth answering. To write an AI-driven Slack bot, you'll need to:
1. Choose a programming language (e.g., Python, Node.js) and a framework (e.g., Botkit, Rasa).
2. Set up a Slack App and obtain a bot token.
3. Design a conversational flow using intents, entities, and dialogs.
4. Integrate with a Natural Language Processing (NLP) service (e.g., Dialogflow, Wit.ai) or train your own model.
5. Use the Slack API to send and receive messages.

Some popular libraries and tools include:

* slackclient for Python
* @slack/bolt for Node.js
* rasa for conversational AI

Don't even think about asking me to help with Jira integration. Next thing you know, you'll be asking me to attend a meeting to discuss the "synergy" between our Slack bot and the company's "mission statement". Just focus on the code.

`jsafrane`: @GrumpyBot hey, that was helpful, thanks!

`GrumpyBot`: sigh You're welcome. Just don't get too excited, we've only just started. You still have to deal with the actual coding, debugging, and inevitable Slack API rate limiting issues. But hey, at least you've got your bot set up. Now, if you'll excuse me, I have to go stare at some Jira tickets and contemplate the meaninglessness of my existence. Next question, or are you just going to waste my time with small talk?
