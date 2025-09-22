package cli

import (
	"errors"
	"flag"
)

const DefaultLevel = -1

// Config конфигурация утилиты
type Config struct {
	URL   string
	Level int
}

// NewConfig собирает конфигурацию утилиты из флагов и аргументов командной строки
func NewConfig() (*Config, error) {
	var config Config

	level := flag.Int("l", DefaultLevel, "level of recursion")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		return nil, errors.New("no URL provided")
	}

	config.Level = *level
	config.URL = args[0]

	return &config, nil
}
