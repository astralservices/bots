package commands

import (
	"fmt"
	"os"

	db "github.com/astralservices/bots/pkg/database/supabase"
	"github.com/astralservices/bots/pkg/types"
	"github.com/astralservices/bots/pkg/utils"
	"github.com/astralservices/dgc"
	"github.com/bwmarrin/discordgo"
)

var Region = &dgc.Command{
	Name:        "region",
	Domain:      "astral.bot.region",
	Aliases:     []string{"region", "r"},
	Description: "Retrieves the bot's region, or fetches a provided one.",
	Category:    "Bot",
	Usage:       "region [bot]",
	Slash:       true,
	SlashGuilds: []string{os.Getenv("DEV_GUILD")},
	Handler: func(ctx *dgc.Ctx) {
		database := db.New()

		self := ctx.CustomObjects.MustGet("self").(types.Bot)

		var region types.Region

		err := ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
			Title: "Fetching region...",
			Color: 0xffff00,
		}))

		if err != nil {
			utils.ErrorHandler(err)
		}

		if ctx.Arguments.Amount() == 0 {
			r, err := database.GetRegion(self.Region)

			if err != nil {
				ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
					Title:       "Error",
					Description: "An error occurred while fetching the region.",
					Color:       0xff0000,
				}))
				return
			}

			region = r
		} else {
			r, err := database.GetRegion(ctx.Arguments.Get(0).Raw())

			if err != nil {
				ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
					Title:       "Region Not Found",
					Description: "The region you provided was not found.",
					Color:       0xff0000,
				}))

				return
			}

			region = r
		}

		_, err = ctx.Session.ChannelMessageEditEmbed(ctx.Message.ChannelID, ctx.Message.ID, utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
			Title: "Region Info for " + region.Flag + " `" + region.ID + "`",
			Color: 0x00ff00,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:  "Location",
					Value: fmt.Sprintf("%s, %s, %s", region.City, region.Region, region.Country),
				},
			},
		}))

		if err != nil {
			utils.ErrorHandler(err)
		}
	},
}
