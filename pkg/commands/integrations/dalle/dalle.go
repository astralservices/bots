package dalle

import (
	"github.com/astralservices/bots/pkg/utils"
	"github.com/astralservices/dgc"
	"github.com/bwmarrin/discordgo"
)

type DalleIntegrationSettings struct {
	APIKey    string `json:"apiKey"`
	Ratelimit string `json:"rateLimit"`
}

var DalleIntegrationID = "9d76f6a8-7a71-4faf-81d2-5d9f8e65b2af"

var DalleCommand = &dgc.Command{
	Name:        "dalle",
	Aliases:     []string{"dalle"},
	Description: "Parent command of all dalle commands.",
	Usage:       "dalle <subcommand>",
	Example:     "dalle generate <prompt>",
	Category:    "Dalle",
	Domain:      "astral.integrations.dalle",
	Handler: func(ctx *dgc.Ctx) {
		ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
			Title:       "Error",
			Description: "You must provide a subcommand.",
			Color:       0xff0000,
		}))
	},
	SubCommands: []*dgc.Command{
		GenerateCommand,
	},
}
