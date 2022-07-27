package commands

import (
	"os"
	"runtime"
	"strings"

	"github.com/astralservices/bots/pkg/constants"
	db "github.com/astralservices/bots/pkg/database/supabase"
	"github.com/astralservices/bots/pkg/types"
	"github.com/astralservices/bots/pkg/utils"
	"github.com/astralservices/dgc"
	"github.com/bwmarrin/discordgo"

	"github.com/mackerelio/go-osstat/cpu"
	"github.com/mackerelio/go-osstat/uptime"
)

var Info = &dgc.Command{
	Name:        "info",
	Description: "Get information about the bot.",
	Usage:       "info",
	Category:    "Bot",
	Aliases:     []string{"info", "i", "about"},
	Domain:      "astral.bot.info",
	Example:     "info",
	Handler: func(ctx *dgc.Ctx) {
		mu := utils.GetMemoryUsage()
		uptime, err := uptime.Get()
		if err != nil {
			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:       "Error",
				Color:       0xff0000,
				Description: "Unable to get uptime information.",
			}))
			return
		}

		isWindows := runtime.GOOS == "windows"
		cpu, err := cpu.Get()
		if err != nil && !isWindows {
			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:       "Error",
				Color:       0xff0000,
				Description: "Unable to get CPU information.",
			}))
			return
		}

		var cpuString string
		if isWindows {
			cpuString = "N/A"
		} else {
			cpuString = "**User**: " + string(cpu.User) + "%"
		}

		embed := discordgo.MessageEmbed{}
		embed.Title = "Bot Information"
		embed.Description = "This bot is running via [Astral](https://astralapp.io/).\n\nFramework: DiscordGo v" + discordgo.VERSION
		embed.Fields = []*discordgo.MessageEmbedField{
			{
				Name:  "Region",
				Value: "Fetching...",
			},
			{
				Name:  "Version",
				Value: constants.VERSION,
			},
			{
				Name:  "Developers",
				Value: "AmusedGrape#0001",
			},
			{
				Name:  "Server",
				Value: os.Getenv("SERVER"),
			},
			{
				Name:  "Memory Usage",
				Value: "**Used**: " + mu.Allocated + "\n**Total**: " + mu.Sys,
			},
			{
				Name:  "CPU Usage",
				Value: cpuString,
			},
			{
				Name:  "Uptime",
				Value: uptime.String(),
			},
		}
		err = ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, embed))

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

		embed.Fields[0].Value = regionName

		_, err = ctx.Session.ChannelMessageEditEmbed(ctx.Message.ChannelID, ctx.Message.ID, utils.GenerateEmbed(*ctx, embed))
		if err != nil {
			utils.ErrorHandler(err)
		}
	},
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
