package redisstream

import "os"

// BuildConsumerName returns a stable consumer name for the current host or a provided fallback.
func BuildConsumerName(fallback string) string {
	hostName, err := os.Hostname()
	if err != nil || hostName == "" {
		return fallback
	}

	return hostName
}
