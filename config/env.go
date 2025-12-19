package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func GetString(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}

func MustString(key string) string {
	v, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(v) == "" {
		panic(fmt.Sprintf("missing required env: %s", key))
	}
	return v
}

func GetInt(key string, def int) int {
	v, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	n, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil {
		return def
	}
	return n
}

func GetBool(key string, def bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	b, err := strconv.ParseBool(strings.TrimSpace(v))
	if err != nil {
		return def
	}
	return b
}

func GetDuration(key string, def time.Duration) time.Duration {
	v, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	d, err := time.ParseDuration(strings.TrimSpace(v))
	if err != nil {
		return def
	}
	return d
}

func GetValue[T any](key string, def T, parse func(string) (T, error)) T {
	v, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	out, err := parse(strings.TrimSpace(v))
	if err != nil {
		return def
	}
	return out
}
