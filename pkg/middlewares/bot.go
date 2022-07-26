package middlewares

import (
	"github.com/astralservices/bots/utils"
	"github.com/zekroTJA/shireikan"
)

type Bot struct {
	Settings utils.IBot
}

// Handle is the Middlewares handler.
func (m *Bot) Handle(cmd shireikan.Command, ctx shireikan.Context, layer shireikan.MiddlewareLayer) (bool, error) {
	ctx.SetObject("bot", m.Settings)

	return true, nil
}

// GetLayer returns the execution layer.
func (m *Bot) GetLayer() shireikan.MiddlewareLayer {
	return shireikan.LayerBeforeCommand
}
