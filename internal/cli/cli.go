package cli

import (
	"errors"
	"flag"
)

// Config конфигурация утилиты
type Config struct {
	URL string
}

// NewConfig собирает конфигурацию утилиты из флагов и аргументов командной строки
func NewConfig() (*Config, error) {
	flag.Parse()
	args := flag.Args()

	if len(args) == 0 {
		return nil, errors.New("no URL provided")
	}

	var config Config
	config.URL = args[0]

	return &config, nil
}
