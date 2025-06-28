package redis

import (
	"github.com/rs/zerolog/log"
	"strings"
)

func GET(store *KVStore, tokens []string, connectionCounter int) string {
	if len(tokens) > 2 {
		log.Error().Int("connection_counter", connectionCounter).Msgf("GET query can only specify keyname, got %d tokens", len(tokens))
		return ""
	}
	key := strings.Trim(tokens[1], "\n")
	value := store.Get(key)
	if key == "" {
		log.Error().Int("connection_counter", connectionCounter).Msg("GET query has no key specified")
		return ""
	}
	log.Info().Int("connection_counter", connectionCounter).Msgf("GET query for key %s returned value %s", key, value)
	return value
}

func SET(store *KVStore, tokens []string, connectionCounter int) string {
	if len(tokens) > 3 {
		log.Error().Int("connection_counter", connectionCounter).Msgf("SET query can only specify key and value, got %d tokens", len(tokens))
		return ""
	}
	if len(tokens) < 3 {
		log.Error().Int("connection_counter", connectionCounter).Msgf("SET query has too few parameters, got %d tokens", len(tokens))
		return ""
	}
	key := strings.Trim(tokens[1], "\n")
	value := strings.Trim(tokens[2], "\n")
	store.Set(key, value)
	return value
}
