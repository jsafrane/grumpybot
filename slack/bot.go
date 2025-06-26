package slack

import (
	"context"
	"log"
	"regexp"
	"sort"
	"strings"

	"github.com/openai/openai-go"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

// Bot is a wrapper around the slack client that allows to send messages and listen on events.
// It calls a listener with the content of a thread when it receives a message directly tagging the bot.
type Bot struct {
	slackClient *socketmode.Client
	aiClient    *openai.Client
	logger      *log.Logger
	myID        string
	listener    MessageListener
}

type MessageSender string

const (
	MessageSenderUser MessageSender = "user"
	MessageSenderBot  MessageSender = "bot"
)

type Message struct {
	Sender MessageSender
	Text   string
}

type MessageListener func(ctx context.Context, channel, threadTS string, thread []Message)

var removeUser = regexp.MustCompile(`<@[A-Z0-9]+>`)

func NewBot(logger *log.Logger, botToken, appToken, myID string) *Bot {
	api := slack.New(
		botToken,
		slack.OptionLog(logger),
		//slack.OptionDebug(true),
		slack.OptionAppLevelToken(appToken),
	)

	client := socketmode.New(
		api,
		// socketmode.OptionDebug(true),
		socketmode.OptionLog(logger),
	)
	return &Bot{
		slackClient: client,
		logger:      logger,
		myID:        myID,
	}
}

func (b *Bot) SetMessageListener(listener MessageListener) {
	b.listener = listener
}

func (b *Bot) handleMention(ctx context.Context, event *slackevents.AppMentionEvent) {
	msg := removeUser.ReplaceAllString(event.Text, "")
	msg = strings.TrimSpace(msg)
	b.logger.Printf("Received mention event with msg %s", msg)

	thread := event.ThreadTimeStamp
	if thread == "" {
		// This is a new thread
		thread = event.TimeStamp
	}

	var rsp string
	switch msg {
	case "help":
		rsp = "Hello, I'm a badly written AI bot. I send all messages in a slack thread to an LLM of my author's choice. My context is limited to the current thread. I do not store anything anywhere."
	case "ping":
		rsp = "Pong!"
	default:
		b.sendThreadToListener(ctx, event.Channel, thread)
		return
	}
	_, _, err := b.slackClient.PostMessageContext(
		ctx,
		event.Channel,
		slack.MsgOptionText(rsp, false),
		slack.MsgOptionTS(thread),
	)
	if err != nil {
		b.logger.Printf("failed posting message: %v", err)
		return
	}
}

// Send all messages in the thread to the listener
func (b *Bot) sendThreadToListener(ctx context.Context, channel, threadTS string) {
	threadMessages := b.getThreadMessages(ctx, channel, threadTS)

	msgs := []Message{}
	for _, message := range threadMessages {
		var sender MessageSender
		messageText := message.Text
		if message.User == b.myID {
			sender = MessageSenderBot
		} else {
			sender = MessageSenderUser
			// remove "@GrumpyBot" from the message
			messageText = removeUser.ReplaceAllString(messageText, "")
		}
		msgs = append(msgs, Message{
			Sender: sender,
			Text:   messageText,
		})
	}

	if b.listener != nil {
		b.listener(ctx, channel, threadTS, msgs)
	}
}

// Get all messages in a thread
func (b *Bot) getThreadMessages(ctx context.Context, channel, threadTS string) []slack.Message {
	var threadMessages []slack.Message
	var cursor string
	for {
		history, hasMore, nextCursor, err := b.slackClient.GetConversationRepliesContext(ctx, &slack.GetConversationRepliesParameters{
			ChannelID: channel,
			Timestamp: threadTS,
			Limit:     900,
			Inclusive: true,
			Cursor:    cursor,
		})
		if err != nil {
			b.logger.Printf("failed getting conversation history: %v", err)
			return nil
		}
		b.logger.Printf("Got %d messages in this thread", len(history))

		for _, message := range history {
			if message.ThreadTimestamp == threadTS || message.Timestamp == threadTS {
				threadMessages = append(threadMessages, message)
			}
		}
		if !hasMore {
			break
		}
		cursor = nextCursor
	}
	// Sort messages by timestamp to get chronological order
	sort.Slice(threadMessages, func(i, j int) bool {
		return threadMessages[i].Timestamp < threadMessages[j].Timestamp
	})
	return threadMessages
}

func (b *Bot) Run(ctx context.Context) {
	go b.slackClient.RunContext(ctx)

	// Process events from slack
	for {
		select {
		case <-ctx.Done():
			return
		case evt := <-b.slackClient.Events:
			switch evt.Type {
			case socketmode.EventTypeConnecting:
				b.logger.Println("Connecting to Slack with Socket Mode...")
			case socketmode.EventTypeConnectionError:
				b.logger.Println("Connection failed. Retrying later...")
			case socketmode.EventTypeConnected:
				b.logger.Println("Connected to Slack with Socket Mode.")
			case socketmode.EventTypeEventsAPI:
				eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
				if !ok {
					b.logger.Printf("Ignored %+v\n", evt)
					continue
				}

				b.slackClient.Ack(*evt.Request)

				switch eventsAPIEvent.Type {
				case slackevents.CallbackEvent:
					innerEvent := eventsAPIEvent.InnerEvent
					switch ev := innerEvent.Data.(type) {
					case *slackevents.AppMentionEvent:
						b.handleMention(ctx, ev)
					case *slackevents.MemberJoinedChannelEvent:
						b.logger.Printf("user %q joined to channel %q", ev.User, ev.Channel)
					}
				default:
					b.logger.Printf("unsupported Events API event received")
				}

			case socketmode.EventTypeHello:
				b.logger.Printf("Hello received!")
			default:
				b.logger.Printf("Unexpected event type received: %s\n", evt.Type)
			}
		}
	}
}

func (b *Bot) SendMessage(ctx context.Context, channel, threadTS, message string) error {
	_, _, err := b.slackClient.PostMessageContext(
		ctx,
		channel,
		slack.MsgOptionText(message, false),
		slack.MsgOptionTS(threadTS),
	)
	return err
}
