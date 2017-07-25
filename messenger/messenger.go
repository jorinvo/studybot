// Package messenger implements the Messenger bot
// and handles all the user interaction.
package messenger

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/jorinvo/studybot/brain"
	"github.com/jorinvo/studybot/fbot"
)

// Feedback describes a message from a user a human has to react to
type Feedback struct {
	ChatID   int64
	Username string
	Message  string
}

// Bot is a messenger bot handling webhook events and notifications.
// Use New to setup and use register Bot as a http.Handler.
type Bot struct {
	store        brain.Store
	setup        bool
	err          *log.Logger
	info         *log.Logger
	client       fbot.Client
	verifyToken  string
	feedback     chan<- Feedback
	notifyTimers map[int64]*time.Timer
	http.Handler
}

// Setup sends greetings and the getting started message to Facebook.
func Setup(b *Bot) {
	b.setup = true
}

// LogInfo is an option to set the info logger of the bot.
func LogInfo(l *log.Logger) func(*Bot) {
	return func(b *Bot) {
		b.info = l
	}
}

// LogErr is an option to set the error logger of the bot.
func LogErr(l *log.Logger) func(*Bot) {
	return func(b *Bot) {
		b.err = l
	}
}

// Verify is an option to enable verification of the webhook.
func Verify(token string) func(*Bot) {
	return func(b *Bot) {
		b.verifyToken = token
	}
}

// GetFeedback sets up user feedback to be sent to the given channel.
func GetFeedback(f chan<- Feedback) func(*Bot) {
	return func(b *Bot) {
		b.feedback = f
	}
}

// Notify enables sending notifications when studies are ready.
func Notify(b *Bot) {
	b.notifyTimers = map[int64]*time.Timer{}
}

// New creates a Bot.
// It can be used as a HTTP handler for the webhook.
// The options Setup, LogInfo, LogErr, Notify, Verify, GetFeedback can be used.
func New(store brain.Store, token string, options ...func(*Bot)) (Bot, error) {
	client := fbot.New(token)
	b := Bot{
		store:  store,
		client: client,
	}

	for _, option := range options {
		option(&b)
	}
	if b.info == nil {
		b.info = log.New(ioutil.Discard, "", 0)
	}
	if b.err == nil {
		b.err = log.New(ioutil.Discard, "", 0)
	}
	b.Handler = client.Webhook(b.HandleEvent, b.verifyToken)

	if b.setup {
		if err := b.client.SetGreetings(map[string]string{"default": greeting}); err != nil {
			return b, fmt.Errorf("failed to set greeting: %v", err)
		}
		b.info.Println("Greeting set")
		if err := b.client.SetGetStartedPayload(payloadGetStarted); err != nil {
			return b, fmt.Errorf("failed to enable Get Started button: %v", err)
		}
		b.info.Printf("Get Started button activated")
	}

	if b.notifyTimers != nil {
		if err := b.store.EachActiveChat(b.scheduleNotify); err != nil {
			return b, err
		}
		b.info.Println("Notifications enabled")
	}

	return b, nil
}

// SendMessage sends a message to a specific user.
func (b Bot) SendMessage(id int64, msg string) error {
	if err := b.client.Send(id, msg, nil); err != nil {
		return err
	}
	b.send(b.messageStartMenu(id))
	return nil
}
