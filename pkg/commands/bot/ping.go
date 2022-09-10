package commands

import (
	"fmt"
	"os"
	"strings"
	"time"

	db "github.com/astralservices/bots/pkg/database/supabase"
	"github.com/astralservices/bots/pkg/types"
	"github.com/astralservices/bots/pkg/utils"
	"github.com/astralservices/dgc"
	"github.com/bwmarrin/discordgo"
)

var Ping = &dgc.Command{
	Name:        "ping",
	Domain:      "astral.bot.ping",
	Aliases:     []string{"ping", "p"},
	Description: "Retrieves the bot's API and gateway latency to Discord's servers.",
	Category:    "Bot",
	Usage:       "ping",
	Slash:       true,
	SlashGuilds: []string{os.Getenv("DEV_GUILD")},
	Handler: func(ctx *dgc.Ctx) {
		start := time.Now()

		err := ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
			Title: "Pinging...",
			Color: 0xffff00,
		}))

		if err != nil {
			utils.ErrorHandler(err)
		}

		end := time.Now()
		diff := end.Sub(start)

		_, err = ctx.Session.ChannelMessageEditEmbed(ctx.Message.ChannelID, ctx.Message.ID, utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
			Title:       "Pong!",
			Description: fmt.Sprintf(":ping_pong: Gateway Ping: `%dms`\n:desktop: API Ping: `%dms`", ctx.Session.HeartbeatLatency().Milliseconds(), diff.Milliseconds()),
			Color:       0x00ff00,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:  "Region",
					Value: "Fetching...",
				},
			},
		}))

		if err != nil {
			utils.ErrorHandler(err)
		}

		database := db.New()

		self := ctx.CustomObjects.MustGet("self").(types.Bot)

		var regionName string

		if self.Region == "" {
			regionName = "Unknown"
		} else {
			r, err := database.GetRegion(self.Region)

			if err != nil {
				utils.ErrorHandler(err)
				regionName = "Error Fetching Region"
			} else {
				regionName = r.Flag + " " + strings.ToUpper(strings.Split(r.ID, ".")[0]) + " (" + r.City + ", " + r.Region + ", " + r.Country + ")"
			}
		}

		_, err = ctx.Session.ChannelMessageEditEmbed(ctx.Message.ChannelID, ctx.Message.ID, utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
			Title:       "Pong!",
			Description: fmt.Sprintf(":ping_pong: Gateway Ping: `%dms`\n:desktop: API Ping: `%dms`", ctx.Session.HeartbeatLatency().Milliseconds(), diff.Milliseconds()),
			Color:       0x00ff00,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:  "Region",
					Value: regionName,
				},
			},
		}))

		if err != nil {
			utils.ErrorHandler(err)
		}
	},
}
