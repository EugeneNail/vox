package redis

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	redisclient "github.com/redis/go-redis/v9"
)

const messageEventsConsumerGroup = "message-service"

type streamPayloadHandler func(context.Context, string) bool

func listenAndConsumeStream(
	ctx context.Context,
	client *redisclient.Client,
	stream string,
	consumerName string,
	handler streamPayloadHandler,
) error {
	if err := ensureStreamGroup(ctx, client, stream); err != nil {
		return err
	}

	for {
		streams, err := client.XReadGroup(ctx, &redisclient.XReadGroupArgs{
			Group:    messageEventsConsumerGroup,
			Consumer: consumerName,
			Streams:  []string{stream, ">"},
			Count:    10,
			Block:    5 * time.Second,
		}).Result()
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}

			if errors.Is(err, redisclient.Nil) {
				continue
			}

			return err
		}

		for _, streamResult := range streams {
			for _, message := range streamResult.Messages {
				shouldAck := handleStreamMessage(ctx, stream, message, handler)
				if !shouldAck {
					continue
				}

				if err := client.XAck(ctx, stream, messageEventsConsumerGroup, message.ID).Err(); err != nil {
					log.Printf("acknowledging %s event %q: %v", stream, message.ID, err)
				}
			}
		}
	}
}

func ensureStreamGroup(ctx context.Context, client *redisclient.Client, stream string) error {
	if err := client.XGroupCreateMkStream(ctx, stream, messageEventsConsumerGroup, "0").Err(); err != nil {
		if strings.Contains(err.Error(), "BUSYGROUP") {
			return nil
		}

		return err
	}

	return nil
}

func handleStreamMessage(ctx context.Context, stream string, message redisclient.XMessage, handler streamPayloadHandler) bool {
	rawPayload, ok := message.Values["payload"]
	if !ok {
		log.Printf("%s stream entry %q does not contain payload", stream, message.ID)
		return true
	}

	payload, ok := rawPayload.(string)
	if !ok {
		log.Printf("%s stream entry %q contains non-string payload", stream, message.ID)
		return true
	}

	return handler(ctx, payload)
}

func buildConsumerName() string {
	hostName, err := os.Hostname()
	if err != nil || hostName == "" {
		return "message-service"
	}

	return hostName
}
