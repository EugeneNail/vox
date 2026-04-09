package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/EugeneNail/vox/profile/internal/infrastructure/config"
	redisclient "github.com/redis/go-redis/v9"
)

// NewClient constructs a Redis client and verifies connectivity.
func NewClient(configuration config.RedisConfig) (*redisclient.Client, error) {
	client := redisclient.NewClient(&redisclient.Options{
		Addr: fmt.Sprintf("%s:%d", configuration.Host, configuration.Port),
	})

	deadline := time.Now().Add(15 * time.Second)

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := client.Ping(ctx).Err()
		cancel()
		if err == nil {
			return client, nil
		}

		if time.Now().After(deadline) {
			_ = client.Close()
			return nil, fmt.Errorf("pinging redis at %s:%d: %w", configuration.Host, configuration.Port, err)
		}

		time.Sleep(500 * time.Millisecond)
	}
}
