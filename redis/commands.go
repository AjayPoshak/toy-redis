package redis

import (
	"github.com/rs/zerolog/log"
	"strings"
)

func GET(store *KVStore, tokens []string, connectionCounter int64) string {
	if len(tokens) < 2 {
		log.Error().Int64("connection_counter", connectionCounter).Msgf("GET query requires a key, got %d tokens", len(tokens))
		return ""
	}
	if len(tokens) > 2 {
		log.Error().Int64("connection_counter", connectionCounter).Msgf("GET query can only specify keyname, got %d tokens", len(tokens))
		return ""
	}
	key := strings.Trim(tokens[1], "\n")
	if key == "" {
		log.Error().Int64("connection_counter", connectionCounter).Msg("GET query has no key specified")
		return ""
	}
	value := store.Get(key)
	log.Info().Int64("connection_counter", connectionCounter).Msgf("GET query for key %s returned value %s", key, value)
	return value
}

func SET(store *KVStore, tokens []string, connectionCounter int64) string {
	if len(tokens) > 3 {
		log.Error().Int64("connection_counter", connectionCounter).Msgf("SET query can only specify key and value, got %d tokens", len(tokens))
		return ""
	}
	if len(tokens) < 3 {
		log.Error().Int64("connection_counter", connectionCounter).Msgf("SET query has too few parameters, got %d tokens", len(tokens))
		return ""
	}
	key := strings.Trim(tokens[1], "\n")
	value := strings.Trim(tokens[2], "\n")
	store.Set(key, value)
	return value
}
