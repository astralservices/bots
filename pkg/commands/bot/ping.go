package commands

import (
	"github.com/astralservices/bots/pkg/utils"
	"github.com/astralservices/dgc"
	"github.com/bwmarrin/discordgo"
)

var Ping = dgc.Command{
	Name:        "ping",
	Description: "Ping!",
	Usage:       "ping",
	Handler: func(ctx *dgc.Ctx) {
		ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
			Title: "Pong!",
		}))
	},
}
