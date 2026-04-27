package redisstream

import "os"

// BuildConsumerName returns a consumer instance name, not a service name.
// Redis Streams uses it to distinguish concrete running replicas inside one consumer group.
func BuildConsumerName(fallback string) string {
	hostName, err := os.Hostname()
	if err != nil || hostName == "" {
		return fallback
	}

	return hostName
}
