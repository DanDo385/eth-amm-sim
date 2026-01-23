// Package bots defines the Bot interface
package bots

import (
	"context"

	"eth-amm-sim/internal/config"
)

// Bot is the interface all trading bots implement
type Bot interface {
	Run(ctx context.Context)
	Stop()
	Nickname() string
	Type() config.BotType
}
