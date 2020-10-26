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

	// whenever the message written is "ping", the bot replies "pong"
	if msg.Content == "ping" {
		if _, err := msg.Reply(noCtx, s, "pong"); err != nil {
			log.Error(fmt.Errorf("failed to reply to ping. %w", err))
		}
	}
}

func main() {
	const prefix = "!"

	client := disgord.New(disgord.Config{
		ProjectName: "MyBot",
		BotToken:    os.Getenv("DISCORD_TOKEN"),
		Logger:      log,
		IgnoreEvents: []string{
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
	})
	defer client.StayConnectedUntilInterrupted(context.Background())

	logFilter, _ := std.NewLogFilter(client)
	filter, _ := std.NewMsgFilter(context.Background(), client)
	filter.SetPrefix(prefix)

	// create a handler and bind it to new message events
	client.
		Event().
		WithMdlw(
			filter.NotByBot,    // ignore bot messages
			filter.HasPrefix,   // read original
			logFilter.LogMsg,         // log command message
			filter.StripPrefix, // write copy
		).
		MessageCreate(replyPongToPing)
}
