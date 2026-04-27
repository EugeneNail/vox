package redisstream

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	redisclient "github.com/redis/go-redis/v9"
)

// PayloadHandler handles a raw payload read from a Redis stream.
type PayloadHandler func(context.Context, string) bool

// ListenAndConsume reads events from a Redis stream consumer group and acknowledges handled entries.
func ListenAndConsume(
	ctx context.Context,
	client *redisclient.Client,
	stream string,
	consumerGroup string,
	consumerName string,
	handler PayloadHandler,
) error {
	if err := ensureStreamGroup(ctx, client, stream, consumerGroup); err != nil {
		return fmt.Errorf("ensuring %q consumer group for stream %q: %w", consumerGroup, stream, err)
	}

	for {
		streams, err := client.XReadGroup(ctx, &redisclient.XReadGroupArgs{
			Group:    consumerGroup,
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

			return fmt.Errorf("reading stream %q for consumer group %q: %w", stream, consumerGroup, err)
		}

		for _, streamResult := range streams {
			for _, message := range streamResult.Messages {
				shouldAck := handleStreamMessage(ctx, stream, message, handler)
				if !shouldAck {
					continue
				}

				if err := client.XAck(ctx, stream, consumerGroup, message.ID).Err(); err != nil {
					log.Printf("acknowledging %s event %q for group %q: %v", stream, message.ID, consumerGroup, err)
				}
			}
		}
	}
}

func ensureStreamGroup(ctx context.Context, client *redisclient.Client, stream string, consumerGroup string) error {
	if err := client.XGroupCreateMkStream(ctx, stream, consumerGroup, "0").Err(); err != nil {
		if strings.Contains(err.Error(), "BUSYGROUP") {
			return nil
		}

		return err
	}

	return nil
}

func handleStreamMessage(ctx context.Context, stream string, message redisclient.XMessage, handler PayloadHandler) bool {
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
