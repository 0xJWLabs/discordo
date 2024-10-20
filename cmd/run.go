package cmd

import (
	"github.com/0xJWLabs/discordo/internal/config"
	"github.com/0xJWLabs/discordo/internal/logger"
)

var (
	discordState *State
	layout       *Layout
)

func Run(token string) error {
	if err := logger.Load(); err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	layout = newLayout(cfg)
	return layout.run(token)
}
