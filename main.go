package main

import (
	"context"
	"fmt"
	"os"

	"github.com/andersfylling/disgord"
	"github.com/andersfylling/disgord/std"
	"github.com/sirupsen/logrus"
)

var log = &logrus.Logger{
	Out:       os.Stderr,
	Formatter: new(logrus.TextFormatter),
	Hooks:     make(logrus.LevelHooks),
	Level:     logrus.ErrorLevel,
}

var noCtx = context.Background()

// replyPongToPing is a handler that replies pong to ping messages
func replyPongToPing(s disgord.Session, data *disgord.MessageCreate) {
	msg := data.Message
	if msg.Content != "ping" {
		return
	}

	// whenever the message written is "ping", the bot replies "pong"
	if _, err := msg.Reply(noCtx, s, "pong"); err != nil {
		log.Error(fmt.Errorf("failed to reply to ping. %w", err))
	}
}

func main() {
	const prefix = "!"

	client := disgord.New(disgord.Config{
		ProjectName: "MyBot",
		BotToken:    os.Getenv("DISCORD_TOKEN"),
		Logger:      log,
		RejectEvents: []string{
			// rarely used, and causes unnecessary spam
			disgord.EvtTypingStart,

			// these require special privilege
			// https://discord.com/developers/docs/topics/gateway#privileged-intents
			disgord.EvtPresenceUpdate,
			disgord.EvtGuildMemberAdd,
			disgord.EvtGuildMemberUpdate,
			disgord.EvtGuildMemberRemove,
		},
		Presence: &disgord.UpdateStatusPayload{
			Game: &disgord.Activity{
				Name: "write " + prefix + "ping",
			},
		},
		// I WANT DM EVENTS
		// these are events sent directly to your bot through
		// direct messaging.
		// I recommend just not using these unless you have a specific use case.
		// WARNING: you only specify intents when you want DM capabilities. 
		//  For anything else, populate the RejectEvents setting above.
		//Intents: []disgord.Intent{
		//	disgord.IntentDirectMessageReactions,
		//	disgord.IntentDirectMessageTyping,
		//	disgord.IntentDirectMessages,
		//},
	})
	defer client.Gateway().StayConnectedUntilInterrupted()

	logFilter, _ := std.NewLogFilter(client)
	filter, _ := std.NewMsgFilter(context.Background(), client)
	filter.SetPrefix(prefix)

	// create a handler and bind it to new message events
	// thing about the middlewares are whitelists or passthrough functions.
	client.
		Gateway().
		WithMiddleware(
			filter.NotByBot,    // ignore bot messages
			filter.HasPrefix,   // message must have the given prefix
			logFilter.LogMsg,   // log command message
			filter.StripPrefix, // remove the command prefix from the message
		).
		MessageCreate(replyPongToPing)
}
