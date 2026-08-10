package main

import (
	"errors"
	"os"
	"time"
)

type config struct {
	databaseURL string
	httpAddr    string

	readTimeout  time.Duration
	writeTimeout time.Duration
	idleTimeout  time.Duration
}

var (
	defaultReadTimeout  = 5 * time.Second
	defaultWriteTimeout = 10 * time.Second
	defaultIdleTimeout  = 60 * time.Second
)

func loadConfig() (config, error) {
	readTimeout, err := getDuration("READ_TIMEOUT", defaultReadTimeout)
	if err != nil {
		return config{}, err
	}

	writeTimeout, err := getDuration("WRITE_TIMEOUT", defaultWriteTimeout)
	if err != nil {
		return config{}, err
	}

	idleTimeout, err := getDuration("IDLE_TIMEOUT", defaultIdleTimeout)
	if err != nil {
		return config{}, err
	}

	cfg := config{
		databaseURL:  os.Getenv("DATABASE_URL"),
		httpAddr:     os.Getenv("HTTP_ADDR"),
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
		idleTimeout:  idleTimeout,
	}

	if cfg.databaseURL == "" {
		return config{}, errors.New("DATABASE_URL is required")
	}

	if cfg.httpAddr == "" {
		cfg.httpAddr = ":8080"
	}

	return cfg, nil
}

func getDuration(envKey string, defaultValue time.Duration) (time.Duration, error) {
	envValue := os.Getenv(envKey)

	if envValue == "" {
		return defaultValue, nil
	}

	parsedTime, err := time.ParseDuration(envValue)

	if err != nil {
		return defaultValue, err
	}

	return parsedTime, nil
}
