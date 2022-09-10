package utility

import (
	"fmt"
	"os"

	"github.com/astralservices/bots/pkg/utils"
	"github.com/astralservices/dgc"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var ServerInfoCommand = &dgc.Command{
	Name:        "serverinfo",
	Aliases:     []string{"serverinfo", "server", "guild", "guildinfo"},
	Domain:      "astral.utility.serverinfo",
	Category:    "Utility",
	Usage:       "serverinfo",
	Description: "Get information about the server.",
	Slash:       true,
	SlashGuilds: []string{os.Getenv("DEV_GUILD")},
	Handler: func(ctx *dgc.Ctx) {
		guild, err := ctx.Session.GuildWithCounts(ctx.Event.GuildID)

		if err != nil {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
			return
		}

		guildOwner, err := ctx.Session.User(guild.OwnerID)

		if err != nil {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
			return
		}

		p := message.NewPrinter(language.English)

		ts, err := discordgo.SnowflakeTimestamp(guild.ID)

		if err != nil {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
			return
		}

		channels, err := ctx.Session.GuildChannels(guild.ID)

		if err != nil {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
			return
		}

		ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
			Title:     "Server Information",
			Thumbnail: &discordgo.MessageEmbedThumbnail{URL: guild.IconURL()},
			Fields: []*discordgo.MessageEmbedField{
				{Name: "Name", Value: guild.Name, Inline: true},
				{Name: "ID", Value: guild.ID, Inline: true},
				{Name: "Owner", Value: guildOwner.Mention(), Inline: true},
				{Name: "Members", Value: p.Sprintf("%d", guild.ApproximateMemberCount), Inline: true},
				{Name: "Channels", Value: p.Sprintf("%d", len(channels)), Inline: true},
				{Name: "Roles", Value: p.Sprintf("%d", len(guild.Roles)), Inline: true},
				{Name: "Created", Value: fmt.Sprintf("<t:%d>", ts.Unix()), Inline: true},
			},
		}))
	},
}
