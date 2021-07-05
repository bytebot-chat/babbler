package main

import (
	"context"
	"encoding/json"
	"strings"

	disco "github.com/bytebot-chat/gateway-discord/model"
	irc "github.com/bytebot-chat/gateway-irc/model"
	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"
	"github.com/satori/go.uuid"
)

func replyIRC(ctx context.Context, m irc.Message, rdb *redis.Client, topic, reply string) {
	if !strings.HasPrefix(m.To, "#") { // DMs go back to source, channel goes back to channel
		m.To = m.From
	}
	m.From = ""
	m.Metadata.Dest = m.Metadata.Source
	m.Metadata.Source = "babbler"
	m.Content = reply
	m.Metadata.ID = uuid.Must(uuid.NewV4(), *new(error))
	stringMsg, _ := json.Marshal(m)
	rdb.Publish(ctx, topic, stringMsg)

	log.Debug().
		RawJSON("message", stringMsg).
		Msg("Reply")

	return
}

func replyDiscord(ctx context.Context, m disco.Message, rdb *redis.Client, topic, reply string) {
	metadata := disco.Metadata{
		Dest:   m.Metadata.Source,
		Source: "babbler",
		ID:     uuid.Must(uuid.NewV4(), *new(error)),
	}

	stringMsg, _ := m.MarshalReply(metadata, m.ChannelID, reply)
	rdb.Publish(ctx, topic, stringMsg)
	log.Debug().
		RawJSON("message", stringMsg).
		Msg("Reply")

	return
}

type stringArrayFlags []string

func (i *stringArrayFlags) String() string {
	return "String array flag"
}

func (i *stringArrayFlags) Set(s string) error {
	*i = append(*i, s)
	return nil
}
