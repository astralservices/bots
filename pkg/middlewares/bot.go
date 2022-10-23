package middlewares

import (
	"github.com/astralservices/bots/pkg/types"
	"github.com/astralservices/dgc"
)

type Bot struct {
	Bot   types.Bot
	Cache map[string]interface{}
}

func (b *Bot) BotMiddleware(next dgc.ExecutionHandler) dgc.ExecutionHandler {
	return func(ctx *dgc.Ctx) {
		ctx.CustomObjects.Set("self", b.Bot)
		ctx.CustomObjects.Set("cache", b.Cache)

		next(ctx)
	}
}
