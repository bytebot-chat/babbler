package main

import (
	"context"
	"flag"
	"math/rand"
	"sync"
	"time"

	disco "github.com/bytebot-chat/gateway-discord/model"
	irc "github.com/bytebot-chat/gateway-irc/model"
	"github.com/go-redis/redis/v8"
	"github.com/mb-14/gomarkov"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var ircInbound stringArrayFlags
var ircOutbound stringArrayFlags
var discordInbound stringArrayFlags
var discordOutbound stringArrayFlags

var addr = flag.String("redis", "localhost:6379", "Redis server address")

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	flag.Var(&ircInbound, "irc-inbound", "IRC topic to listen to. May be repeated. Example: -irc-inbound=irc1 -irc-inbound=irc2")
	flag.Var(&ircOutbound, "irc-outbound", "IRC topic to publish to. May be repeated. Example: -irc-outbound=irc1 -irc-outbound=irc2")

	flag.Var(&discordInbound, "discord-inbound", "Discord topic to listen to. May be repeated. Example: -discord-inbound=discord1 -discord-inbound=discord2")
	flag.Var(&discordOutbound, "discord-outbound", "Discord topic to publish to. May be repeated. Example: -discord-outbound=discord1 -discord-outbound=discord2")
}

func main() {

	flag.Parse()

	log.Info().
		Str("Redis address", *addr).
		Msg("Bytebot Party Pack starting up!")

	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr: *addr,
		DB:   0,
	})

	err := rdb.Ping(ctx).Err()
	if err != nil {
		log.Warn().Msg("Ping timeout, trying to connect to redis again...")
		time.Sleep(3 * time.Second)
		err := rdb.Ping(ctx).Err()
		if err != nil {
			log.Fatal().Err(err).
				Msg("Couldn't connect to redis server")
		}
	}

	log.Info().Msg("Initializing Markov model...")
	model := model{
		chain: gomarkov.NewChain(1),
		mutex: &sync.RWMutex{},
	}
	model.train("")

	log.Info().Msg("Subscribing to topics...")
	var wg sync.WaitGroup

	for _, topic := range ircInbound {
		log.Info().Msg("Launching worker for " + topic + "...")
		wg.Add(1)
		go subscribeIRC(ctx, &wg, rdb, topic, ircOutbound, model)
	}

	for _, topic := range discordInbound {
		log.Info().Msg("Launching worker for " + topic + "...")
		wg.Add(1)
		go subscribeDiscord(ctx, &wg, rdb, topic, discordOutbound, model)
	}

	log.Info().Msg("Workers launched. Listening for messages.")
	wg.Wait()

}

func subscribeIRC(ctx context.Context, wg *sync.WaitGroup, rdb *redis.Client, topic string, outbound []string, model model) {
	defer wg.Done()

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	log.Info().Msg("Subscribing to " + topic)
	sub := rdb.Subscribe(ctx, topic)
	log.Info().Msg("Subscribed!")
	channel := sub.Channel()
	for msg := range channel {
		m := &irc.Message{}
		err := m.Unmarshal([]byte(msg.Payload))
		if err != nil {
			log.Error().
				Str("message payload", msg.Payload).
				Err(err)
		}
		log.Debug().
			RawJSON("Received message", []byte(msg.Payload)).
			Msg("Received message")

		if r1.Intn(100) < 2 {
			for _, q := range outbound {
				replyIRC(ctx, *m, rdb, q, model.babble())
			}
		}
		// Begin training the model
		model.train(m.Content)
	}
}

func subscribeDiscord(ctx context.Context, wg *sync.WaitGroup, rdb *redis.Client, topic string, outbound []string, model model) {
	defer wg.Done()

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	log.Info().Msg("Subscribing to " + topic)
	sub := rdb.Subscribe(ctx, topic)
	log.Info().Msg("Subscribed!")
	channel := sub.Channel()
	for msg := range channel {
		m := &disco.Message{}
		err := m.Unmarshal([]byte(msg.Payload))
		if err != nil {
			log.Error().
				Str("message payload", msg.Payload).
				Err(err)
		}
		log.Debug().
			RawJSON("Received message", []byte(msg.Payload)).
			Msg("Received message")

		if r1.Intn(100) < 2 {
			for _, q := range outbound {
				replyDiscord(ctx, *m, rdb, q, model.babble())
			}
		}
		// Begin training the model
		model.train(m.Content)
	}
}
